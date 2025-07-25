package checkout

import (
	"os"
	"path/filepath"
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

func Test_Remove_Unknown_Or_Empty(t *testing.T) {
	co := mockCheckout(t)
	if err := co.Remove("X"); err == nil {
		t.Error("expected error removing unknown SKU")
	}
	if err := co.Remove("A"); err == nil {
		t.Error("expected error removing when none scanned")
	}
}

func Test_Remove_And_Recalculate(t *testing.T) {
	co := mockCheckout(t)
	for _, sku := range []string{"A", "A", "A", "B"} {
		if err := co.Scan(sku); err != nil {
			t.Fatalf("scan %q: %v", sku, err)
		}
	}
	if total := co.GetTotalPrice(); total != 160 {
		t.Fatalf("before remove: got %d, want %d", total, 160)
	}

	if err := co.Remove("A"); err != nil {
		t.Fatalf("remove A: %v", err)
	}
	if total := co.GetTotalPrice(); total != 130 {
		t.Errorf("after remove: got %d, want %d", total, 130)
	}
}

func Test_Remove_Beyond_Zero(t *testing.T) {
	co := mockCheckout(t)
	if err := co.Scan("C"); err != nil {
		t.Fatalf("scan C: %v", err)
	}
	if err := co.Remove("C"); err != nil {
		t.Fatalf("remove C: %v", err)
	}
	if err := co.Remove("C"); err == nil {
		t.Error("expected error removing C beyond zero")
	}
}

func Test_Pricing_Yaml_Changing(t *testing.T) {
	// 1) Write initial pricing (no special for A)

	dir := t.TempDir()
	config := filepath.Join(dir, "pricing.yaml")
	initial := `
				items:
				- sku: "A"
					unit_price: 50
				- sku: "B"
					unit_price: 30
				`
	if err := os.WriteFile(config, []byte(initial), 0644); err != nil {
		t.Fatalf("write initial config: %v", err)
	}

	// 2) Load checkout and scan 3 A's
	co, err := NewCheckout(config)
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}

	for range 3 {
		if err := co.Scan("A"); err != nil {
			t.Fatalf("scan A: %v", err)
		}
	}

	if got := co.GetTotalPrice(); got != 3*50 {
		t.Errorf("before reload: got %d, want %d", got, 3*50)
	}

	// 3) Overwrite YAML to add special 3-for-130 on A
	updated := `
				items:
				- sku: "A"
					unit_price: 50
					special_price:
					count: 3
					price: 130
				- sku: "B"
					unit_price: 30
				`

	if err := os.WriteFile(config, []byte(updated), 0644); err != nil {
		t.Fatalf("write updated config: %v", err)
	}

	// 4) The next GetTotalPrice should pick up the special
	if got := co.GetTotalPrice(); got != 130 {
		t.Errorf("after reload: got %d, want %d", got, 130)
	}
}
