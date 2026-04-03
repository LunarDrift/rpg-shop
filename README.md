# RPG-SHOP
A RESTful API and CLI client for an RPG shop inventory system, built with Go and PostgreSQL. The API handles all the business logic and database operations. The CLI client provides a terminal interface for browsing and purchasing items.

![Example CLI output](ex.png)

## Tech Stack
- **Go:** HTTP server and CLI client
- **PostgreSQL:** persistent storage
- **goose:** database migrations
- **sqlc:** type-safe SQL query generation

## Features
- Browse shop inventory with item names, prices, and current quantity
- Buy items by index number (stock decreases on purchase, fails if out of stock)
- Restock items by index with a specified quantity
- Full REST API with JSON responses

## Project Structure:
```
rpg-shop/
├── main.go               # Server entry point
├── server.go             # Server struct and helpers
├── handler_items.go      # HTTP handlers
├── items.go              # Item type definition
├── cmd/
│   └── shop/
│       └── main.go       # CLI client entry point
├── internal/
│   └── database/         # sqlc generated code
├── sql/
│   ├── queries/          # SQL queries
│   └── schema/           # goose migrations
└── sqlc.yaml
```

## Prerequisites
- Go 1.21+
- PostgreSQL
- goose: `go install github.com/pressly/goose/v3/cmd/goose@latest`
- sqlc: `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`

## Setup
1. Clone the repository
```bash
git clone https://github.com/LunarDrift/rpg-shop
cd rpg-shop
```

2. Create the database
```bash
createdb rpg_shop
```

3. Run migrations
```bash
goose -dir sql/schema postgres "postgres://localhost/rpg_shop?sslmode=disable" up
```

4. Set up the environment variables
  - Rename the `.env.example` file to `.env` and enter your database name

5. Install dependencies
```bash
go mod tidy
```

## Running the Server
```bash
go run .
```

The server starts on `http://localhost:8080`

## Using the CLI
In a separate terminal:
```bash
# Browse all items in the shop
go run ./cmd/shop browse

# Buy an item by its index number and optional quantity
go run ./cmd/shop buy 2 5

# Restock an item (shopkeeper)
go run ./cmd/shop/restock 2 10
```

### Example output


## API Endpoints


|Method|Endpoint|Description|
|---|---|---|
|GET|`/health`|Health check|
|GET|`/items`|List all items|
|GET|`/items/{id}`|Get item by ID|
|POST|`/items`|Create a new item|
|POST|`/items/buy/{id}`|Buy an item|
|PATCH|`/items/restock/{id}`|Restock an item|
|DELETE|`/items/{id}`|Delete an item|

### Seeding the Database
To add items directly via psql:
```bash
psql rpg_shop
```

```sql
INSERT INTO items (name, description, price, quantity)
VALUES
  ('Iron Sword', 'A sturdy blade forged from iron', 100, 5),
  ('Health Potion', 'Restores 50 HP', 25, 10),
  ('Magic Staff', 'Channels arcane energy', 200, 2);
```
