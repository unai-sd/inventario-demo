-- Total stock of every product across all warehouses, used by low-stock alerts.
CREATE VIEW product_stock_totals AS
SELECT p.id,
       p.sku,
       p.name,
       p.unit,
       p.min_stock,
       p.created_at,
       COALESCE(SUM(s.quantity), 0)::INTEGER AS total
FROM products p
LEFT JOIN stock s ON s.product_id = p.id
GROUP BY p.id;
