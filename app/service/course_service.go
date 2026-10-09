package service

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"siakad-mini/app/model"
	"siakad-mini/app/repository"
	"siakad-mini/helper"
)

type CourseService struct {
	repo repository.CourseRepository
}

func NewCourseService(repo repository.CourseRepository) *CourseService {
	return &CourseService{repo: repo}
}

// GetCourses: GET /api/v1/courses (semua role yang sudah login)
func (s *CourseService) GetCourses(c *fiber.Ctx) error {
	// semester boleh kosong, tetapi kalau diisi harus angka positif
	semester := 0
	if raw := c.Query("semester"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 {
			return helper.FailValidation(c, map[string][]string{
				"semester": {"Semester harus berupa angka positif"},
			})
		}
		semester = n
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	courses, err := s.repo.FindAll(ctx, model.CourseListParams{
		Semester:      semester,
		Search:        strings.TrimSpace(c.Query("search")),
		OnlyAvailable: c.Query("available") == "true",
		TahunAkademik: strings.TrimSpace(c.Query("tahun_akademik")),
	})
	if err != nil {
		return fmt.Errorf("get courses: %w", err)
	}

	// sisa_kuota = kuota - terisi
	for i := range courses {
		sisa := courses[i].Kuota - courses[i].Terisi
		if sisa < 0 {
			sisa = 0
		}
		courses[i].SisaKuota = sisa
	}

	return helper.Success(c, fiber.StatusOK, "Data mata kuliah berhasil diambil", courses)
}