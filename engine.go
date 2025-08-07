package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"time"
)

type Order struct {
	Cmd       string    `json:"cmd"`
	ID        string    `json:"id"`
	Side      string    `json:"side,omitempty"`  // BUY or SELL
	Type      string    `json:"type,omitempty"`  // LIMIT or MARKET
	Price     *float64  `json:"price,omitempty"` // nil for MARKET
	Qty       float64   `json:"qty,omitempty"`
	CreatedAt time.Time `json:"createdat"`
}


type Trade struct {
	BuyID  string  `json:"buyId"`
	SellID string  `json:"sellId"`
	Price  float64 `json:"price"`
	Qty    float64 `json:"qty"`
	ExecID int     `json:"execId"`
}

type Output struct {
	Trades    []Trade         `json:"trades"`
	OrderBook map[string][]*Order `json:"orderBook"` // bids and asks
}

var (
	orderByID = map[string]*Order{}
	trades    []Trade
	execID    = 1
)

var latencyBuckets = make(map[string]int)

func recordLatency(d time.Duration) {
	ns := d.Nanoseconds()

	switch {
	case ns < 10000:        // < 0.01ms
		latencyBuckets["<0.01ms"]++
	case ns < 50000:        // 0.01–0.05ms
		latencyBuckets["0.01–0.05ms"]++
	case ns < 100000:       // 0.05–0.1ms
		latencyBuckets["0.05–0.1ms"]++
	case ns < 200000:       // 0.1–0.2ms
		latencyBuckets["0.1–0.2ms"]++
	case ns < 500000:       // 0.2–0.5ms
		latencyBuckets["0.2–0.5ms"]++
	case ns < 1000000:     // 0.5–1ms
		latencyBuckets["0.5–1ms"]++
	case ns < 2000000:     // 1–2ms
		latencyBuckets["1–2ms"]++
	default:                 // > 2ms
		latencyBuckets[">2ms"]++
	}
}




func addOrder(o *Order) {
		if o == nil {
        return
    }
    if o.ID == "" {
        return
    }
    if o.Qty <= 0 {
        return
    }
    if o.Side != "BUY" && o.Side != "SELL" {
        return
    }
    if o.Type != "LIMIT" && o.Type != "MARKET" {
        return
    }

	orderByID[o.ID] = o

	matchOrder(o)

	if o.Qty == 0 {
		delete(orderByID, o.ID)
	}
}

func matchOrder(o *Order) {
	var candidates []*Order
	for _, other := range orderByID {
		if o.Side == "BUY" && other.Side == "SELL"{
			// MARKET NO NEED FILTER, LIMIT NEED PRICE FILTER 
			if o.Type == "MARKET" || ((other.Type=="LIMIT"&&o.Price != nil && other.Price != nil && *o.Price >= *other.Price)||
			other.Type=="MARKET") {
				candidates = append(candidates, other)
			}
		}
		if o.Side == "SELL" && other.Side == "BUY"{
			if o.Type == "MARKET" || ((other.Type=="LIMIT"&&o.Price != nil && other.Price != nil && *o.Price <= *other.Price)||
			other.Type=="MARKET") {
				candidates = append(candidates, other)
			}
		}
	}

	// Sort best price first
	priceAsc := o.Side == "BUY" // BUY Lowest price, SELL Highest price
	sortOrders(&candidates, priceAsc)


	for _, match := range candidates {
		if o.Qty <= 0 {
			break
		}

		var price float64
		if match.Price == nil {
			// Candidate is MARKET → use o.Price
			// Skip MARKET-to-MARKET match since both have undefined price
			if o.Price == nil {
				continue // prevent MARKET-MARKET match
			}
			price = *o.Price
		} else {
			price = *match.Price
		}

		q := math.Min(o.Qty, match.Qty)
		recordTrade(o, match, q, price)
		o.Qty -= q
		match.Qty -= q

		if match.Qty == 0 {
			delete(orderByID, match.ID)
		}
	}

}


