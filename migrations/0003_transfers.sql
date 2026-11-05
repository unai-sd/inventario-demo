CREATE TABLE transfers (
    id                BIGSERIAL PRIMARY KEY,
    product_id        BIGINT      NOT NULL REFERENCES products (id),
    from_warehouse_id BIGINT      NOT NULL REFERENCES warehouses (id),
    to_warehouse_id   BIGINT      NOT NULL REFERENCES warehouses (id),
    quantity          INTEGER     NOT NULL CHECK (quantity > 0),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (from_warehouse_id <> to_warehouse_id)
);
