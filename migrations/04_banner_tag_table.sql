-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS banner_tag
(
    banner_id INTEGER,
    tag_id INTEGER,
    PRIMARY KEY(banner_id, tag_id),
    FOREIGN KEY(banner_id) REFERENCES banner(id),
    FOREIGN KEY(tag_id) REFERENCES tag(id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE banner_tag;
-- +goose StatementEnd
