package checkout

import (
	"fmt"
	"os"
	"sync"

	"gopkg.in/yaml.v2"
)

type ICheckout interface {
	Scan(sku string) error
	Remove(sku string) error
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
	path    string
	mu      sync.RWMutex
	rules   map[string]ItemRule
	scanned map[string]int
}

func NewCheckout(path string) (ICheckout, error) {
	c := &checkout{
		path:    path,
		rules:   make(map[string]ItemRule),
		scanned: make(map[string]int),
	}

	if err := c.checkIfPriceChanged(); err != nil {
		return nil, err
	}

	return c, nil
}

// simulating that if the yaml file did changed
// we're prepared for it to change on the fly
// in an ideal world, you would have it managed by an external service, so we wouldn't need to check it manually
func (c *checkout) checkIfPriceChanged() error {
	data, err := os.ReadFile(c.path)
	if err != nil {
		return err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return err
	}

	newRules := make(map[string]ItemRule, len(config.Items))
	for _, it := range config.Items {
		newRules[it.SKU] = it
	}

	c.mu.Lock()
	c.rules = newRules
	c.mu.Unlock()

	return nil
}

func (c *checkout) Scan(sku string) error {
	c.checkIfPriceChanged()

	c.mu.RLock()
	_, ok := c.rules[sku]
	c.mu.RUnlock()

	if !ok {
		return fmt.Errorf("unknown SKU %q", sku)
	}

	c.mu.Lock()
	c.scanned[sku]++
	c.mu.Unlock()

	return nil
}

func (c *checkout) Remove(sku string) error {
	c.checkIfPriceChanged()

	c.mu.RLock()
	_, ok := c.rules[sku]
	c.mu.RUnlock()

	if !ok {
		return fmt.Errorf("unknown SKU %q", sku)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.scanned[sku] == 0 {
		return fmt.Errorf("no %q in cart to remove", sku)
	}

	c.scanned[sku]--
	return nil
}

func (c *checkout) GetTotalPrice() int {
	c.checkIfPriceChanged()

	c.mu.RLock()
	defer c.mu.RUnlock()

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
