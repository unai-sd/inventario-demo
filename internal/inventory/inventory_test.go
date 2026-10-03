package inventory

import (
	"errors"
	"testing"
)

func TestProductValidate(t *testing.T) {
	cases := []struct {
		name    string
		product Product
		field   string
	}{
		{"valid", Product{SKU: " ace-500 ", Name: "Aceite 500 ml"}, ""},
		{"bad sku", Product{SKU: "a", Name: "Aceite"}, "sku"},
		{"no name", Product{SKU: "ACE-500"}, "name"},
		{"bad unit", Product{SKU: "ACE-500", Name: "Aceite", Unit: "barril"}, "unit"},
		{"negative min", Product{SKU: "ACE-500", Name: "Aceite", MinStock: -1}, "min_stock"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := tc.product
			p.Normalize()
			err := p.Validate()
			if tc.field == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if p.SKU != "ACE-500" || p.Unit != "ud" {
					t.Fatalf("not normalized: %+v", p)
				}
				return
			}
			var verr *ValidationError
			if !errors.As(err, &verr) || verr.Field != tc.field {
				t.Fatalf("want validation error on %q, got %v", tc.field, err)
			}
		})
	}
}

func TestWarehouseValidate(t *testing.T) {
	w := Warehouse{Code: "mad-01", Name: "Central Madrid"}
	w.Normalize()
	if err := w.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	bad := Warehouse{Code: "MADRID", Name: "x"}
	if err := bad.Validate(); err == nil {
		t.Fatal("expected error for bad code")
	}
}

func TestMovementValidate(t *testing.T) {
	ok := Movement{ProductID: 1, WarehouseID: 1, Kind: KindAdjust, Quantity: 0}
	if err := ok.Validate(); err != nil {
		t.Fatalf("adjust to zero should be valid: %v", err)
	}
	for _, m := range []Movement{
		{ProductID: 1, WarehouseID: 1, Kind: KindIn, Quantity: 0},
		{ProductID: 1, WarehouseID: 1, Kind: "gift", Quantity: 1},
		{ProductID: 0, WarehouseID: 1, Kind: KindOut, Quantity: 1},
		{ProductID: 1, WarehouseID: 0, Kind: KindOut, Quantity: 1},
	} {
		if err := m.Validate(); err == nil {
			t.Errorf("expected error for %+v", m)
		}
	}
}

func TestApplyMovement(t *testing.T) {
	cases := []struct {
		current, qty, want int
		kind               MovementKind
		err                error
	}{
		{10, 5, 15, KindIn, nil},
		{10, 4, 6, KindOut, nil},
		{10, 10, 0, KindOut, nil},
		{10, 11, 10, KindOut, ErrInsufficientStock},
		{10, 3, 3, KindAdjust, nil},
	}
	for _, tc := range cases {
		got, err := ApplyMovement(tc.current, tc.kind, tc.qty)
		if !errors.Is(err, tc.err) || got != tc.want {
			t.Errorf("ApplyMovement(%d, %s, %d) = %d, %v; want %d, %v",
				tc.current, tc.kind, tc.qty, got, err, tc.want, tc.err)
		}
	}
}
