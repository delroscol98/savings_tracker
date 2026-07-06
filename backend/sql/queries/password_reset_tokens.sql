-- name: CreatePasswordResetToken :one
INSERT INTO password_reset_tokens(id, user_id, token_hash, created_at, expires_at)
VALUES (
    GEN_RANDOM_UUID(),
    $1,
    $2,
    NOW(),
    NOW() + INTERVAL '30 minutes'
)
RETURNING *;

-- name: GetPasswordResetTokenByHash :one
SELECT * FROM password_reset_tokens
WHERE token_hash = $1;

-- name: ConsumePasswordResetToken :exec
UPDATE password_reset_tokens
SET consumed_at = NOW()
WHERE id = $1;

-- name: DeactivateUserTokens :exec
UPDATE password_reset_tokens
SET consumed_at = NOW()
WHERE user_id = $1 AND consumed_at IS NULL;
