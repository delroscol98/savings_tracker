-- name: CreateGoal :one
INSERT INTO goals (id, target, deadline, created_at, updated_at, user_id)
VALUES (
    get_random_uuid(),
    $1,
    $2,
    NOW(),
    NOW(),
    $3
)
RETURNING *;
