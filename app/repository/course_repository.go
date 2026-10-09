package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"siakad-mini/app/model"
)

type CourseRepository interface {
	FindAll(ctx context.Context, p model.CourseListParams) ([]model.Course, error)
}

type coursePostgresRepository struct {
	pool *pgxpool.Pool
}

func NewCourseRepository(pool *pgxpool.Pool) CourseRepository {
	return &coursePostgresRepository{pool: pool}
}

// FindAll mengambil daftar mata kuliah beserta jumlah yang sudah mengambil
// (terisi). Semua nilai dari pengguna dikirim sebagai parameter, bukan
// digabung ke teks SQL.
func (r *coursePostgresRepository) FindAll(ctx context.Context, p model.CourseListParams) ([]model.Course, error) {
	args := []any{}

	// Syarat tahun akademik masuk ke JOIN, supaya mata kuliah tanpa peserta
	// tetap muncul dengan terisi = 0.
	join := "LEFT JOIN enrollments e ON e.course_id = c.id"
	if p.TahunAkademik != "" {
		args = append(args, p.TahunAkademik)
		join += fmt.Sprintf(" AND e.tahun_akademik = $%d", len(args))
	}

	conds := []string{}
	if p.Semester > 0 {
		args = append(args, p.Semester)
		conds = append(conds, fmt.Sprintf("c.semester = $%d", len(args)))
	}
	if p.Search != "" {
		args = append(args, "%"+p.Search+"%")
		conds = append(conds, fmt.Sprintf("(c.kode_mk ILIKE $%[1]d OR c.nama_mk ILIKE $%[1]d)", len(args)))
	}

	where := ""
	if len(conds) > 0 {
		where = "WHERE " + strings.Join(conds, " AND ")
	}

	// available=true: hanya yang jumlah pengambilnya masih di bawah kuota.
	having := ""
	if p.OnlyAvailable {
		having = "HAVING COUNT(e.id) < c.kuota"
	}

	query := fmt.Sprintf(
		`SELECT c.id, c.kode_mk, c.nama_mk, c.sks, c.semester, c.kuota, COUNT(e.id)
		 FROM courses c
		 %s
		 %s
		 GROUP BY c.id
		 %s
		 ORDER BY c.semester ASC, c.kode_mk ASC`,
		join, where, having,
	)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("mengambil daftar mata kuliah: %w", err)
	}
	defer rows.Close()

	courses := []model.Course{}
	for rows.Next() {
		var c model.Course
		if err := rows.Scan(&c.ID, &c.KodeMK, &c.NamaMK, &c.SKS, &c.Semester, &c.Kuota, &c.Terisi); err != nil {
			return nil, fmt.Errorf("membaca baris mata kuliah: %w", err)
		}
		courses = append(courses, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("membaca daftar mata kuliah: %w", err)
	}
	return courses, nil
}