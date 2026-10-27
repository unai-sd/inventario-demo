package store

import (
	"context"

	"github.com/jackc/pgx/v5"

	"distribuciones-demo.example/inventario/internal/inventory"
)

// LowStock returns the products whose total stock requires an alert.
func (s *Store) LowStock(ctx context.Context) ([]inventory.LowStockItem, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, sku, name, unit, min_stock, created_at, total
		 FROM product_stock_totals ORDER BY sku`)
	if err != nil {
		return nil, err
	}
	items := []inventory.LowStockItem{}
	var p inventory.Product
	var total int
	_, err = pgx.ForEachRow(rows,
		[]any{&p.ID, &p.SKU, &p.Name, &p.Unit, &p.MinStock, &p.CreatedAt, &total},
		func() error {
			if inventory.IsLowStock(total, p.MinStock) {
				items = append(items, inventory.NewLowStockItem(p, total))
			}
			return nil
		})
	return items, err
}
