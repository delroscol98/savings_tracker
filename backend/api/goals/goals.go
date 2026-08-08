package goals

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/delroscol98/savings_tracker/backend/internal/database"
	"github.com/delroscol98/savings_tracker/backend/internal/middleware"
	"github.com/delroscol98/savings_tracker/backend/internal/response"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Queries interface {
	GetGoals(ctx context.Context, userID uuid.UUID) ([]database.Goal, error)
	CreateGoal(ctx context.Context, arg database.CreateGoalParams) (database.Goal, error)
	GetGoalById(ctx context.Context, id uuid.UUID) (database.Goal, error)
	UpdateGoal(ctx context.Context, arg database.UpdateGoalParams) (database.Goal, error)
	DeleteGoal(ctx context.Context, arg database.DeleteGoalParams) error

	CreateDeposit(ctx context.Context, arg database.CreateDepositParams) (database.Deposit, error)
}

type GoalsConfig struct {
	Queries Queries
}

func (g *GoalsConfig) GetGoalsHandler(w http.ResponseWriter, r *http.Request) {
	userId, ok := middleware.GetUserId(r.Context())
	if !ok {
		response.RespondWithError(w, http.StatusUnauthorized, "User not found")
		return
	}

	dbGoals, err := g.Queries.GetGoals(r.Context(), userId)
	if err != nil {
		log.Print(err)
		response.RespondWithError(w, http.StatusInternalServerError, "error fetching goals")
		return
	}

	goals := make([]Goal, len(dbGoals))
	for i, goal := range dbGoals {
		goals[i] = Goal{
			Id:       goal.ID,
			Target:   goal.Target,
			Deadline: goal.Deadline,
			UserId:   goal.UserID,
		}
	}

	response.RespondWithJSON(w, http.StatusOK, goals)
}

func (g *GoalsConfig) CreateGoalHandler(w http.ResponseWriter, r *http.Request) {
	userId, ok := middleware.GetUserId(r.Context())
	if !ok {
		response.RespondWithError(w, http.StatusUnauthorized, "User not found")
		return
	}

	// Creating Goal
	params := CreateGoalParams{}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		log.Print(err)
		response.RespondWithError(w, http.StatusBadRequest, "Error decoding body")
		return
	}

	if userId != params.UserId {
		response.RespondWithError(w, http.StatusUnauthorized, "User ID does not match")
		return
	}

	fieldErrors := ValidateGoalParams(params.GoalFields)
	if fieldErrors != nil {
		log.Print(fieldErrors)
		response.RespondWithValidationError(w, http.StatusBadRequest, response.ValidationErrorBody{
			Error:  "Invalid parameters to create new goal",
			Fields: fieldErrors,
		})
		return
	}

	goal, err := g.Queries.CreateGoal(r.Context(), database.CreateGoalParams{
		Target:   params.Target,
		Deadline: params.Deadline,
		UserID:   params.UserId,
	})
	if err != nil {
		// PostgreSQL's foreign key violation code is 23503
		var pqe *pq.Error
		if errors.As(err, &pqe) && pqe.Code == "23503" {
			response.RespondWithError(w, http.StatusBadRequest, "Error creating goal")
			return
		}
		log.Print(err)
		response.RespondWithError(w, http.StatusInternalServerError, "Error creating goal")
		return
	}

	response.RespondWithJSON(w, http.StatusCreated, Goal{
		Id:       goal.ID,
		Target:   goal.Target,
		Deadline: goal.Deadline,
		UserId:   goal.UserID,
	})
}

func (g *GoalsConfig) UpdateGoalHandler(w http.ResponseWriter, r *http.Request) {
	userId, ok := middleware.GetUserId(r.Context())
	if !ok {
		response.RespondWithError(w, http.StatusUnauthorized, "User not found")
		return
	}

	goalIdString := r.PathValue("goalId")

	goalId, err := uuid.Parse(goalIdString)
	if err != nil {
		log.Print(err)
		response.RespondWithError(w, http.StatusBadRequest, "error parsing goal id")
		return
	}

	goal, err := g.Queries.GetGoalById(r.Context(), goalId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			response.RespondWithError(w, http.StatusNotFound, "error finding goal")
			return
		}
		log.Print(err)
		response.RespondWithError(w, http.StatusInternalServerError, "error finding goal")
		return
	}

	if userId != goal.UserID {
		response.RespondWithError(w, http.StatusForbidden, "mismatch user id")
		return
	}

	body := UpdateGoalParams{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&body)
	if err != nil {
		log.Print(err)
		response.RespondWithError(w, http.StatusBadRequest, "error decoding body")
		return
	}

	fieldErrors := ValidateGoalParams(body.GoalFields)
	if fieldErrors != nil {
		log.Print(fieldErrors)
		response.RespondWithValidationError(w, http.StatusBadRequest, response.ValidationErrorBody{
			Error:  "Invalid parameters to create new goal",
			Fields: fieldErrors,
		})
		return
	}

	goal, err = g.Queries.UpdateGoal(r.Context(), database.UpdateGoalParams{
		Target:   body.Target,
		Deadline: body.Deadline,
		ID:       goalId,
		UserID:   userId,
	})
	if err != nil {
		log.Print(err)
		response.RespondWithError(w, http.StatusInternalServerError, "error updating goal")
		return
	}

	response.RespondWithJSON(w, http.StatusOK, Goal{
		Id:       goal.ID,
		Target:   goal.Target,
		Deadline: goal.Deadline,
		UserId:   goal.UserID,
	})
}

func (g *GoalsConfig) DeleteGoalHandler(w http.ResponseWriter, r *http.Request) {
	userId, ok := middleware.GetUserId(r.Context())
	if !ok {
		response.RespondWithError(w, http.StatusUnauthorized, "User not found")
		return
	}

	goalIdString := r.PathValue("goalId")

	goalId, err := uuid.Parse(goalIdString)
	if err != nil {
		log.Print(err)
		response.RespondWithError(w, http.StatusBadRequest, "error parsing goal id")
		return
	}

	goal, err := g.Queries.GetGoalById(r.Context(), goalId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			response.RespondWithError(w, http.StatusNotFound, "error finding goal")
			return
		}
		log.Print(err)
		response.RespondWithError(w, http.StatusInternalServerError, "error finding goal")
		return
	}

	if userId != goal.UserID {
		response.RespondWithError(w, http.StatusForbidden, "mismatch user id")
		return
	}

	err = g.Queries.DeleteGoal(r.Context(), database.DeleteGoalParams{
		ID:     goalId,
		UserID: userId,
	})
	if err != nil {
		log.Print(err)
		response.RespondWithError(w, http.StatusInternalServerError, "error deleting goal")
		return
	}

	response.RespondWithJSON(
		w, http.StatusOK, struct {
			Message string `json:"message"`
		}{
			Message: "Goal successfully deleted",
		},
	)
}
