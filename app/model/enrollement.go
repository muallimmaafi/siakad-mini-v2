package model

import "time"

type Enrollment struct {
	ID            int       `json:"id"`
	StudentID     int       `json:"student_id"`
	CourseID      int       `json:"course_id"`
	TahunAkademik string    `json:"tahun_akademik"`
	CreatedAt     time.Time `json:"created_at"`
}

type EnrollmentCourse struct {
	KodeMK   string `json:"kode_mk"`
	NamaMK   string `json:"nama_mk"`
	SKS      int    `json:"sks"`
	Semester int    `json:"semester"`
}

type CreateEnrollmentRequest struct {
	CourseID      int    `json:"course_id"`
	TahunAkademik string `json:"tahun_akademik"`
}

// EnrollmentResult adalah respons POST /enrollments.
type EnrollmentResult struct {
	Enrollment
	Course   EnrollmentCourse `json:"course"`
	TotalSKS int              `json:"total_sks"`
	BatasSKS int              `json:"batas_sks"`
	SisaSKS  int              `json:"sisa_sks"`
}