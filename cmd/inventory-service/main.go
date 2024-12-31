package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
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
		dbPort = "5435"
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
		dbName = "inventory_service"
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

func reserveInventory(c *gin.Context) {
	var products []models.Product
	if err := c.ShouldBindJSON(&products); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx, err := db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to begin transaction"})
		return
	}
	defer tx.Rollback()

	for _, product := range products {
		var currentQuantity int
		err := tx.QueryRow("SELECT quantity FROM inventory WHERE product_id = $1 FOR UPDATE", product.ID).Scan(&currentQuantity)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			return
		}

		if currentQuantity < product.Quantity {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient inventory"})
			return
		}

		_, err = tx.Exec("UPDATE inventory SET quantity = quantity - $1 WHERE product_id = $2",
			product.Quantity, product.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update inventory"})
			return
		}
	}

	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Inventory reserved successfully"})
}

func rollbackReservation(c *gin.Context) {
	var products []models.Product
	if err := c.ShouldBindJSON(&products); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx, err := db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to begin transaction"})
		return
	}
	defer tx.Rollback()

	for _, product := range products {
		_, err = tx.Exec("UPDATE inventory SET quantity = quantity + $1 WHERE product_id = $2",
			product.Quantity, product.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to rollback inventory"})
			return
		}
	}

	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Inventory rollback successful"})
}

func main() {
	r := gin.Default()
	
	// Trust only loopback proxies
	r.SetTrustedProxies([]string{"127.0.0.1"})
	
	r.POST("/api/inventory/reserve", reserveInventory)
	r.POST("/api/inventory/rollback", rollbackReservation)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	r.Run(":" + port)
}
