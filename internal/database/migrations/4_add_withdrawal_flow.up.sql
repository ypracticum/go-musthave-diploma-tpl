CREATE TABLE withdrawal_flow (
    id           uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id     text NOT NULL,
    user_id      uuid REFERENCES users NOT NULL,
    amount       numeric(15, 2) NOT NULL DEFAULT 0,
    processed_at timestamp NOT NULL DEFAULT current_timestamp
);
