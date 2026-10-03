package httpapi

import (
	"context"
	"sort"
	"time"

	"distribuciones-demo.example/inventario/internal/inventory"
)

type stockKey struct{ product, warehouse int64 }

// fakeStore is an in-memory Store for handler tests.
type fakeStore struct {
	products   map[int64]inventory.Product
	warehouses map[int64]inventory.Warehouse
	stock      map[stockKey]int
	movements  []inventory.Movement
	nextID     int64
}

func newFakeStore() *fakeStore {
	return &fakeStore{
		products:   map[int64]inventory.Product{},
		warehouses: map[int64]inventory.Warehouse{},
		stock:      map[stockKey]int{},
	}
}

func (f *fakeStore) id() int64 { f.nextID++; return f.nextID }

func (f *fakeStore) CreateProduct(_ context.Context, p inventory.Product) (inventory.Product, error) {
	for _, existing := range f.products {
		if existing.SKU == p.SKU {
			return p, inventory.ErrConflict
		}
	}
	p.ID, p.CreatedAt = f.id(), time.Now()
	f.products[p.ID] = p
	return p, nil
}

func (f *fakeStore) ListProducts(context.Context) ([]inventory.Product, error) {
	out := []inventory.Product{}
	for _, p := range f.products {
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].SKU < out[j].SKU })
	return out, nil
}

func (f *fakeStore) GetProduct(_ context.Context, id int64) (inventory.Product, error) {
	p, ok := f.products[id]
	if !ok {
		return p, inventory.ErrNotFound
	}
	return p, nil
}

func (f *fakeStore) CreateWarehouse(_ context.Context, w inventory.Warehouse) (inventory.Warehouse, error) {
	w.ID = f.id()
	f.warehouses[w.ID] = w
	return w, nil
}

func (f *fakeStore) ListWarehouses(context.Context) ([]inventory.Warehouse, error) {
	out := []inventory.Warehouse{}
	for _, w := range f.warehouses {
		out = append(out, w)
	}
	return out, nil
}

func (f *fakeStore) RecordMovement(_ context.Context, m inventory.Movement) (inventory.Movement, int, error) {
	if _, ok := f.products[m.ProductID]; !ok {
		return m, 0, inventory.ErrNotFound
	}
	if _, ok := f.warehouses[m.WarehouseID]; !ok {
		return m, 0, inventory.ErrNotFound
	}
	key := stockKey{m.ProductID, m.WarehouseID}
	next, err := inventory.ApplyMovement(f.stock[key], m.Kind, m.Quantity)
	if err != nil {
		return m, f.stock[key], err
	}
	f.stock[key] = next
	m.ID, m.CreatedAt = f.id(), time.Now()
	f.movements = append(f.movements, m)
	return m, next, nil
}

func (f *fakeStore) ListMovements(_ context.Context, productID int64) ([]inventory.Movement, error) {
	out := []inventory.Movement{}
	for _, m := range f.movements {
		if productID == 0 || m.ProductID == productID {
			out = append(out, m)
		}
	}
	return out, nil
}

func (f *fakeStore) StockLevels(_ context.Context, warehouseID int64) ([]inventory.StockLevel, error) {
	out := []inventory.StockLevel{}
	for k, q := range f.stock {
		if warehouseID == 0 || k.warehouse == warehouseID {
			out = append(out, inventory.StockLevel{
				ProductID: k.product, SKU: f.products[k.product].SKU,
				WarehouseID: k.warehouse, WarehouseCode: f.warehouses[k.warehouse].Code, Quantity: q,
			})
		}
	}
	return out, nil
}
