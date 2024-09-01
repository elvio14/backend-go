package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v79"
	"github.com/stripe/stripe-go/v79/checkout/session"
)

func createCheckoutSession(c *gin.Context) string {
	orderID := c.Param("order_id")
	SQL := `SELECT total FROM orders WHERE order_id = ?`
	row, err := db.Query(SQL, orderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "error getting price", "error": err.Error()})
		return ""
	}
	defer row.Close()

	var total int

	if row.Next() {
		err := row.Scan(&total)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "error scanning row", "error": err.Error()})
			return ""
		}
	}

	domain := "http://localhost:5000"
	params := &stripe.CheckoutSessionParams{
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String("cad"),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String(orderID),
					},
					UnitAmount: stripe.Int64(int64(total)),
				},
				Quantity: stripe.Int64(1),
			},
		},
		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL: stripe.String(domain + "/success.html"),
		CancelURL:  stripe.String(domain + "/cancel.html"),
	}

	s, err := session.New(params)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return ""
	}

	c.Redirect(303, s.URL)

	paymentIntentID := s.PaymentIntent.ID

	return paymentIntentID
}
