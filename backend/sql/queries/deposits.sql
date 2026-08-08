-- name: CreateDeposit :one
INSERT INTO deposits (id, amount, note, created_at, goal_id, user_id)
VALUES (
    gen_random_uuid(),
    $1,
    $2,
    NOW(),
    $3,
    $4
)
RETURNING *;
