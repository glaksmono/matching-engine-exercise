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

	OrderSideBuy        = "BUY"
	OrderSideSell       = "SELL"
	OrderCommandNew     = "NEW"
	OrderCommandCancel  = "CANCEL"
	OrderCommandReplace = "REPLACE"

	OrderTypeLimit  = "LIMIT"
	OrderTypeMarket = "MARKET"
)

type Order struct {
	Cmd        string          `json:"cmd"`
	ID         string          `json:"id"`
	Side       string          `json:"side"`
	Type       string          `json:"type"`
	Price      decimal.Decimal `json:"price"`
	Qty        decimal.Decimal `json:"qty"`
	SequenceID int             `json:"-"`
	Canceled   bool            `json:"-"`
}

func (o Order) IsMarketType() bool {
	return o.Type == OrderTypeMarket
}

func (o Order) IsLimitType() bool {
	return o.Type == OrderTypeLimit
}

type Trade struct {
	BuyID  string          `json:"buyId"`
	SellID string          `json:"sellId"`
	Price  decimal.Decimal `json:"price"`
	Qty    decimal.Decimal `json:"qty"`
	Exec   uint64          `json:"execId"`
}

type OrderBookData struct {
	ID    string          `json:"id"`
	Price decimal.Decimal `json:"price"`
	Qty   decimal.Decimal `json:"qty"`
}

type OrderBook struct {
	Bids []OrderBookData `json:"bids"`
	Asks []OrderBookData `json:"asks"`
}

type BuyOrders []*Order

