CREATE SCHEMA IF NOT EXISTS portaldb;

CREATE TABLE IF NOT EXISTS portaldb.categories (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(120) NOT NULL UNIQUE,
    description VARCHAR(255) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS portaldb.products (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(160) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    sku VARCHAR(80) NOT NULL UNIQUE,
    price NUMERIC(10,2) NOT NULL,
    stock INTEGER NOT NULL,
    category_id BIGINT NOT NULL REFERENCES portaldb.categories(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS portaldb.customers (
    id BIGSERIAL PRIMARY KEY,
    first_name VARCHAR(80) NOT NULL,
    last_name VARCHAR(80) NOT NULL,
    email VARCHAR(160) NOT NULL UNIQUE,
    phone VARCHAR(30) NOT NULL DEFAULT '',
    address TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS portaldb.orders (
    id BIGSERIAL PRIMARY KEY,
    customer_id BIGINT NOT NULL REFERENCES portaldb.customers(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    total_amount NUMERIC(10,2) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS portaldb.order_items (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL REFERENCES portaldb.orders(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    product_id BIGINT NOT NULL REFERENCES portaldb.products(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    quantity INTEGER NOT NULL,
    unit_price NUMERIC(10,2) NOT NULL,
    line_total NUMERIC(10,2) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_products_category_id ON portaldb.products (category_id);
CREATE INDEX IF NOT EXISTS idx_orders_customer_id ON portaldb.orders (customer_id);
CREATE INDEX IF NOT EXISTS idx_orders_status ON portaldb.orders (status);
CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON portaldb.order_items (order_id);
CREATE INDEX IF NOT EXISTS idx_order_items_product_id ON portaldb.order_items (product_id);
