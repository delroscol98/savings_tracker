package goals_test

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/delroscol98/savings_tracker/backend/internal/database"
)

type mockDB struct {
	GetGoalsErr    error
	CreateGoalErr  error
	GetGoalByIdErr error
	UpdateGoalErr  error
	DeleteGoalErr  error
	Goals          map[string]database.Goal
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
		return database.Goal{}, errors.New("goal not found")
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
