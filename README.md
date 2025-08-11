# Bursa Crypto Indonesia – Coding Exercise
**Build a Minimal Crypto-Exchange Matching Engine**

This project is an implementation of a crypto-exchange matching engine. It's designed to be simple, correct, and production-quality.

---

## How to Run

1.  **Build the engine:**
    ```bash
    make build
    ```
    This will create an executable file named `engine` in the project root.

2.  **Run the engine:**
    ```bash
    ./engine path/to/orders.json
    ```
    Replace `path/to/orders.json` with the actual path to your input JSON file. For example:
    ```bash
    ./engine test/test1_gemini.json
    ```

---

## How to Test

Run the full test suite using Make:

```bash
make test
```

This command will run all unit tests, including the test cases located in the `test/` directory.

---

## How to Add a New Test Case

To add a new test case, you need to create two files: an input file and an output file.

1.  **Create an input file:**
    -   Create a new JSON file in the `test/` directory (e.g., `test/new_test.json`).
    -   This file should contain a JSON array of order commands.

    **Example: `test/new_test.json`**
    ```json
    [
      {"cmd":"NEW","id":"O-1","side":"BUY","type":"LIMIT","price":65000,"qty":0.5},
      {"cmd":"REPLACE","id":"O-1","price":65100,"qty":0.6},
      {"cmd":"NEW","id":"O-2","side":"SELL","type":"MARKET","qty":0.2}
    ]
    ```

2.  **Create an output file:**
    -   Create a corresponding JSON file in the `output/` directory with the same name (e.g., `output/new_test.json`).
    -   This file should contain the expected JSON output after running the input commands.

    **Example: `output/new_test.json`**
    ```json
    {
     "trades": [
      {
       "buyId": "O-1",
       "sellId": "O-2",
       "price": 65100,
       "qty": "0.2",
       "execId": 1
      }
     ],
     "orderBook": {
      "bids": [
       {
        "id": "O-1",
        "price": 65100,
        "qty": "0.4"
       }
      ],
      "asks": []
     }
    }
    ```

3.  **Run the tests:**
    The new test case will be automatically picked up and executed when you run `make test`.
