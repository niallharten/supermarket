package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"supermarket/checkout"
)

func printMenu() {
	fmt.Println(`commands:
				1. scan <SKU>     — add a item
				2. remove <SKU>   — remove a item
				3. total          — show current total
				4. checkout       — show total and exit
				5. exit     — exit without checking out
				
				Enter a command to get started`)
}

func main() {
	configPath := flag.String("config", "pricing.yaml", "path to pricing YAML")
	flag.Parse()

	co, err := checkout.NewCheckout(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading config: %v\n", err)
		os.Exit(1)
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Type 'help' for commands")
	printMenu()

	for {
		fmt.Print("> ")
		line, _ := reader.ReadString('\n')
		parts := strings.Fields(strings.TrimSpace(line))
		if len(parts) == 0 {
			continue
		}

		switch cmd, args := parts[0], parts[1:]; cmd {

		case "scan":
			if len(args) != 1 {
				fmt.Println("Command use: scan <SKU>")
				continue
			}

			if err := co.Scan(args[0]); err != nil {
				fmt.Println("error:", err)
			}

		case "total":
			t := co.GetTotalPrice()
			fmt.Printf("Total: %d\n", t)

		case "checkout":
			t := co.GetTotalPrice()
			fmt.Printf("Final total: %d\n", t)
			return

		case "help":
			printMenu()

		case "exit":
			return

		default:
			fmt.Println("unknown command; type 'help'")
		}
	}
}
