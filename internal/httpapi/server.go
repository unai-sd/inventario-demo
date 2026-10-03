// Package httpapi exposes the inventory over HTTP using the standard library
// router (method and wildcard patterns, Go 1.22+).
package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"distribuciones-demo.example/inventario/internal/inventory"
)

// Store is the persistence the handlers need.
type Store interface {
	CreateProduct(ctx context.Context, p inventory.Product) (inventory.Product, error)
	ListProducts(ctx context.Context) ([]inventory.Product, error)
	GetProduct(ctx context.Context, id int64) (inventory.Product, error)
	CreateWarehouse(ctx context.Context, w inventory.Warehouse) (inventory.Warehouse, error)
	ListWarehouses(ctx context.Context) ([]inventory.Warehouse, error)
	RecordMovement(ctx context.Context, m inventory.Movement) (inventory.Movement, int, error)
	ListMovements(ctx context.Context, productID int64) ([]inventory.Movement, error)
	StockLevels(ctx context.Context, warehouseID int64) ([]inventory.StockLevel, error)
}

type server struct {
	store   Store
	version string
}

// NewHandler returns the HTTP handler of the API.
func NewHandler(store Store, version string) http.Handler {
	s := &server{store: store, version: version}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.health)
	mux.HandleFunc("GET /products", s.listProducts)
	mux.HandleFunc("POST /products", s.createProduct)
	mux.HandleFunc("GET /products/{id}", s.getProduct)
	mux.HandleFunc("GET /warehouses", s.listWarehouses)
	mux.HandleFunc("POST /warehouses", s.createWarehouse)
	mux.HandleFunc("GET /movements", s.listMovements)
	mux.HandleFunc("POST /movements", s.createMovement)
	mux.HandleFunc("GET /stock", s.stockLevels)
	return mux
}

func (s *server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "version": s.version})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, err error) {
	var verr *inventory.ValidationError
	switch {
	case errors.As(err, &verr):
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": verr.Error(), "field": verr.Field})
	case errors.Is(err, inventory.ErrNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
	case errors.Is(err, inventory.ErrConflict):
		writeJSON(w, http.StatusConflict, map[string]string{"error": "already exists"})
	case errors.Is(err, inventory.ErrInsufficientStock):
		writeJSON(w, http.StatusConflict, map[string]string{"error": "insufficient stock"})
	default:
		slog.Error("request failed", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
	}
}

func decode(w http.ResponseWriter, r *http.Request, dst any) error {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return &inventory.ValidationError{Field: "body", Message: "invalid JSON: " + err.Error()}
	}
	return nil
}

// queryID parses an optional numeric query parameter (0 when absent).
func queryID(r *http.Request, name string) (int64, error) {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		return 0, nil
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id < 0 {
		return 0, &inventory.ValidationError{Field: name, Message: "must be a positive integer"}
	}
	return id, nil
}
