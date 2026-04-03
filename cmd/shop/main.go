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
		fmt.Println("Commands: browse, buy <idx> [quantity], restock <idx> <quantity>")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "browse":
		browseItems()
	case "buy":
		if len(os.Args) < 3 {
			fmt.Println("Usage: shop buy <item-id> [quantity]")
			os.Exit(1)
		}
		quantity := "1"
		if len(os.Args) >= 4 {
			quantity = os.Args[3]
		}
		buyItem(os.Args[2], quantity)
	case "restock":
		restockItem(os.Args[2], os.Args[3])
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
	}
}

func checkResponseError(r *http.Response, errMsg string) bool {
	if r.StatusCode != http.StatusOK {
		var errResp struct {
			Error string `json:"error"`
		}
		json.NewDecoder(r.Body).Decode(&errResp)
		fmt.Println(Red+errMsg+":"+Reset, errResp.Error)
		return true
	}
	return false
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
