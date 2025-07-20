package main

import (
	"flag"
	"fmt"
	"os"
	"supermarket/checkout"

	"gopkg.in/yaml.v2"
)

func main() {
	configPath := flag.String("config", "pricing.yaml", "path to pricing YAML")
	flag.Parse()

	data, err := os.ReadFile(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read config %q: %v\n", *configPath, err)
		os.Exit(1)
	}

	var cfg checkout.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse YAML: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Loaded pricing rules:")
	for _, it := range cfg.Items {
		fmt.Printf("  SKU %s: unit_price=%d", it.SKU, it.UnitPrice)
		if it.SpecialPrice != nil {
			fmt.Printf(", special=%d for %d", it.SpecialPrice.Count, it.SpecialPrice.Price)
		}
	}
}
