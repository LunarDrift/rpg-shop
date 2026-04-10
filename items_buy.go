package main

import (
	"encoding/json"
	"net/http"

	"github.com/LunarDrift/rpg-shop/internal/database"
	"github.com/google/uuid"
)

func (s *Server) handlerBuyItem(w http.ResponseWriter, r *http.Request) {
	// token validation from context
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

	// Quantity checking
	if userItemRow.Quantity <= 0 {
		respondWithError(w, http.StatusBadRequest, "Item out of stock", err)
		return
	}
	if params.Quantity > userItemRow.Quantity {
		respondWithError(w, http.StatusBadRequest, "Not enough in stock", err)
		return
	}
	// Balance check
	totalPrice := userItemRow.Price * int32(params.Quantity)
	if userItemRow.Balance < totalPrice {
		respondWithError(w, http.StatusBadRequest, "Not enough gold", err)
		return
	}

	// Build param structs for update queries
	addToInventoryParams := database.AddToInventoryParams{
		UserID:   userID,
		ItemID:   itemID,
		Quantity: params.Quantity,
	}
	qtyParams := database.UpdateQuantityParams{
		Quantity: userItemRow.Quantity - params.Quantity,
		ID:       itemID,
	}

	balanceParams := database.UpdateBalanceParams{
		ID:      userID,
		Balance: userItemRow.Balance - totalPrice,
	}

	// Make calls to database
	// TODO: 3 separate calls to database is bad. Learn how to use transactions so that data stays consistent if anything fails
	updatedQty, err := s.db.UpdateQuantity(r.Context(), qtyParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not update item", err)
		return
	}
	updatedUser, err := s.db.UpdateBalance(r.Context(), balanceParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not update user", err)
		return
	}
	err = s.db.AddToInventory(r.Context(), addToInventoryParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not update inventory", err)
		return
	}

	type purchaseResponse struct {
		ItemName  string `json:"item_name"`
		Quantity  int32  `json:"quantity"`
		TotalCost int32  `json:"total_cost"`
		Balance   int32  `json:"balance"`
	}
	respondWithJSON(w, http.StatusOK, purchaseResponse{
		ItemName:  updatedQty.Name,
		Quantity:  updatedQty.Quantity,
		TotalCost: totalPrice,
		Balance:   updatedUser.Balance,
	})
}
