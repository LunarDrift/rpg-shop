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

func buyItem(idx string, quantity string) {
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
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal("Could not reach shop:", err)
	}
	defer resp.Body.Close()

	if checkResponseError(resp, "Purchase failed") {
		return
	}

	var item Item
	json.NewDecoder(resp.Body).Decode(&item)
	fmt.Printf("You purchased "+Bold+Blue+"%dx %s"+Reset+" for "+Yellow+"%dg"+Reset+"\n", qtyInt, item.Name, item.Price*int32(qtyInt))
}

func restockItem(idx string, quantity string) {
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
// --------- USER COMMANDS ----------
// ----------------------------------------------------------------------------------------------------

type User struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

func login(name string) {
	// make request to api
	resp, err := http.Get(baseURL + "/users?name=" + name)
	if err != nil {
		log.Fatal("Could not reach server:", err)
	}
	defer resp.Body.Close()

	if checkResponseError(resp, "User not found") {
		return
	}

	// set user
	var user User
	json.NewDecoder(resp.Body).Decode(&user)

	cfg := Config{}
	err = cfg.SetUser(user.ID, user.Name)
	if err != nil {
		log.Fatal("Could not set user in config", err)
	}
	fmt.Printf("Logged in as %s\n", user.Name)
}

func register(name string) {
	// make request to api
	body := strings.NewReader(fmt.Sprintf(`{"name": "%s"}`, name))
	req, err := http.NewRequest("POST", baseURL+"/users", body)
	if err != nil {
		log.Fatal("Could not reach server:", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal("Could not send request:", err)
	}
	defer resp.Body.Close()
	if checkResponseError(resp, "Could not create user") {
		return
	}

	// save as current user in config
	var user User
	json.NewDecoder(resp.Body).Decode(&user)

	cfg := Config{}
	if err = cfg.SetUser(user.ID, user.Name); err != nil {
		log.Fatal("Could not set user in config")
	}
	fmt.Println("User created and logged in")
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
