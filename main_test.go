package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/shopspring/decimal"
)

type TestCase struct {
	name  string
	input []Order
	want  Result
}

func TestSimpleMatch(t *testing.T) {
	engine := new(MatchingEngine)

	engine.Init()

	orders := []Order{
		Order{Cmd: "NEW", ID: "B1", Side: OrderSideBuy, Type: OrderTypeLimit, Price: 100, Qty: decimal.NewFromInt(10)},
		Order{Cmd: "NEW", ID: "S1", Side: OrderSideSell, Type: OrderTypeLimit, Price: 100, Qty: decimal.NewFromInt(10)},
	}

	for _, order := range orders {
		switch order.Cmd {
		case OrderCommandNew:
			engine.commandNew(&order)
		case OrderCommandCancel:
			engine.commandCancel(&order)
		default:
			fmt.Printf("Unknown command: %s\n", order.Cmd)
		}
	}

	result := engine.processOutput()

	eq := cmp.Equal(result, Result{
		Trades: []Trade{
			Trade{BuyID: "B1", SellID: "S1", Price: 100, Qty: decimal.NewFromInt(10), Exec: 1},
		},
		OrderBook: OrderBook{
			Bids: []OrderBookData{},
			Asks: []OrderBookData{},
		},
	})

	if !eq {
		t.Error("Result not match")
	}
}

func TestMatch(t *testing.T) {
	cases := []TestCase{}

	err := filepath.Walk("./test/", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			orders := []Order{}
			err := loadJSON(path, &orders)
			if err != nil {
				fmt.Print("Error loading json from file", path, err)
			}

			output := Result{}
			err = loadJSON("./output/"+filepath.Base(path), &output)
			if err != nil {
				fmt.Print("Error loading json from file", path, err)
			}

			cases = append(cases, TestCase{
				name:  filepath.Base(path),
				input: orders,
				want:  output,
			})
		}

		return nil
	})
	if err != nil {
		t.Error(err)
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Running test case %s", t.Name())
			engine := new(MatchingEngine)

			engine.Init()

			for _, order := range tc.input {
				switch order.Cmd {
				case OrderCommandNew:
					engine.commandNew(&order)
				case OrderCommandCancel:
					engine.commandCancel(&order)
				default:
					fmt.Printf("Unknown command: %s\n", order.Cmd)
				}
			}

			result := engine.processOutput()
			eq := cmp.Diff(result, tc.want)

			if eq != "" {
				t.Error("Result not match ", eq)
			}

		})
	}
}

func TestSpeed(t *testing.T) {
	engine := new(MatchingEngine)

	t.Logf("Load huge json")
	orders := []Order{}
	err := loadJSON("./huge_orders.json", &orders)
	if err != nil {
		fmt.Print("Error loading json from file ./generator/huge_orders.json", err)
	}
	t.Logf("json loaded with %d orders", len(orders))

	engine.Init()
	initTime := time.Now()

	t.Logf("Running speed test")
	for _, order := range orders {
		switch order.Cmd {
		case OrderCommandNew:
			engine.commandNew(&order)
		case OrderCommandCancel:
			engine.commandCancel(&order)
		default:
			fmt.Printf("Unknown command: %s\n", order.Cmd)
		}
	}

	engine.processOutput()

	now := time.Now()

	duration := now.Sub(initTime)
	if duration > time.Second*30 {
		t.Error("Speed test took too long")
	}
}
