package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"

	"github.com/erply/api-go-wrapper/pkg/api"
	"github.com/erply/api-go-wrapper/pkg/api/customers"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"
	httpSwagger "github.com/swaggo/http-swagger"
)


var err error
var db *sql.DB
var store *sessions.CookieStore
var (
	clients map[string]*api.Client
	mutex   sync.Mutex
)

func main() {

	createJSONTable()
	store = sessions.NewCookieStore([]byte("your-secret-key"))
	store.Options = &sessions.Options{
		MaxAge: 0,
	}
	router := mux.NewRouter()
	//router.Use(checkSessionMiddleware)

	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)
	router.HandleFunc("/auth", authHandle).Methods("POST")
	router.HandleFunc("/write", writeHandle).Methods("POST")
	router.HandleFunc("/read", readHandle).Methods("GET")

	log.Println("Server started on http://localhost:8080")
	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	// Catches termination and cleans up
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			if sig == os.Interrupt {
				cleanup()
				os.Exit(1)
			}
		}
	}()
	log.Fatal(server.ListenAndServe())
}

func init() {
	clients = make(map[string]*api.Client)
}

func cleanup() {
	_, err := db.Exec("DROP TABLE IF EXISTS customer_search_cache")
	if err != nil {
		log.Println("Error dropping table:", err)
	}
	if err := db.Close(); err != nil {
		log.Println("Error closing database connection:", err)
	}
}

// @Summary Authenticate User
// @Description Authenticate user with username, password, and code
// @Tags Authentication
// @Accept json
// @Produce json
// @Param username formData string true "Username"
// @Param password formData string true "Password"
// @Param code formData string true "Code"
// @Success 200 {string} string "Authentication successful"
// @Failure 400 {string} string "Invalid credentials"
// @Failure 401 {string} string "Authentication failed"
// @Router /auth [post]
func authHandle(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")
	code := r.FormValue("code")

	// Validate the user credentials
	if username == "" || password == "" || code == "" {
		http.Error(w, "Invalid credentials", http.StatusBadRequest)
		return
	}

	// Create a client for the user with their credentials
	client, err := api.NewClientFromCredentials(username, password, code, nil)
	if err != nil {
		http.Error(w, "Authentication failed", http.StatusUnauthorized)
		return
	}

	// Create a session for the authenticated user
	session, err := store.Get(r, "session-name")
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	// Store the user's code in the session
	session.Values["code"] = code

	// Save the session
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		return
	}

	mutex.Lock()
	clients[code] = client
	mutex.Unlock()

	fmt.Println("auth successful", code, username, password)
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Authentication successful")
}

/*
// Checks session every request. GetSession gives new session if expired
func checkSessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()
		if r.URL.Path == "/auth" || r.URL.Path == "/swagger/index.html"{
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		session, err := store.Get(r, "session-name")
		if err != nil {
			http.Error(w, "Failed to get session", http.StatusInternalServerError)
			return
		}

		code, ok := session.Values["code"].(string)
		if !ok {
			http.Error(w, "Code not found. Unauthorized", http.StatusInternalServerError)
			return
		}
		fmt.Println("code: ", code)

		// Retrieve the Client for the user
		mutex.Lock()
		client := clients[code]
		mutex.Unlock()
		if client == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		_, err = client.GetSession()
		if err != nil {
			http.Error(w, "Unauthorized from erply, use /auth", http.StatusUnauthorized)
			return
		}
		// Attach the client to the request context
		ctx = context.WithValue(r.Context(), "client", client)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)

	})
}
*/
// @Summary Read Customer Data
// @Description Retrieve customer data based on query parameters
// @Tags Customers
// @Accept json
// @Produce json
// @Param param1 query string false "Parameter 1"
// @Param param2 query string false "Parameter 2"
// @Success 200 {array} customers.Customer
// @Failure 500 {string} string "Internal Server Error"
// @Router /read [get]
// Gets request, checks database if isn't cached sends API request and caches response
func readHandle(w http.ResponseWriter, r *http.Request) {
	client := r.Context().Value("client").(*api.Client)
	queryParams := r.URL.Query()
	filters := make(map[string]string)
	for key, values := range queryParams {
		if len(values) > 0 {
			filters[key] = values[0]
		}
	}
	fmt.Println("filters: ", filters)
	searchParamsJSON, err := json.Marshal(filters)
	if err != nil {
		fmt.Println(err)
	}
	customers, err := queryCustomer(searchParamsJSON)
	if err != nil {
		customers, err = client.CustomerManager.GetCustomers(r.Context(), filters)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		fmt.Println("API pinged")
		insertCustomers(customers, searchParamsJSON)
	}
	fmt.Println(customers)
	responseJSON, err := json.Marshal(customers)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(responseJSON)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// @Summary Write Data
// @Description Write data with query parameters
// @Tags Data
// @Accept json
// @Produce json
// @Param param1 query string true "Parameter 1"
// @Param param2 query string true "Parameter 2"
// @Success 200 {string} string "Data written successfully"
// @Failure 400 {string} string "Invalid parameters"
// @Failure 500 {string} string "Internal Server Error"
// @Router /write [post]
// Only calls SaveCustomer
func writeHandle(w http.ResponseWriter, r *http.Request) {
	client := r.Context().Value("client").(*api.Client)
	queryParams := r.URL.Query()
	filters := make(map[string]string)
	for key, values := range queryParams {
		if len(values) > 0 {
			filters[key] = values[0]
		}
	}

	report, err := client.CustomerManager.SaveCustomer(r.Context(), filters)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println("report ", report)
	responseJSON, err := json.Marshal(report)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(responseJSON)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// Checking cache
func queryCustomer(searchParamsJSON []byte) ([]customers.Customer, error) {
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
func insertCustomers(customers []customers.Customer, searchParamsJSON []byte) {
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

func createJSONTable() {
	db, err = sql.Open("sqlite3", "./customers.db")
	if err != nil {
		log.Println(err)
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
	}
}
