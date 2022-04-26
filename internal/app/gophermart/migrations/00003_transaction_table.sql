-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS "transactions" (
    id bigserial NOT NULL PRIMARY KEY,
    login text NOT NULL,
    number bigserial NOT NULL,
    amount numeric DEFAULT 0,
    processed_at timestamptz NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS "transactions";
-- +goose StatementEnd