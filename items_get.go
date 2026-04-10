package main

import (
	"net/http"

	"github.com/google/uuid"
)

func (s *Server) handlerGetItems(w http.ResponseWriter, r *http.Request) {
	items, err := s.db.GetAllItems(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not fetch items", err)
		return
	}
	respondWithJSON(w, http.StatusOK, items)
}

func (s *Server) handlerGetItemByID(w http.ResponseWriter, r *http.Request) {
	itemIDStr := r.PathValue("item_id")

	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid item ID", err)
		return
	}

	item, err := s.db.GetItemByID(r.Context(), itemID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not fetch item", err)
		return
	}
	respondWithJSON(w, http.StatusOK, item)
}
