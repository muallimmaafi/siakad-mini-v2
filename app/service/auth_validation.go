package service

import (
	"net/mail"
	"strings"

	"siakad-mini/app/model"
)

const minPasswordLength = 8

// ValidateLogin memeriksa isi body login dan mengembalikan peta
// nama field -> daftar pesan. Peta kosong berarti lolos.
func ValidateLogin(req model.LoginRequest) map[string][]string {
	errs := map[string][]string{}

	email := strings.TrimSpace(req.Email)
	if email == "" {
		errs["email"] = append(errs["email"], "Email wajib diisi")
	} else if addr, err := mail.ParseAddress(email); err != nil || addr.Address != email {
		errs["email"] = append(errs["email"], "Format email tidak valid")
	}

	if req.Password == "" {
		errs["password"] = append(errs["password"], "Password wajib diisi")
	} else if len(req.Password) < minPasswordLength {
		errs["password"] = append(errs["password"], "Password minimal 8 karakter")
	}

	return errs
}