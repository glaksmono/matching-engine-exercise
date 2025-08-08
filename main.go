package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/shopspring/decimal"
)

const (
	OrderStatusWaiting = 0
	OrderStatusSuccess = 1

	OrderSideBuy       = "BUY"
	OrderSideSell      = "SELL"
	OrderCommandNew    = "NEW"
	OrderCommandCancel = "CANCEL"

	OrderTypeLimit  = "LIMIT"
	OrderTypeMarket = "MARKET"
)

type OrderJSON struct {
	Cmd    string  `json:"cmd"`
	ID     string  `json:"id"`
	Side   *string `json:"side"`
	Type   *string `json:"type"`
	Price  int     `json:"price"`
	Qty    float64 `json:"qty"`
	Status int     `json:"-"`
}

type Order struct {
	Cmd        string          `json:"cmd"`
	ID         string          `json:"id"`
	Side       *string         `json:"side"`
	Type       *string         `json:"type"`
	Price      int             `json:"price"`
	Qty        decimal.Decimal `json:"qty"`
	InitialQty decimal.Decimal `json:"-"`
	Canceled   bool            `json:"-"`
}

type Trade struct {
	BuyID  string          `json:"buyId"`
	SellID string          `json:"sellId"`
	Price  int             `json:"price"`
	Qty    decimal.Decimal `json:"qty"`
	Exec   int64           `json:"execId"`
}

type OrderBookData struct {
	ID    string          `json:"id"`
	Price int             `json:"price"`
	Qty   decimal.Decimal `json:"qty"`
}

type OrderBook struct {
	Bids []OrderBookData `json:"bids"`
	Asks []OrderBookData `json:"asks"`
}

var waitingBuyOrders []Order
var waitingSellOrders []Order

var trades []Trade

func main() {
	if len(os.Args) < 2 {
		fmt.Println("please add paths json as arguments")
		os.Exit(1)
	}

	file := os.Args[1]
	orders, err := loadJSON(file)
	if err != nil {
		fmt.Printf("error loading JSON from file %s: %v\n", file, err)
	}

	for _, order := range orders {
		switch order.Cmd {
		case OrderCommandNew:
			commandNew(&order)
		case OrderCommandCancel:
			commandCancel(&order)
		default:
			fmt.Printf("Unknown command: %s\n", order.Cmd)
		}
	}

	orderBook := OrderBook{
		Bids: []OrderBookData{},
		Asks: []OrderBookData{},
	}
	for _, o := range waitingBuyOrders {
		if o.Qty.Sign() > 0 && !o.Canceled {
			orderBook.Bids = append(orderBook.Bids, OrderBookData{
				ID:    o.ID,
				Price: o.Price,
				Qty:   o.Qty,
			})
		}
	}
	for _, o := range waitingSellOrders {
		if o.Qty.Sign() > 0 && !o.Canceled {
			orderBook.Asks = append(orderBook.Asks, OrderBookData{
				ID:    o.ID,
				Price: o.Price,
				Qty:   o.Qty,
			})
		}
	}

	finalOutput := struct {
		Trades    []Trade   `json:"trades"`
		OrderBook OrderBook `json:"orderBook"`
	}{
		Trades:    trades,
		OrderBook: orderBook,
	}

	var buf []byte
	buf, err = json.MarshalIndent(finalOutput, "", " ")
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("%s\n", string(buf))
}

func commandNew(order *Order) {
	// Initial quantity before order execution
	order.InitialQty = order.Qty.Copy()

	switch *order.Side {
	case OrderSideBuy:
		orderBuy(order)
	case OrderSideSell:
		orderSell(order)
	default:
		fmt.Printf("Unknown side: %s\n", *order.Side)
		return
	}
}

func commandCancel(order *Order) {
		for idx := range waitingSellOrders {
			if waitingSellOrders[idx].ID == order.ID && waitingSellOrders[idx].Qty != decimal.Zero && !waitingSellOrders[idx].Canceled {
				waitingSellOrders[idx].Canceled = true
				return
			}
		}

		for idx := range waitingBuyOrders{
			if waitingBuyOrders[idx].ID == order.ID && waitingBuyOrders[idx].Qty != decimal.Zero && !waitingBuyOrders[idx].Canceled {

				waitingBuyOrders[idx].Canceled = true
				return
			}
		}
}

