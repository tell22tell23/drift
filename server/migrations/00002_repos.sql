-- +goose Up
-- +goose StatementBegin
CREATE EXTENSTION IF NOT EXISTS "pgcrypto";

CREATE TABLE repos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    url TEXT NOT NULL,
    owner_id UUID REFERENCES users(id) ON DELETE CASCADE,
    access_user_ids UUID[] DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX idx_repos_name ON repos(name);
-- +goose StatementEnd
-- +goose Down

-- +goose StatementBegin
DROP TABLE IF EXISTS repos;
-- +goose StatementEnd
