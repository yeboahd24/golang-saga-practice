package main

import (
	"fmt"
	"log"

	"github.com/google/uuid"
)

// Order represents an order in the system
type Order struct {
	ID       string
	UserID   string
	Amount   float64
	Status   string
	Products []string
}

// OrderService handles order-related operations
type OrderService struct{}

func (s *OrderService) CreateOrder(userID string, amount float64, products []string) (*Order, error) {
	order := &Order{
		ID:       uuid.New().String(),
		UserID:   userID,
		Amount:   amount,
		Status:   "pending",
		Products: products,
	}
	log.Printf("Order created: %v\n", order)
	return order, nil
}

func (s *OrderService) UpdateOrderStatus(orderID, status string) error {
	log.Printf("Order %s status updated to: %s\n", orderID, status)
	return nil
}

// PaymentService handles payment operations
type PaymentService struct{}

func (s *PaymentService) ProcessPayment(orderID string, amount float64) error {
	log.Printf("Processing payment for order %s, amount: %.2f\n", orderID, amount)
	// Simulate payment processing
	return nil
}

func (s *PaymentService) RollbackPayment(orderID string) error {
	log.Printf("Rolling back payment for order %s\n", orderID)
	return nil
}

// InventoryService handles inventory operations
type InventoryService struct{}

func (s *InventoryService) ReserveProducts(orderID string, products []string) error {
	log.Printf("Reserving products for order %s: %v\n", orderID, products)
	return nil
}

func (s *InventoryService) RollbackReservation(orderID string, products []string) error {
	log.Printf("Rolling back product reservation for order %s: %v\n", orderID, products)
	return nil
}

// OrderSaga coordinates the entire order process
type OrderSaga struct {
	orderService     *OrderService
	paymentService   *PaymentService
	inventoryService *InventoryService
}

func NewOrderSaga() *OrderSaga {
	return &OrderSaga{
		orderService:     &OrderService{},
		paymentService:   &PaymentService{},
		inventoryService: &InventoryService{},
	}
}

func (s *OrderSaga) Execute(userID string, amount float64, products []string) error {
	// Step 1: Create Order
	order, err := s.orderService.CreateOrder(userID, amount, products)
	if err != nil {
		return fmt.Errorf("failed to create order: %v", err)
	}

	// Step 2: Reserve Inventory
	err = s.inventoryService.ReserveProducts(order.ID, products)
	if err != nil {
		// Rollback order creation
		s.orderService.UpdateOrderStatus(order.ID, "failed")
		return fmt.Errorf("failed to reserve products: %v", err)
	}

	// Step 3: Process Payment
	err = s.paymentService.ProcessPayment(order.ID, amount)
	if err != nil {
		// Rollback inventory reservation
		s.inventoryService.RollbackReservation(order.ID, products)
		// Update order status
		s.orderService.UpdateOrderStatus(order.ID, "failed")
		return fmt.Errorf("failed to process payment: %v", err)
	}

	// Update order status to completed
	s.orderService.UpdateOrderStatus(order.ID, "completed")
	return nil
}

func main() {
	saga := NewOrderSaga()

	// Example order
	userID := "user123"
	amount := 99.99
	products := []string{"product1", "product2"}

	fmt.Println("Starting order saga...")
	err := saga.Execute(userID, amount, products)
	if err != nil {
		log.Printf("Saga failed: %v\n", err)
		return
	}
	fmt.Println("Order saga completed successfully!")
}