func orderBuy(order *Order) {
	reqQty := order.Qty

	sortWaitingOrderSell()

	deletedOrders := 0

	for index := range waitingSellOrders {
		idx := index - deletedOrders

		if waitingSellOrders[idx].Qty.Sign() == 0 || waitingSellOrders[idx].Canceled {
			continue
		}

		canMatch := false

		if *order.Type == OrderTypeMarket {
			canMatch = true
		} else if waitingSellOrders[idx].Price <= order.Price {
			canMatch = true
		}

		if !canMatch {
			continue
		}

		var qty decimal.Decimal
		if waitingSellOrders[idx].Qty.Cmp(reqQty) > 0 {
			qty = reqQty
			waitingSellOrders[idx].Qty = waitingSellOrders[idx].Qty.Sub(reqQty)
		} else {
			qty = waitingSellOrders[idx].Qty
			waitingSellOrders[idx].Qty = decimal.Zero
		}

		reqQty = reqQty.Sub(qty)

		if qty.Sign() > 0 {
			trades = append(trades, Trade{
				BuyID:  order.ID,
				SellID: waitingSellOrders[idx].ID,
				Price:  waitingSellOrders[idx].Price,
				Qty:    qty,
				Exec:   int64(len(trades) + 1),
			})

			if waitingSellOrders[idx].Qty == decimal.Zero {
				waitingSellOrders = deleteByIndex(waitingSellOrders, idx)
				deletedOrders++
			}
		}

		if reqQty.Sign() == 0 {
			break
		}
	}

	if reqQty.Sign() > 0 && *order.Type == OrderTypeLimit {
		order.Qty = reqQty
		waitingBuyOrders = append(waitingBuyOrders, *order)
	}
}

func orderSell(order *Order) {
	reqQty := order.Qty

	sortWaitingOrderBuy()
	deletedOrders := 0 

	for index := range waitingBuyOrders {
		idx := index - deletedOrders

		if waitingBuyOrders[idx].Qty.Sign() == 0 || waitingBuyOrders[idx].Canceled {
			continue
		}

		canMatch := false

		if *order.Type == OrderTypeMarket {
			canMatch = true
		} else if waitingBuyOrders[idx].Price >= order.Price {
			canMatch = true
		}

		if !canMatch {
			continue
		}

		var qty decimal.Decimal
		if waitingBuyOrders[idx].Qty.Cmp(reqQty) > 0 {
			qty = reqQty
			waitingBuyOrders[idx].Qty = waitingBuyOrders[idx].Qty.Sub(reqQty)
		} else {
			qty = waitingBuyOrders[idx].Qty
			waitingBuyOrders[idx].Qty = decimal.Zero
		}

		reqQty = reqQty.Sub(qty)

		if qty.Sign() > 0 {
			trades = append(trades, Trade{
				BuyID:  waitingBuyOrders[idx].ID,
				SellID: order.ID,
				Price:  waitingBuyOrders[idx].Price,
				Qty:    qty,
				Exec:   int64(len(trades) + 1),
			})

			if waitingBuyOrders[idx].Qty == decimal.Zero {
				waitingBuyOrders = deleteByIndex(waitingBuyOrders, idx)
				deletedOrders++
			}
		}

		if reqQty.Sign() == 0 {
			break
		}
	}

	if reqQty.Sign() > 0  && *order.Type == OrderTypeLimit {
		order.Qty = reqQty
		waitingSellOrders = append(waitingSellOrders, *order)
	}
}

func sortWaitingOrderBuy() {
	sort.Slice(waitingBuyOrders, func(i, j int) bool {
		return waitingBuyOrders[i].Price > waitingBuyOrders[j].Price && !waitingBuyOrders[i].Canceled
	})
}

func sortWaitingOrderSell() {
	sort.Slice(waitingSellOrders, func(i, j int) bool {
		return waitingSellOrders[i].Price < waitingSellOrders[j].Price && !waitingSellOrders[i].Canceled
	})
}

func loadJSON(filePath string) ([]Order, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file %s does not exist", filePath)
	} else if err != nil {
		return nil, fmt.Errorf("error reading file %s: %v", filePath, err)
	}

	buf, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file %s: %v", filePath, err)
	}

	dataTradeJSON := []Order{}

	err = json.Unmarshal(buf, &dataTradeJSON)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON from file %s: %v", filePath, err)
	}

	return dataTradeJSON, nil
}

func deleteByIndex(orders []Order,idx int) []Order {
	if idx == len(orders) {
		orders = orders[:idx]
	}else {
		orders = append(orders[:idx], orders[idx+1:]...)
	}

	return orders
}
