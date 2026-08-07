package api_test

import (
	"context"
	"testing"
	"time"

	"github.com/delroscol98/savings_tracker/backend/internal/auth"
	"github.com/delroscol98/savings_tracker/backend/internal/database"
	"github.com/google/uuid"
)

func seedUser(t *testing.T, email string) database.User {
	t.Helper()

	return seedUserWithPassword(t, email, "ThisIsATestPassword")
}

func seedUserWithPassword(t *testing.T, email, password string) database.User {
	t.Helper()

	hashedPw, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	user, err := dbQueries.CreateUser(context.Background(), database.CreateUserParams{
		Email:          email,
		HashedPassword: hashedPw,
		FullName:       "John Smith",
	})
	if err != nil {
		t.Fatalf("failed to seed user %q: %v", email, err)
	}

	return user
}

func seedGoal(t *testing.T, userID uuid.UUID, target int32, deadline time.Time) database.Goal {
	t.Helper()

	goal, err := dbQueries.CreateGoal(context.Background(), database.CreateGoalParams{
		Target:   target,
		Deadline: deadline,
		UserID:   userID,
	})
	if err != nil {
		t.Fatalf("failed to seed goal for user %v: %v", userID, err)
	}

	return goal
}

func seedUserWithGoal(t *testing.T, email string, target int32, deadline time.Time) (database.User, database.Goal) {
	t.Helper()

	user := seedUser(t, email)
	goal := seedGoal(t, user.ID, target, deadline)

	return user, goal
}
