package service

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"siakad-mini/app/model"
	"siakad-mini/app/repository"
	"siakad-mini/helper"
)

const (
	defaultPerPage = 10
	maxPerPage     = 50 // batas maksimal sesuai soal
)

type StudentService struct {
	repo repository.StudentRepository
}

func NewStudentService(repo repository.StudentRepository) *StudentService {
	return &StudentService{repo: repo}
}

// studentError menerjemahkan error repository menjadi respons yang sesuai.
func studentError(c *fiber.Ctx, err error, action string) error {
	if errors.Is(err, repository.ErrNotFound) {
		return helper.Fail(c, fiber.StatusNotFound, "Mahasiswa tidak ditemukan")
	}
	return fmt.Errorf("%s: %w", action, err) // jadi 500 lewat ErrorHandler
}

// GetStudents: GET /api/v1/students (admin)
func (s *StudentService) GetStudents(c *fiber.Ctx) error {
	errs := map[string][]string{}

	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		errs["page"] = []string{"Page harus berupa angka positif"}
	}

	perPage, err := strconv.Atoi(c.Query("per_page", strconv.Itoa(defaultPerPage)))
	if err != nil || perPage < 1 {
		errs["per_page"] = []string{"Per_page harus berupa angka positif"}
	} else if perPage > maxPerPage {
		perPage = maxPerPage
	}

	angkatan := 0
	if raw := c.Query("angkatan"); raw != "" {
		angkatan, err = strconv.Atoi(raw)
		if err != nil || angkatan < 1 {
			errs["angkatan"] = []string{"Angkatan harus berupa angka"}
		}
	}

	sortBy := c.Query("sort")
	if sortBy != "" && sortBy != "nama" && sortBy != "-ipk_terakhir" {
		errs["sort"] = []string{"Sort hanya boleh: nama atau -ipk_terakhir"}
	}

	if len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	students, total, err := s.repo.FindAll(ctx, model.StudentListParams{
		Page:     page,
		PerPage:  perPage,
		Prodi:    strings.TrimSpace(c.Query("prodi")),
		Angkatan: angkatan,
		Search:   strings.TrimSpace(c.Query("search")),
		Sort:     sortBy,
	})
	if err != nil {
		return fmt.Errorf("get students: %w", err)
	}

	lastPage := (total + perPage - 1) / perPage
	if lastPage < 1 {
		lastPage = 1
	}

	return helper.SuccessList(c, "Data mahasiswa berhasil diambil", students, model.Meta{
		CurrentPage: page,
		PerPage:     perPage,
		Total:       total,
		LastPage:    lastPage,
	})
}

// CreateStudent: POST /api/v1/students (admin)
func (s *StudentService) CreateStudent(c *fiber.Ctx) error {
	var req model.CreateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.FailValidation(c, map[string][]string{"body": {"Format JSON tidak valid"}})
	}
	req.NIM = strings.TrimSpace(req.NIM)
	req.Nama = strings.TrimSpace(req.Nama)
	req.Prodi = strings.TrimSpace(req.Prodi)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	if errs := ValidateCreateStudent(req, time.Now().Year()); len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	// NIM atau email kembar dijawab 422 dengan pesan di field yang bersangkutan.
	errs := map[string][]string{}
	if exists, err := s.repo.ExistsNIM(ctx, req.NIM); err != nil {
		return fmt.Errorf("cek nim: %w", err)
	} else if exists {
		errs["nim"] = []string{"NIM sudah terdaftar"}
	}
	if exists, err := s.repo.ExistsEmail(ctx, req.Email); err != nil {
		return fmt.Errorf("cek email: %w", err)
	} else if exists {
		errs["email"] = []string{"Email sudah terdaftar"}
	}
	if len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	// Password awal = NIM (di-hash).
	hash, err := helper.HashPassword(req.NIM)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	student, err := s.repo.Create(ctx, req, hash)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			// Jaring pengaman bila ada dua request kembar bersamaan.
			return helper.FailValidation(c, map[string][]string{
				"nim": {"NIM atau email sudah terdaftar"},
			})
		}
		return fmt.Errorf("create student: %w", err)
	}

	return helper.Success(c, fiber.StatusCreated, "Mahasiswa berhasil ditambahkan",
		model.StudentCreated{Student: student, Email: req.Email})
}

// GetStudentByID: GET /api/v1/students/:id (admin, atau mahasiswa untuk data sendiri)
func (s *StudentService) GetStudentByID(c *fiber.Ctx) error {
	id, ok := helper.ParamID(c)
	if !ok {
		return helper.Fail(c, fiber.StatusNotFound, "Mahasiswa tidak ditemukan")
	}

	authUser, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "Token tidak valid")
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	// cari dulu (404 kalau tidak ada), baru cek hak akses (403)
	student, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return studentError(c, err, "get student")
	}

	if !CanAccessStudent(authUser.Role, authUser.UserID, student.UserID) {
		return helper.Fail(c, fiber.StatusForbidden, "Anda tidak boleh mengakses data mahasiswa lain")
	}

	courses, err := s.repo.FindEnrollments(ctx, student.ID, strings.TrimSpace(c.Query("tahun_akademik")))
	if err != nil {
		return fmt.Errorf("get student krs: %w", err)
	}

	totalSKS := 0
	for _, course := range courses {
		totalSKS += course.SKS
	}

	return helper.Success(c, fiber.StatusOK, "Detail mahasiswa berhasil diambil", model.StudentDetail{
		Student:    student,
		MataKuliah: courses,
		TotalSKS:   totalSKS,
		BatasSKS:   BatasSKS(student.IPKTerakhir),
	})
}

// UpdateStudent: PUT /api/v1/students/:id (admin)
func (s *StudentService) UpdateStudent(c *fiber.Ctx) error {
	id, ok := helper.ParamID(c)
	if !ok {
		return helper.Fail(c, fiber.StatusNotFound, "Mahasiswa tidak ditemukan")
	}

	var req model.UpdateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.FailValidation(c, map[string][]string{"body": {"Format JSON tidak valid"}})
	}
	req.Nama = strings.TrimSpace(req.Nama)
	req.Prodi = strings.TrimSpace(req.Prodi)

	if errs := ValidateUpdateStudent(req, time.Now().Year()); len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return studentError(c, err, "update student")
	}

	// NIM tidak boleh diubah.
	if req.NIM != nil && strings.TrimSpace(*req.NIM) != current.NIM {
		return helper.FailValidation(c, map[string][]string{
			"nim": {"NIM tidak dapat diubah"},
		})
	}

	updated, err := s.repo.Update(ctx, id, req)
	if err != nil {
		return studentError(c, err, "update student")
	}

	return helper.Success(c, fiber.StatusOK, "Data mahasiswa berhasil diperbarui", updated)
}

// DeleteStudent: DELETE /api/v1/students/:id (admin) - soft delete
func (s *StudentService) DeleteStudent(c *fiber.Ctx) error {
	id, ok := helper.ParamID(c)
	if !ok {
		return helper.Fail(c, fiber.StatusNotFound, "Mahasiswa tidak ditemukan")
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	if err := s.repo.SoftDelete(ctx, id); err != nil {
		return studentError(c, err, "delete student")
	}
	return helper.NoContent(c)
}