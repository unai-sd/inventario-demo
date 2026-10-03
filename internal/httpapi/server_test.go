package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func do(t *testing.T, h http.Handler, method, path, body string) (*httptest.ResponseRecorder, map[string]any) {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	var out map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &out)
	return rec, out
}

func seeded(t *testing.T) (http.Handler, *fakeStore) {
	t.Helper()
	store := newFakeStore()
	h := NewHandler(store, "test")
	if rec, _ := do(t, h, "POST", "/products", `{"sku":"ace-500","name":"Aceite 500 ml","min_stock":10}`); rec.Code != http.StatusCreated {
		t.Fatalf("create product: %d %s", rec.Code, rec.Body)
	}
	if rec, _ := do(t, h, "POST", "/warehouses", `{"code":"MAD-01","name":"Central","city":"Madrid"}`); rec.Code != http.StatusCreated {
		t.Fatalf("create warehouse: %d %s", rec.Code, rec.Body)
	}
	if rec, _ := do(t, h, "POST", "/warehouses", `{"code":"VLC-01","name":"Levante","city":"Valencia"}`); rec.Code != http.StatusCreated {
		t.Fatalf("create warehouse: %d %s", rec.Code, rec.Body)
	}
	return h, store
}

func TestHealth(t *testing.T) {
	rec, body := do(t, NewHandler(newFakeStore(), "1.2.3"), "GET", "/healthz", "")
	if rec.Code != http.StatusOK || body["version"] != "1.2.3" {
		t.Fatalf("unexpected health: %d %v", rec.Code, body)
	}
}

func TestProducts(t *testing.T) {
	h, _ := seeded(t)
	rec, body := do(t, h, "GET", "/products/1", "")
	if rec.Code != http.StatusOK || body["sku"] != "ACE-500" || body["unit"] != "ud" {
		t.Fatalf("unexpected product: %d %v", rec.Code, body)
	}
	if rec, _ := do(t, h, "POST", "/products", `{"sku":"ACE-500","name":"Otro"}`); rec.Code != http.StatusConflict {
		t.Fatalf("duplicate sku: want 409, got %d", rec.Code)
	}
	if rec, body := do(t, h, "POST", "/products", `{"sku":"x","name":"Otro"}`); rec.Code != http.StatusUnprocessableEntity || body["field"] != "sku" {
		t.Fatalf("invalid sku: got %d %v", rec.Code, body)
	}
	if rec, _ := do(t, h, "POST", "/products", `{"sku":"ABC-1","name":"x","color":"red"}`); rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("unknown field: got %d", rec.Code)
	}
	if rec, _ := do(t, h, "GET", "/products/999", ""); rec.Code != http.StatusNotFound {
		t.Fatalf("missing product: got %d", rec.Code)
	}
	if rec, _ := do(t, h, "DELETE", "/products/1", ""); rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("delete: want 405, got %d", rec.Code)
	}
}

func TestMovementsUpdateStock(t *testing.T) {
	h, store := seeded(t)
	rec, body := do(t, h, "POST", "/movements", `{"product_id":1,"warehouse_id":2,"kind":"in","quantity":30,"reason":"albarán 1001"}`)
	if rec.Code != http.StatusCreated || body["stock"] != float64(30) {
		t.Fatalf("entry: %d %v", rec.Code, body)
	}
	rec, body = do(t, h, "POST", "/movements", `{"product_id":1,"warehouse_id":2,"kind":"out","quantity":12}`)
	if rec.Code != http.StatusCreated || body["stock"] != float64(18) {
		t.Fatalf("exit: %d %v", rec.Code, body)
	}
	if rec, _ := do(t, h, "POST", "/movements", `{"product_id":1,"warehouse_id":2,"kind":"out","quantity":19}`); rec.Code != http.StatusConflict {
		t.Fatalf("insufficient stock: want 409, got %d", rec.Code)
	}
	if rec, _ := do(t, h, "POST", "/movements", `{"product_id":1,"warehouse_id":99,"kind":"in","quantity":1}`); rec.Code != http.StatusNotFound {
		t.Fatalf("unknown warehouse: want 404, got %d", rec.Code)
	}
	if len(store.movements) != 2 {
		t.Fatalf("want 2 movements, got %d", len(store.movements))
	}
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/stock?warehouse_id=2", nil))
	if !strings.Contains(rec.Body.String(), `"quantity":18`) {
		t.Fatalf("stock levels: %s", rec.Body)
	}
	if rec, _ := do(t, h, "GET", "/movements?product_id=abc", ""); rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("bad filter: want 422, got %d", rec.Code)
	}
}
