package httpapi

import (
	"net/http"

	"distribuciones-demo.example/inventario/internal/inventory"
)

func (s *server) createTransfer(w http.ResponseWriter, r *http.Request) {
	var t inventory.Transfer
	if err := decode(w, r, &t); err != nil {
		writeError(w, err)
		return
	}
	if err := t.Validate(); err != nil {
		writeError(w, err)
		return
	}
	created, err := s.store.Transfer(r.Context(), t)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}
