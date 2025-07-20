package checkout

import (
	"os"
	"testing"
)

func Test_NewCheckout_Success(t *testing.T) {
	_, err := NewCheckout("test-pricing.yaml")
	if err != nil {
		t.Fatalf("expected no error loading valid config, got %v", err)
	}
}

func Test_NewCheckout_MissingFile(t *testing.T) {
	if _, err := NewCheckout("no-such-file.yml"); err == nil {
		t.Error("expected error when config file is missing")
	}
}

func Test_NewCheckout_InvalidYAML(t *testing.T) {
	tmp, err := os.CreateTemp("", "bad-parse.yml")
	if err != nil {
		t.Fatalf("create temp: %v", err)
	}
	tmp.WriteString("[][]][][]")

	if _, err := NewCheckout(tmp.Name()); err == nil {
		t.Error("expected error parsing invalid YAML")
	}
}

func Test_GetTotalPrice_DefaultZero(t *testing.T) {
	co, _ := NewCheckout("test-pricing.yaml")
	if total := co.GetTotalPrice(); total != 0 {
		t.Errorf("expected default total 0, got %d", total)
	}
}
