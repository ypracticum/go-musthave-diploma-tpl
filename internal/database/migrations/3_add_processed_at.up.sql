ALTER TABLE accrual_flow
ADD COLUMN processed_at timestamp NOT NULL DEFAULT current_timestamp;
