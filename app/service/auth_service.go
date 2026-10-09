package service

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"

	"siakad-mini/app/model"
	"siakad-mini/app/repository"
	"siakad-mini/helper"
)

type AuthService struct {
	users repository.UserRepository
	jwt   *helper.JWTManager
}

func NewAuthService(users repository.UserRepository, jwt *helper.JWTManager) *AuthService {
	return &AuthService{users: users, jwt: jwt}
}

// Login: POST /api/v1/auth/login
func (s *AuthService) Login(c *fiber.Ctx) error {
	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.FailValidation(c, map[string][]string{
			"body": {"Format JSON tidak valid"},
		})
	}
	req.Email = strings.TrimSpace(req.Email)

	// 1. cek isi body, kalau salah kirim 422
	if errs := ValidateLogin(req); len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	// 2. cari user berdasarkan email
	user, err := s.users.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			// Tetap menghitung hash supaya waktu respons mirip kasus password salah.
			helper.VerifyDummyPassword(req.Password)
			return helper.Fail(c, fiber.StatusUnauthorized, "Email atau password salah")
		}
		return fmt.Errorf("login: %w", err)
	}

	// 3. cocokkan password dengan hash di database
	if !helper.VerifyPassword(user.Password, req.Password) {
		return helper.Fail(c, fiber.StatusUnauthorized, "Email atau password salah")
	}

	// 4. buat access token
	token, err := s.jwt.GenerateAccess(user)
	if err != nil {
		return fmt.Errorf("membuat token: %w", err)
	}

	return helper.Success(c, fiber.StatusOK, "Login berhasil", model.LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int(s.jwt.AccessTTL().Seconds()),
		User:        model.LoginUser{ID: user.ID, Email: user.Email, Role: user.Role},
	})
}

// Me: GET /api/v1/auth/me
func (s *AuthService) Me(c *fiber.Ctx) error {
	authUser, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "Token tidak valid")
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	user, err := s.users.FindByID(ctx, authUser.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Fail(c, fiber.StatusUnauthorized, "Akun tidak ditemukan")
		}
		return fmt.Errorf("me: %w", err)
	}

	data := model.MeResponse{ID: user.ID, Email: user.Email, Role: user.Role}

	// kalau mahasiswa, tambahkan data dari tabel students
	if user.Role == "mahasiswa" {
		profile, err := s.users.FindStudentProfile(ctx, user.ID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return helper.Fail(c, fiber.StatusUnauthorized, "Akun tidak ditemukan")
			}
			return fmt.Errorf("me profil: %w", err)
		}
		data.Student = &profile
	}

	return helper.Success(c, fiber.StatusOK, "Data user berhasil diambil", data)
}