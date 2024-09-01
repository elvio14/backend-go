package main

import (
	_ "modernc.org/sqlite"
)

type OrderItem struct {
	OrderItemID int
	OrderID     string
	ProductID   string
	Size        string
	Price       int
	Quantity    int
	Notes       string
}

func initializeOrderItemsTable() error {
	SQL := `
		CREATE TABLE IF NOT EXISTS order_items (
			item_id INTEGER PRIMARY KEY AUTOINCREMENT,
			order_id TEXT,
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
