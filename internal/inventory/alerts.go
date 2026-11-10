package inventory

// LowStockItem is a product whose total stock across warehouses is at or
// below its configured minimum.
type LowStockItem struct {
	ProductID int64  `json:"product_id"`
	SKU       string `json:"sku"`
	Name      string `json:"name"`
	MinStock  int    `json:"min_stock"`
	Total     int    `json:"total"`
	Missing   int    `json:"missing"`
}

// IsLowStock reports whether a product must raise a low-stock alert. Products
// with a minimum of zero never raise alerts.
func IsLowStock(total, minStock int) bool {
	return minStock > 0 && total <= minStock
}

// NewLowStockItem builds the alert line for a product.
func NewLowStockItem(p Product, total int) LowStockItem {
	return LowStockItem{
		ProductID: p.ID,
		SKU:       p.SKU,
		Name:      p.Name,
		MinStock:  p.MinStock,
		Total:     total,
		Missing:   max(p.MinStock-total, 0),
	}
}
