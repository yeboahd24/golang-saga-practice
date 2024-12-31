package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"saga/internal/models"
)

var db *sql.DB

type InventoryRequest struct {
	OrderID   string             `json:"order_id"`
	Products  []ProductReserve   `json:"products"`
}

type ProductReserve struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type PaymentRequest struct {
	OrderID string  `json:"order_id"`
	Amount  float64 `json:"amount"`
	UserID  string  `json:"user_id"`
}

func init() {
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5433"
	}
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "postgres"
	}
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "postgres"
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "order_service"
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	var err error
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}
}

func createOrder(c *gin.Context) {
	var order models.Order
	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order.ID = uuid.New().String()
	order.Status = "pending"

	tx, err := db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to begin transaction"})
		return
	}
	defer tx.Rollback()

	// Insert order
	_, err = tx.Exec(`
		INSERT INTO orders (id, user_id, amount, status)
		VALUES ($1, $2, $3, $4)
	`, order.ID, order.UserID, order.Amount, order.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
		return
	}

	// Insert order products
	for _, product := range order.Products {
		_, err = tx.Exec(`
			INSERT INTO order_products (order_id, product_id, quantity, price)
			VALUES ($1, $2, $3, $4)
		`, order.ID, product.ID, product.Quantity, product.Price)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order products"})
			return
		}
	}

	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	// Start saga - Call inventory service
	go startOrderSaga(order)

	c.JSON(http.StatusCreated, order)
}

func startOrderSaga(order models.Order) {
	// Step 1: Reserve Inventory
	inventoryReq := InventoryRequest{
		OrderID:  order.ID,
		Products: make([]ProductReserve, len(order.Products)),
	}
	for i, product := range order.Products {
		inventoryReq.Products[i] = ProductReserve{
			ProductID: product.ID,
			Quantity:  product.Quantity,
		}
	}
	
	inventoryBody, err := json.Marshal(inventoryReq)
	if err != nil {
		updateOrderStatus(order.ID, "failed")
		return
	}
	
	resp, err := http.Post("http://localhost:8081/api/inventory/reserve", 
		"application/json", 
		bytes.NewBuffer(inventoryBody))
	if err != nil || resp.StatusCode != http.StatusOK {
		updateOrderStatus(order.ID, "failed")
		return
	}

	// Step 2: Process Payment
	paymentReq := PaymentRequest{
		OrderID: order.ID,
		Amount:  order.Amount,
		UserID:  order.UserID,
	}
	
	paymentBody, err := json.Marshal(paymentReq)
	if err != nil {
		// Rollback inventory
		rollbackReq := InventoryRequest{
			OrderID:  order.ID,
			Products: inventoryReq.Products,
		}
		rollbackBody, _ := json.Marshal(rollbackReq)
		http.Post("http://localhost:8081/api/inventory/rollback", 
			"application/json", 
			bytes.NewBuffer(rollbackBody))
		updateOrderStatus(order.ID, "failed")
		return
	}
	
	resp, err = http.Post("http://localhost:8082/api/payments/process", 
		"application/json", 
		bytes.NewBuffer(paymentBody))
	if err != nil || resp.StatusCode != http.StatusOK {
		// Rollback inventory
		rollbackReq := InventoryRequest{
			OrderID:  order.ID,
			Products: inventoryReq.Products,
		}
		rollbackBody, _ := json.Marshal(rollbackReq)
		http.Post("http://localhost:8081/api/inventory/rollback", 
			"application/json", 
			bytes.NewBuffer(rollbackBody))
		updateOrderStatus(order.ID, "failed")
		return
	}

	updateOrderStatus(order.ID, "completed")
}

func updateOrderStatus(orderID, status string) {
	_, err := db.Exec("UPDATE orders SET status = $1 WHERE id = $2", status, orderID)
	if err != nil {
		log.Printf("Failed to update order status: %v", err)
	}
}

func main() {
	r := gin.Default()
	
	// Trust only loopback proxies
	r.SetTrustedProxies([]string{"127.0.0.1"})
	
	r.POST("/api/orders", createOrder)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
