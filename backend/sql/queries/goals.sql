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

-- name: UpdateGoal :one
UPDATE goals
SET target = $1, deadline = $2, updated_at = NOW()
WHERE id = $3 AND user_id = $4
RETURNING *;

-- name: DeleteGoal :exec
DELETE FROM goals
WHERE id = $1 AND user_id = $2;

-- name: GetGoals :many
SELECT 
    goals.id,
    goals.target,
    goals.deadline,
    goals.created_at, 
    goals.updated_at,
    goals.user_id,
    COALESCE(SUM(deposits.amount), 0)::bigint as progress
FROM goals
LEFT JOIN deposits ON goals.id = deposits.goal_id
WHERE goals.user_id = $1
GROUP BY goals.id, goals.target, goals.deadline, goals.created_at, goals.updated_at, goals.user_id;
