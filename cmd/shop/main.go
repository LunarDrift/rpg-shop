package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/google/uuid"
)

const (
	baseURL = "http://localhost:8080"
	// ANSI Color codes
	Red    = "\033[31m"
	Green  = "\033[32m"
	Blue   = "\033[34m"
	Cyan   = "\033[36m"
	Yellow = "\033[33m"
	Reset  = "\033[0m"
	Bold   = "\033[1m"
	Dim    = "\033[2m"
)

type Item struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       int32     `json:"price"`
	Quantity    int32     `json:"quantity"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: shop <command>")
		fmt.Println("Commands: browse, buy <id>")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "browse":
		browseItems()
	case "buy":
		if len(os.Args) < 3 {
			fmt.Println("Usage: shop buy <item-id>")
			os.Exit(1)
		}
		buyItem(os.Args[2])
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
	}
}

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

func buyItem(idx string) {
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

	resp, err := http.Post(baseURL+"/items/buy/"+items[idxInt-1].ID.String(), "application/json", nil)
	if err != nil {
		log.Fatal("Could not reach shop:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]string
		json.NewDecoder(resp.Body).Decode(&errResp)
		fmt.Println(Red+"Purchase failed:"+Reset, errResp["error"])
		return
	}

	var item Item
	json.NewDecoder(resp.Body).Decode(&item)
	fmt.Printf("You purchased %s for %d gold!\n", item.Name, item.Price)
}
