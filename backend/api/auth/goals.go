package auth

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/delroscol98/savings_tracker/backend/internal/database"
	"github.com/delroscol98/savings_tracker/backend/internal/response"
	"github.com/google/uuid"
)

func (a *AuthConfig) CreateGoalHandler(w http.ResponseWriter, r *http.Request) {
	// JWT validation
	token, err := GetBearerToken(r.Header)
	if err != nil {
		log.Print(err)
		response.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	userId, err := ValidateJWT(token, a.JWTSecret)
	if err != nil {
		log.Print(err)
		response.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Creating Goal
	params := CreateGoalParams{}

	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&params)
	if err != nil {
		log.Print(err)
		response.RespondWithError(w, http.StatusBadRequest, "Error decoding body")
		return
	}

	if userId != params.UserId {
		response.RespondWithError(w, http.StatusUnauthorized, "User ID does not match")
		return
	}

	params, fieldErrors := ValidateCreateGoalParams(params)
	if fieldErrors != nil {
		log.Print(fieldErrors)
		response.RespondWithValidationError(w, http.StatusBadRequest, response.ValidationErrorBody{
			Error:  "Invalid parameters to create new goal",
			Fields: fieldErrors,
		})
		return
	}

	goal, err := a.Queries.CreateGoal(r.Context(), database.CreateGoalParams{
		Target:   params.Target,
		Deadline: params.Deadline,
		UserID:   params.UserId,
	})
	if err != nil {
		log.Print(err)
		response.RespondWithError(w, http.StatusBadRequest, "Error creating goal")
		return
	}

	response.RespondWithJSON(w, http.StatusCreated, Goal{
		Id:        goal.ID,
		Target:    goal.Target,
		Deadline:  goal.Deadline,
		CreatedAt: goal.CreatedAt,
		UpdatedAt: goal.UpdatedAt,
		UserId:    goal.UserID,
	})
}

func (a *AuthConfig) DeleteGoalHandler(w http.ResponseWriter, r *http.Request) {
	// JWT validation
	token, err := GetBearerToken(r.Header)
	if err != nil {
		log.Print(err)
		response.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	userId, err := ValidateJWT(token, a.JWTSecret)
	if err != nil {
		log.Print(err)
		response.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	goalIdString := r.PathValue("goalId")

	goalId, err := uuid.Parse(goalIdString)
	if err != nil {
		log.Print(err)
		response.RespondWithError(w, http.StatusBadRequest, "error parsing goal id")
		return
	}

	goal, err := a.Queries.GetGoalById(r.Context(), goalId)
	if err != nil {
		log.Print(err)
		response.RespondWithError(w, http.StatusNotFound, "error finding goal")
		return
	}

	if userId != goal.UserID {
		response.RespondWithError(w, http.StatusForbidden, "mismatch user id")
		return
	}

	err = a.Queries.DeleteGoal(r.Context(), goalId)
	if err != nil {
		log.Print(err)
		response.RespondWithError(w, http.StatusNotFound, "error deleting goal")
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
