package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/LunarDrift/rpg-shop/internal/auth"
	"github.com/LunarDrift/rpg-shop/internal/database"
	"github.com/google/uuid"
)

func (s *Server) handlerRegisterUser(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Name     string `json:"name"`
		Password string `json:"password"`
	}
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not hash password", err)
		return
	}

	createUserParams := database.CreateUserParams{
		Name:           params.Name,
		HashedPassword: hashedPassword,
	}

	user, err := s.db.CreateUser(r.Context(), createUserParams)
	if err != nil {
		// not the best solution; would be better to use pq's error types to check specific postgres error codes
		if strings.Contains(err.Error(), "unique constraint") {
			respondWithError(w, http.StatusBadRequest, "Username already taken", err)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Could not create user", err)
		return
	}
	respondWithJSON(w, http.StatusCreated, user)
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

func (s *Server) handlerGetUser(w http.ResponseWriter, r *http.Request) {
	userName := r.URL.Query().Get("name")

	// Fetch all users if no username specified
	if userName == "" {
		users, err := s.db.GetAllUsers(r.Context())
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Could not fetch users", err)
			return
		}
		respondWithJSON(w, http.StatusOK, users)
		return
	}

	user, err := s.db.GetUserByName(r.Context(), userName)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "User not found", err)
		return
	}
	respondWithJSON(w, http.StatusOK, user)
}
