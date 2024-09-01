package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	_ "modernc.org/sqlite"

	"golang.org/x/crypto/bcrypt"
)

func hashPassword(password string) (string, error) {

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func checkHashedPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))

	return err == nil
}

func generateJWT(userID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id":   userID,
		"logged_in": true,
		"exp":       time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func validateJWT(tokenString string) (*jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return &claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

func validateLoggedIn(token string) (string, error) {
	claims, err := validateJWT(token)
	if err != nil {
		return "", err
	}

	loggedInID := (*claims)["user_id"].(string)
	return loggedInID, nil
}

func validateAdmin(c *gin.Context) bool {
	token := c.GetHeader("token")

	id, err := validateLoggedIn(token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error_1": err.Error()})
	}

	SQL_id := `SELECT is_admin FROM users WHERE id = ?`
	rows, err := db.Query(SQL_id, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error_2": err.Error()})
	}
	defer rows.Close()

	var isAdmin bool
	if rows.Next() {
		err = rows.Scan(&isAdmin)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error_3": err.Error()})
		}
	}

	if !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{"message": "admin access only"})
	}

	return isAdmin
}
