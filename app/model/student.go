package model

type Student struct {
	ID          int     `json:"id"`
	UserID      int     `json:"-"` // dipakai untuk cek kepemilikan, tidak ikut ke JSON
	NIM         string  `json:"nim"`
	Nama        string  `json:"nama"`
	Prodi       string  `json:"prodi"`
	Angkatan    int     `json:"angkatan"`
	IPKTerakhir float64 `json:"ipk_terakhir"`
}

// StudentCreated adalah respons POST /students (data mahasiswa + email akunnya).
type StudentCreated struct {
	Student
	Email string `json:"email"`
}

// CourseTaken adalah satu mata kuliah pada KRS mahasiswa.
type CourseTaken struct {
	EnrollmentID  int    `json:"enrollment_id"`
	CourseID      int    `json:"course_id"`
	KodeMK        string `json:"kode_mk"`
	NamaMK        string `json:"nama_mk"`
	SKS           int    `json:"sks"`
	Semester      int    `json:"semester"`
	TahunAkademik string `json:"tahun_akademik"`
}

type StudentDetail struct {
	Student
	MataKuliah []CourseTaken `json:"mata_kuliah"`
	TotalSKS   int           `json:"total_sks"`
	BatasSKS   int           `json:"batas_sks"`
}

type CreateStudentRequest struct {
	NIM         string   `json:"nim"`
	Nama        string   `json:"nama"`
	Email       string   `json:"email"`
	Prodi       string   `json:"prodi"`
	Angkatan    int      `json:"angkatan"`
	IPKTerakhir *float64 `json:"ipk_terakhir"`
}

type UpdateStudentRequest struct {
	NIM         *string  `json:"nim"` // bila dikirim, harus sama dengan NIM lama
	Nama        string   `json:"nama"`
	Prodi       string   `json:"prodi"`
	Angkatan    int      `json:"angkatan"`
	IPKTerakhir *float64 `json:"ipk_terakhir"`
}

// StudentListParams adalah parameter query untuk GET /students.
type StudentListParams struct {
	Page     int
	PerPage  int
	Prodi    string
	Angkatan int
	Search   string
	Sort     string
}