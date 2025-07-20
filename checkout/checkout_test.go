package checkout

import (
	"os"
	"testing"
)

const (
	priceA        = 50
	priceB        = 30
	priceC        = 20
	priceD        = 15
	specialACount = 3
	specialAPrice = 130
	specialBCount = 2
	specialBPrice = 45
)

func mockCheckout(t *testing.T) ICheckout {
	co, err := NewCheckout("test-pricing.yaml")
	if err != nil {
		t.Fatalf("failed loading config: %v", err)
	}
	return co
}

func Test_NewCheckout_Success(t *testing.T) {
	if _, err := NewCheckout("test-pricing.yaml"); err != nil {
		t.Fatalf("expected no error loading valid config, got %v", err)
	}
}

func Test_NewCheckout_MissingFile(t *testing.T) {
	if _, err := NewCheckout("no-such-file.yaml"); err == nil {
		t.Error("expected error when config file is missing")
	}
}

func Test_NewCheckout_InvalidYAML(t *testing.T) {
	tmp, err := os.CreateTemp("", "bad-parse-*.yaml")
	if err != nil {
		t.Fatalf("create temp: %v", err)
	}
	defer os.Remove(tmp.Name())
	os.WriteFile(tmp.Name(), []byte("[][]][][]"), 0644)

	if _, err := NewCheckout(tmp.Name()); err == nil {
		t.Error("expected error parsing invalid YAML")
	}
}

func Test_Scan_Unknown_SKU(t *testing.T) {
	co := mockCheckout(t)
	if err := co.Scan("X"); err == nil {
		t.Error("expected error when scanning unknown SKU")
	}
}

func Test_Totals(t *testing.T) {
	cases := []struct {
		name string
		skus []string
		want int
	}{
		{"nothing", []string{}, 0},
		{"single A", []string{"A"}, priceA},
		{"C+D", []string{"C", "D"}, priceC + priceD},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			co := mockCheckout(t)
			for _, sku := range tc.skus {
				if err := co.Scan(sku); err != nil {
					t.Fatalf("scan %q: %v", sku, err)
				}
			}
			if got := co.GetTotalPrice(); got != tc.want {
				t.Errorf("got %d, want %d", got, tc.want)
			}
		})
	}
}

func Test_Mixed_Items(t *testing.T) {
	co := mockCheckout(t)
	seq := []string{
		"A", "A", "A", "A",
		"B", "B", "B",
		"C",
		"D",
	}

	for _, sku := range seq {
		if err := co.Scan(sku); err != nil {
			t.Fatalf("scan %q: %v", sku, err)
		}
	}

	want := (specialAPrice + priceA) + (specialBPrice + priceB) + priceC + priceD
	if got := co.GetTotalPrice(); got != want {
		t.Errorf("got %d, want %d", got, want)
	}
}

func Test_Special_Pricing(t *testing.T) {
	belowA := specialACount - 1
	belowB := specialBCount - 1
	someC := 5
	someD := 4

	cases := []struct {
		name      string
		scanCount map[string]int
		want      int
	}{
		{
			name:      "under special offer threshold A",
			scanCount: map[string]int{"A": belowA},
			want:      belowA * priceA,
		},
		{
			name:      "under special offer threshold B",
			scanCount: map[string]int{"B": belowB},
			want:      belowB * priceB,
		},
		{
			name:      "no special C",
			scanCount: map[string]int{"C": someC},
			want:      someC * priceC,
		},
		{
			name:      "no special D",
			scanCount: map[string]int{"D": someD},
			want:      someD * priceD,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			co := mockCheckout(t)
			for sku, cnt := range tc.scanCount {
				for range cnt {
					if err := co.Scan(sku); err != nil {
						t.Fatalf("scan %q: %v", sku, err)
					}
				}
			}
			if got := co.GetTotalPrice(); got != tc.want {
				t.Errorf("%q: got %d, want %d", tc.name, got, tc.want)
			}
		})
	}
}

func Test_Multiple_Specials(t *testing.T) {
	cases := []struct {
		name      string
		scanCount map[string]int
		want      int
	}{
		{
			name:      "3 A + 2 B",
			scanCount: map[string]int{"A": specialACount, "B": specialBCount},
			want:      specialAPrice + specialBPrice,
		},
		{
			name:      "6 A + 4 B",
			scanCount: map[string]int{"A": 2 * specialACount, "B": 2 * specialBCount},
			want:      2*specialAPrice + 2*specialBPrice,
		},
		{
			name:      "7 A + 5 B",
			scanCount: map[string]int{"A": 2*specialACount + 1, "B": 2*specialBCount + 1},
			want:      2*specialAPrice + priceA + 2*specialBPrice + priceB,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			co := mockCheckout(t)
			for sku, cnt := range tc.scanCount {
				for range cnt {
					if err := co.Scan(sku); err != nil {
						t.Fatalf("scan %q: %v", sku, err)
					}
				}
			}
			if got := co.GetTotalPrice(); got != tc.want {
				t.Errorf("%s: got %d, want %d", tc.name, got, tc.want)
			}
		})
	}
}
