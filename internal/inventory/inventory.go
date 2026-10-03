// Package inventory holds the domain model and the pure business rules of the
// warehouse inventory: products, warehouses, stock movements and validation.
package inventory

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// MovementKind is the type of a stock movement.
type MovementKind string

const (
	// KindIn adds units to a warehouse (goods received).
	KindIn MovementKind = "in"
	// KindOut removes units from a warehouse (goods shipped).
	KindOut MovementKind = "out"
	// KindAdjust sets the absolute quantity after a physical count.
	KindAdjust MovementKind = "adjust"
)

var (
	// ErrNotFound is returned when a referenced entity does not exist.
	ErrNotFound = errors.New("not found")
	// ErrConflict is returned when a unique value already exists.
	ErrConflict = errors.New("already exists")
	// ErrInsufficientStock is returned when an outgoing movement exceeds the stock.
	ErrInsufficientStock = errors.New("insufficient stock")
)

// ValidationError describes an invalid input field.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

func invalid(field, message string) error {
	return &ValidationError{Field: field, Message: message}
}

// Product is an item the distributor stores and ships.
type Product struct {
	ID        int64     `json:"id"`
	SKU       string    `json:"sku"`
	Name      string    `json:"name"`
	Unit      string    `json:"unit"`
	MinStock  int       `json:"min_stock"`
	CreatedAt time.Time `json:"created_at"`
}

// Warehouse is a physical location that holds stock.
type Warehouse struct {
	ID   int64  `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
	City string `json:"city"`
}

// Movement is an entry in the stock ledger of one product in one warehouse.
type Movement struct {
	ID          int64        `json:"id"`
	ProductID   int64        `json:"product_id"`
	WarehouseID int64        `json:"warehouse_id"`
	Kind        MovementKind `json:"kind"`
	Quantity    int          `json:"quantity"`
	Reason      string       `json:"reason"`
	CreatedAt   time.Time    `json:"created_at"`
}

// StockLevel is the current quantity of a product in a warehouse.
type StockLevel struct {
	ProductID     int64  `json:"product_id"`
	SKU           string `json:"sku"`
	WarehouseID   int64  `json:"warehouse_id"`
	WarehouseCode string `json:"warehouse_code"`
	Quantity      int    `json:"quantity"`
}

var (
	skuPattern           = regexp.MustCompile(`^[A-Z0-9-]{3,20}$`)
	warehouseCodePattern = regexp.MustCompile(`^[A-Z]{3}-[0-9]{2}$`)
	validUnits           = map[string]bool{"ud": true, "kg": true, "l": true, "caja": true}
)

// Normalize trims the fields and fills defaults.
func (p *Product) Normalize() {
	p.SKU = strings.ToUpper(strings.TrimSpace(p.SKU))
	p.Name = strings.TrimSpace(p.Name)
	if p.Unit == "" {
		p.Unit = "ud"
	}
}

// Validate checks a product before it is stored.
func (p *Product) Validate() error {
	switch {
	case !skuPattern.MatchString(p.SKU):
		return invalid("sku", "must be 3-20 uppercase letters, digits or dashes")
	case p.Name == "":
		return invalid("name", "is required")
	case !validUnits[p.Unit]:
		return invalid("unit", "must be one of ud, kg, l, caja")
	case p.MinStock < 0:
		return invalid("min_stock", "cannot be negative")
	}
	return nil
}

// Normalize trims the fields of a warehouse.
func (w *Warehouse) Normalize() {
	w.Code = strings.ToUpper(strings.TrimSpace(w.Code))
	w.Name = strings.TrimSpace(w.Name)
	w.City = strings.TrimSpace(w.City)
}

// Validate checks a warehouse before it is stored.
func (w *Warehouse) Validate() error {
	switch {
	case !warehouseCodePattern.MatchString(w.Code):
		return invalid("code", "must look like MAD-01")
	case w.Name == "":
		return invalid("name", "is required")
	}
	return nil
}

// Validate checks a movement request.
func (m *Movement) Validate() error {
	switch m.Kind {
	case KindIn, KindOut:
		if m.Quantity <= 0 {
			return invalid("quantity", "must be greater than zero")
		}
	case KindAdjust:
		if m.Quantity < 0 {
			return invalid("quantity", "cannot be negative")
		}
	default:
		return invalid("kind", "must be one of in, out, adjust")
	}
	if m.ProductID <= 0 {
		return invalid("product_id", "is required")
	}
	if m.WarehouseID <= 0 {
		return invalid("warehouse_id", "is required")
	}
	return nil
}

// ApplyMovement returns the stock that results from applying a movement to the
// current quantity.
func ApplyMovement(current int, kind MovementKind, quantity int) (int, error) {
	switch kind {
	case KindIn:
		return current + quantity, nil
	case KindOut:
		if quantity > current {
			return current, ErrInsufficientStock
		}
		return current - quantity, nil
	case KindAdjust:
		return quantity, nil
	}
	return current, invalid("kind", "unknown movement kind")
}
