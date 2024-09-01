package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "modernc.org/sqlite"
)

type CartItem struct {
	CartItemID int
	CartID     string
	ProductID  string
	Size       string
	Price      int
	Quantity   int
	Notes      string
}

func initializeCartItemsTable() error {
	SQL := `
		CREATE TABLE IF NOT EXISTS cart_items (
			item_id INTEGER PRIMARY KEY AUTOINCREMENT,
			cart_id TEXT,
			product_id TEXT,
			size TEXT,
			price TEXT,
			quantity INT,
			notes TEXT
		)`
	_, err := db.Exec(SQL)
	if err != nil {
		return err
	} else {
		return nil
	}
}

func addToCart(c *gin.Context) {
	SQL := `INSERT INTO cart_items (cart_id, product_id, size, price, quantity, notes) VALUES (?, ?, ?, ?, ?, ?);`

	var cartItem CartItem
	if err := c.ShouldBindBodyWithJSON(&cartItem); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error_1": err.Error()})
		return
	}

	_, err := db.Exec(SQL, cartItem.CartID, cartItem.ProductID, cartItem.Size, cartItem.Price, cartItem.Quantity, cartItem.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error_2": err.Error()})
		return
	}

	message := fmt.Sprintf(`item %s added to cart %s`, cartItem.ProductID, cartItem.CartID)

	c.JSON(http.StatusOK, gin.H{"message": message})
}

func removeFromCart(c *gin.Context) {
	cartName := c.Param("cart_name")
	itemID := c.Param("item_id")
	SQL := fmt.Sprintf(`DELETE FROM %s WHERE item_id = ?`, cartName)

	_, err := db.Exec(SQL, itemID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error_1": err.Error()})
		return
	}

	message := fmt.Sprintf(`removed %s from cart %s`, itemID, cartName)

	c.JSON(http.StatusOK, gin.H{"message": message})
}

func changeQuantity(c *gin.Context) {
	cartName := c.Param("cart_name")
	itemID := c.Param("item_id")
	quantity := c.Param("quantity")
	SQL := fmt.Sprintf(`UPDATE %s SET quantity = ? WHERE item_id = ?`, cartName)

	_, err := db.Exec(SQL, quantity, itemID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	message := fmt.Sprintf(`changed %s's quantity to %s in cart %s`, itemID, quantity, cartName)

	c.JSON(http.StatusOK, gin.H{"message": message})
}
