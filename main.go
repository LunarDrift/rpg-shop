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

	// get env variables
	connStr := os.Getenv("DB_URL")
	jwtSecret := os.Getenv("JWT_SECRET")

	// open connection to database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Could not connect to database:", err)
	}
	defer db.Close()

	queries := database.New(db)
	s := NewServer(queries, jwtSecret)

	mux := http.NewServeMux()

	// -- Item routes --
	mux.HandleFunc("GET /health", handlerHealth)
	mux.HandleFunc("POST /items", s.handlerCreateItem)
	mux.HandleFunc("GET /items", s.handlerGetItems)
	mux.HandleFunc("GET /items/{item_id}", s.handlerGetItemByID)
	mux.HandleFunc("POST /items/buy/{item_id}", s.middlewareAuth(s.handlerBuyItem))
	mux.HandleFunc("PATCH /items/restock/{item_id}", s.middlewareAuth(s.handlerRestockItem))
	mux.HandleFunc("DELETE /items/{item_id}", s.middlewareAuth(s.handlerDeleteItemByID))

	// -- User routes --
	mux.HandleFunc("POST /users", s.handlerRegisterUser)
	mux.HandleFunc("POST /users/login", s.handlerLogIn)
	mux.HandleFunc("GET /users/{id}", s.handlerGetUserByID)
	mux.HandleFunc("GET /users", s.handlerGetUser)

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
