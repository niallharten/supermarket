package checkout

import (
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
	return nil
}

func (c *checkout) GetTotalPrice() int {
	return 0
}
