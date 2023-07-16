package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/erply/api-go-wrapper/pkg/api"
)

func TestAuthHandle(t *testing.T) {
	// Create a new request with form values
	form := url.Values{}
	form.Set("username", "rolandlehes@gmail.com")
	form.Set("password", "Buddernorton55")
	form.Set("code", "532358")

	req, err := http.NewRequest("POST", "/auth", strings.NewReader(form.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(authHandle)

	handler.ServeHTTP(rr, req)

	// Check the response status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body
	expectedBody := "Authentication successful"
	if rr.Body.String() != expectedBody {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expectedBody)
	}
}

func TestReadHandle(t *testing.T) {
	// Create a mock client for testing
	mockClient := &api.Client{} // Create a mock implementation of the api.Client interface

	// Create a mock request with query parameters
	req, err := http.NewRequest("GET", "/read", nil)
	if err != nil {
		t.Fatal(err)
	}
	query := req.URL.Query()
	query.Set("searchName", "roland lehes")
	req.URL.RawQuery = query.Encode()

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(readHandle)

	// Set the mock client in the request context
	ctx := context.WithValue(req.Context(), "client", mockClient)
	req = req.WithContext(ctx)

	handler.ServeHTTP(rr, req)

	// Check the response status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

/*
func TestWriteHandle(t *testing.T) {
	// Create a mock client for testing
	mockClient := &api.Client{} // Create a mock implementation of the api.Client interface

	// Create a mock request with query parameters
	req, err := http.NewRequest("POST", "/write", nil)
	if err != nil {
		t.Fatal(err)
	}
	query := req.URL.Query()
	query.Set("param1", "value1")
	query.Set("param2", "value2")
	req.URL.RawQuery = query.Encode()

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(writeHandle)

	// Set the mock client in the request context
	ctx := context.WithValue(req.Context(), "client", mockClient)
	req = req.WithContext(ctx)

	handler.ServeHTTP(rr, req)

	// Check the response status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestQueryCustomer(t *testing.T) {
	// Set up a test database with sample data
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Create the customer_search_cache table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS customer_search_cache (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			search_parameters JSONB NOT NULL,
			search_results JSONB NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatal(err)
	}

	// Insert a sample row into the customer_search_cache table
	searchParams := `{"param1": "value1", "param2": "value2"}`
	searchResults := `[{"id": 1, "name": "John Doe"}, {"id": 2, "name": "Jane Smith"}]`
	_, err = db.Exec("INSERT INTO customer_search_cache (search_parameters, search_results) VALUES (?, ?)", searchParams, searchResults)
	if err != nil {
		t.Fatal(err)
	}

	// Call the queryCustomer function to retrieve the cached results
	customers, err := queryCustomer([]byte(searchParams), db)
	if err != nil {
		t.Fatal(err)
	}

	// Check the retrieved customers
	expectedCustomers := []customers.Customer{
		{ID: 1, Name: "John Doe"},
		{ID: 2, Name: "Jane Smith"},
	}
	if !reflect.DeepEqual(customers, expectedCustomers) {
		t.Errorf("queryCustomer returned unexpected results: got %v want %v", customers, expectedCustomers)
	}
}

func TestInsertCustomers(t *testing.T) {
	// Set up a test database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Create the customer_search_cache table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS customer_search_cache (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			search_parameters JSONB NOT NULL,
			search_results JSONB NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatal(err)
	}

	// Insert sample customers
	customers := []customers.Customer{
		{ID: 1, Name: "John Doe"},
		{ID: 2, Name: "Jane Smith"},
	}

	// Convert customers to JSON
	customersJSON, err := json.Marshal(customers)
	if err != nil {
		t.Fatal(err)
	}

	// Insert the customers into the database
	searchParams := `{"param1": "value1", "param2": "value2"}`
	err = insertCustomers(customers, []byte(searchParams), db)
	if err != nil {
		t.Fatal(err)
	}

	// Query the customer_search_cache table to retrieve the inserted results
	row := db.QueryRow("SELECT search_results FROM customer_search_cache WHERE search_parameters = ?", searchParams)
	var searchResultsJSON string
	err = row.Scan(&searchResultsJSON)
	if err != nil {
		t.Fatal(err)
	}

	// Compare the retrieved search results with the expected results
	if searchResultsJSON != string(customersJSON) {
		t.Errorf("insertCustomers failed to insert the customers correctly")
	}
}

func TestCreateJSONTable(t *testing.T) {
	// Set up a test database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Create the JSON table using the createJSONTable function
	err = createJSONTable(db)
	if err != nil {
		t.Fatal(err)
	}

	// Check if the customer_search_cache table exists
	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table' AND name='customer_search_cache'")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	if !rows.Next() {
		t.Errorf("createJSONTable failed to create the customer_search_cache table")
	}
}

func TestCheckSessionMiddleware(t *testing.T) {
	// Set up a test server and router
	router := mux.NewRouter()
	router.Use(checkSessionMiddleware)
	router.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Create a mock request
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	// Call the checkSessionMiddleware directly with the mock request and response recorder
	checkSessionMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rr, req)

	// Check the response status code
	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("checkSessionMiddleware returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
	}
}
*/
