-- name: CreateGoal :one
INSERT INTO goals (id, target, deadline, created_at, updated_at, user_id)
VALUES (
    gen_random_uuid(),
    $1,
    $2,
    NOW(),
    NOW(),
    $3
)
RETURNING *;

-- name: GetGoalById :one
SELECT * FROM goals
WHERE id = $1;

-- name: DeleteGoal :exec
DELETE FROM goals
WHERE id = $1;
