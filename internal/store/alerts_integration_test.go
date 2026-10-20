//go:build integration

package store

import (
	"context"
	"testing"

	"distribuciones-demo.example/inventario/internal/inventory"
)

func TestLowStock(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	p, w := fixtures(t, s)

	items, err := s.LowStock(ctx)
	if err != nil || len(items) != 1 || items[0].Missing != 10 {
		t.Fatalf("empty stock must alert: %+v %v", items, err)
	}
	if _, _, err := s.RecordMovement(ctx, inventory.Movement{ProductID: p.ID, WarehouseID: w.ID, Kind: inventory.KindIn, Quantity: 11}); err != nil {
		t.Fatal(err)
	}
	if items, _ := s.LowStock(ctx); len(items) != 0 {
		t.Fatalf("no alert expected: %+v", items)
	}
}
