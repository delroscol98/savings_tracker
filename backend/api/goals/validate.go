package goals

import (
	"time"

	"github.com/delroscol98/savings_tracker/backend/internal/response"
)

func ValidateGoalParams(params GoalFields) response.FieldErrors {
	fieldsErrors := make(response.FieldErrors)

	// Target Validation
	if params.Target < 0 {
		fieldsErrors["target"] = append(fieldsErrors["target"], "Goal target cannot be negative")
	}

	// Deadline Validation
	if time.Now().After(params.Deadline) {
		fieldsErrors["deadline"] = append(fieldsErrors["deadline"], "Deadline cannot be in the past")
	}

	// Check for any error messages
	if len(fieldsErrors) == 0 {
		return nil
	}

	return fieldsErrors
}
