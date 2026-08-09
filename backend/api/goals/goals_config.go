package goals

import (
	"context"

	"github.com/delroscol98/savings_tracker/backend/internal/database"
	"github.com/google/uuid"
)

type Queries interface {
	GetGoals(ctx context.Context, userID uuid.UUID) ([]database.Goal, error)
	CreateGoal(ctx context.Context, arg database.CreateGoalParams) (database.Goal, error)
	GetGoalById(ctx context.Context, id uuid.UUID) (database.Goal, error)
	UpdateGoal(ctx context.Context, arg database.UpdateGoalParams) (database.Goal, error)
	DeleteGoal(ctx context.Context, arg database.DeleteGoalParams) error

	CreateDeposit(ctx context.Context, arg database.CreateDepositParams) (database.Deposit, error)
	GetDepositsByGoalAndUser(ctx context.Context, arg database.GetDepositsByGoalAndUserParams) ([]database.Deposit, error)
}

type GoalsConfig struct {
	Queries Queries
}
