package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"

	"siakad-mini/helper"
)

// LoginRateLimiter membatasi percobaan login yang GAGAL dari satu alamat IP.
// Lebih dari 5 kali dalam 1 menit ditolak dengan status 429.
// SkipSuccessfulRequests: login yang berhasil tidak ikut dihitung.
func LoginRateLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:                    5,
		Expiration:             1 * time.Minute,
		SkipSuccessfulRequests: true,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			c.Set("Retry-After", "60")
			return helper.Fail(c, fiber.StatusTooManyRequests,
				"Terlalu banyak percobaan login, coba lagi dalam satu menit")
		},
	})
}