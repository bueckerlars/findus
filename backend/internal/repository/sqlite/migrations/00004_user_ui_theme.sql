-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ADD COLUMN ui_theme TEXT NOT NULL DEFAULT 'default';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 1;
-- +goose StatementEnd
