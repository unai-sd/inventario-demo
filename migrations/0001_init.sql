CREATE TABLE products (
    id         BIGSERIAL PRIMARY KEY,
    sku        TEXT        NOT NULL UNIQUE,
    name       TEXT        NOT NULL,
    unit       TEXT        NOT NULL DEFAULT 'ud',
    min_stock  INTEGER     NOT NULL DEFAULT 0 CHECK (min_stock >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE warehouses (
    id   BIGSERIAL PRIMARY KEY,
    code TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    city TEXT NOT NULL DEFAULT ''
);

CREATE TABLE stock (
    product_id   BIGINT  NOT NULL REFERENCES products (id),
    warehouse_id BIGINT  NOT NULL REFERENCES warehouses (id),
    quantity     INTEGER NOT NULL DEFAULT 0 CHECK (quantity >= 0),
    PRIMARY KEY (product_id, warehouse_id)
);

CREATE TABLE movements (
    id           BIGSERIAL PRIMARY KEY,
    product_id   BIGINT      NOT NULL REFERENCES products (id),
    warehouse_id BIGINT      NOT NULL REFERENCES warehouses (id),
    kind         TEXT        NOT NULL CHECK (kind IN ('in', 'out', 'adjust')),
    quantity     INTEGER     NOT NULL CHECK (quantity >= 0),
    reason       TEXT        NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX movements_product_idx ON movements (product_id, created_at);
