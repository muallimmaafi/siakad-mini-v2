package repository

import (
	"errors"
	"fmt"
)

// Sentinel error dari lapisan repository. Service memeriksanya dengan
// errors.Is, sehingga lapisan atas tidak perlu tahu detail pgx.
var (
	ErrNotFound  = errors.New("data tidak ditemukan")
	ErrDuplicate = errors.New("data sudah ada")

	// khusus KRS
	ErrStudentNotFound = errors.New("mahasiswa tidak ditemukan atau tidak aktif")
	ErrCourseNotFound  = errors.New("mata kuliah tidak ditemukan")
	ErrAlreadyEnrolled = errors.New("mata kuliah sudah diambil pada tahun akademik ini")
	ErrCourseFull      = errors.New("kuota mata kuliah sudah penuh")
)

// SKSExceededError membawa angka-angka yang dibutuhkan untuk pesan
// "sisa SKS" kepada pengguna.
type SKSExceededError struct {
	Batas     int
	Diambil   int
	Sisa      int
	SKSMatkul int
}

func (e *SKSExceededError) Error() string {
	return fmt.Sprintf("total SKS melebihi batas (batas %d, diambil %d, sisa %d, matkul %d)",
		e.Batas, e.Diambil, e.Sisa, e.SKSMatkul)
}