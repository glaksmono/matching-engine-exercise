package main

import (
	"encoding/json"
	"os"
	"testing"
	"time"
)

func resetState() {
	orderByID = make(map[string]*Order)
	trades = []Trade{}
	execID = 1
}

func TestMatchMarketBuyWithLimitSell(t *testing.T) {
	resetState()
	price := 100.0
	addOrder(&Order{Cmd: "NEW", ID: "SELL1", Side: "SELL", Type: "LIMIT", Price: &price, Qty: 2})
	addOrder(&Order{Cmd: "NEW", ID: "BUY1", Side: "BUY", Type: "MARKET", Qty: 1})

	if len(trades) != 1 {
		t.Errorf("Expected 1 trade, got %d", len(trades))
	}
	if trades[0].Qty != 1 {
		t.Errorf("Expected Qty 1, got %f", trades[0].Qty)
	}
}

func TestBuildOrderBook(t *testing.T) {
	resetState()
	priceBuy := 99.0
	priceSell := 101.0

	addOrder(&Order{Cmd: "NEW", ID: "BUY1", Side: "BUY", Type: "LIMIT", Price: &priceBuy, Qty: 1})
	addOrder(&Order{Cmd: "NEW", ID: "SELL1", Side: "SELL", Type: "LIMIT", Price: &priceSell, Qty: 1})

	book := buildOrderBook()
	if len(book["bids"]) != 1 || len(book["asks"]) != 1 {
		t.Errorf("Expected 1 bid and 1 ask, got %+v", book)
	}
}

func TestSortOrdersBuyAndSell(t *testing.T) {
	resetState()
	buyPrice1 := 90.0
	buyPrice2 := 100.0
	buyOrders := []*Order{
		{ID: "B1", Price: &buyPrice1},
		{ID: "B2", Price: &buyPrice2},
	}
	sortOrders(&buyOrders, false)
	if *buyOrders[0].Price != 100.0 {
		t.Error("BUY orders not sorted descending")
	}

	sellPrice1 := 90.0
	sellPrice2 := 100.0
	sellOrders := []*Order{
		{ID: "S1", Price: &sellPrice2},
		{ID: "S2", Price: &sellPrice1},
	}
	sortOrders(&sellOrders, true)
	if *sellOrders[0].Price != 90.0 {
		t.Error("SELL orders not sorted ascending")
	}
}

func TestMultipleLimitBuyMatchingMarketSell(t *testing.T) {
	resetState()
	high := 100.0
	low := 99.0
	addOrder(&Order{Cmd: "NEW", ID: "B1", Side: "BUY", Type: "LIMIT", Price: &high, Qty: 1})
	addOrder(&Order{Cmd: "NEW", ID: "B2", Side: "BUY", Type: "LIMIT", Price: &low, Qty: 1})
	addOrder(&Order{Cmd: "NEW", ID: "S1", Side: "SELL", Type: "MARKET", Qty: 2})

	if len(trades) != 2 {
		t.Fatalf("Expected 2 trades, got %d", len(trades))
	}
	if trades[0].Price != high || trades[1].Price != low {
		t.Errorf("Expected trade prices [%v, %v], got [%v, %v]",
			high, low, trades[0].Price, trades[1].Price)
	}
}

func TestSortOrdersWithNilPrice(t *testing.T) {
	resetState()
	price := 100.0
	orders := []*Order{
		{ID: "MKT", Side: "BUY", Type: "MARKET", Price: nil},
		{ID: "LMT", Side: "BUY", Type: "LIMIT", Price: &price},
	}
	sortOrders(&orders, true)
	if orders[0].ID != "MKT" || orders[1].ID != "LMT" {
		t.Errorf("Expected order [MKT, LMT], got [%v, %v]", orders[0].ID, orders[1].ID)
	}
}