func sortOrders(orders *[]*Order, priceAsc bool) {
	sort.SliceStable(*orders, func(i, j int) bool {
		o1 := (*orders)[i]
		o2 := (*orders)[j]

		// MARKET orders first
		if o1.Price == nil && o2.Price != nil {
			return true
		}
		if o1.Price != nil && o2.Price == nil {
			return false
		}
		// if both market, sort from earliest time
		if o1.Price == nil && o2.Price == nil {
			return o1.CreatedAt.Before(o2.CreatedAt)
		}

		// if price same, sort from earliest time
		if *o1.Price == *o2.Price {
			return o1.CreatedAt.Before(o2.CreatedAt)
		}

		if priceAsc {
			return *o1.Price < *o2.Price // lowest price first
		}
		return *o1.Price > *o2.Price // highest price first
	})
}



func recordTrade(o1, o2 *Order, qty float64, price float64) {
	var buyID, sellID string
	if o1.Side == "BUY" {
		buyID = o1.ID
		sellID = o2.ID
	} else {
		buyID = o2.ID
		sellID = o1.ID
	}
	trades = append(trades, Trade{
		BuyID:  buyID,
		SellID: sellID,
		Price:  price,
		Qty:    qty,
		ExecID: execID,
	})
	execID++
}

func cancelOrder(id string) {
	delete(orderByID, id)
}

//cancels and re-submits an order with the same ID.
//Only works if the order is still open (not fully matched). Otherwise, the command is skipped.
func replaceOrder(o *Order) {

    _, exists := orderByID[o.ID]

    if !exists {
        return
    }
    addOrder(o)
}

func buildOrderBook() map[string][]*Order {
	bids := []*Order{}
	asks := []*Order{}
	for _, o := range orderByID {
		if o.Side == "BUY" {
			bids = append(bids, o)
		} else {
			asks = append(asks, o)
		}
	}
	sortOrders(&bids, false) // bids side → highest price first
	sortOrders(&asks, true)  // asks side → lowest price first

	return map[string][]*Order{
		"bids": bids,
		"asks": asks,
	}
}

func (o Order) MarshalJSON() ([]byte, error) {
	type Alias Order
	return json.Marshal(&struct {
		ID    string   `json:"id"`
		Side  string   `json:"side,omitempty"`
		Type  string   `json:"type,omitempty"`
		Price *float64 `json:"price,omitempty"`
		Qty   float64  `json:"qty"`
	}{
		ID:    o.ID,
		Side:  o.Side,
		Type:  o.Type,
		Price: o.Price,
		Qty:   math.Round(o.Qty*1e8) / 1e8, // round to 8 decimal places
	})
}


func main() {
	start := time.Now()
	if len(os.Args) < 2 {
		fmt.Println("Usage: ./engine orders.json")
		return
	}

	file, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}

	var orders []Order
	if err := json.Unmarshal(file, &orders); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
		return
	}

	for i, o := range orders {
		
		// This guarantees unique CreatedAt timestamps, even for orders with the same price
		baseTime := time.Now()
		o.CreatedAt = baseTime.Add(time.Duration(i) * time.Nanosecond)

		switch o.Cmd {
		case "NEW":
			copy := o
			addOrder(&copy)

		case "CANCEL":
			cancelOrder(o.ID)

		case "REPLACE":
			copy := o
			replaceOrder(&copy)
		}
		
		duration := time.Since(o.CreatedAt)
		recordLatency(duration)
	}

	output := Output{
		Trades:    trades,
		OrderBook: buildOrderBook(),
	}

	jsonOut, _ := json.MarshalIndent(output, "", "  ")
	fmt.Println(string(jsonOut))

	fmt.Println("\n--- Latency Histogram ---")
	for bucket, count := range latencyBuckets {
		fmt.Printf("%s: %d\n", bucket, count)
	}

	duration := time.Since(start)
	fmt.Printf("Total execution time: %s\n", duration)
}
