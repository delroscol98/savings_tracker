-- name: Ping :one
SELECT 1;

-- name: CreateUser :one
INSERT INTO users(id, created_at, updated_at, email, hashed_password, full_name)
VALUES(
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2,
    $3
)
RETURNING *;

-- name: Login :one
SELECT id, created_at, updated_at, email, hashed_password, full_name FROM users
WHERE $1 = email;