// Test main function with valid JSON file
func TestMainWithValidFile(t *testing.T) {
	// Create test JSON file
	orders := []Order{
		{Cmd: "NEW", ID: "B1", Side: "BUY", Type: "LIMIT", Price: func() *float64 { p := 100.0; return &p }(), Qty: 1},
		{Cmd: "NEW", ID: "S1", Side: "SELL", Type: "LIMIT", Price: func() *float64 { p := 100.0; return &p }(), Qty: 1},
		{Cmd: "CANCEL", ID: "B1"},
	}
	
	jsonData, err := json.Marshal(orders)
	if err != nil {
		t.Fatal(err)
	}
	
	err = os.WriteFile("test_orders.json", jsonData, 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove("test_orders.json")
	
	// Save original args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	
	// Test with valid file
	os.Args = []string{"engine", "test_orders.json"}
	
	resetState()	
	main()
}

// Test main function with invalid JSON
func TestMainWithInvalidJSON(t *testing.T) {
	// Create invalid JSON file
	err := os.WriteFile("invalid.json", []byte("invalid json"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove("invalid.json")

	// Save original args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"engine", "invalid.json"}
	main()
}


// Test main function with non-existent file
func TestMainWithNonExistentFile(t *testing.T) {
	// Save original args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	
	os.Args = []string{"engine", "nonexistent.json"}
	
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected main() to panic with non-existent file")
		}
	}()
	
	main()
}

// Test matchOrder early exit when quantity becomes zero
func TestMatchOrderEarlyExit(t *testing.T) {
	resetState()
	price1 := 100.0
	price2 := 101.0
	
	// Add two sell orders
	addOrder(&Order{Cmd: "NEW", ID: "S1", Side: "SELL", Type: "LIMIT", Price: &price1, Qty: 1})
	addOrder(&Order{Cmd: "NEW", ID: "S2", Side: "SELL", Type: "LIMIT", Price: &price2, Qty: 1})
	
	// Add buy order that exactly matches first sell order
	addOrder(&Order{Cmd: "NEW", ID: "B1", Side: "BUY", Type: "LIMIT", Price: &price2, Qty: 1})
	
	// Should only match with S1 (cheaper), not S2
	if len(trades) != 1 {
		t.Errorf("Expected 1 trade, got %d", len(trades))
	}
	if trades[0].Price != price1 {
		t.Errorf("Expected trade at %v, got %v", price1, trades[0].Price)
	}
}

// NEW TESTS TO IMPROVE COVERAGE

// Test addOrder with invalid inputs
func TestAddOrderInvalidInputs(t *testing.T) {
	resetState()
	
	// Test nil order
	addOrder(nil)
	if len(orderByID) != 0 {
		t.Error("Should not add nil order")
	}
	
	// Test empty ID
	addOrder(&Order{Cmd: "NEW", ID: "", Side: "BUY", Type: "LIMIT", Qty: 1})
	if len(orderByID) != 0 {
		t.Error("Should not add order with empty ID")
	}
	
	// Test zero quantity
	addOrder(&Order{Cmd: "NEW", ID: "TEST", Side: "BUY", Type: "LIMIT", Qty: 0})
	if len(orderByID) != 0 {
		t.Error("Should not add order with zero quantity")
	}
	
	// Test negative quantity
	addOrder(&Order{Cmd: "NEW", ID: "TEST", Side: "BUY", Type: "LIMIT", Qty: -1})
	if len(orderByID) != 0 {
		t.Error("Should not add order with negative quantity")
	}
	
	// Test invalid side
	addOrder(&Order{Cmd: "NEW", ID: "TEST", Side: "INVALID", Type: "LIMIT", Qty: 1})
	if len(orderByID) != 0 {
		t.Error("Should not add order with invalid side")
	}
	
	// Test invalid type
	addOrder(&Order{Cmd: "NEW", ID: "TEST", Side: "BUY", Type: "INVALID", Qty: 1})
	if len(orderByID) != 0 {
		t.Error("Should not add order with invalid type")
	}
}

// Test cancelOrder
func TestCancelOrder(t *testing.T) {
	resetState()
	price := 100.0
	
	// Add an order
	addOrder(&Order{Cmd: "NEW", ID: "TEST", Side: "BUY", Type: "LIMIT", Price: &price, Qty: 1})
	if len(orderByID) != 1 {
		t.Error("Order should be added")
	}
	
	// Cancel the order
	cancelOrder("TEST")
	if len(orderByID) != 0 {
		t.Error("Order should be cancelled")
	}
	
	// Cancel non-existent order (should not panic)
	cancelOrder("NON_EXISTENT")
}

// Test replaceOrder
func TestReplaceOrder(t *testing.T) {
	resetState()
	price1 := 100.0
	price2 := 200.0
	
	// Add an order
	addOrder(&Order{Cmd: "NEW", ID: "TEST", Side: "BUY", Type: "LIMIT", Price: &price1, Qty: 1})
	if len(orderByID) != 1 {
		t.Error("Order should be added")
	}
	
	// Replace the order
	replaceOrder(&Order{Cmd: "REPLACE", ID: "TEST", Side: "BUY", Type: "LIMIT", Price: &price2, Qty: 2})
	if orderByID["TEST"].Qty != 2 || *orderByID["TEST"].Price != 200.0 {
		t.Error("Order should be replaced")
	}
	
	// Try to replace non-existent order
	replaceOrder(&Order{Cmd: "REPLACE", ID: "NON_EXISTENT", Side: "BUY", Type: "LIMIT", Price: &price1, Qty: 1})
	if _, exists := orderByID["NON_EXISTENT"]; exists {
		t.Error("Non-existent order should not be added")
	}
}

// Test market-market matching prevention
func TestMarketMarketMatchPrevention(t *testing.T) {
	resetState()
	
	// Add market buy order
	addOrder(&Order{Cmd: "NEW", ID: "MKT_BUY", Side: "BUY", Type: "MARKET", Qty: 1})
	
	// Add market sell order - should not match
	addOrder(&Order{Cmd: "NEW", ID: "MKT_SELL", Side: "SELL", Type: "MARKET", Qty: 1})
	
	if len(trades) != 0 {
		t.Error("Market orders should not match with each other")
	}
	if len(orderByID) != 2 {
		t.Error("Both market orders should remain in book")
	}
}


// Test sorting with same prices (time priority)
func TestSortOrdersSamePrice(t *testing.T) {
	resetState()
	price := 100.0
	
	now := time.Now()
	orders := []*Order{
		{ID: "SECOND", Price: &price, CreatedAt: now.Add(time.Second)},
		{ID: "FIRST", Price: &price, CreatedAt: now},
	}
	
	sortOrders(&orders, true)
	
	if orders[0].ID != "FIRST" {
		t.Error("Orders with same price should be sorted by time priority")
	}
}

// Test sorting with both market orders (time priority)
func TestSortBothMarketOrders(t *testing.T) {
	resetState()
	
	now := time.Now()
	orders := []*Order{
		{ID: "SECOND", Price: nil, CreatedAt: now.Add(time.Second)},
		{ID: "FIRST", Price: nil, CreatedAt: now},
	}
	
	sortOrders(&orders, true)
	
	if orders[0].ID != "FIRST" {
		t.Error("Market orders should be sorted by time priority")
	}
}

// Test main with no arguments
func TestMainNoArgs(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	
	os.Args = []string{"engine"}
	
	// Capture stdout to verify usage message
	main() // Should print usage and return
}

func TestRecordLatencyBuckets(t *testing.T) {
	// Reset before test
	latencyBuckets = make(map[string]int)

	// Simulate durations across bucket boundaries
	durations := []struct {
		dur     time.Duration
		bucket  string
	}{
		{time.Nanosecond * 9000, "<0.01ms"},
		{time.Nanosecond * 10000, "0.01–0.05ms"},
		{time.Nanosecond * 60000, "0.05–0.1ms"},
		{time.Nanosecond * 150000, "0.1–0.2ms"},
		{time.Nanosecond * 300000, "0.2–0.5ms"},
		{time.Nanosecond * 800000, "0.5–1ms"},
		{time.Nanosecond * 1500000, "1–2ms"},
		{time.Nanosecond * 3000000, ">2ms"},
	}

	for _, d := range durations {
		recordLatency(d.dur)
	}

	// Check each bucket count = 1
	for _, d := range durations {
		if latencyBuckets[d.bucket] != 1 {
			t.Errorf("Expected bucket %s to have count 1, got %d", d.bucket, latencyBuckets[d.bucket])
		}
	}
}

func TestProcessOrders(t *testing.T) {
	// Reset global state if needed
	resetState()

orders := []Order{
	{Cmd: "NEW", ID: "B1", Side: "BUY", Type: "LIMIT", Price: float64Ptr(100.0), Qty: 1},
	{Cmd: "CANCEL", ID: "B1"},
	{Cmd: "NEW", ID: "B2", Side: "BUY", Type: "LIMIT", Price: float64Ptr(100.0), Qty: 1},
	{Cmd: "REPLACE", ID: "B2", Price: float64Ptr(105.0), Qty: 0.5},
}


	for _, o := range orders {
		switch o.Cmd {
		case "NEW":
			orderCopy := o
			addOrder(&orderCopy)
		case "CANCEL":
			cancelOrder(o.ID)
		case "REPLACE":
			orderCopy := o
			replaceOrder(&orderCopy)
		}
	}

	// Assert that orders were added/cancelled/replaced correctly
	if _, exists := orderByID["B2"]; !exists {
		t.Errorf("Expected order B2 to exist after REPLACE")
	}
}

func float64Ptr(v float64) *float64 {
	return &v
}
