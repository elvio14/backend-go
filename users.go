package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	_ "modernc.org/sqlite"
)

type User struct {
	ID           string
	IsAdmin      bool
	IsRegistered bool
	Username     string
	Email        string
	Password     string
	CreatedAt    time.Time
}

type UserLogin struct {
	Username string
	Password string
}

func initializeUserTable() error {
	SQL := `
		CREATE TABLE IF NOT EXISTS users (
			id TEXT,
			is_admin BOOLEAN,
			is_registered BOOLEAN,
			username TEXT,
			email TEXT,
			password TEXT,
			created_at TIMESTAMP
		)
	`
	_, err := db.Exec(SQL)
	if err != nil {
		return err
	} else {
		return nil
	}
}

func addCurrentUser(userID string) error {
	SQL := `INSERT INTO users (id, is_admin, is_registered, created_at) VALUES (?, ?, ?, ?)`

	_, err := db.Exec(SQL, userID, 1, 1, time.Now())
	if err != nil {
		return err
	} else {
		return nil
	}
}

func generateUserID() (string, error) {
	count, err := getCount("users")
	if err != nil {
		return "", err
	}

	userID := fmt.Sprintf("%06d", count+1)
	userID = "U" + userID

	return userID, nil
}

func register(c *gin.Context, userID string) {
	var user User
	if err := c.ShouldBindBodyWithJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error_1": err.Error()})
		return
	}

	hashedPassword, err := hashPassword(user.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error_2": err.Error()})
		return
	}

	SQL := `INSERT OR REPLACE INTO users (
				id,
				is_admin,
				is_registered,
				username,
				email,
				password,
				created_at
			)
			VALUES (?, ?, ?, ?, ?, ?, ?)
			`

	_, err_1 := db.Exec(SQL, userID, user.IsAdmin, 1, user.Username, user.Email, hashedPassword, time.Now())
	if err_1 != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error_3": err_1.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User added.", "user_id": userID})
}

func login(c *gin.Context) {
	var login UserLogin
	if err := c.ShouldBindBodyWithJSON(&login); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error_1": err.Error(), "message": "bad request"})
		return
	}

	SQL_id := `SELECT * FROM users WHERE username = ?`
	rows, err := db.Query(SQL_id, login.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error_2": err.Error()})
		return
	}
	defer rows.Close()

	var currentUser User
	if rows.Next() {
		err = rows.Scan(&currentUser.ID, &currentUser.IsAdmin, &currentUser.IsRegistered, &currentUser.Username, &currentUser.Email, &currentUser.Password, &currentUser.CreatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error_3": err.Error()})
			return
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"message": "user doesn't exist"})
		return
	}

	isCorrect := checkHashedPassword(login.Password, currentUser.Password)

	var token string
	if isCorrect {
		result, err := generateJWT(currentUser.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error_4": err.Error()})
			return
		}
		token = result
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"message": "wrong password."})
		return
	}

	loggedInID, err := validateLoggedIn(token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error_5": err.Error()})
		return
	}

	isAdmin, err := getAdmin(loggedInID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error_6": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token, "user_id": loggedInID, "is_admin": isAdmin})
}
