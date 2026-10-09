package model

import "time"

type User struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // tidak pernah keluar sebagai JSON
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// AuthUser adalah identitas yang dibawa access token.
type AuthUser struct {
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginUser struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type LoginResponse struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	ExpiresIn   int       `json:"expires_in"` // detik
	User        LoginUser `json:"user"`
}

// StudentProfile adalah data mahasiswa yang ikut tampil di /auth/me.
type StudentProfile struct {
	NIM      string `json:"nim"`
	Nama     string `json:"nama"`
	Prodi    string `json:"prodi"`
	Angkatan int    `json:"angkatan"`
}

type MeResponse struct {
	ID      int             `json:"id"`
	Email   string          `json:"email"`
	Role    string          `json:"role"`
	Student *StudentProfile `json:"student,omitempty"`
}