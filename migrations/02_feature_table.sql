-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS feature
(
    id INTEGER PRIMARY KEY,
    created_at TIMESTAMP,
    used_at TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE feature;
-- +goose StatementEnd
