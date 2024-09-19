CREATE TYPE order_status AS ENUM ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED');

CREATE TABLE orders (
    id          text PRIMARY KEY,
    user_id     uuid REFERENCES users NOT NULL,
    status      order_status NOT NULL DEFAULT 'NEW',
    uploaded_at timestamp NOT NULL DEFAULT current_timestamp
);

CREATE TABLE accrual_flow (
    id       uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id text REFERENCES orders NOT NULL,
    amount   numeric(15, 2) NOT NULL DEFAULT 0
);
