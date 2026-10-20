package inventory

import "testing"

func TestIsLowStock(t *testing.T) {
	cases := []struct {
		total, min int
		want       bool
	}{
		{0, 0, false},
		{5, 10, true},
		{11, 10, false},
	}
	for _, tc := range cases {
		if got := IsLowStock(tc.total, tc.min); got != tc.want {
			t.Errorf("IsLowStock(%d, %d) = %v, want %v", tc.total, tc.min, got, tc.want)
		}
	}
}

func TestNewLowStockItem(t *testing.T) {
	item := NewLowStockItem(Product{ID: 3, SKU: "ARZ-1K", Name: "Arroz 1 kg", MinStock: 20}, 8)
	if item.Missing != 12 || item.Total != 8 {
		t.Fatalf("unexpected item: %+v", item)
	}
}
