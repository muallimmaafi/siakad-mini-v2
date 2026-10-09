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

type EnrollmentService struct {
	repo repository.EnrollmentRepository
}

func NewEnrollmentService(repo repository.EnrollmentRepository) *EnrollmentService {
	return &EnrollmentService{repo: repo}
}

// enrollmentError menerjemahkan hasil aturan bisnis menjadi status HTTP:
// duplikat 409, kuota penuh dan SKS melebihi batas 422, bukan mahasiswa 403.
func enrollmentError(c *fiber.Ctx, err error) error {
	var sksErr *repository.SKSExceededError

	switch {
	case errors.Is(err, repository.ErrStudentNotFound):
		return helper.Fail(c, fiber.StatusForbidden, "Hanya mahasiswa aktif yang dapat mengambil mata kuliah")
	case errors.Is(err, repository.ErrCourseNotFound):
		return helper.FailValidation(c, map[string][]string{
			"course_id": {"Mata kuliah tidak ditemukan"},
		})
	case errors.Is(err, repository.ErrAlreadyEnrolled):
		return helper.Fail(c, fiber.StatusConflict, "Mata kuliah ini sudah Anda ambil pada tahun akademik yang sama")
	case errors.Is(err, repository.ErrCourseFull):
		return helper.Fail(c, fiber.StatusUnprocessableEntity, "Kuota mata kuliah sudah penuh")
	case errors.As(err, &sksErr):
		return helper.Fail(c, fiber.StatusUnprocessableEntity, fmt.Sprintf(
			"Total SKS melebihi batas. Batas SKS Anda %d, sudah diambil %d, sisa %d SKS, sedangkan mata kuliah ini %d SKS",
			sksErr.Batas, sksErr.Diambil, sksErr.Sisa, sksErr.SKSMatkul))
	default:
		return fmt.Errorf("enrollment: %w", err) // jadi 500 lewat ErrorHandler
	}
}

// CreateEnrollment: POST /api/v1/enrollments (mahasiswa)
func (s *EnrollmentService) CreateEnrollment(c *fiber.Ctx) error {
	authUser, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "Token tidak valid")
	}

	var req model.CreateEnrollmentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.FailValidation(c, map[string][]string{"body": {"Format JSON tidak valid"}})
	}
	req.TahunAkademik = strings.TrimSpace(req.TahunAkademik)

	if errs := ValidateEnrollment(req); len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	result, err := s.repo.Enroll(ctx, authUser.UserID, req.CourseID, req.TahunAkademik, BatasSKS)
	if err != nil {
		return enrollmentError(c, err)
	}

	return helper.Success(c, fiber.StatusCreated, "Mata kuliah berhasil ditambahkan ke KRS", result)
}