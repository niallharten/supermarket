# Supermarket CLI Checkout

A simple interactive checkout application that reads pricing rules from a YAML file and supports “n for y” specials (e.g. 3 × A for 130). It reloads the pricing file on each operation so you can update prices on the fly.

## Prerequisites

* Go 1.22 or later installed ([https://golang.org/dl/](https://golang.org/dl/))

## Project Layout

```text
.
├── go.mod
├── pricing.yaml           # your live pricing rules
├── main.go            
└── checkout/
    ├── checkout.go        # core checkout logic
    ├── checkout_test.go   # unit tests against test-pricing.yaml
    └── test-pricing.yaml  # test fixture for automated tests
```

## Configuration

At the project root, create or edit `pricing.yaml`:

```yaml
items:
  - sku: "A"
    unit_price: 50
    special_price:
      count: 3
      price: 130
  - sku: "B"
    unit_price: 30
    special_price:
      count: 2
      price: 45
  - sku: "C"
    unit_price: 20
  - sku: "D"
    unit_price: 15
```

This file is reloaded on every `scan`, `remove`, or `total` command.

## Build

```bash
go mod tidy
go build -o supermarket main.go
```

## Usage

```bash
./supermarket --config pricing.yaml
```

Commands:

* `scan <SKU>`: add an item
* `remove <SKU>`: remove an item
* `total`: show current total
* `checkout`: show final total and exit
* `help`: display this menu
* `exit`: quit without checkout

## Testing

```bash
go test ./supermarket
```

## Dynamic Pricing Updates

Edit `pricing.yaml` while the CLI is running; the next operation picks up changes.

