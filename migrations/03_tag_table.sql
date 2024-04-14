-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS tag
(
    id INTEGER PRIMARY KEY,
    created_at TIMESTAMP,
    used_at TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE tag;
-- +goose StatementEnd
