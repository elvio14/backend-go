package main

import (
	"fmt"

	_ "modernc.org/sqlite"
)

func initializeCounterTable() error {
	SQL_0 := `CREATE TABLE IF NOT EXISTS counter (
				id INT PRIMARY KEY,
				users INT,
				products INT
			)`

	_, err_0 := db.Exec(SQL_0)
	if err_0 != nil {
		return fmt.Errorf("failed to create counter table: %v", err_0)
	}

	SQL_1 := `INSERT OR IGNORE INTO counter (id, users, products) VALUES (?, ?, ?)`
	_, err_1 := db.Exec(SQL_1, 1, 0, 0)
	if err_1 != nil {
		return fmt.Errorf("failed to create counter table: %v", err_1)
	}

	return nil
}
