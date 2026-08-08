package goals_test

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"

	"github.com/delroscol98/savings_tracker/backend/internal/database"
)

type mockDB struct {
	GetGoalsErr      error
	CreateGoalErr    error
	GetGoalByIdErr   error
	UpdateGoalErr    error
	DeleteGoalErr    error
	CreateDepositErr error

	Users    map[string]database.User
	Goals    map[string]database.Goal
	Deposits map[string]database.Deposit
}

func (m *mockDB) GetGoals(ctx context.Context, userID uuid.UUID) ([]database.Goal, error) {
	if m.GetGoalsErr != nil {
		return nil, m.GetGoalsErr
	}

	if m.Goals == nil {
		m.Goals = make(map[string]database.Goal)
	}

	goals := make([]database.Goal, 0, len(m.Goals))
	for _, goal := range m.Goals {
		if userID == goal.UserID {
			goals = append(goals, goal)
		}
	}

	return goals, nil
}

func (m *mockDB) CreateGoal(ctx context.Context, params database.CreateGoalParams) (database.Goal, error) {
	if m.CreateGoalErr != nil {
		return database.Goal{}, m.CreateGoalErr
	}

	if m.Goals == nil {
		m.Goals = make(map[string]database.Goal)
	}

	goalId := uuid.New()
	goal := database.Goal{
		ID:        goalId,
		Target:    params.Target,
		Deadline:  params.Deadline,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    params.UserID,
	}
	m.Goals[goalId.String()] = goal

	return goal, nil
}

func (m *mockDB) GetGoalById(ctx context.Context, id uuid.UUID) (database.Goal, error) {
	if m.GetGoalByIdErr != nil {
		return database.Goal{}, m.GetGoalByIdErr
	}

	if m.Goals == nil {
		m.Goals = make(map[string]database.Goal)
	}

	goal, ok := m.Goals[id.String()]
	if !ok {
		return database.Goal{}, sql.ErrNoRows
	}

	return goal, nil
}

func (m *mockDB) UpdateGoal(ctx context.Context, arg database.UpdateGoalParams) (database.Goal, error) {
	if m.UpdateGoalErr != nil {
		return database.Goal{}, m.UpdateGoalErr
	}

	if m.Goals == nil {
		m.Goals = make(map[string]database.Goal)
	}

	goal := m.Goals[arg.ID.String()]
	goal.Target = arg.Target
	goal.Deadline = arg.Deadline
	goal.UpdatedAt = time.Now()

	return goal, nil
}

func (m *mockDB) DeleteGoal(ctx context.Context, arg database.DeleteGoalParams) error {
	if m.DeleteGoalErr != nil {
		return m.DeleteGoalErr
	}

	if m.Goals == nil {
		m.Goals = make(map[string]database.Goal)
	}

	delete(m.Goals, arg.ID.String())

	return nil
}

func (m *mockDB) CreateDeposit(ctx context.Context, arg database.CreateDepositParams) (database.Deposit, error) {
	if m.CreateDepositErr != nil {
		return database.Deposit{}, m.CreateDepositErr
	}

	if m.Deposits == nil {
		m.Deposits = make(map[string]database.Deposit)
	}

	depositId := uuid.New()
	deposit := database.Deposit{
		ID:        depositId,
		Amount:    arg.Amount,
		Note:      arg.Note,
		CreatedAt: time.Now(),
		GoalID:    arg.GoalID,
		UserID:    arg.UserID,
	}

	m.Deposits[depositId.String()] = deposit

	return deposit, nil
}
