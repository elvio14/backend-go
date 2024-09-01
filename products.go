package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	_ "modernc.org/sqlite"
)

//DATA MODELS

type Product struct {
	ProductID   string
	Image       string
	Name        string
	Description string
	Category    string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func initializeProductsTable() error {
	SQL := `
		CREATE TABLE IF NOT EXISTS products (
			product_id TEXT PRIMARY KEY,
			image TEXT,
			name TEXT NOT NULL,
			description TEXT,
			category TEXT,
			created_at TIMESTAMP,
			updated_at TIMESTAMP
		);
	`
	_, err := db.Exec(SQL)
	if err != nil {
		return err
	} else {
		return nil
	}
}

// PRODUCTS
func addProduct(c *gin.Context) {
	var count int

	if result, err := getCount("products"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error_1": err.Error()})
	} else {
		count = result
	}

	// Generate Product ID
	productID := fmt.Sprintf("%04d", count+1)

	var product Product
	if err := c.ShouldBindBodyWithJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error_2": err.Error()})
		return
	}

	product.ProductID = productID

	SQL := `
		INSERT INTO products (product_id, image, name, description, category, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?);
	`
	_, err := db.Exec(SQL, productID, product.Image, product.Name, product.Description, product.Category, time.Now(), time.Now())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error_3": err.Error()})
		return
	}

	message := fmt.Sprintf(`product added: %s`, product.Name)

	c.JSON(http.StatusOK, gin.H{"message": message})
}

func getAllProducts(c *gin.Context) {
	SQL := `
		SELECT * FROM products;
	`
	rows, err := db.Query(SQL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"message": "error getting products",
		})
	}
	defer rows.Close()

	products, err := bindProducts(rows)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, products)
}

func getProductsByCategory(c *gin.Context) {
	category := c.Param("category")
	SQL := `SELECT * FROM products WHERE category = ?`
	rows, err := db.Query(SQL, category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	products, err := bindProducts(rows)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	c.IndentedJSON(http.StatusOK, products)
}

func getProductByID(c *gin.Context) {
	ID := c.Param("ID")
	SQL := `SELECT * FROM products WHERE id = ?`
	row, err := db.Query(SQL, ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer row.Close()

	product, err := bindProducts(row)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	c.IndentedJSON(http.StatusOK, product)
}

// EDIT
func editProduct(c *gin.Context) {
	ID := c.Param("ID")
	column := c.Param("column")
	if column == "price" {
		editPriceInt(c, ID, column)
	} else {
		editProductString(c, ID, column)
	}
}

func editProductString(c *gin.Context, ID string, column string) {
	body, err_0 := c.GetRawData()
	if err_0 != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error_1": "Could not read request body."})
		return
	}
	bodyString := string(body)

	SQL := fmt.Sprintf(`UPDATE products SET (%s, updated_at) = (?, ?) WHERE id = ?;`, column)

	_, err_1 := db.Exec(SQL, bodyString, time.Now(), ID)
	if err_1 != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error_2": err_1.Error()})
		return
	}

	message := fmt.Sprintf(`edited product %s's %s`, ID, column)

	c.JSON(http.StatusOK, gin.H{"message": message})

}

func editPriceInt(c *gin.Context, ID string, column string) {
	body, err_0 := c.GetRawData()
	if err_0 != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error_0": "Could not read request body."})
	}
	bodyString := string(body)

	bodyInt, err_1 := strconv.ParseInt(bodyString, 10, 64)
	if err_1 != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error_1": "failed to convert to number"})
		return
	}

	SQL := fmt.Sprintf(`UPDATE products SET (%s, updated_at) = (?, ?) WHERE id = ?;`, column)

	_, err_2 := db.Exec(SQL, bodyInt, time.Now(), ID)
	if err_2 != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error_2": err_2.Error()})
		return
	}

	message := fmt.Sprintf(`edited product %s's %s`, ID, column)

	c.JSON(http.StatusOK, gin.H{"message": message})
}

func bindProducts(rows *sql.Rows) ([]Product, error) {
	products := []Product{}
	for rows.Next() {
		var product Product
		err := rows.Scan(&product.ProductID, &product.Image, &product.Name, &product.Description, &product.Category, &product.CreatedAt, &product.UpdatedAt)
		if err != nil {
			return products, err
		}
		products = append(products, product)
	}

	return products, nil
}
