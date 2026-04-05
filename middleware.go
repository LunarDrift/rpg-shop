package main

import (
	"context"
	"net/http"

	"github.com/LunarDrift/rpg-shop/internal/auth"
)

// custom key type to avoid context key collisions
type contextKey string

const userIDKey contextKey = "userID"

func (s *Server) middlewareAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := auth.GetBearerToken(r.Header)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "Missing or invalid token", err)
			return
		}
		userID, err := auth.ValidateJWT(token, s.jwtSecret)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "Invalid token", err)
			return
		}
		// attach userID to context
		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next(w, r.WithContext(ctx))
	}
}
