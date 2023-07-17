package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	_ "github.com/mattn/go-sqlite3"

	"github.com/erply/api-go-wrapper/pkg/api/customers"
)

var db *sql.DB
var err error

func Cleanup() (error){
	_, err = db.Exec("DROP TABLE IF EXISTS customer_search_cache")
	if err != nil {
		log.Println("Error dropping table:", err)
		return err
	}
	if err = db.Close(); err != nil {
		log.Println("Error closing database connection:", err)
		return err
	}
	return nil
}

// Checking cache
func QueryCustomer(searchParamsJSON []byte) ([]customers.Customer, error) {
	row := db.QueryRow("SELECT search_results FROM customer_search_cache WHERE search_parameters = $1", string(searchParamsJSON))
	var searchResultsJSON string
	err = row.Scan(&searchResultsJSON)
	var customers []customers.Customer
	if err == nil {
		err = json.Unmarshal([]byte(searchResultsJSON), &customers)
		if err != nil {
			return nil, err
		}
		fmt.Println("Cahce pinged")
		return customers, nil
	} else if err != sql.ErrNoRows {
		return nil, err
	}
	return nil, err
}

// Cache results
func InsertCustomers(customers []customers.Customer, searchParamsJSON []byte) {
	searchResultsJSON, err := json.Marshal(customers)
	if err != nil {
		fmt.Println(err)
	}
	_, err = db.Exec("INSERT INTO customer_search_cache (search_parameters, search_results) VALUES ($1, $2)",
		string(searchParamsJSON), string(searchResultsJSON))
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Customer cahced")
}

func CreateJSONTable() (error) {
	db, err = sql.Open("sqlite3", "./customers.db")
	if err != nil {
		log.Println(err)
		return err
	}
	createTable := `
	CREATE TABLE IF NOT EXISTS customer_search_cache (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		search_parameters JSONB NOT NULL,
		search_results JSONB NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)
	`
	_, err = db.Exec(createTable)
	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}
