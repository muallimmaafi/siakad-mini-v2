package service

import (
	"regexp"
	"strconv"

	"siakad-mini/app/model"
)

var tahunAkademikRe = regexp.MustCompile(`^(\d{4})/(\d{4})-(Ganjil|Genap)$`)

// IsTahunAkademik memeriksa format seperti 2026/2027-Ganjil, dengan tahun
// kedua sama dengan tahun pertama ditambah satu.
func IsTahunAkademik(s string) bool {
	m := tahunAkademikRe.FindStringSubmatch(s)
	if m == nil {
		return false
	}
	awal, _ := strconv.Atoi(m[1])
	akhir, _ := strconv.Atoi(m[2])
	return akhir == awal+1
}

// ValidateEnrollment memeriksa body POST /enrollments.
func ValidateEnrollment(req model.CreateEnrollmentRequest) map[string][]string {
	errs := map[string][]string{}

	if req.CourseID < 1 {
		addErr(errs, "course_id", "course_id wajib diisi")
	}

	if req.TahunAkademik == "" {
		addErr(errs, "tahun_akademik", "Tahun akademik wajib diisi")
	} else if !IsTahunAkademik(req.TahunAkademik) {
		addErr(errs, "tahun_akademik", "Format tahun_akademik harus seperti 2026/2027-Ganjil atau 2026/2027-Genap")
	}

	return errs
}