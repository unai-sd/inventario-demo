-- Demo data for Distribuciones Demo. Idempotent: safe to run several times.
-- Usage: make seed (after make db-up migrate).

INSERT INTO warehouses (code, name, city) VALUES
    ('MAD-01', 'Almacén central', 'Madrid'),
    ('VLC-01', 'Almacén Levante', 'Valencia'),
    ('SEV-01', 'Almacén Sur', 'Sevilla')
ON CONFLICT (code) DO NOTHING;

INSERT INTO products (sku, name, unit, min_stock) VALUES
    ('ACE-500', 'Aceite de oliva 500 ml', 'ud', 40),
    ('ARZ-1K', 'Arroz redondo 1 kg', 'ud', 60),
    ('HAR-1K', 'Harina de trigo 1 kg', 'ud', 30),
    ('LEG-GAR', 'Garbanzos cocidos', 'caja', 10),
    ('AGU-15L', 'Agua mineral 1,5 l', 'caja', 25),
    ('CAF-250', 'Café molido 250 g', 'ud', 20)
ON CONFLICT (sku) DO NOTHING;

-- Opening stock, recorded as incoming movements only the first time.
WITH opening (sku, code, quantity) AS (
    VALUES ('ACE-500', 'MAD-01', 120), ('ACE-500', 'VLC-01', 35),
           ('ARZ-1K', 'MAD-01', 45), ('ARZ-1K', 'SEV-01', 10),
           ('HAR-1K', 'MAD-01', 80),
           ('LEG-GAR', 'VLC-01', 6),
           ('AGU-15L', 'MAD-01', 40), ('AGU-15L', 'SEV-01', 30),
           ('CAF-250', 'VLC-01', 20)
), rows AS (
    SELECT p.id AS product_id, w.id AS warehouse_id, o.quantity
    FROM opening o
    JOIN products p ON p.sku = o.sku
    JOIN warehouses w ON w.code = o.code
), inserted AS (
    INSERT INTO stock (product_id, warehouse_id, quantity)
    SELECT product_id, warehouse_id, quantity FROM rows
    ON CONFLICT DO NOTHING
    RETURNING product_id, warehouse_id, quantity
)
INSERT INTO movements (product_id, warehouse_id, kind, quantity, reason)
SELECT product_id, warehouse_id, 'in', quantity, 'carga inicial de demostración' FROM inserted;
