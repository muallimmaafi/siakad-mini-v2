package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"siakad-mini/app/model"
)

type StudentRepository interface {
	FindAll(ctx context.Context, p model.StudentListParams) ([]model.Student, int, error)
	FindByID(ctx context.Context, id int) (model.Student, error)
	FindEnrollments(ctx context.Context, studentID int, tahun string) ([]model.CourseTaken, error)
	ExistsNIM(ctx context.Context, nim string) (bool, error)
	ExistsEmail(ctx context.Context, email string) (bool, error)
	Create(ctx context.Context, in model.CreateStudentRequest, passwordHash string) (model.Student, error)
	Update(ctx context.Context, id int, in model.UpdateStudentRequest) (model.Student, error)
	SoftDelete(ctx context.Context, id int) error
}

type studentPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewStudentRepository(pool *pgxpool.Pool) StudentRepository {
	return &studentPostgresRepository{pool: pool}
}

// ipk_terakhir::float8 memastikan NUMERIC terbaca aman sebagai float64.
const studentColumns = `id, user_id, nim, nama, prodi, angkatan, ipk_terakhir::float8`

func scanStudent(row pgx.Row) (model.Student, error) {
	var s model.Student
	err := row.Scan(&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan, &s.IPKTerakhir)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Student{}, ErrNotFound
		}
		return model.Student{}, fmt.Errorf("membaca mahasiswa: %w", err)
	}
	return s, nil
}

// FindAll: semua kondisi memakai parameter ($1, $2, ...) dan kolom pengurut
// dibatasi whitelist, sehingga input pengguna tidak pernah masuk langsung
// ke teks SQL. "deleted_at IS NULL" menyembunyikan mahasiswa yang di-soft delete.
func (r *studentPostgresRepository) FindAll(ctx context.Context, p model.StudentListParams) ([]model.Student, int, error) {
	conds := []string{"deleted_at IS NULL"}
	args := []any{}
	add := func(cond string, value any) {
		args = append(args, value)
		conds = append(conds, fmt.Sprintf(cond, len(args)))
	}

	if p.Prodi != "" {
		add("prodi ILIKE $%d", p.Prodi)
	}
	if p.Angkatan > 0 {
		add("angkatan = $%d", p.Angkatan)
	}
	if p.Search != "" {
		add("(nim ILIKE $%[1]d OR nama ILIKE $%[1]d)", "%"+p.Search+"%")
	}
	where := "WHERE " + strings.Join(conds, " AND ")

	var total int
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM students "+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("menghitung mahasiswa: %w", err)
	}

	order := "id ASC"
	switch p.Sort {
	case "nama":
		order = "nama ASC, id ASC"
	case "-ipk_terakhir":
		order = "ipk_terakhir DESC, id ASC"
	}

	args = append(args, p.PerPage, (p.Page-1)*p.PerPage)
	query := fmt.Sprintf(
		`SELECT %s FROM students %s ORDER BY %s LIMIT $%d OFFSET $%d`,
		studentColumns, where, order, len(args)-1, len(args),
	)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("mengambil daftar mahasiswa: %w", err)
	}
	defer rows.Close()

	students := []model.Student{}
	for rows.Next() {
		var s model.Student
		if err := rows.Scan(&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan, &s.IPKTerakhir); err != nil {
			return nil, 0, fmt.Errorf("membaca baris mahasiswa: %w", err)
		}
		students = append(students, s)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("membaca daftar mahasiswa: %w", err)
	}
	return students, total, nil
}

func (r *studentPostgresRepository) FindByID(ctx context.Context, id int) (model.Student, error) {
	return scanStudent(r.pool.QueryRow(ctx,
		`SELECT `+studentColumns+` FROM students WHERE id = $1 AND deleted_at IS NULL`, id))
}

