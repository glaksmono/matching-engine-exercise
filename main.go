package main

import (
	"container/heap"
	"encoding/json"
	"fmt"
	"os"

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

type Order struct {
	Cmd        string          `json:"cmd"`
	ID         string          `json:"id"`
	Side       *string         `json:"side"`
	Type       *string         `json:"type"`
	Price      int             `json:"price"`
	Qty        decimal.Decimal `json:"qty"`
	InitialQty decimal.Decimal `json:"-"`
	SequenceID int             `json:"-"`
	Canceled   bool            `json:"-"`
}

func (o Order) IsMarketType() bool {
	return *o.Type == OrderTypeMarket
}

func (o Order) IsLimitType() bool {
	return *o.Type == OrderTypeLimit
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

type BuyOrders []*Order

func (h BuyOrders) Len() int { return len(h) }
func (h BuyOrders) Less(i, j int) bool {
	if h[i].Price > h[j].Price {
		return true
	}

	if h[i].Price == h[j].Price {
		return h[i].SequenceID < h[j].SequenceID
	}

	return false
}

func (h BuyOrders) Swap(i, j int)       { h[i], h[j] = h[j], h[i] }
func (h *BuyOrders) Push(x interface{}) { *h = append(*h, x.(*Order)) }
func (h BuyOrders) GetByIndex(i int) *Order {
	return h[i]
}
func (h *BuyOrders) Pop() interface{} {
	oldOrder := *h
	lenOrder := len(oldOrder)
	order := oldOrder[lenOrder-1]
	*h = oldOrder[0 : lenOrder-1]

	return order
}

type SellOrders []*Order

func (h SellOrders) Len() int { return len(h) }
func (h SellOrders) Less(i, j int) bool {
	if h[i].Price < h[j].Price {
		return true
	}

	if h[i].Price == h[j].Price {
		return h[i].SequenceID < h[j].SequenceID
	}

	return false
}

func (h SellOrders) Swap(i, j int) { h[i], h[j] = h[j], h[i] }
func (h SellOrders) GetByIndex(i int) *Order {
	return h[i]
}

func (h *SellOrders) Push(x interface{}) {
	*h = append(*h, x.(*Order))
}

func (h *SellOrders) Pop() interface{} {
	oldOrder := *h
	lenOrder := len(oldOrder)
	order := oldOrder[lenOrder-1]
	*h = oldOrder[0 : lenOrder-1]

	return order
}

var buyOrders = &BuyOrders{}
var sellOrders = &SellOrders{}
var orderMap = make(map[string]*Order)
var sequenceCounter int = 0

var trades = make([]Trade, 0)

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


	trades = []Trade{}

	heap.Init(buyOrders)
	heap.Init(sellOrders)

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
	for buyOrders.Len() > 0 {
		o := heap.Pop(buyOrders).(*Order)
		if o.Qty.Sign() > 0 && !o.Canceled {

						orderBook.Bids = append(orderBook.Bids, OrderBookData{
				ID:    o.ID,
				Price: o.Price,
				Qty:   o.Qty,
			})
		}
	}
	for sellOrders.Len() > 0 {
		o := heap.Pop(sellOrders).(*Order)
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
	sequenceCounter++
	order.InitialQty = order.Qty.Copy()
	order.SequenceID = sequenceCounter

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
	o, _ := orderMap[order.ID]
	if o == nil {
		return
	}
	o.Canceled = true

	// for idx := range waitingSellOrders {
	// 	if waitingSellOrders[idx].ID == order.ID && waitingSellOrders[idx].Qty != decimal.Zero && !waitingSellOrders[idx].Canceled {
	// 		waitingSellOrders[idx].Canceled = true
	// 		return
	// 	}
	// }

	// for idx := range waitingBuyOrders {
	// 	if waitingBuyOrders[idx].ID == order.ID && waitingBuyOrders[idx].Qty != decimal.Zero && !waitingBuyOrders[idx].Canceled {
	//
	// 		waitingBuyOrders[idx].Canceled = true
	// 		return
	// 	}
	// }
}

func orderBuy(order *Order) {
	reqQty := order.Qty

	for sellOrders.Len() > 0 && reqQty.Cmp(decimal.Zero) > 0 {
		sellOrder := sellOrders.GetByIndex(0)

		if sellOrder.Canceled {
			heap.Pop(sellOrders)
			continue
		}

		canMatch := order.IsMarketType() || sellOrder.Price <= order.Price

		if !canMatch {
			break
		}

		var qty decimal.Decimal
		if sellOrder.Qty.Cmp(reqQty) > 0 {
			qty = reqQty
			sellOrder.Qty = sellOrder.Qty.Sub(reqQty)
		} else {
			qty = sellOrder.Qty
			sellOrder.Qty = decimal.Zero

			heap.Pop(sellOrders)
			delete(orderMap, sellOrder.ID)
		}

				reqQty = reqQty.Sub(qty)

		if qty.Sign() > 0 {
			trades = append(trades, Trade{
				BuyID:  order.ID,
				SellID: sellOrder.ID,
				Price:  sellOrder.Price,
				Qty:    qty,
				Exec:   int64(len(trades) + 1),
			})
		}
	}

	if reqQty.Sign() > 0 && order.IsLimitType() {
		order.Qty = reqQty
		orderMap[order.ID] = order
		heap.Push(buyOrders, order)
	}
}

func orderSell(order *Order) {
	reqQty := order.Qty

	for buyOrders.Len() > 0 && reqQty.Cmp(decimal.Zero) > 0 {
		bestOrder := buyOrders.GetByIndex(0)

		if bestOrder.Canceled {
			heap.Pop(buyOrders)
			continue
		}

		canMatch := order.IsMarketType() || bestOrder.Price >= order.Price

		if !canMatch {
			break
		}

		var qty decimal.Decimal
		if bestOrder.Qty.Cmp(reqQty) > 0 {
			qty = reqQty
			bestOrder.Qty = bestOrder.Qty.Sub(reqQty)
		} else {
			qty = bestOrder.Qty
			bestOrder.Qty = decimal.Zero

			heap.Pop(buyOrders)
			delete(orderMap, bestOrder.ID)
		}

				reqQty = reqQty.Sub(qty)

		if qty.Sign() > 0 {
			trades = append(trades, Trade{
				BuyID:  bestOrder.ID,
				SellID: order.ID,
				Price:  bestOrder.Price,
				Qty:    qty,
				Exec:   int64(len(trades) + 1),
			})
		}
	}

	if reqQty.Sign() > 0 && *order.Type == OrderTypeLimit {
		order.Qty = reqQty
		orderMap[order.ID] = order
		heap.Push(sellOrders, order)
	}
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
