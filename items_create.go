package main

import (
	"encoding/json"
	"net/http"

	"github.com/LunarDrift/rpg-shop/internal/database"
	"github.com/google/uuid"
)

func (s *Server) handlerCreateItem(w http.ResponseWriter, r *http.Request) {
	// token validation from context
	userID, ok := r.Context().Value(userIDKey).(uuid.UUID)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Could not get user ID", nil)
		return
	}

	// check if admin
	dbUser, err := s.db.GetUserByID(r.Context(), userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not fetch user", err)
		return
	}
	if !dbUser.IsAdmin {
		respondWithError(w, http.StatusForbidden, "Not authorized to do that", nil)
		return
	}

	var params struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Price       int32  `json:"price"`
		Quantity    int32  `json:"quantity"`
	}

	err = json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	createItemParams := database.CreateItemParams{
		Name:        params.Name,
		Description: params.Description,
		Price:       params.Price,
		Quantity:    params.Quantity,
	}

	item, err := s.db.CreateItem(r.Context(), createItemParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create item", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, item)
}
