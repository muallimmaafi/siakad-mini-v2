package middleware

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"

	"siakad-mini/helper"
)

// RequireAuth memeriksa access token di header Authorization.
// Kalau token sah, identitas user disimpan di Locals agar bisa dibaca
// service dan middleware berikutnya.
func RequireAuth(jwt *helper.JWTManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token, ok := bearerToken(c)
		if !ok {
			return helper.Fail(c, fiber.StatusUnauthorized, "Token tidak ditemukan atau format salah")
		}

		authUser, err := jwt.Parse(token)
		if err != nil {
			if errors.Is(err, helper.ErrExpiredToken) {
				return helper.Fail(c, fiber.StatusUnauthorized, "Token sudah kedaluwarsa")
			}
			return helper.Fail(c, fiber.StatusUnauthorized, "Token tidak valid")
		}

		c.Locals(helper.LocalsAuthUser, authUser)
		return c.Next()
	}
}

// bearerToken mengambil token dari header "Authorization: Bearer <token>".
func bearerToken(c *fiber.Ctx) (string, bool) {
	parts := strings.SplitN(c.Get(fiber.HeaderAuthorization), " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", false
	}
	token := strings.TrimSpace(parts[1])
	return token, token != ""
}