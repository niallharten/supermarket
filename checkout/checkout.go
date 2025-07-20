package checkout

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type ICheckout interface {
	Scan(sku string) error
	GetTotalPrice() int
}

type SpecialPrice struct {
	Count int `yaml:"count"`
	Price int `yaml:"price"`
}

type ItemRule struct {
	SKU          string        `yaml:"sku"`
	UnitPrice    int           `yaml:"unit_price"`
	SpecialPrice *SpecialPrice `yaml:"special_price,omitempty"`
}

type Config struct {
	Items []ItemRule `yaml:"items"`
}

type checkout struct {
	rules   map[string]ItemRule
	scanned map[string]int
}

func NewCheckout(path string) (ICheckout, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	rules := make(map[string]ItemRule, len(cfg.Items))
	for _, it := range cfg.Items {
		rules[it.SKU] = it
	}
	return &checkout{rules: rules, scanned: make(map[string]int)}, nil
}

func (c *checkout) Scan(sku string) error {
	if _, ok := c.rules[sku]; !ok {
		return fmt.Errorf("unknown SKU %q", sku)
	}
	c.scanned[sku]++
	return nil
}

func (c *checkout) GetTotalPrice() int {
	var total int
	for sku, count := range c.scanned {
		rule := c.rules[sku]

		// if there's a special deal and we've scanned enough
		if sp := rule.SpecialPrice; sp != nil && count >= sp.Count {
			bundles := count / sp.Count
			total += bundles * sp.Price
			count -= bundles * sp.Count
		}
		total += count * rule.UnitPrice
	}
	return total
}
