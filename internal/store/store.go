// Package store persists the inventory in PostgreSQL using pgx.
package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"distribuciones-demo.example/inventario/internal/inventory"
)

// Store is the PostgreSQL implementation of the inventory repository.
type Store struct {
	pool *pgxpool.Pool
}

// New wraps a connection pool.
func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

// mapError translates PostgreSQL errors into domain errors.
func mapError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique_violation
			return inventory.ErrConflict
		case "23503": // foreign_key_violation
			return inventory.ErrNotFound
		}
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return inventory.ErrNotFound
	}
	return err
}

// CreateProduct inserts a product.
func (s *Store) CreateProduct(ctx context.Context, p inventory.Product) (inventory.Product, error) {
	err := s.pool.QueryRow(ctx,
		`INSERT INTO products (sku, name, unit, min_stock) VALUES ($1, $2, $3, $4)
		 RETURNING id, created_at`,
		p.SKU, p.Name, p.Unit, p.MinStock,
	).Scan(&p.ID, &p.CreatedAt)
	return p, mapError(err)
}

// ListProducts returns every product ordered by SKU.
func (s *Store) ListProducts(ctx context.Context) ([]inventory.Product, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, sku, name, unit, min_stock, created_at FROM products ORDER BY sku`)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByPos[inventory.Product])
}

// GetProduct returns one product.
func (s *Store) GetProduct(ctx context.Context, id int64) (inventory.Product, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, sku, name, unit, min_stock, created_at FROM products WHERE id = $1`, id)
	if err != nil {
		return inventory.Product{}, err
	}
	p, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByPos[inventory.Product])
	return p, mapError(err)
}

// CreateWarehouse inserts a warehouse.
func (s *Store) CreateWarehouse(ctx context.Context, w inventory.Warehouse) (inventory.Warehouse, error) {
	err := s.pool.QueryRow(ctx,
		`INSERT INTO warehouses (code, name, city) VALUES ($1, $2, $3) RETURNING id`,
		w.Code, w.Name, w.City,
	).Scan(&w.ID)
	return w, mapError(err)
}

// ListWarehouses returns every warehouse ordered by code.
func (s *Store) ListWarehouses(ctx context.Context) ([]inventory.Warehouse, error) {
	rows, err := s.pool.Query(ctx, `SELECT id, code, name, city FROM warehouses ORDER BY code`)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByPos[inventory.Warehouse])
}

// RecordMovement applies a movement to the stock and appends it to the ledger
// in a single transaction. It returns the stored movement and the new quantity.
func (s *Store) RecordMovement(ctx context.Context, m inventory.Movement) (inventory.Movement, int, error) {
	var quantity int
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		var err error
		m, quantity, err = applyMovement(ctx, tx, m)
		return err
	})
	return m, quantity, mapError(err)
}

func applyMovement(ctx context.Context, tx pgx.Tx, m inventory.Movement) (inventory.Movement, int, error) {
	if _, err := tx.Exec(ctx,
		`INSERT INTO stock (product_id, warehouse_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		m.ProductID, m.WarehouseID); err != nil {
		return m, 0, err
	}
	var current int
	if err := tx.QueryRow(ctx,
		`SELECT quantity FROM stock WHERE product_id = $1 AND warehouse_id = $2 FOR UPDATE`,
		m.ProductID, m.WarehouseID).Scan(&current); err != nil {
		return m, 0, err
	}
	next, err := inventory.ApplyMovement(current, m.Kind, m.Quantity)
	if err != nil {
		return m, current, err
	}
	if _, err := tx.Exec(ctx,
		`UPDATE stock SET quantity = $3 WHERE product_id = $1 AND warehouse_id = $2`,
		m.ProductID, m.WarehouseID, next); err != nil {
		return m, current, err
	}
	err = tx.QueryRow(ctx,
		`INSERT INTO movements (product_id, warehouse_id, kind, quantity, reason)
		 VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at`,
		m.ProductID, m.WarehouseID, m.Kind, m.Quantity, m.Reason,
	).Scan(&m.ID, &m.CreatedAt)
	return m, next, err
}

// ListMovements returns the ledger, newest first. A productID of 0 lists all.
func (s *Store) ListMovements(ctx context.Context, productID int64) ([]inventory.Movement, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, product_id, warehouse_id, kind, quantity, reason, created_at
		 FROM movements WHERE $1 = 0 OR product_id = $1 ORDER BY id DESC LIMIT 500`, productID)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByPos[inventory.Movement])
}

// StockLevels returns the stock per product and warehouse. A warehouseID of 0
// lists every warehouse.
func (s *Store) StockLevels(ctx context.Context, warehouseID int64) ([]inventory.StockLevel, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT s.product_id, p.sku, s.warehouse_id, w.code, s.quantity
		 FROM stock s
		 JOIN products p ON p.id = s.product_id
		 JOIN warehouses w ON w.id = s.warehouse_id
		 WHERE $1 = 0 OR s.warehouse_id = $1
		 ORDER BY w.code, p.sku`, warehouseID)
	if err != nil {
		return nil, fmt.Errorf("stock levels: %w", err)
	}
	return pgx.CollectRows(rows, pgx.RowToStructByPos[inventory.StockLevel])
}
