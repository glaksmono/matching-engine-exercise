# Bursa Crypto Indonesia – Coding Exercise  
**Build a Minimal Crypto-Exchange Matching Engine**

Welcome to the open-source interview exercise repository for **ICE (Exchange)**, **CACI (Clearing)**, and **ICC (Custodian)**.  
Your task is to design and implement the *smallest possible* **matching engine** that still behaves correctly and is production-quality.

---

## 1. Why we ask for this
| # | Skill we observe | How this exercise reveals it |
|---|------------------|------------------------------|
| 1 | **Leverage AI** | You may (and should) use any AI tool to bootstrap code, write tests, docs, etc. |
| 2 | **Review AI output** | Bad AI code left un-fixed is an automatic red flag. |
| 3 | **Design & architect** | Clean boundaries, extensible abstractions, thoughtful trade-offs. |
| 4 | **Code fundamentals** | Readability, idiomatic style, algorithmic correctness. |
| 5 | **“Gets things done”** | Working software in a PR, not just half-finished sketches. |
| 6 | **Discipline** | Unit tests, basic security hygiene, and explanations of scalability / availability. |

Keep it **simple**—we’d rather see a tiny, bullet-proof core than a kitchen-sink project.

---

## 2. Functional spec (MVP)

1. **Order types**  
   * `LIMIT` – price & quantity  
   * `MARKET` – quantity only (price = best available)  
2. **Fields**

| Field | Type | Example |
|-------|------|---------|
| `id`  | string | `"O-1001"` |
| `side`| `"BUY"` &#124; `"SELL"` | |
| `type`| `"LIMIT"` &#124; `"MARKET"` | |
| `price` | decimal (omit for MARKET) | `65000.00` |
| `qty`   | decimal | `0.25` |

3. **Matching algorithm**  
   * **Price-time priority** (best price, then earliest time).  
   * Partial fills allowed.  

4. **Lifecycle commands**  
   * `NEW` – add order  
   * `CANCEL` – remove by `id` (ignore if already filled)  
   * **Optional**: `REPLACE` (cancel-and-new in one step) – implement if you have time.  

---

## 3. I/O contract (CLI)

Run your engine from the project root:

```bash
./engine path/to/orders.json
```

	•	Input – a JSON array of commands.
	•	Output – two JSON arrays printed to stdout:
	1.	trades – every execution in the order it happened.
	2.	orderBook – final resting state (best BID ↔ best ASK).

Sample input (orders.json)

```json
[
  {"cmd":"NEW","id":"O-1","side":"BUY","type":"LIMIT","price":65000,"qty":0.5},
  {"cmd":"NEW","id":"O-2","side":"SELL","type":"LIMIT","price":65500,"qty":0.3},
  {"cmd":"NEW","id":"O-3","side":"SELL","type":"MARKET","qty":0.2},
  {"cmd":"CANCEL","id":"O-2"}
]
```

Expected output

```json
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

<!-- 📖 Sample execution walk-through -->

> ### What happens in the example above?
>
> 1. **`O-3` – MARKET SELL 0.2 BTC**  
>    * Because it is a *market* order, it crosses the book and takes the best bid.  
>    * Best bid at that moment = **`O-1` BUY LIMIT 65 000 USDT**.  
>    * A trade prints at **65 000 USDT** for **0.2 BTC**.
>
> 2. **`O-1` after the fill**  
>    * Started with **0.5 BTC**.  
>    * **0.2 BTC** was executed.  
>    * **0.3 BTC** remains resting on the bid side of the book.
>
> 3. **`O-2` – SELL LIMIT 65 500 USDT**  
>    * Still resting when `CANCEL O-2` arrives.  
>    * It is removed, so it never appears in the final order-book snapshot.
>
> **Result**  
> * `trades` array shows a single execution (`O-1` ↔ `O-3`).  
> * Final `orderBook` holds one bid (`O-1`, 0.3 BTC @ 65 000) and **no asks**.

⸻

## 4. Non-functional expectations

| Area            | Minimal bar                                                                                                                         |
|-----------------|-------------------------------------------------------------------------------------------------------------------------------------|
| **Language**    | Any mainstream language with an open-source compiler/interpreter. Idiomatic style and consistent formatting expected.               |
| **Build**       | One-liner bootstrap (`make`, `npm ci`, `go run`, etc.). The repo must clone, build, and run on a clean machine without manual tweaks.|
| **Tests**       | Automated unit tests for happy paths *and* edge cases. Aim for ≥ 80 % line coverage or equivalent meaningful metric.                |
| **Security**    | No obvious injections, overflows, race conditions, or hard-coded secrets. Follow the language’s standard secure-coding guidelines.  |
| **Performance** | Big-O analysis documented in code comments. The engine must process 10 k orders < 1 s on a laptop-class machine (non-debug build).   |
| **Scalability** | Include an `ARCHITECTURE.md` explaining how you would evolve this into a highly available, low-latency service (sharding, HA, etc.).|
| **Observability**| Basic logging + one metrics hook (e.g., order-match latency histogram).                                                            |
| **Docs**        | Clear `README` usage instructions and inline code comments.                                                                         |

> **Tip:** Keep the implementation tiny. Demonstrate discipline with tests, security hygiene, and a short scalability note rather than over-engineering


⸻

## 5. Submission checklist

1. **Fork** this repo.  
2. Create a **feature branch** `your-name/matching-engine`.  
3. Commit your solution (code, tests, docs).  
4. Open a **Pull Request** to `main` that includes:  
   * A one-paragraph design rationale (why you built it this way).  
   * Links or citations to any AI prompts/tools you used (transparency ≠ penalty).  
   * Clear instructions to run tests & a ≤ 30 s smoke benchmark.  
5. Keep the PR self-contained—CI should pass from a clean clone.

⸻

## 6. What *not* to worry about

* **GUI, REST API, persistence layer, auth** – focus only on the core matching logic.  
* **Perfect performance tuning** – clarity beats micro-optimisations for this demo.  
* **Edge-case regulations** (lot sizes, funding rates, etc.) – assume a simplified environment.

⸻

## 7. Need help?

If anything is unclear, open an **Issue** in this repository—concise questions only, please.

Good luck, and have fun!

