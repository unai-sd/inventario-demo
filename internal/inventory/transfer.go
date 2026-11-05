package inventory

import "time"

// Transfer moves units of a product from one warehouse to another. It is
// recorded as an outgoing and an incoming movement plus the transfer itself.
type Transfer struct {
	ID              int64     `json:"id"`
	ProductID       int64     `json:"product_id"`
	FromWarehouseID int64     `json:"from_warehouse_id"`
	ToWarehouseID   int64     `json:"to_warehouse_id"`
	Quantity        int       `json:"quantity"`
	CreatedAt       time.Time `json:"created_at"`
}

// Validate checks a transfer request.
func (t *Transfer) Validate() error {
	switch {
	case t.ProductID <= 0:
		return invalid("product_id", "is required")
	case t.FromWarehouseID <= 0 || t.ToWarehouseID <= 0:
		return invalid("warehouse", "origin and destination are required")
	case t.FromWarehouseID == t.ToWarehouseID:
		return invalid("to_warehouse_id", "must differ from the origin warehouse")
	case t.Quantity <= 0:
		return invalid("quantity", "must be greater than zero")
	}
	return nil
}
