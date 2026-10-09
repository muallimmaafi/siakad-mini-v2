package service

import (
	"net/mail"

	"siakad-mini/app/model"
)

func addErr(errs map[string][]string, field, msg string) {
	errs[field] = append(errs[field], msg)
}

func isAllDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// ValidateCreateStudent memeriksa body POST /students.
func ValidateCreateStudent(req model.CreateStudentRequest, tahunSekarang int) map[string][]string {
	errs := map[string][]string{}

	if req.NIM == "" {
		addErr(errs, "nim", "NIM wajib diisi")
	} else if len(req.NIM) != 12 || !isAllDigits(req.NIM) {
		addErr(errs, "nim", "NIM harus 12 digit angka")
	}

	if req.Email == "" {
		addErr(errs, "email", "Email wajib diisi")
	} else if addr, err := mail.ParseAddress(req.Email); err != nil || addr.Address != req.Email {
		addErr(errs, "email", "Format email tidak valid")
	}

	validateStudentFields(errs, req.Nama, req.Prodi, req.Angkatan, req.IPKTerakhir, tahunSekarang)
	return errs
}

// ValidateUpdateStudent memeriksa body PUT /students/:id (NIM tidak diubah).
func ValidateUpdateStudent(req model.UpdateStudentRequest, tahunSekarang int) map[string][]string {
	errs := map[string][]string{}
	validateStudentFields(errs, req.Nama, req.Prodi, req.Angkatan, req.IPKTerakhir, tahunSekarang)
	return errs
}

// validateStudentFields memuat aturan yang sama untuk tambah dan ubah.
func validateStudentFields(errs map[string][]string, nama, prodi string, angkatan int, ipk *float64, tahunSekarang int) {
	if nama == "" {
		addErr(errs, "nama", "Nama wajib diisi")
	}
	if prodi == "" {
		addErr(errs, "prodi", "Prodi wajib diisi")
	}

	if angkatan == 0 {
		addErr(errs, "angkatan", "Angkatan wajib diisi")
	} else if angkatan < 1000 || angkatan > tahunSekarang {
		addErr(errs, "angkatan", "Angkatan harus 4 digit dan tidak lebih dari tahun berjalan")
	}

	if ipk != nil && (*ipk < 0 || *ipk > 4) {
		addErr(errs, "ipk_terakhir", "IPK harus antara 0,00 sampai 4,00")
	}
}