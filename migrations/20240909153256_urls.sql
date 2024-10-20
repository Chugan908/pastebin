-- +goose Up
-- +goose StatementBegin
CREATE TABLE text_urls(
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    hashed_url TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS text_urls;
-- +goose StatementEnd
