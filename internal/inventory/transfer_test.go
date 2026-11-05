package inventory

import "testing"

func TestTransferValidate(t *testing.T) {
	ok := Transfer{ProductID: 1, FromWarehouseID: 1, ToWarehouseID: 2, Quantity: 5}
	if err := ok.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, tr := range []Transfer{
		{ProductID: 1, FromWarehouseID: 1, ToWarehouseID: 1, Quantity: 5},
		{ProductID: 1, FromWarehouseID: 1, ToWarehouseID: 2, Quantity: 0},
		{ProductID: 0, FromWarehouseID: 1, ToWarehouseID: 2, Quantity: 5},
		{ProductID: 1, FromWarehouseID: 0, ToWarehouseID: 2, Quantity: 5},
	} {
		if err := tr.Validate(); err == nil {
			t.Errorf("expected error for %+v", tr)
		}
	}
}
