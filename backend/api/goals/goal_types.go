package goals

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
	Progress int64     `json:"progress"`
}

type CreateDepositParams struct {
	Amount int32  `json:"amount"`
	Note   string `json:"note"`
}

type Deposit struct {
	Id        uuid.UUID `json:"id"`
	Amount    int32     `json:"amount"`
	Note      string    `json:"note"`
	CreatedAt time.Time `json:"created_at"`
}
