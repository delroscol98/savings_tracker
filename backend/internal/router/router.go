package router

import (
	"net/http"

	"github.com/delroscol98/savings_tracker/backend/api/goals"
	"github.com/delroscol98/savings_tracker/backend/api/health"
	"github.com/delroscol98/savings_tracker/backend/api/users"
	"github.com/delroscol98/savings_tracker/backend/internal/middleware"
)

type Dependencies struct {
	Users  *users.UsersConfig
	Goals  *goals.GoalsConfig
	Health *health.HealthConfig
}

func NewRouter(deps Dependencies) *http.ServeMux {
	serveMux := http.NewServeMux()

	// HEALTH
	serveMux.Handle("GET /health", middleware.Log(http.HandlerFunc(deps.Health.CheckHealthHandler)))

	// USERS
	serveMux.Handle("POST /api/users", middleware.Log(http.HandlerFunc(deps.Users.CreateUserHandler)))
	serveMux.Handle("POST /api/login", middleware.Log(http.HandlerFunc(deps.Users.LoginUserHandler)))
	serveMux.Handle("POST /api/forgot-password", middleware.Log(http.HandlerFunc(deps.Users.RequestPasswordResetHandler)))
	serveMux.Handle("POST /api/reset-password", middleware.Log(http.HandlerFunc(deps.Users.ResetPasswordHandler)))

	// GOALS
	serveMux.Handle("GET /api/goals", middleware.Log(middleware.RequireAuth(http.HandlerFunc(deps.Goals.GetGoalsHandler))))
	serveMux.Handle("POST /api/goals", middleware.Log(middleware.RequireAuth(http.HandlerFunc(deps.Goals.CreateGoalHandler))))
	serveMux.Handle("PUT /api/goals/{goalId}", middleware.Log(middleware.RequireAuth(http.HandlerFunc(deps.Goals.UpdateGoalHandler))))
	serveMux.Handle("DELETE /api/goals/{goalId}", middleware.Log(middleware.RequireAuth(http.HandlerFunc(deps.Goals.DeleteGoalHandler))))
	serveMux.Handle("POST /api/goals/{goalId}/deposits", middleware.Log(middleware.RequireAuth(http.HandlerFunc(deps.Goals.CreateDepositHandler))))

	return serveMux
}
