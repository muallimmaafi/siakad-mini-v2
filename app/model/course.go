package model

type Course struct {
	ID        int    `json:"id"`
	KodeMK    string `json:"kode_mk"`
	NamaMK    string `json:"nama_mk"`
	SKS       int    `json:"sks"`
	Semester  int    `json:"semester"`
	Kuota     int    `json:"kuota"`
	Terisi    int    `json:"terisi"`
	SisaKuota int    `json:"sisa_kuota"`
}

// CourseListParams adalah parameter query untuk GET /courses.
type CourseListParams struct {
	Semester      int
	Search        string
	OnlyAvailable bool
	TahunAkademik string
}