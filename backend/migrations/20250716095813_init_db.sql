-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS orders (
    order_uid UUID PRIMARY KEY,
    track_number VARCHAR(32) NOT NULL,
    entry VARCHAR(10),
    locale VARCHAR(5),
    internal_signature TEXT,
    customer_id VARCHAR(50),
    delivery_service VARCHAR(50),
    shardkey VARCHAR(5),
    sm_id INT,
    date_created TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    oof_shard VARCHAR(5)
);

CREATE TABLE IF NOT EXISTS delivery (
    order_uid UUID PRIMARY KEY REFERENCES orders(order_uid) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    phone VARCHAR(15) NOT NULL,
    zip VARCHAR(10) NOT NULL,
    city VARCHAR(100) NOT NULL,
    address TEXT NOT NULL,
    region VARCHAR(100) NOT NULL,
    email VARCHAR(100) NOT NULL
);

CREATE TABLE IF NOT EXISTS payment (
    transaction TEXT PRIMARY KEY,
    order_uid UUID NOT NULL REFERENCES orders(order_uid) ON DELETE CASCADE,
    request_id TEXT,
    currency VARCHAR(5) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    amount INT NOT NULL,
    payment_dt BIGINT NOT NULL,
    bank VARCHAR(50) NOT NULL,
    delivery_cost INT NOT NULL,
    goods_total INT NOT NULL,
    custom_fee INT NOT NULL
);

CREATE TABLE IF NOT EXISTS items (
    id SERIAL PRIMARY KEY,
    order_id UUID NOT NULL REFERENCES orders(order_uid) ON DELETE CASCADE,
    chrt_id BIGINT NOT NULL,
    track_number VARCHAR(32) NOT NULL,
    price INT NOT NULL,
    rid TEXT NOT NULL,
    name VARCHAR(100) NOT NULL,
    sale INT NOT NULL,
    size VARCHAR(10),
    total_price INT NOT NULL,
    nm_id BIGINT NOT NULL,
    brand VARCHAR(100) NOT NULL,
    status INT NOT NULL
);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS orders;

DROP TABLE IF EXISTS delivery;

DROP TABLE IF EXISTS payment;

DROP TABLE IF EXISTS items;

-- +goose StatementEnd