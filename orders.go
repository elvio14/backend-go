package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	_ "modernc.org/sqlite"
)

type Order struct {
	OrderID         string
	IsDelivery      string
	DeliveryAddress string
	ReadyDate       time.Time
	PaymentID       string
	Notes           string
	Subtotal        int
	DeliveryFee     int
	Tax             int
	TotalPrice      int
	Status          string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	//13 fields
}

func initializeOrdersTable() error {
	SQL := `CREATE TABLE IF NOT EXISTS orders (
				order_id TEXT,
				id_delivery BOOLEAN,
				delivery_address TEXT,
				ready_date TIMESTAMP,
				payment_id TEXT,
				notes TEXT,
				subtotal INT,
				delivery_fee INT,
				tax INT GENERATED ALWAYS AS (CAST((subtotal + delivery_fee) * 0.13 AS INT)) STORED,
				total_price INT GENERATED ALWAYS AS (subtotal + tax) STORED,
				status TEXT,
				created_at TIMESTAMP,
				updated_at TIMESTAMP)
		`

	_, err := db.Exec(SQL)
	if err != nil {
		return err
	} else {
		return nil
	}
}

func getNumberOfOrders(c *gin.Context) {
	rows, err := countRows("orders")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	} else {
		c.JSON(http.StatusOK, gin.H{"number_of_orders": rows})
	}
}

func editOrderStatus(c *gin.Context) {
	orderID := c.Param("order_id")
	status := c.Param("status")
	SQL := `UPDATE orders SET status = ? WHERE order_id = ?`
	_, err := db.Exec(SQL, status, orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	message := fmt.Sprintf(`updated order %s's status to %s`, orderID, status)

	c.JSON(http.StatusOK, gin.H{"message": message})
}
