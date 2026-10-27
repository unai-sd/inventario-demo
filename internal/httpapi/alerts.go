package httpapi

import "net/http"

func (s *server) lowStock(w http.ResponseWriter, r *http.Request) {
	items, err := s.store.LowStock(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}
