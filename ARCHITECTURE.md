### Architectural Choices and Thought Process

My main goal when building this matching engine was to make it both **fast** and **correct**. Here’s a breakdown of my architectural decisions and the specific Go packages I chose to achieve that.

---

### The Core Problem: How to Match Orders Quickly?

The biggest challenge is efficiently matching new orders with existing ones. A **buy** order wants the lowest-priced **sell** order, while a **sell** order wants the highest-priced **buy** order.

To solve this, I split the order book into two separate data structures: one for **buy** orders and one for **sell** orders. The key is to keep them sorted by price, with the best-priced order always at the top.

### Why I Chose `container/heap`: The Heart of the Engine

This problem led me directly to Go's built-in `container/heap` package. A heap is Go's implementation of a priority queue, and it's perfect for this task.

* **Finding the best order is instant ($O(1)$).** The best-priced order is always at the top of the heap, ready to be matched. This means there's no need to search through a list of orders.
* **Adding or removing orders is very fast ($O(\log n)$).** When a new order arrives, the heap efficiently places it in the correct spot. This is exponentially faster than re-sorting the entire list of orders.

This is the main reason the engine can handle so many orders so quickly—it's built to do the minimum amount of work necessary to find the next match.

---

### The Next Problem: How to Cancel Orders Quickly?

The heap is great for finding the best price, but what about when a user wants to cancel an order by its ID? Searching for that ID inside the heap would be slow. I needed a way to find any order by its ID instantly.

My solution was to use an **`orderMap`**. This is a simple Go `map` where the key is the order's string ID and the value is a pointer to the actual `Order` object. Now, when a cancel request comes in, I can find the order in the map instantly ($O(1)$) and simply mark it as "canceled." The next time that order reaches the top of the heap, the engine will see the canceled flag and discard it.

---

### Other Key Packages

#### `shopspring/decimal`: For Handling Money Safely

Using floating-point numbers for financial calculations is a bad idea due to potential rounding errors. To avoid this, I used the `shopspring/decimal` package, which is designed to handle high-precision decimal values. It ensures all financial calculations are accurate and reliable.

#### `google/go-cmp`: For Better Testing

When a test fails, I need to know exactly what went wrong. Go's built-in `==` operator on complex structs only tells you if two things are identical, which isn't very helpful for debugging. The `go-cmp` library provides a clean, readable "diff" that shows exactly which fields in a struct are different, making it much easier to debug and fix tests.


### Scalability
Scalability and Event-Driven Architecture

The core logic of this engine is designed to be highly adaptable to a scalable, event-driven architecture. Here’s how it works:

How It Scales

This design is very friendly to horizontal scaling. Instead of running a single, monolithic engine, we can distribute the workload by having a separate matching engine instance for each trading pair (e.g., BTC/USD, ETH/USD).

- Partitioning the Order Book: Each instance of the matching engine would be responsible for only one specific trading pair. This means each engine has a smaller, more manageable buy heap and sell heap, which keeps processing fast.

- Parallel Processing: Multiple matching engines can run simultaneously on different servers, processing orders for different trading pairs independently. This allows the system to handle a high volume of orders across many markets without a single point of failure or performance bottleneck.

Integrating with an Event-Driven System

The engine's logic is a natural fit for an event-driven model. The entire process can be framed as a series of events and responses.

1. Incoming Event: A new order arrives. This is an event.
2. Engine Action: The matching engine consumes this event and processes the order.
3. Outgoing Events:
    - Match Event: If a match is found, the engine publishes a "trade executed" event.
    - Book Update Event: If the order is not fully matched, it's added to the order book, and the engine publishes an "order book updated" event.
    - Cancellation Event: If an order is canceled, the engine publishes an "order canceled" event.

This architecture ensures that the system is decoupled. Components don't need to know about each other; they just react to the events they care about. For example, a separate service can listen for "trade executed" events to handle transactions, while another service can listen for "order book updated" events to broadcast changes to clients.

In summary, the combination of a fast, heap-based core logic with a decoupled, event-driven architecture makes this engine not only efficient for single-pair trading but also highly scalable for a complex, multi-market exchange.
