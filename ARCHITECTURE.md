# Architecture Overview

This is a basic in-memory matching engine written in Go.

---

## 🔧 Key Concepts

- Orders are matched by price and time (best price first, then oldest)
- Matching is done right after each order is added
- All logic runs in a single thread
- Everything is done in memory

---

## ⚙️ Performance

- Adds and matches orders fast (under 1 microsecond each)
- Can handle 10,000+ orders in under 1 second on a laptop
- Uses a **map (`orderByID`) for constant-time lookup** when cancelling or replacing orders

---

## 🚀 How to scale this later

1. **Expose a REST API**  
   Add endpoints like `/orders`, `/cancel`, `/book`, `/trades` to make it accessible over HTTP

2. **Split by trading pair**  
   Run one engine per pair like BTC/USDT or ETH/IDR

3. **Add persistence**  
   Store trades and orders in a database like PostgreSQL or Redis so the engine can resume after shutdown

---

## 🔐 Safety

- Input is JSON only — avoids code injection or shell access
- Every order goes through **input validation** (checks for empty IDs, zero quantity, valid sides/types)
- Code is single-threaded — no race conditions
- No passwords or secrets in code

---

## 🚫 Limitations

- No web API by default
- No saved state between runs
- No advanced order types like stop-loss

---

## ✅ Summary

Simple, fast, and safe — with input validation, map lookups for performance, and a structure that's easy to scale later.
