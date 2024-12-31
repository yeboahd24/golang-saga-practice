package main

import (
	"database/sql"
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

func init() {
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5434"
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
		dbName = "payment_service"
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

func processPayment(c *gin.Context) {
	var payment models.Payment
	if err := c.ShouldBindJSON(&payment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payment.ID = uuid.New().String()
	payment.Status = "processing"

	tx, err := db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to begin transaction"})
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		INSERT INTO payments (id, order_id, amount, status)
		VALUES ($1, $2, $3, $4)
	`, payment.ID, payment.OrderID, payment.Amount, payment.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create payment"})
		return
	}

	// Simulate payment processing
	// In a real application, you would integrate with a payment gateway
	payment.Status = "completed"
	_, err = tx.Exec("UPDATE payments SET status = $1 WHERE id = $2", payment.Status, payment.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update payment status"})
		return
	}

	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, payment)
}

func rollbackPayment(c *gin.Context) {
	var payment models.Payment
	if err := c.ShouldBindJSON(&payment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := db.Exec("UPDATE payments SET status = 'failed' WHERE order_id = $1", payment.OrderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to rollback payment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Payment rollback successful"})
}

func main() {
	r := gin.Default()
	
	// Trust only loopback proxies
	r.SetTrustedProxies([]string{"127.0.0.1"})
	
	r.POST("/api/payments/process", processPayment)
	r.POST("/api/payments/rollback", rollbackPayment)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}
	r.Run(":" + port)
}
