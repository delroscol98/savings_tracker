package goals

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/delroscol98/savings_tracker/backend/internal/database"
	"github.com/delroscol98/savings_tracker/backend/internal/middleware"
	"github.com/delroscol98/savings_tracker/backend/internal/response"
	"github.com/google/uuid"
)

func (g *GoalsConfig) CreateDepositHandler(w http.ResponseWriter, r *http.Request) {
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

	body := CreateDepositParams{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&body)
	if err != nil {
		log.Print(err)
		response.RespondWithError(w, http.StatusBadRequest, "Error decoding body")
		return
	}

	fieldErrors := ValidateDepositParams(body)
	if fieldErrors != nil {
		log.Print(fieldErrors)
		response.RespondWithValidationError(w, http.StatusBadRequest, response.ValidationErrorBody{
			Error:  "Invalid parameters to create new deposit",
			Fields: fieldErrors,
		})
		return
	}

	deposit, err := g.Queries.CreateDeposit(r.Context(), database.CreateDepositParams{
		Amount: body.Amount,
		Note: sql.NullString{
			String: body.Note,
			Valid:  true,
		},
		GoalID: goalId,
		UserID: userId,
	})
	if err != nil {
		log.Print(err)
		response.RespondWithError(w, http.StatusInternalServerError, "error creating a deposit")
		return
	}

	response.RespondWithJSON(w, http.StatusCreated, Deposit{
		Id:        deposit.ID,
		Amount:    deposit.Amount,
		Note:      deposit.Note.String,
		CreatedAt: deposit.CreatedAt,
	})
}
