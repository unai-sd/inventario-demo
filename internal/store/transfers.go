package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"distribuciones-demo.example/inventario/internal/inventory"
)

// Transfer moves stock between two warehouses atomically.
func (s *Store) Transfer(ctx context.Context, t inventory.Transfer) (inventory.Transfer, error) {
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx,
			`INSERT INTO transfers (product_id, from_warehouse_id, to_warehouse_id, quantity)
			 VALUES ($1, $2, $3, $4) RETURNING id, created_at`,
			t.ProductID, t.FromWarehouseID, t.ToWarehouseID, t.Quantity,
		).Scan(&t.ID, &t.CreatedAt); err != nil {
			return err
		}
		reason := fmt.Sprintf("transfer #%d", t.ID)
		legs := []inventory.Movement{
			{ProductID: t.ProductID, WarehouseID: t.FromWarehouseID, Kind: inventory.KindOut},
			{ProductID: t.ProductID, WarehouseID: t.ToWarehouseID, Kind: inventory.KindIn},
		}
		for _, leg := range legs {
			leg.Quantity, leg.Reason = t.Quantity, reason
			if _, _, err := applyMovement(ctx, tx, leg); err != nil {
				return err
			}
		}
		return nil
	})
	return t, mapError(err)
}
