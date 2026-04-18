package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// ----------------------------------------------------------------------------------------------------
// --------- USER COMMANDS ----------
// ----------------------------------------------------------------------------------------------------

type User struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Balance int32     `json:"balance"`
	Token   string    `json:"token"`
}

func login(name, password string) {
	// make request to api
	body := strings.NewReader(fmt.Sprintf(`{"name": "%s", "password": "%s"}`, name, password))
	req, err := http.NewRequest("POST", baseURL+"/users/login", body)
	if err != nil {
		log.Fatal("Could not log in:", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
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
	err = cfg.SetUser(user.ID, user.Name, user.Token)
	if err != nil {
		log.Fatal(Red+"Could not set user in config:"+Reset, err)
	}
	fmt.Printf("Logged in as "+Bold+Blue+"%s"+Reset+"\n", user.Name)
}

func register(name, password string) {
	// make request to api
	body := strings.NewReader(fmt.Sprintf(`{"name": "%s", "password": "%s"}`, name, password))
	req, err := http.NewRequest("POST", baseURL+"/users", body)
	if err != nil {
		log.Fatal("Could not reach server:", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(Red+"Could not send request:"+Reset, err)
	}
	defer resp.Body.Close()

	if checkResponseError(resp, "Could not create user") {
		return
	}
	login(name, password)
}

func whoami() {
	cfg, err := Read()
	if err != nil || cfg.Token == "" {
		fmt.Println("Not logged in. Use 'shop login <name> <password>' first")
		return
	}

	resp, err := http.Get(baseURL + "/users/" + cfg.CurrentUserID.String())
	if err != nil {
		log.Fatal(Red+"Could not reach server:"+Reset, err)
	}
	defer resp.Body.Close()

	var user User
	json.NewDecoder(resp.Body).Decode(&user)

	fmt.Printf("Logged in as "+Bold+Blue+"%s"+Reset+" | Balance: "+Yellow+"%dg"+Reset+"\n", user.Name, user.Balance)
}

func logout() {
	cfg := Config{}
	err := write(cfg)
	if err != nil {
		log.Fatal(Red+"Could not log out:"+Reset, err)
	}
	fmt.Println("Logged out")
}

func explore() {
	cfg, err := Read()
	if err != nil || cfg.Token == "" {
		fmt.Println("Not logged in. Use 'shop login <name> <password>' first")
		return
	}

	req, err := http.NewRequest("PATCH", baseURL+"/users/earn", nil)
	if err != nil {
		log.Fatal(Red+"Error making request:"+Reset, err)
	}
	req.Header.Set("Authorization", "Bearer "+cfg.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(Red+"Could not reach server:"+Reset, err)
	}
	defer resp.Body.Close()

	if checkResponseError(resp, "Could not earn gold") {
		return
	}

	var user User
	json.NewDecoder(resp.Body).Decode(&user)

	fmt.Println("You found some items to sell while out exploring...")
	fmt.Printf("New Balance: "+Yellow+"%dg"+Reset+"\n", user.Balance)
}

func inventory() {
	cfg, err := Read()
	if err != nil || cfg.Token == "" {
		fmt.Println("Not logged in. Use 'shop login <name> <password>' first")
		return
	}

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
		Name     string `json:"name"`
		Price    int32  `json:"price"`
		Quantity int32  `json:"quantity"`
	}
	var inventory []InventoryItem
	json.NewDecoder(resp.Body).Decode(&inventory)

	fmt.Println("Current Inventory:")
	for i, item := range inventory {
		fmt.Printf("%d. "+Bold+Blue+"%s"+Reset+"     (qty: %d)  "+Yellow+"%dg"+Reset+"\n", i+1, item.Name, item.Quantity, item.Price*item.Quantity)
	}
}
