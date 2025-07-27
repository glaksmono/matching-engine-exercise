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

	•	Input – a JSON array of commands.
	•	Output – two JSON arrays printed to stdout:
	1.	trades – every execution in the order it happened.
	2.	orderBook – final resting state (best BID ↔ best ASK).

Sample input (orders.json)

[
  {"cmd":"NEW","id":"O-1","side":"BUY","type":"LIMIT","price":65000,"qty":0.5},
  {"cmd":"NEW","id":"O-2","side":"SELL","type":"LIMIT","price":65500,"qty":0.3},
  {"cmd":"NEW","id":"O-3","side":"SELL","type":"MARKET","qty":0.2},
  {"cmd":"CANCEL","id":"O-2"}
]

Expected output

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

O-3 is a MARKET sell → hits best BID O-1 at 65 000.
O-2 is later cancelled, leaving only the partial O-1 in the book.

⸻

4. Non-functional expectations

Area	Minimal bar
Language	Anything with a mainstream compiler / interpreter & open licence.
Build	1-line setup (make, npm install, etc.).
Tests	Unit tests covering happy path + edge cases.
Security	No obvious injection / overflow / race conditions.
Scalability note	Add a short ARCHITECTURE.md describing how you’d evolve this into HA, low-latency production service (sharding, persistence, etc.). A diagram is nice but optional.


⸻

5. Submission checklist
	1.	Fork this repo.
	2.	Create a feature branch your-name/matching-engine.
	3.	Commit your solution (code, tests, docs).
	4.	Open a Pull Request to main with:
	•	A one-paragraph design rationale.
	•	Link to any AI prompts or tools you used (transparency ≠ penalty).
	•	How to run tests & benchmarks (30 s smoke test is enough).
	5.	Keep the PR self-contained—CI should pass with a clean clone.

⸻

6. What not to worry about
	•	GUI, REST API, persistence layer, auth.
	•	Perfect performance tuning—clarity beats micro-optimisations here.
	•	Edge-case regulations (lot sizes, funding rates, etc.)—this is a demo.

⸻

7. Need help?

If something is unclear, open an Issue in the repo—concise questions only, please.
Good luck, and have fun!