func (h BuyOrders) Len() int { return len(h) }
func (h BuyOrders) Less(i, j int) bool {
	if h[i].Price.GreaterThan(h[j].Price) {
		return true
	}

	if h[i].Price.Equal(h[j].Price) {
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
	if h[i].Price.LessThan(h[j].Price) {
		return true
	}

	if h[i].Price.Equal(h[j].Price) {
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

type Result struct {
	Trades    []Trade   `json:"trades"`
	OrderBook OrderBook `json:"orderBook"`
}

type MatchingEngine struct {
	trades     []Trade
	buyOrders  *BuyOrders
	sellOrders *SellOrders
	sequence   int
	orderMap   map[string]*Order
}

func (me *MatchingEngine) Init() {
	me.trades = make([]Trade, 0)
	me.buyOrders = &BuyOrders{}
	me.sellOrders = &SellOrders{}
	me.orderMap = make(map[string]*Order)
	me.sequence = 0

	heap.Init(me.buyOrders)
	heap.Init(me.sellOrders)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("please add paths json as arguments")
		os.Exit(1)
	}

	file := os.Args[1]
	orders := []Order{}

	err := loadJSON(file, &orders)
	if err != nil {
		fmt.Printf("error loading JSON from file %s: %v\n", file, err)
	}

	engine := &MatchingEngine{}
	engine.Init()

	for _, order := range orders {
		switch order.Cmd {
		case OrderCommandNew:
			engine.commandNew(&order)
		case OrderCommandCancel:
			engine.commandCancel(&order)
		case OrderCommandReplace:
			engine.commandReplace(&order)
		default:
			fmt.Printf("Unknown command: %s\n", order.Cmd)
		}
	}

	result := engine.processOutput()

	var buf []byte

	buf, err = json.MarshalIndent(result, "", " ")
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("%s\n", string(buf))
}

func (me *MatchingEngine) commandNew(order *Order) {
	me.sequence++
	order.SequenceID = me.sequence

	switch order.Side {
	case OrderSideBuy:
		me.orderBuy(order)
	case OrderSideSell:
		me.orderSell(order)
	default:
		fmt.Printf("Unknown side: %s\n", order.Side)
		return
	}
}

func (me *MatchingEngine) commandCancel(order *Order) {
	o, isExist := me.orderMap[order.ID]
	if !isExist {
		return
	}

	o.Canceled = true
}

func (me *MatchingEngine) commandReplace(order *Order) {
	o, isExist := me.orderMap[order.ID]
	if !isExist {
		return
	}

	me.sequence++

	o.Price = order.Price
	o.Qty = order.Qty

	me.orderMap[o.ID] = o
	if o.Side == OrderSideBuy {
		buyOrder := heap.Pop(me.buyOrders).(*Order)
		heap.Push(me.buyOrders, buyOrder)
	}

	if o.Side == OrderSideSell {
		sellOrder := heap.Pop(me.sellOrders).(*Order)

		heap.Push(me.sellOrders, sellOrder)
	}
}

func (me *MatchingEngine) orderBuy(order *Order) {
	reqQty := order.Qty

	for me.sellOrders.Len() > 0 && reqQty.Cmp(decimal.Zero) > 0 {
		bestOrder := me.sellOrders.GetByIndex(0)

		if bestOrder.Canceled {
			heap.Pop(me.sellOrders)

			continue
		}

		canMatch := order.IsMarketType() || bestOrder.Price.LessThanOrEqual(order.Price)

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

			heap.Pop(me.sellOrders)

			delete(me.orderMap, bestOrder.ID)
		}

		reqQty = reqQty.Sub(qty)

		if qty.Sign() > 0 {
			me.trades = append(me.trades, Trade{
				BuyID:  order.ID,
				SellID: bestOrder.ID,
				Price:  bestOrder.Price,
				Qty:    qty,
				Exec:   uint64(len(me.trades) + 1),
			})
		}
	}

	if reqQty.Sign() > 0 && order.IsLimitType() {
		order.Qty = reqQty

		me.orderMap[order.ID] = order

		heap.Push(me.buyOrders, order)
	}
}

func (me *MatchingEngine) orderSell(order *Order) {
	reqQty := order.Qty

	for me.buyOrders.Len() > 0 && reqQty.Cmp(decimal.Zero) > 0 {
		bestOrder := me.buyOrders.GetByIndex(0)

		if bestOrder.Canceled {
			heap.Pop(me.buyOrders)

			continue
		}

		canMatch := order.IsMarketType() || bestOrder.Price.GreaterThanOrEqual(order.Price)

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

			heap.Pop(me.buyOrders)

			delete(me.orderMap, bestOrder.ID)
		}

		reqQty = reqQty.Sub(qty)

		if qty.Sign() > 0 {
			me.trades = append(me.trades, Trade{
				BuyID:  bestOrder.ID,
				SellID: order.ID,
				Price:  bestOrder.Price,
				Qty:    qty,
				Exec:   uint64(len(me.trades) + 1),
			})
		}
	}

	if reqQty.Sign() > 0 && order.Type == OrderTypeLimit {
		order.Qty = reqQty

		me.orderMap[order.ID] = order

		heap.Push(me.sellOrders, order)
	}
}

func loadJSON[T any](filePath string, data *T) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file %s does not exist", filePath)
	} else if err != nil {
		return fmt.Errorf("error reading file %s: %v", filePath, err)
	}

	buf, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading file %s: %v", filePath, err)
	}

	err = json.Unmarshal(buf, data)
	if err != nil {
		return fmt.Errorf("error unmarshalling JSON from file %s: %v", filePath, err)
	}

	return nil
}

func (me *MatchingEngine) processOutput() Result {
	orderBook := OrderBook{
		Bids: []OrderBookData{},
		Asks: []OrderBookData{},
	}

	for me.buyOrders.Len() > 0 {
		o := heap.Pop(me.buyOrders).(*Order)

		if o.Qty.Sign() > 0 && !o.Canceled {
			orderBook.Bids = append(orderBook.Bids, OrderBookData{
				ID:    o.ID,
				Price: o.Price,
				Qty:   o.Qty,
			})
		}
	}
	for me.sellOrders.Len() > 0 {
		o := heap.Pop(me.sellOrders).(*Order)

		if o.Qty.Sign() > 0 && !o.Canceled {
			orderBook.Asks = append(orderBook.Asks, OrderBookData{
				ID:    o.ID,
				Price: o.Price,
				Qty:   o.Qty,
			})
		}
	}

	finalOutput := Result{
		Trades:    me.trades,
		OrderBook: orderBook,
	}

	return finalOutput
}
