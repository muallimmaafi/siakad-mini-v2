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
	studentRepo := repository.NewStudentRepository(deps.Pool)
	courseRepo := repository.NewCourseRepository(deps.Pool)
	enrollmentRepo := repository.NewEnrollmentRepository(deps.Pool)

	authService := service.NewAuthService(userRepo, deps.JWT)
	studentService := service.NewStudentService(studentRepo)
	courseService := service.NewCourseService(courseRepo)
	enrollmentService := service.NewEnrollmentService(enrollmentRepo)

	authRequired := middleware.RequireAuth(deps.JWT)
	adminOnly := middleware.RequireRole("admin")
	mahasiswaOnly := middleware.RequireRole("mahasiswa")

	api := app.Group("/api/v1")

	// --- publik ---
	api.Get("/health", healthCheck(deps.Pool))

	// --- autentikasi ---
	auth := api.Group("/auth", middleware.RequireJSON)
	auth.Post("/login", middleware.LoginRateLimiter(), authService.Login)
	auth.Get("/me", authRequired, authService.Me)

	// --- mahasiswa: semua endpoint wajib login, sebagian khusus admin ---
	students := api.Group("/students", authRequired, middleware.RequireJSON)
	students.Get("/", adminOnly, studentService.GetStudents)
	students.Post("/", adminOnly, studentService.CreateStudent)
	students.Get("/:id", middleware.RequireRole("admin", "mahasiswa"), studentService.GetStudentByID) // admin atau pemilik data, dicek di service
	students.Put("/:id", adminOnly, studentService.UpdateStudent)
	students.Delete("/:id", adminOnly, studentService.DeleteStudent)

	// --- mata kuliah: semua role yang sudah login ---
	api.Get("/courses", authRequired, courseService.GetCourses)

	// --- KRS: khusus mahasiswa ---
	enrollments := api.Group("/enrollments", authRequired, mahasiswaOnly, middleware.RequireJSON)
	enrollments.Post("/", enrollmentService.CreateEnrollment)
	enrollments.Post("/", enrollmentService.CreateEnrollment)
	enrollments.Delete("/:id", enrollmentService.DeleteEnrollment)
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

