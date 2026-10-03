//go:build integration

package store

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"distribuciones-demo.example/inventario/internal/inventory"
	"distribuciones-demo.example/inventario/internal/migrate"
	"distribuciones-demo.example/inventario/migrations"
)

// newTestStore connects to DATABASE_URL, applies the migrations and empties
// every table. Integration tests share one database, so run them with -p 1.
func newTestStore(t *testing.T) *Store {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Fatal("DATABASE_URL is required for integration tests (see make db-up)")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	all, err := migrate.Load(migrations.FS)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := migrate.Up(ctx, pool, all); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `TRUNCATE products, warehouses, stock, movements RESTART IDENTITY CASCADE`); err != nil {
		t.Fatal(err)
	}
	return New(pool)
}

func fixtures(t *testing.T, s *Store) (inventory.Product, inventory.Warehouse) {
	t.Helper()
	ctx := context.Background()
	p, err := s.CreateProduct(ctx, inventory.Product{SKU: "ACE-500", Name: "Aceite 500 ml", Unit: "ud", MinStock: 10})
	if err != nil {
		t.Fatal(err)
	}
	w, err := s.CreateWarehouse(ctx, inventory.Warehouse{Code: "MAD-01", Name: "Central", City: "Madrid"})
	if err != nil {
		t.Fatal(err)
	}
	return p, w
}

func TestMigrationsAreIdempotent(t *testing.T) {
	s := newTestStore(t)
	all, _ := migrate.Load(migrations.FS)
	applied, err := migrate.Up(context.Background(), s.pool, all)
	if err != nil || len(applied) != 0 {
		t.Fatalf("second run must apply nothing, got %v, %v", applied, err)
	}
}

func TestProductsAndWarehouses(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	p, _ := fixtures(t, s)
	got, err := s.GetProduct(ctx, p.ID)
	if err != nil || got.SKU != "ACE-500" || got.MinStock != 10 {
		t.Fatalf("get product: %+v %v", got, err)
	}
	if _, err := s.CreateProduct(ctx, inventory.Product{SKU: "ACE-500", Name: "x", Unit: "ud"}); !errors.Is(err, inventory.ErrConflict) {
		t.Fatalf("duplicate sku: want ErrConflict, got %v", err)
	}
	if _, err := s.GetProduct(ctx, 999); !errors.Is(err, inventory.ErrNotFound) {
		t.Fatalf("missing product: want ErrNotFound, got %v", err)
	}
	products, _ := s.ListProducts(ctx)
	warehouses, _ := s.ListWarehouses(ctx)
	if len(products) != 1 || len(warehouses) != 1 {
		t.Fatalf("lists: %d products, %d warehouses", len(products), len(warehouses))
	}
}

func TestRecordMovement(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	p, w := fixtures(t, s)

	_, qty, err := s.RecordMovement(ctx, inventory.Movement{ProductID: p.ID, WarehouseID: w.ID, Kind: inventory.KindIn, Quantity: 40})
	if err != nil || qty != 40 {
		t.Fatalf("entry: %d %v", qty, err)
	}
	_, qty, err = s.RecordMovement(ctx, inventory.Movement{ProductID: p.ID, WarehouseID: w.ID, Kind: inventory.KindOut, Quantity: 15})
	if err != nil || qty != 25 {
		t.Fatalf("exit: %d %v", qty, err)
	}
	_, _, err = s.RecordMovement(ctx, inventory.Movement{ProductID: p.ID, WarehouseID: w.ID, Kind: inventory.KindOut, Quantity: 26})
	if !errors.Is(err, inventory.ErrInsufficientStock) {
		t.Fatalf("want ErrInsufficientStock, got %v", err)
	}
	_, _, err = s.RecordMovement(ctx, inventory.Movement{ProductID: 999, WarehouseID: w.ID, Kind: inventory.KindIn, Quantity: 1})
	if !errors.Is(err, inventory.ErrNotFound) {
		t.Fatalf("unknown product: want ErrNotFound, got %v", err)
	}

	movements, _ := s.ListMovements(ctx, p.ID)
	if len(movements) != 2 || movements[0].Kind != inventory.KindOut {
		t.Fatalf("ledger: %+v", movements)
	}
	levels, _ := s.StockLevels(ctx, w.ID)
	if len(levels) != 1 || levels[0].Quantity != 25 || levels[0].WarehouseCode != "MAD-01" {
		t.Fatalf("levels: %+v", levels)
	}
}
