-- +goose Up
CREATE TABLE IF NOT EXISTS password_reset_tokens(
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    token_hash TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    consumed_at TIMESTAMP,

    FOREIGN KEY (user_id) REFERENCES users(id)
);

-- +goose Down
DROP TABLE password_reset_tokens;
    
