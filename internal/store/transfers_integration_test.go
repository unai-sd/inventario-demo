//go:build integration

package store

import (
	"context"
	"errors"
	"testing"

	"distribuciones-demo.example/inventario/internal/inventory"
)

func TestTransfer(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	p, from := fixtures(t, s)
	to, err := s.CreateWarehouse(ctx, inventory.Warehouse{Code: "VLC-01", Name: "Levante", City: "Valencia"})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.RecordMovement(ctx, inventory.Movement{ProductID: p.ID, WarehouseID: from.ID, Kind: inventory.KindIn, Quantity: 30}); err != nil {
		t.Fatal(err)
	}

	tr, err := s.Transfer(ctx, inventory.Transfer{ProductID: p.ID, FromWarehouseID: from.ID, ToWarehouseID: to.ID, Quantity: 12})
	if err != nil || tr.ID == 0 {
		t.Fatalf("transfer: %+v %v", tr, err)
	}
	levels, _ := s.StockLevels(ctx, 0)
	got := map[string]int{}
	for _, l := range levels {
		got[l.WarehouseCode] = l.Quantity
	}
	if got["MAD-01"] != 18 || got["VLC-01"] != 12 {
		t.Fatalf("levels after transfer: %v", got)
	}

	_, err = s.Transfer(ctx, inventory.Transfer{ProductID: p.ID, FromWarehouseID: from.ID, ToWarehouseID: to.ID, Quantity: 100})
	if !errors.Is(err, inventory.ErrInsufficientStock) {
		t.Fatalf("want ErrInsufficientStock, got %v", err)
	}
	if movements, _ := s.ListMovements(ctx, p.ID); len(movements) != 3 {
		t.Fatalf("failed transfer must not leave movements: %d", len(movements))
	}
}
