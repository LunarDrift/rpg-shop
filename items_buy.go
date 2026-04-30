package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/LunarDrift/rpg-shop/internal/database"
	"github.com/google/uuid"
)

type PurchaseResult struct {
	ItemName  string `json:"item_name"`
	Quantity  int32  `json:"quantity"`
	TotalCost int32  `json:"total_cost"`
	Balance   int32  `json:"balance"`
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

	result, err := s.ProcessPurchase(r.Context(), userID, itemID, params.Quantity)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error(), err)
	}

	respondWithJSON(w, http.StatusOK, result)
}

func (s *Server) ProcessPurchase(ctx context.Context, userID, itemID uuid.UUID, quantity int32) (PurchaseResult, error) {
	// Fetch user and item
	userItemRow, err := s.db.GetUserAndItem(ctx, database.GetUserAndItemParams{ID: userID, ID_2: itemID})
	if err != nil {
		return PurchaseResult{}, fmt.Errorf("user/item not found: %w", err)
	}

	// Quantity checking
	if userItemRow.Quantity <= 0 {
		return PurchaseResult{}, errors.New("item out of stock")
	}
	if quantity > userItemRow.Quantity {
		return PurchaseResult{}, errors.New("not enough in stock")
	}
	// Balance check
	totalPrice := userItemRow.Price * quantity
	if userItemRow.Balance < totalPrice {
		return PurchaseResult{}, errors.New("not enough gold")
	}

	// Transaction for the three db writes
	tx, err := s.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return PurchaseResult{}, err
	}
	defer tx.Rollback()

	qtx := s.db.WithTx(tx)

	// Make calls to database
	updatedQty, err := qtx.UpdateQuantity(ctx, database.UpdateQuantityParams{
		Quantity: userItemRow.Quantity - quantity,
		ID:       itemID,
	})
	if err != nil {
		return PurchaseResult{}, err
	}
	updatedUser, err := qtx.UpdateBalance(ctx, database.UpdateBalanceParams{
		ID:      userID,
		Balance: userItemRow.Balance - totalPrice,
	})
	if err != nil {
		return PurchaseResult{}, err
	}
	err = qtx.AddToInventory(ctx, database.AddToInventoryParams{
		UserID:   userID,
		ItemID:   itemID,
		Quantity: quantity,
	})
	if err != nil {
		return PurchaseResult{}, err
	}

	if err = tx.Commit(); err != nil {
		return PurchaseResult{}, err
	}

	return PurchaseResult{
		ItemName:  updatedQty.Name,
		Quantity:  updatedQty.Quantity,
		TotalCost: totalPrice,
		Balance:   updatedUser.Balance,
	}, nil
}
