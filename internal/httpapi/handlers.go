package httpapi

import (
	"net/http"
	"strconv"

	"distribuciones-demo.example/inventario/internal/inventory"
)

func (s *server) listProducts(w http.ResponseWriter, r *http.Request) {
	products, err := s.store.ListProducts(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, products)
}

func (s *server) createProduct(w http.ResponseWriter, r *http.Request) {
	var p inventory.Product
	if err := decode(w, r, &p); err != nil {
		writeError(w, err)
		return
	}
	p.Normalize()
	if err := p.Validate(); err != nil {
		writeError(w, err)
		return
	}
	created, err := s.store.CreateProduct(r.Context(), p)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *server) getProduct(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, inventory.ErrNotFound)
		return
	}
	p, err := s.store.GetProduct(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (s *server) listWarehouses(w http.ResponseWriter, r *http.Request) {
	warehouses, err := s.store.ListWarehouses(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, warehouses)
}

func (s *server) createWarehouse(w http.ResponseWriter, r *http.Request) {
	var wh inventory.Warehouse
	if err := decode(w, r, &wh); err != nil {
		writeError(w, err)
		return
	}
	wh.Normalize()
	if err := wh.Validate(); err != nil {
		writeError(w, err)
		return
	}
	created, err := s.store.CreateWarehouse(r.Context(), wh)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *server) listMovements(w http.ResponseWriter, r *http.Request) {
	productID, err := queryID(r, "product_id")
	if err != nil {
		writeError(w, err)
		return
	}
	movements, err := s.store.ListMovements(r.Context(), productID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, movements)
}

type movementResponse struct {
	Movement inventory.Movement `json:"movement"`
	Stock    int                `json:"stock"`
}

func (s *server) createMovement(w http.ResponseWriter, r *http.Request) {
	var m inventory.Movement
	if err := decode(w, r, &m); err != nil {
		writeError(w, err)
		return
	}
	if err := m.Validate(); err != nil {
		writeError(w, err)
		return
	}
	created, stock, err := s.store.RecordMovement(r.Context(), m)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, movementResponse{Movement: created, Stock: stock})
}

func (s *server) stockLevels(w http.ResponseWriter, r *http.Request) {
	warehouseID, err := queryID(r, "warehouse_id")
	if err != nil {
		writeError(w, err)
		return
	}
	levels, err := s.store.StockLevels(r.Context(), warehouseID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, levels)
}
