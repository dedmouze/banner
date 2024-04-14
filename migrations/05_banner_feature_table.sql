-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS banner_feature
(
    banner_id INTEGER,
    feature_id INTEGER,
    PRIMARY KEY(banner_id, feature_id),
    FOREIGN KEY(banner_id) REFERENCES banner(id),
    FOREIGN KEY(feature_id) REFERENCES feature(id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE banner_feature;
-- +goose StatementEnd
