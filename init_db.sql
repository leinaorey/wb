-- Создание базы данных
CREATE DATABASE orders;

-- Удаление таблиц при необходимости (для повторного запуска)
DROP TABLE IF EXISTS order_items CASCADE;
DROP TABLE IF EXISTS orders CASCADE;

-- Основная таблица заказов
CREATE TABLE orders (
    id BIGSERIAL PRIMARY KEY,
    order_uid TEXT UNIQUE NOT NULL,
    track_number TEXT NOT NULL,
    entry TEXT,
    -- delivery
    delivery_name TEXT,
    delivery_phone TEXT,
    delivery_zip TEXT,
    delivery_city TEXT,
    delivery_address TEXT,
    delivery_region TEXT,
    delivery_email TEXT,
    -- payment
    payment_transaction TEXT,
    payment_request_id TEXT,
    payment_currency CHAR(3),
    payment_provider TEXT,
    payment_amount INTEGER,         
    payment_dt TIMESTAMPTZ,
    payment_bank TEXT,
    delivery_cost INTEGER,
    goods_total INTEGER,
    custom_fee INTEGER,
    locale TEXT,
    internal_signature TEXT,
    customer_id TEXT,
    delivery_service TEXT,
    shardkey TEXT,
    sm_id INTEGER,
    date_created TIMESTAMPTZ,
    oof_shard TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Таблица позиций заказа
CREATE TABLE order_items (
    id BIGSERIAL PRIMARY KEY,
    order_uid TEXT NOT NULL REFERENCES orders(order_uid) ON DELETE CASCADE,
    chrt_id BIGINT,
    track_number TEXT,
    price INTEGER,
    rid TEXT,
    name TEXT,
    sale INTEGER,
    size TEXT,
    total_price INTEGER,
    nm_id BIGINT,
    brand TEXT,
    status INTEGER,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Индексы
CREATE INDEX idx_orders_order_uid ON orders(order_uid);
CREATE INDEX idx_orders_track_number ON orders(track_number);
CREATE INDEX idx_orders_customer_id ON orders(customer_id);
CREATE INDEX idx_orders_date_created ON orders(date_created);
CREATE INDEX idx_order_items_order_uid ON order_items(order_uid);
CREATE INDEX idx_order_items_nm_id ON order_items(nm_id);

-- Триггер для updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at = NOW();
   RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_orders_updated_at
    BEFORE UPDATE ON orders
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

