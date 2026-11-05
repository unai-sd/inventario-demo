package httpapi

import (
	"context"
	"net/http"
	"testing"
	"time"

	"distribuciones-demo.example/inventario/internal/inventory"
)

func (f *fakeStore) Transfer(ctx context.Context, t inventory.Transfer) (inventory.Transfer, error) {
	from := stockKey{t.ProductID, t.FromWarehouseID}
	if f.stock[from] < t.Quantity {
		return t, inventory.ErrInsufficientStock
	}
	out := inventory.Movement{ProductID: t.ProductID, WarehouseID: t.FromWarehouseID, Kind: inventory.KindOut, Quantity: t.Quantity}
	in := inventory.Movement{ProductID: t.ProductID, WarehouseID: t.ToWarehouseID, Kind: inventory.KindIn, Quantity: t.Quantity}
	for _, m := range []inventory.Movement{out, in} {
		if _, _, err := f.RecordMovement(ctx, m); err != nil {
			return t, err
		}
	}
	t.ID, t.CreatedAt = f.id(), time.Now()
	return t, nil
}

func TestTransfers(t *testing.T) {
	h, store := seeded(t)
	do(t, h, "POST", "/movements", `{"product_id":1,"warehouse_id":2,"kind":"in","quantity":20}`)
	rec, _ := do(t, h, "POST", "/transfers", `{"product_id":1,"from_warehouse_id":2,"to_warehouse_id":3,"quantity":8}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("transfer: %d %s", rec.Code, rec.Body)
	}
	if store.stock[stockKey{1, 2}] != 12 || store.stock[stockKey{1, 3}] != 8 {
		t.Fatalf("unexpected stock: %v", store.stock)
	}
	if rec, _ := do(t, h, "POST", "/transfers", `{"product_id":1,"from_warehouse_id":2,"to_warehouse_id":3,"quantity":50}`); rec.Code != http.StatusConflict {
		t.Fatalf("insufficient: want 409, got %d", rec.Code)
	}
	if rec, _ := do(t, h, "POST", "/transfers", `{"product_id":1,"from_warehouse_id":2,"to_warehouse_id":2,"quantity":1}`); rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("same warehouse: want 422, got %d", rec.Code)
	}
}
