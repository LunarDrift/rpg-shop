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
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Token string    `json:"token"`
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
		log.Fatal("Could not set user in config:", err)
	}
	fmt.Printf("Logged in as %s\n", user.Name)
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
		log.Fatal("Could not send request:", err)
	}
	defer resp.Body.Close()

	if checkResponseError(resp, "Could not create user") {
		return
	}
	login(name, password)
}
