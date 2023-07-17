package main

import (
	"erply-api/modules/database"
	"erply-api/modules/handlers"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/gorilla/mux"
)

func main() {

	database.CreateJSONTable()

	router := mux.NewRouter()
	router.Use(handlers.CheckSessionMiddleware)

	router.HandleFunc("/auth", handlers.AuthHandle).Methods("POST")
	router.HandleFunc("/write", handlers.WriteHandle).Methods("POST")
	router.HandleFunc("/read", handlers.ReadHandle).Methods("GET")

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
				database.Cleanup()
				os.Exit(1)
			}
		}
	}()
	log.Fatal(server.ListenAndServe())
}
