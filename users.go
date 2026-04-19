package main

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/LunarDrift/rpg-shop/internal/auth"
	"github.com/LunarDrift/rpg-shop/internal/database"
	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Balance   int32     `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

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

	dbUser, err := s.db.CreateUser(r.Context(), createUserParams)
	if err != nil {
		// not the best solution; would be better to use pq's error types to check specific postgres error codes
		if strings.Contains(err.Error(), "unique constraint") {
			respondWithError(w, http.StatusBadRequest, "Username already taken", err)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Could not create user", err)
		return
	}

	// map to main.User so we're not returning hashedPassword in the payload
	user := User{
		ID:        dbUser.ID,
		Name:      dbUser.Name,
		Balance:   dbUser.Balance,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
	}
	respondWithJSON(w, http.StatusCreated, user)
}

func (s *Server) handlerLogIn(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Name     string `json:"name"`
		Password string `json:"password"`
	}
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	dbUser, err := s.db.GetUserByName(r.Context(), params.Name)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect name or password", err)
		return
	}

	// check password against stored hash
	match, err := auth.CheckPasswordHash(params.Password, dbUser.HashedPassword)
	if err != nil || !match {
		respondWithError(w, http.StatusUnauthorized, "Incorrect name or password", err)
		return
	}

	// build token
	accessToken, err := auth.MakeJWT(dbUser.ID, s.jwtSecret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not make token", err)
		return
	}

	type loginResponse struct {
		ID        uuid.UUID `json:"id"`
		Name      string    `json:"name"`
		Balance   int32     `json:"balance"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Token     string    `json:"token"`
	}
	loginResp := loginResponse{
		ID:        dbUser.ID,
		Name:      dbUser.Name,
		Balance:   dbUser.Balance,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Token:     accessToken,
	}
	respondWithJSON(w, http.StatusOK, loginResp)
}

func (s *Server) handlerGetUserByID(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.PathValue("id")

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	dbUser, err := s.db.GetUserByID(r.Context(), userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not fetch user", err)
		return
	}

	user := User{
		ID:        dbUser.ID,
		Name:      dbUser.Name,
		Balance:   dbUser.Balance,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
	}
	respondWithJSON(w, http.StatusOK, user)
}

func (s *Server) handlerGetUser(w http.ResponseWriter, r *http.Request) {
	userName := r.URL.Query().Get("name")

	// Fetch all users if no username specified
	if userName == "" {
		dbUsers, err := s.db.GetAllUsers(r.Context())
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Could not fetch users", err)
			return
		}

		users := []User{}
		for _, user := range dbUsers {
			users = append(users, User{
				ID:        user.ID,
				Name:      user.Name,
				Balance:   user.Balance,
				CreatedAt: user.CreatedAt,
				UpdatedAt: user.UpdatedAt,
			})
		}
		respondWithJSON(w, http.StatusOK, users)
		return
	}

	dbUser, err := s.db.GetUserByName(r.Context(), userName)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "User not found", err)
		return
	}

	user := User{
		ID:        dbUser.ID,
		Name:      dbUser.Name,
		Balance:   dbUser.Balance,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
	}
	respondWithJSON(w, http.StatusOK, user)
}

func (s *Server) handlerEarnGold(w http.ResponseWriter, r *http.Request) {
	// get ID from Context
	userID, ok := r.Context().Value(userIDKey).(uuid.UUID)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Could not get user ID", nil)
		return
	}
	// Get user by ID for current balance
	dbUser, err := s.db.GetUserByID(r.Context(), userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not fetch user", err)
		return
	}
	// Generate random reward with rand.Intn
	reward := max(10, rand.Intn(100)) // random between 10-100 gold

	// Build UpdateBalanceParams
	balanceParams := database.UpdateBalanceParams{
		ID:      userID,
		Balance: dbUser.Balance + int32(reward),
	}

	// Call Query
	updatedDBUser, err := s.db.UpdateBalance(r.Context(), balanceParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not update user", err)
		return
	}

	// convert to main.User
	user := User{
		ID:        updatedDBUser.ID,
		Name:      updatedDBUser.Name,
		Balance:   updatedDBUser.Balance,
		CreatedAt: updatedDBUser.CreatedAt,
		UpdatedAt: updatedDBUser.UpdatedAt,
	}
	respondWithJSON(w, http.StatusOK, user)
}

func (s *Server) handlerGetInventory(w http.ResponseWriter, r *http.Request) {
	// get ID from Context
	userID, ok := r.Context().Value(userIDKey).(uuid.UUID)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Could not get user ID", nil)
		return
	}

	inventoryRow, err := s.db.GetUserInventory(r.Context(), userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get user inventory", err)
		return
	}

	respondWithJSON(w, http.StatusOK, inventoryRow)
}

func (s *Server) handlerSellItem(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(userIDKey).(uuid.UUID)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Could not get user ID", nil)
		return
	}

	var params struct {
		Quantity int32 `json:"quantity"`
	}
	err := json.NewDecoder(r.Body).Decode(&params)
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

	userItemParams := database.GetUserAndItemParams{ID: userID, ID_2: itemID}
	userItemRow, err := s.db.GetUserAndItem(r.Context(), userItemParams)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "User/Item not found", err)
		return
	}

	currentItem, err := s.db.GetUserInventoryItem(r.Context(), database.GetUserInventoryItemParams{
		UserID: userID,
		ItemID: itemID,
	})
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Item not in inventory", err)
		return
	}

	if params.Quantity > currentItem.Quantity {
		respondWithError(w, http.StatusBadRequest, "Not enough in inventory", nil)
		return
	}

	updateInventoryParams := database.UpdateInventoryQuantityParams{
		UserID:   userID,
		ItemID:   itemID,
		Quantity: currentItem.Quantity - params.Quantity,
	}
	qtyParams := database.UpdateQuantityParams{
		Quantity: userItemRow.Quantity + params.Quantity,
		ID:       itemID,
	}
	totalPrice := userItemRow.Price * int32(params.Quantity)
	balanceParams := database.UpdateBalanceParams{
		ID:      userID,
		Balance: userItemRow.Balance + totalPrice,
	}

	updatedInventory, err := s.db.UpdateInventoryQuantity(r.Context(), updateInventoryParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not update inventory", err)
		return
	}
	updatedQty, err := s.db.UpdateQuantity(r.Context(), qtyParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not update item quantity", err)
		return
	}
	updatedBalance, err := s.db.UpdateBalance(r.Context(), balanceParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not update balance", err)
		return
	}

	type transactionResponse struct {
		ItemName          string `json:"item_name"`
		ShopQuantity      int32  `json:"shop_quantity"`
		InventoryQuantity int32  `json:"inventory_quantity"`
		TotalPrice        int32  `json:"total_price"`
		Balance           int32  `json:"balance"`
	}
	respondWithJSON(w, http.StatusOK, transactionResponse{
		ItemName:          updatedQty.Name,
		ShopQuantity:      updatedQty.Quantity,
		InventoryQuantity: updatedInventory.Quantity,
		TotalPrice:        totalPrice,
		Balance:           updatedBalance.Balance,
	})
}
