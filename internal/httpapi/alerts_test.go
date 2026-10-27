package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"distribuciones-demo.example/inventario/internal/inventory"
)

func (f *fakeStore) LowStock(context.Context) ([]inventory.LowStockItem, error) {
	totals := map[int64]int{}
	for k, q := range f.stock {
		totals[k.product] += q
	}
	items := []inventory.LowStockItem{}
	for _, p := range f.products {
		if inventory.IsLowStock(totals[p.ID], p.MinStock) {
			items = append(items, inventory.NewLowStockItem(p, totals[p.ID]))
		}
	}
	return items, nil
}

func TestLowStockAlerts(t *testing.T) {
	h, _ := seeded(t)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/alerts/low-stock", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"missing":10`) {
		t.Fatalf("empty stock must alert: %d %s", rec.Code, rec.Body)
	}
	do(t, h, "POST", "/movements", `{"product_id":1,"warehouse_id":2,"kind":"in","quantity":25}`)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/alerts/low-stock", nil))
	if strings.TrimSpace(rec.Body.String()) != "[]" {
		t.Fatalf("no alert expected: %s", rec.Body)
	}
}
