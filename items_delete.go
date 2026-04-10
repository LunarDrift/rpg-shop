package main

import (
	"net/http"

	"github.com/google/uuid"
)

func (s *Server) handlerDeleteItemByID(w http.ResponseWriter, r *http.Request) {
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

	itemIDStr := r.PathValue("item_id")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid item ID", err)
		return
	}

	err = s.db.DeleteItemByID(r.Context(), itemID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not delete item", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
