package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"siakad-mini/app/repository"
	"siakad-mini/app/service"
	"siakad-mini/helper"
	"siakad-mini/middleware"
)

// Dependencies mengumpulkan seluruh dependensi jadi satu struct, supaya
// penambahan service berikutnya tidak mengubah tanda tangan Register.
type Dependencies struct {
	Pool *pgxpool.Pool
	JWT  *helper.JWTManager
}

func Register(app *fiber.App, deps Dependencies) {
	// Susun rantai: repository -> service
	userRepo := repository.NewUserRepository(deps.Pool)
	authService := service.NewAuthService(userRepo, deps.JWT)

	authRequired := middleware.RequireAuth(deps.JWT)

	api := app.Group("/api/v1")

	// --- publik ---
	api.Get("/health", healthCheck(deps.Pool))

	// --- autentikasi ---
	auth := api.Group("/auth", middleware.RequireJSON)
	auth.Post("/login", middleware.LoginRateLimiter(), authService.Login)
	auth.Get("/me", authRequired, authService.Me)
}

// healthCheck melaporkan kondisi layanan beserta databasenya.
func healthCheck(pool *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()
		if err := pool.Ping(ctx); err != nil {
			return helper.Fail(c, fiber.StatusServiceUnavailable,
				"database tidak dapat dihubungi")
		}
		return helper.Success(c, fiber.StatusOK, "server dan database berjalan",
			fiber.Map{"timestamp": time.Now()})
	}
}