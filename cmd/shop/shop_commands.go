package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// ----------------------------------------------------------------------------------------------------
// --------- ITEM COMMANDS ---------
// ----------------------------------------------------------------------------------------------------

func browseItems() {
	resp, err := http.Get(baseURL + "/items")
	if err != nil {
		log.Fatal(Red+"Could not reach shop:"+Reset, err)
	}
	defer resp.Body.Close()

	var items []Item
	json.NewDecoder(resp.Body).Decode(&items)

	fmt.Println(Bold + Green + "========= Welcome to the Shop =========" + Reset)
	for i, item := range items {
		fmt.Printf("%d. "+Bold+Blue+"%-20s"+Reset+Yellow+"%dg"+Reset+"  (qty: %d)\n  %s\n\n",
			i+1,
			item.Name,
			item.Price,
			item.Quantity,
			item.Description,
		)
	}
}

func sellItem(idx, quantity string) {
	// make sure user is logged in
	cfg, err := Read()
	if err != nil || cfg.Token == "" {
		fmt.Println("Not logged in. Run 'shop login <name> <password>' first")
		os.Exit(1)
	}

	// fetch users inventory
	req, err := http.NewRequest("GET", baseURL+"/users/inventory", nil)
	if err != nil {
		log.Fatal(Red+"Error making request:"+Reset, err)
	}
	req.Header.Set("Authorization", "Bearer "+cfg.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(Red+"Could not reach server"+Reset, err)
	}
	defer resp.Body.Close()

	if checkResponseError(resp, "Could not get inventory") {
		return
	}

	type InventoryItem struct {
		ID       uuid.UUID `json:"id"`
		Name     string    `json:"name"`
		Price    int32     `json:"price"`
		Quantity int32     `json:"quantity"`
	}
	var inventory []InventoryItem
	json.NewDecoder(resp.Body).Decode(&inventory)

	// index bounds check
	idxInt, err := strconv.Atoi(idx)
	if err != nil {
		fmt.Println("Please enter a valid number")
		os.Exit(1)
	}
	if idxInt < 1 || idxInt > len(inventory) {
		fmt.Println("Invalid item number")
		os.Exit(1)
	}
	qtyInt, err := strconv.Atoi(quantity)
	if err != nil {
		fmt.Println("Please enter a valid number")
		os.Exit(1)
	}

	// make POST request with the quantity
	body := strings.NewReader(fmt.Sprintf(`{"quantity": %d}`, qtyInt))
	req2, err := http.NewRequest("POST", baseURL+"/users/inventory/sell/"+inventory[idxInt-1].ID.String(), body)
	if err != nil {
		log.Fatal("Error making request:", err)
	}
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", "Bearer "+cfg.Token)
	resp2, err := http.DefaultClient.Do(req2)
	if err != nil {
		log.Fatal("Could not reach server:", err)
	}
	defer resp2.Body.Close()

	fmt.Println("Status:", resp2.StatusCode)
	if checkResponseError(resp2, "Could not sell item") {
		return
	}

	var sale struct {
		ItemName          string `json:"item_name"`
		InventoryQuantity int32  `json:"inventory_quantity"`
		TotalPrice        int32  `json:"total_price"`
		Balance           int32  `json:"balance"`
	}
	json.NewDecoder(resp2.Body).Decode(&sale)
	fmt.Printf("You sold "+Bold+Blue+"%dx %s"+Reset+" for "+Yellow+"%dg"+Reset+"\n", qtyInt, sale.ItemName, sale.TotalPrice)
	fmt.Printf("Updated balance: "+Yellow+"%dg"+Reset+"\n", sale.Balance)
}

