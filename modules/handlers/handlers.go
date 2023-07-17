package handlers

import (
	"context"
	"encoding/json"
	"erply-api/modules/database"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/erply/api-go-wrapper/pkg/api"
	"github.com/gorilla/sessions"
)

var mutex sync.Mutex
var store *sessions.CookieStore
var clients map[string]*api.Client

func init() {
	clients = make(map[string]*api.Client)
	store = sessions.NewCookieStore([]byte("your-secret-key"))
	store.Options = &sessions.Options{
		MaxAge: 0,
	}
}

// Takes all values necessary to use NewClientFromCredentials and stores in session
func AuthHandle(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")
	code := r.FormValue("code")

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

	session, err := store.Get(r, "session-name")
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	session.Values["code"] = code

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

// Checks session every request. GetSession gives new session if expired
func CheckSessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()
		if r.URL.Path == "/auth" {
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

// Gets request, checks database if isn't cached sends API request and caches response
func ReadHandle(w http.ResponseWriter, r *http.Request) {
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

	customers, err := database.QueryCustomer(searchParamsJSON)
	if err != nil {
		customers, err = client.CustomerManager.GetCustomers(r.Context(), filters)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		fmt.Println("API pinged")
		database.InsertCustomers(customers, searchParamsJSON)
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

// Only calls SaveCustomer
func WriteHandle(w http.ResponseWriter, r *http.Request) {
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
