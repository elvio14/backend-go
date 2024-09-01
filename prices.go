package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	_ "modernc.org/sqlite"
)

type Price struct {
	PriceID   int
	ProductID string
	Size      string
	Price     int
	CreatedAt time.Time
	UpdatedAt time.Time
}

func initializePriceTable() error {
	SQL := `CREATE TABLE IF NOT EXISTS prices (
				price_id TEXT,
				product_id TEXT,
				size TEXT,
				price INT,
				CreatedAt TIMESTAMP,
				UpdatedAt TIMESTAMP
		)`
	_, err := db.Exec(SQL)
	if err != nil {
		return err
	} else {
		return nil
	}
}

func setPrice(c *gin.Context) {
	var price Price
	if err := c.ShouldBindBodyWithJSON(&price); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	SQL := `INSERT INTO prices (price_id, product_id, size, price, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?)`
	_, err := db.Exec(SQL, price.ProductID, price.Size, price.Price, time.Now(), time.Now())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
}

func updatePriceBySize(c *gin.Context) {
	productID := c.Param("product_id")
	price := c.Param("price")
	size := c.Param("size")

	priceint, err_0 := parseToInt(price)
	if err_0 != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error_0": err_0.Error()})
	}
	SQL := `UPDATE prices SET (price, updated_at) = (?, ?) WHERE (product_id, size) = (?, ?)`
	_, err := db.Exec(SQL, priceint, time.Now(), productID, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	message := fmt.Sprintf(`updated %s's size %s price to %s`, productID, size, price)

	c.JSON(http.StatusOK, gin.H{"message": message})
}