func buyItem(idx, quantity string) {
	// make sure user is logged in
	cfg, err := Read()
	if err != nil || cfg.Token == "" {
		fmt.Println("Not logged in. Run 'shop login <name> <password>' first")
		os.Exit(1)
	}
	// fetch all items
	itemsResp, err := http.Get(baseURL + "/items")
	if err != nil {
		log.Fatal("Could not reach shop:", err)
	}
	defer itemsResp.Body.Close()
	var items []Item
	json.NewDecoder(itemsResp.Body).Decode(&items)

	// index bounds check
	idxInt, err := strconv.Atoi(idx)
	if err != nil {
		fmt.Println("Please enter a valid number")
		os.Exit(1)
	}
	if idxInt < 1 || idxInt > len(items) {
		fmt.Println("Invalid item number")
		os.Exit(1)
	}
	qtyInt, err := strconv.Atoi(quantity)
	if err != nil {
		fmt.Println("Please enter a valid number")
		os.Exit(1)
	}

	body := strings.NewReader(fmt.Sprintf(`{"quantity": %d}`, qtyInt))
	req, err := http.NewRequest("POST", baseURL+"/items/buy/"+items[idxInt-1].ID.String(), body)
	if err != nil {
		log.Fatal("Error making request:", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.Token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal("Could not reach shop:", err)
	}
	defer resp.Body.Close()

	if checkResponseError(resp, "Purchase failed") {
		return
	}

	var purchase struct {
		ItemName  string `json:"item_name"`
		Quantity  int32  `json:"quantity"`
		TotalCost int32  `json:"total_cost"`
		Balance   int32  `json:"balance"`
	}
	json.NewDecoder(resp.Body).Decode(&purchase)
	fmt.Printf("You purchased "+Bold+Blue+"%dx %s"+Reset+" for "+Yellow+"%dg"+Reset+"\n", qtyInt, purchase.ItemName, purchase.TotalCost)
	fmt.Printf("Updated balance: "+Yellow+"%dg"+Reset+"\n", purchase.Balance)
}

func restockItem(idx, quantity string) {
	cfg, err := Read()
	if err != nil || cfg.Token == "" {
		fmt.Println("Not logged in. Run 'shop login <name> <password>' first")
		os.Exit(1)
	}
	itemsResp, err := http.Get(baseURL + "/items")
	if err != nil {
		log.Fatal("Could not reach shop:", err)
	}
	defer itemsResp.Body.Close()
	var items []Item
	json.NewDecoder(itemsResp.Body).Decode(&items)

	idxInt, err := strconv.Atoi(idx)
	if err != nil {
		fmt.Println("Please enter a valid number")
		os.Exit(1)
	}
	if idxInt < 1 || idxInt > len(items) {
		fmt.Println("Invalid item number")
		os.Exit(1)
	}
	qtyInt, err := strconv.Atoi(quantity)
	if err != nil {
		fmt.Println("Please enter a valid number")
		os.Exit(1)
	}

	body := strings.NewReader(fmt.Sprintf(`{"quantity": %d}`, qtyInt))
	req, err := http.NewRequest("PATCH", baseURL+"/items/restock/"+items[idxInt-1].ID.String(), body)
	if err != nil {
		log.Fatal("Error making request:", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.Token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal("Could not reach shop:", err)
	}
	defer resp.Body.Close()

	if checkResponseError(resp, "Restock failed") {
		return
	}

	var item Item
	json.NewDecoder(resp.Body).Decode(&item)

	fmt.Printf("Restocked: "+Bold+Blue+"%-20s"+Reset+"  (qty: %d)\n", item.Name, item.Quantity)
}

// ----------------------------------------------------------------------------------------------------
// --------- RESPONSE CHECK ----------
// ----------------------------------------------------------------------------------------------------
func checkResponseError(r *http.Response, errMsg string) bool {
	if r.StatusCode != http.StatusOK && r.StatusCode != http.StatusCreated {
		var errResp struct {
			Error string `json:"error"`
		}
		json.NewDecoder(r.Body).Decode(&errResp)
		fmt.Println(Red+errMsg+":"+Reset, errResp.Error)
		return true
	}
	return false
}
