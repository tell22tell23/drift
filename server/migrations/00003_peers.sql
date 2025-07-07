-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE peers {
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    peer_id TEXT NOT NULL,
}
-- +goose StatementEnd
-- +goose Down

-- +goose StatementBegin
DROP TABLE IF EXISTS peers;
-- +goose StatementEnd
