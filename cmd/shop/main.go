package main

import (
	"fmt"
	"os"

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

	case "login":
		if len(os.Args) < 4 {
			fmt.Println("Usage: shop login <name> <password>")
			os.Exit(1)
		}
		login(os.Args[2], os.Args[3])

	case "register":
		if len(os.Args) < 4 {
			fmt.Println("Usage: shop register <name> <password>")
			os.Exit(1)
		}
		register(os.Args[2], os.Args[3])

	case "whoami":
		whoami()

	case "logout":
		logout()

	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
	}
}
