package main

import (
	"encoding/json"
	"net/http"

	"github.com/LunarDrift/rpg-shop/internal/database"
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
