package main

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	_ "modernc.org/sqlite"
)

type Cart struct {
	CartID    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func initializeCartsTable() error {
	SQL := `
		CREATE TABLE IF NOT EXISTS carts (
			cart_id TEXT,
			created_at TIMESTAMP,
			updated_at TIMESTAMP
		)`
	_, err := db.Exec(SQL)
	if err != nil {
		return err
	} else {
		return nil
	}
}

func createCart(userID string) (string, error) {
	cartID := "C" + userID[1:]
	SQL := `INSERT INTO carts (cart_id, created_at, updated_at) VALUES (?, ?, ?)`
	_, err := db.Exec(SQL, cartID, time.Now(), time.Now())
	if err != nil {
		return "", err
	} else {
		return cartID, nil
	}
}

func checkoutCart(c *gin.Context) {
	cartID := c.Param("cart_id")
	SQL := `SELECT * FROM cart_items WHERE cart_id = ?`

	rows, err := db.Query(SQL, cartID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "error fetching from db", "error": err.Error()})
	}

	items, err := bindCartItems(rows)
	defer rows.Close()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "error binding cart", "error": err.Error()})
		return
	}

	orderID := "O" + cartID[1:]

	if err := pushOrderItems(items, orderID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "error pushing order items", "error": err.Error()})
		return
	}

	if err := pushOrder(c, orderID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "error pushing order", "error": err.Error()})
	}

	paymentID := createCheckoutSession(c)
	if paymentID != "" {
		err := pushPaymentID(orderID, paymentID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "error pushing payment id", "error": err.Error()})
		}
	} else {
		err := paymentFailedStatus(orderID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "error updating order status", "error": err.Error()})
		}
	}

}

func bindCartItems(rows *sql.Rows) ([]CartItem, error) {
	items := []CartItem{}
	for rows.Next() {
		var item CartItem
		err := rows.Scan(&item.CartItemID, &item.CartID, &item.ProductID, &item.Size, &item.Price, &item.Quantity, &item.Notes)
		if err != nil {
			return items, err
		}
		items = append(items, item)
	}

	return items, nil
}

func pushOrderItems(items []CartItem, orderID string) error {
	SQL := `INSERT INTO order_items (order_id, product_id, size, price, quantity, notes) VALUES (?, ?, ?, ?, ?, ?)`
	for _, v := range items {
		_, err := db.Exec(SQL, orderID, v.ProductID, v.Size, v.Price, v.Quantity, v.Notes)
		if err != nil {
			return err
		}
	}
	return nil
}

func pushOrder(c *gin.Context, orderID string) error {
	var order Order
	if err := c.ShouldBindBodyWithJSON(&order); err != nil {
		return err
	}
	SQL := `INSERT INTO orders (
				order_id,
				is_delivery,
				delivery_address,
				ready_date,
				notes,
				subtotal,
				delivery_fee,
				status,
				created_at,
				updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := db.Exec(SQL, orderID, order.IsDelivery, order.DeliveryAddress, order.ReadyDate, order.Notes, order.Subtotal, order.DeliveryFee, order.Status, time.Now(), time.Now())
	if err != nil {
		return err
	} else {
		return nil
	}
}

func pushPaymentID(orderID string, paymentID string) error {
	SQL := `UPDATE orders SET payment_id = ? WHERE order_id = ?`
	_, err := db.Exec(SQL, paymentID, orderID)
	if err != nil {
		return err
	} else {
		return nil
	}
}

func paymentFailedStatus(orderID string) error {
	SQL := `UPDATE orders SET status = ? WHERE order_id = ?`
	_, err := db.Exec(SQL, "payment failed", orderID)
	if err != nil {
		return err
	} else {
		return nil
	}
}
