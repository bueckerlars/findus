-- +goose Up
-- +goose StatementBegin
CREATE TABLE item_qr_token_reservations (
    item_id TEXT PRIMARY KEY,
    qr_token TEXT NOT NULL UNIQUE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS item_qr_token_reservations;
-- +goose StatementEnd
