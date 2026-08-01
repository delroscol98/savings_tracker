package auth

import (
	"time"

	"github.com/google/uuid"
)

type GoalFields struct {
	Target   int32     `json:"target"`
	Deadline time.Time `json:"deadline"`
}

type CreateGoalParams struct {
	GoalFields
	UserId uuid.UUID `json:"user_id"`
}

type UpdateGoalParams struct {
	GoalFields
}

type Goal struct {
	Id       uuid.UUID `json:"id"`
	Target   int32     `json:"target"`
	Deadline time.Time `json:"deadline"`
	UserId   uuid.UUID `json:"user_id"`
}
