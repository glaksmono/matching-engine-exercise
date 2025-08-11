package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"
)

type OrderJSON struct {
	Cmd   string  `json:"cmd"`
	ID    string  `json:"id"`
	Side  string  `json:"side"`
	Type  string  `json:"type"`
	Price int     `json:"price"`
	Qty   float64 `json:"qty"`
}

func main() {
	numOrders := 1000000
	orders := make([]OrderJSON, numOrders)
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < numOrders; i++ {
		side := "BUY"
		if rand.Intn(2) == 1 {
			side = "SELL"
		}

		orders[i] = OrderJSON{
			Cmd:   "NEW",
			ID:    fmt.Sprintf("O-%d", i+1),
			Side:  side,
			Type:  "LIMIT",
			Price: 95 + rand.Intn(10), // Price between 95 and 104
			Qty:   float64(rand.Intn(10) + 1), // Qty between 1 and 10
		}
	}

	file, err := os.Create("huge_orders.json")
	if err != nil {
		fmt.Printf("Error creating file: %v\n", err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(orders); err != nil {
		fmt.Printf("Error encoding JSON: %v\n", err)
	}

	fmt.Println("Successfully generated huge_orders.json with 10,000 orders.")
}
