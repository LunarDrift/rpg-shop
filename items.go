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

	dbItemParams := database.CreateItemParams{
		Name:        params.Name,
		Description: params.Description,
		Price:       params.Price,
		Quantity:    params.Quantity,
	}

	item, err := s.db.CreateItem(r.Context(), dbItemParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create item", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, item)
}

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

	qtyParams := database.UpdateQuantityParams{
		Quantity: userItemRow.Quantity - params.Quantity,
		ID:       itemID,
	}

	balanceParams := database.UpdateBalanceParams{
		ID:      userID,
		Balance: userItemRow.Balance - totalPrice,
	}

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
