package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/LunarDrift/rpg-shop/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Couldn't find environment file")
		os.Exit(1)
	}
	connStr := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Could not connect to database:", err)
	}
	defer db.Close()

	queries := database.New(db)
	server := NewServer(queries)

	mux := http.NewServeMux()

	// Routes go here
	mux.HandleFunc("GET /health", handlerHealth)
	mux.HandleFunc("POST /items", server.handlerCreateItem)
	mux.HandleFunc("GET /items", server.handlerGetItems)
	mux.HandleFunc("GET /items/{item_id}", server.handlerGetItemByID)
	mux.HandleFunc("POST /items/buy/{item_id}", server.handlerBuyItem)
	mux.HandleFunc("PATCH /items/restock/{item_id}", server.handlerRestockItem)
	mux.HandleFunc("DELETE /items/{item_id}", server.handlerDeleteItemByID)

	httpServer := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Println("Server starting on port", httpServer.Addr)
	log.Fatal(httpServer.ListenAndServe())
}

func handlerHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintln(w, `{"status":"ok"}`)
}
