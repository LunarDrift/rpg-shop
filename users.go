package main

import (
	"encoding/json"
	"net/http"

	"github.com/LunarDrift/rpg-shop/internal/database"
	"github.com/google/uuid"
)

func (s *Server) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Name    string `json:"name"`
		Balance int32  `json:"balance"`
	}
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	dbUserParams := database.CreateUserParams{
		Name:    params.Name,
		Balance: params.Balance,
	}

	user, err := s.db.CreateUser(r.Context(), dbUserParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create user", err)
		return
	}
	respondWithJSON(w, http.StatusCreated, user)
}

func (s *Server) handlerGetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := s.db.GetAllUsers(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not fetch users", err)
		return
	}
	respondWithJSON(w, http.StatusOK, users)
}

func (s *Server) handlerGetUserByID(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.PathValue("id")

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	user, err := s.db.GetUserByID(r.Context(), userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not fetch user", err)
		return
	}
	respondWithJSON(w, http.StatusOK, user)
}
