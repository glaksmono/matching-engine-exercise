# ICE Matching Engine

A simple crypto exchange matching engine built in Go.

## ✅ What it does

- Handles `LIMIT` and `MARKET` orders
- Matches orders using best price and earliest time
- Allows partial fills
- Supports `NEW`, `CANCEL`, and `REPLACE` commands
- Shows all trades and the final order book
- Tracks how fast each order was processed
- Has full test coverage (94%+)

## 🚀 How to run

```
go run engine.go order.json              # Standard input (basic sample)
go run engine.go order_10k_example.json  # Large dataset for performance testing

```

## 🧪 How to test

```
go test -v -cover
```

To see detailed coverage:

```
go test -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## 📥 Example input (orders.json)

```
[
  {"cmd":"NEW","id":"O-1","side":"BUY","type":"LIMIT","price":65000,"qty":0.5},
  {"cmd":"NEW","id":"O-2","side":"SELL","type":"LIMIT","price":65500,"qty":0.3},
  {"cmd":"NEW","id":"O-3","side":"SELL","type":"MARKET","qty":0.2},
  {"cmd":"CANCEL","id":"O-2"}
]
```

## 📤 Example output

```
{
  "trades": [
    {"buyId":"O-1","sellId":"O-3","price":65000,"qty":0.2,"execId":1}
  ],
  "orderBook": {
    "bids": [
      {"id":"O-1","price":65000,"qty":0.3}
    ],
    "asks": []
  }
}
```

## 💡 How it works

- Each order is processed one at a time
- Matching is done right after an order is added
- Different parts of the system (add, match, trade, log) are split into small functions
- Easy to read and test

## 📊 Example latency output

```
--- Latency Histogram ---
<0.01ms: 9235
0.01–0.05ms: 756
...
Total execution time: 635.1ms
```
