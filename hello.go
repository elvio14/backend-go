package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"log"

	"os"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v79"

	_ "modernc.org/sqlite"

	"github.com/joho/godotenv"

	"github.com/gin-contrib/cors"
)

var jwtSecret []byte
var db *sql.DB

func main() {

	fmt.Println("Hello, World!")

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	port := os.Getenv("SECRET_CODE")

	jwtSecret = []byte(os.Getenv("JWT_SECRET"))

	// DATABASE INIT
	db, err = InitializeDB()
	if err != nil {
		log.Fatalf("failed to initialize the database: %v", err)
	}

	userID, cartID := initializeAllTables()

	//REST API
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:5173"},
		AllowMethods: []string{"GET", "PUT", "POST", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "token"},
	}))

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.GET("/ids", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"user_id": userID,
			"cart_id": cartID,
		})
	})

	//PRODUCTS
	r.GET("/products", func(c *gin.Context) {
		getAllProducts(c)
	})

	r.GET("/products/category/:category", func(c *gin.Context) {
		getProductsByCategory(c)
	})

	r.GET("products/ID/:ID", func(c *gin.Context) {
		getProductByID(c)
	})

	r.GET("/products/prices", func(c *gin.Context) {
		getPrices(c)
	})

	r.PUT("/products/ID/:ID/:column", func(c *gin.Context) {
		isAdmin, err := validateAdmin(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}

		if isAdmin {
			editProductString(c)
		}
	})

	r.POST("/products", func(c *gin.Context) {
		//takes an array or products in body JSON
		isAdmin, err := validateAdmin(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}

		if isAdmin {
			addProducts(c)
		}
	})

	r.POST("/products/price", func(c *gin.Context) {
		isAdmin, err := validateAdmin(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}

		if isAdmin {
			setPrice(c)
		}
	})

	r.PUT("/products/price/ID/:product_id/:size/:price", func(c *gin.Context) {
		isAdmin, err := validateAdmin(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}

		if isAdmin {
			updatePriceBySize(c)
		}
	})

	r.DELETE("/products/:id", func(c *gin.Context) {
		isAdmin, err := validateAdmin(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}

		if isAdmin {
			deleteProduct(c)
		}
	})

	//CART
	r.POST("/carts", func(c *gin.Context) {
		addToCart(c)
	})

	r.DELETE("/carts/:cart_id/:item_id", func(c *gin.Context) {
		removeFromCart(c)
	})

	r.POST("/checkout", func(c *gin.Context) {
		checkoutCart(c)
	})

	r.PUT("/carts/:cart_id/:item_id/:quantity", func(c *gin.Context) {
		changeQuantity(c)
	})

	//CART ITEMS
	r.GET("/cart_items/:cart_id", func(c *gin.Context) {
		getCartItems(c)
	})
	//USERS
	r.POST("/users", func(c *gin.Context) {
		register(c, userID)
	})

	r.POST("/login", func(c *gin.Context) {
		login(c)
	})

	r.GET("/validate_admin", func(c *gin.Context) {
		isAdmin, err := validateAdmin(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		c.JSON(http.StatusOK, gin.H{"is_admin": isAdmin})
	})

	//ORDERS
	r.GET("/orders/count", func(c *gin.Context) {
		getNumberOfOrders(c)
	})

	r.PUT("/orders/:order_id/:status", func(c *gin.Context) {
		editOrderStatus(c)
	})

	//PAYMENTS
	stripe.Key = os.Getenv("STRIPE_API_KEY")
	//
	r.Run("localhost:" + port)
	db.Exec("PRAGMA busy_timeout = 5000")
	CloseDB()
}

//----------------------------------------------------------------------------------------

// utils
func InitializeDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite", "./main.db")
	if err != nil {
		log.Fatalf("Failed to open main.db. %v", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatalf("Ping error: %v", err)
	}

	log.Println("Connected to main.db.")

	return db, nil
}

func initializeAllTables() (string, string) {
	if err := initializeCounterTable(); err != nil {
		log.Fatal(`error initializing counter table`, err)
		return "", ""
	}

	if err := initializeUserTable(); err != nil {
		log.Fatal(`error initializing User table`, err)
		return "", ""
	}

	var userID string

	if id, err := generateUserID(); err != nil {
		log.Fatal(`error generating user id`)
		return "", ""
	} else {
		userID = id
	}

	fmt.Println(userID)

	if err := initializeProductsTable(); err != nil {
		log.Fatal(`error initializing product table`, err)
		return "", ""
	}
	if err := initializePriceTable(); err != nil {
		log.Fatal(`error initializing rrice table`, err)
		return "", ""
	}

	if err := addCurrentUser(userID); err != nil {
		log.Fatal(`error adding user`, err)
		return "", ""
	}

	if err := initializeCartsTable(); err != nil {
		log.Fatal(`error initializing carts table`, err)
		return "", ""
	}

	if err := initializeCartItemsTable(); err != nil {
		log.Fatal(`error initializing cart_items table`, err)
		return "", ""
	}

	var cartID string
	if result, err := createCart(userID); err != nil {
		log.Fatal(`error creating cart`, err)
		return "", ""
	} else {
		cartID = result
	}

	if err := initializeOrdersTable(); err != nil {
		log.Fatal(`error initializing orders table`, err)
		return "", ""
	}

	if err := initializeOrderItemsTable(); err != nil {
		log.Fatal(`error initializing order_items table`, err)
		return "", ""
	}

	return userID, cartID
}

// Utils

func CloseDB() {
	defer db.Close()
}

func countRows(table string) (int, error) {
	var count int
	SQL := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
	row := db.QueryRow(SQL)
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count rows: %v", err)
	}

	return count, nil
}

func getCount(column string) (int, error) {
	//Generate temporary UserID
	SQL := fmt.Sprintf(`SELECT %s FROM counter WHERE id = ?`, column)
	rows, err := db.Query(SQL, 1)
	if err != nil {
		return 0, err
	}

	var count int
	for rows.Next() {
		err := rows.Scan(&count)
		if err != nil {
			return 0, err
		}
	}

	SQL_1 := fmt.Sprintf(`UPDATE counter SET %s = %d WHERE id = ?`, column, count+1)
	_, err_1 := db.Query(SQL_1, 1)
	if err_1 != nil {
		return 0, err_1
	}

	return count, nil
}

func parseToInt(value string) (int64, error) {
	valueInt, err_1 := strconv.ParseInt(value, 10, 64)
	if err_1 != nil {
		return 0, fmt.Errorf(err_1.Error())
	} else {
		return valueInt, nil
	}
}
