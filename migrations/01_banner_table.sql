-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS banner
(
    id SERIAL PRIMARY KEY,
    content TEXT UNIQUE,
    is_active BOOLEAN,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_content ON banner(content);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE banner;
-- +goose StatementEnd
