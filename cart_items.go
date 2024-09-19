package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "modernc.org/sqlite"
)

type CartItem struct {
	CartItemID int
	CartID     string `json:"cart_id"`
	ProductID  string `json:"product_id"`
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
	cartID := c.Param("cart_id")
	itemID := c.Param("item_id")
	SQL := `DELETE FROM cart_items WHERE (cart_id, item_id) = (?, ?)`

	_, err := db.Exec(SQL, itemID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error_1": err.Error()})
		return
	}

	message := fmt.Sprintf(`removed %s from cart %s`, itemID, cartID)

	c.JSON(http.StatusOK, gin.H{"message": message})
}

func changeQuantity(c *gin.Context) {
	cartID := c.Param("cart_id")
	itemID := c.Param("item_id")
	quantity := c.Param("quantity")
	SQL := `UPDATE cart_items SET quantity = ? WHERE (cart_id, item_id) = (?, ?)`

	_, err := db.Exec(SQL, quantity, cartID, itemID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	message := fmt.Sprintf(`changed %s's quantity to %s in cart %s`, itemID, quantity, cartID)

	c.JSON(http.StatusOK, gin.H{"message": message})
}

func getCartItems(c *gin.Context) {
	cartID := c.Param("cart_id")
	SQL := `SELECT * FROM cart_items WHERE cart_id = ?`

	rows, err := db.Query(SQL, cartID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	items := []CartItem{}
	for rows.Next() {
		var item CartItem
		err := rows.Scan(&item.CartItemID, &item.CartID, &item.ProductID, &item.Size, &item.Price, &item.Quantity, &item.Notes)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		items = append(items, item)
	}

	c.IndentedJSON(http.StatusOK, items)
}
