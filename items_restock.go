package main

import (
	"encoding/json"
	"net/http"

	"github.com/LunarDrift/rpg-shop/internal/database"
	"github.com/google/uuid"
)

func (s *Server) handlerRestockItem(w http.ResponseWriter, r *http.Request) {
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
		Quantity int32 `json:"quantity"`
	}

	err = json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	itemIDStr := r.PathValue("item_id")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid item ID", err)
		return
	}

	item, err := s.db.GetItemByID(r.Context(), itemID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Item not found", err)
		return
	}

	updateParams := database.UpdateQuantityParams{
		ID:       itemID,
		Quantity: item.Quantity + params.Quantity,
	}

	updatedItem, err := s.db.UpdateQuantity(r.Context(), updateParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not update item", err)
		return
	}

	respondWithJSON(w, http.StatusOK, updatedItem)
}
