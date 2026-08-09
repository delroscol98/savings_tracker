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

-- name: GetDepositsByGoalAndUser :many
SELECT * FROM deposits
WHERE goal_id = $1 AND user_id = $2;