// FindEnrollments mengambil mata kuliah yang diambil mahasiswa.
// tahun (opsional) menyaring ke satu tahun akademik.
func (r *studentPostgresRepository) FindEnrollments(ctx context.Context, studentID int, tahun string) ([]model.CourseTaken, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT e.id, c.id, c.kode_mk, c.nama_mk, c.sks, c.semester, e.tahun_akademik
		 FROM enrollments e
		 JOIN courses c ON c.id = e.course_id
		 WHERE e.student_id = $1 AND ($2::text = '' OR e.tahun_akademik = $2::text)
		 ORDER BY e.tahun_akademik, c.kode_mk`,
		studentID, tahun,
	)
	if err != nil {
		return nil, fmt.Errorf("mengambil KRS: %w", err)
	}
	defer rows.Close()

	courses := []model.CourseTaken{}
	for rows.Next() {
		var c model.CourseTaken
		if err := rows.Scan(&c.EnrollmentID, &c.CourseID, &c.KodeMK, &c.NamaMK, &c.SKS, &c.Semester, &c.TahunAkademik); err != nil {
			return nil, fmt.Errorf("membaca baris KRS: %w", err)
		}
		courses = append(courses, c)
	}
	return courses, rows.Err()
}

// ExistsNIM sengaja TIDAK menyaring deleted_at: NIM milik mahasiswa yang
// sudah di-soft delete tetap dianggap terpakai (unique constraint masih menghitungnya).
func (r *studentPostgresRepository) ExistsNIM(ctx context.Context, nim string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM students WHERE nim = $1)`, nim).Scan(&exists)
	return exists, err
}

func (r *studentPostgresRepository) ExistsEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM users WHERE LOWER(email) = LOWER($1))`, email).Scan(&exists)
	return exists, err
}

// Create membuat user dan student dalam SATU transaction.
// Kalau salah satu gagal, keduanya dibatalkan (rollback).
func (r *studentPostgresRepository) Create(ctx context.Context, in model.CreateStudentRequest, passwordHash string) (model.Student, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return model.Student{}, fmt.Errorf("memulai transaction: %w", err)
	}
	defer tx.Rollback(ctx) // tidak berefek kalau sudah di-Commit

	var userID int
	err = tx.QueryRow(ctx,
		`INSERT INTO users (email, password, role) VALUES ($1, $2, 'mahasiswa') RETURNING id`,
		in.Email, passwordHash,
	).Scan(&userID)
	if err != nil {
		return model.Student{}, mapWriteError(fmt.Errorf("membuat user: %w", err))
	}

	student, err := scanStudent(tx.QueryRow(ctx,
		`INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir)
		 VALUES ($1, $2, $3, $4, $5, COALESCE($6::numeric, 0))
		 RETURNING `+studentColumns,
		userID, in.NIM, in.Nama, in.Prodi, in.Angkatan, in.IPKTerakhir,
	))
	if err != nil {
		return model.Student{}, mapWriteError(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return model.Student{}, fmt.Errorf("commit transaction: %w", err)
	}
	return student, nil
}

func (r *studentPostgresRepository) Update(ctx context.Context, id int, in model.UpdateStudentRequest) (model.Student, error) {
	return scanStudent(r.pool.QueryRow(ctx,
		`UPDATE students
		 SET nama = $2, prodi = $3, angkatan = $4,
		     ipk_terakhir = COALESCE($5::numeric, ipk_terakhir)
		 WHERE id = $1 AND deleted_at IS NULL
		 RETURNING `+studentColumns,
		id, in.Nama, in.Prodi, in.Angkatan, in.IPKTerakhir,
	))
}

// SoftDelete mengisi deleted_at, barisnya tidak dihapus.
func (r *studentPostgresRepository) SoftDelete(ctx context.Context, id int) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE students SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		return fmt.Errorf("menghapus mahasiswa: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// mapWriteError menerjemahkan pelanggaran unique (kode 23505) menjadi ErrDuplicate.
func mapWriteError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return ErrDuplicate
	}
	return err
}