package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

type ProductDelete struct {
	ProductID string
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
func addProducts(c *gin.Context) {
	var count int

	if result, err := getCount("products"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error_1": err.Error()})
	} else {
		count = result
	}

	var products []Product
	if err := c.ShouldBindBodyWithJSON(&products); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error_2": err.Error()})
		return
	}

	SQL := `
		INSERT INTO products (product_id, image, name, description, category, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?);
	`
	var productsString string
	countAdd := 1
	for _, product := range products {
		productID := fmt.Sprintf("%04d", count+countAdd)
		countAdd++
		product.ProductID = productID
		_, err := db.Exec(SQL, product.ProductID, product.Image, product.Name, product.Description, product.Category, time.Now(), time.Now())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error_3": err.Error()})
			return
		}
		productsString = productsString + product.Name + ","
	}

	message := fmt.Sprintf(`product added: %s`, productsString)

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
	//Body: "new-value" : value
	body, err_0 := io.ReadAll(c.Request.Body)
	if err_0 != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error_0": "Could not read request body."})
		return
	}

	var data map[string]interface{}

	if err := json.Unmarshal(body, &data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	value, ok := data["new-value"].(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "reading value"})
		return
	}

	SQL := fmt.Sprintf(`UPDATE products SET (%s, updated_at) = (?, ?) WHERE id = ?;`, column)

	_, err_2 := db.Exec(SQL, value, time.Now(), ID)
	if err_2 != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error_2": err_2.Error()})
		return
	}

	message := fmt.Sprintf(`edited product %s's %s`, ID, column)

	c.JSON(http.StatusOK, gin.H{"message": message})

}

func editPriceInt(c *gin.Context, ID string, column string) {
	//Body: "new-value" : value
	body, err_0 := io.ReadAll(c.Request.Body)
	if err_0 != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error_0": "Could not read request body."})
		return
	}

	var data map[string]interface{}

	if err := json.Unmarshal(body, &data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	value, ok := data["new-value"].(int)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "reading value"})
		return
	}

	SQL := fmt.Sprintf(`UPDATE products SET (%s, updated_at) = (?, ?) WHERE id = ?;`, column)

	_, err_2 := db.Exec(SQL, value, time.Now(), ID)
	if err_2 != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error_2": err_2.Error()})
		return
	}

	message := fmt.Sprintf(`edited product %s's %s`, ID, column)

	c.JSON(http.StatusOK, gin.H{"message": message})
}

func deleteProduct(c *gin.Context) {
	productID := c.Param("id")
	SQL := `DELETE FROM products WHERE product_id = ?`

	_, err := db.Exec(SQL, productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, "product deleted: "+productID)
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
