-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS "orders" (
    id bigserial NOT NULL PRIMARY KEY,
    user_id bigint NOT NULL,
    number text NOT NULL UNIQUE,
    status text NOT NULL,
    accrual numeric DEFAULT 0,
    uploaded_at timestamptz NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS "orders";
-- +goose StatementEnd