package repository

import "errors"

// Sentinel error dari lapisan repository. Service memeriksanya dengan
// errors.Is, sehingga lapisan atas tidak perlu tahu detail pgx.
var (
	ErrNotFound  = errors.New("data tidak ditemukan")
	ErrDuplicate = errors.New("data sudah ada")
)