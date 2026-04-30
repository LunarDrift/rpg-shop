package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/LunarDrift/rpg-shop/internal/database"

	_ "github.com/lib/pq"
)

type Server struct {
	db        *database.Queries
	sqlDB     *sql.DB
	jwtSecret string
}

func NewServer(db *database.Queries, sqlDB *sql.DB, jwtSecret string) *Server {
	return &Server{
		db:        db,
		sqlDB:     sqlDB,
		jwtSecret: jwtSecret,
	}
}

func respondWithJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func respondWithError(w http.ResponseWriter, status int, message string, err error) {
	if err != nil {
		log.Println(err)
	}
	type errorResponse struct {
		Error string `json:"error"`
	}
	respondWithJSON(w, status, errorResponse{Error: message})
}
