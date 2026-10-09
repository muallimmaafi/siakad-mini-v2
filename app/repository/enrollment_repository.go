package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"siakad-mini/app/model"
)

type EnrollmentRepository interface {
	// Enroll mengambil satu mata kuliah untuk mahasiswa yang sedang login.
	// Semua pengecekan dan penyimpanan dilakukan dalam satu transaction.
	// Aturan hitung batas SKS dititipkan lewat parameter batasSKS (diisi service).
	Enroll(ctx context.Context, userID, courseID int, tahun string, batasSKS func(ipk float64) int) (model.EnrollmentResult, error)
	FindOwnerUserID(ctx context.Context, enrollmentID int) (int, error)
	Delete(ctx context.Context, enrollmentID int) error
}

type enrollmentPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewEnrollmentRepository(pool *pgxpool.Pool) EnrollmentRepository {
	return &enrollmentPostgresRepository{pool: pool}
}

func (r *enrollmentPostgresRepository) Enroll(ctx context.Context, userID, courseID int, tahun string, batasSKS func(ipk float64) int) (model.EnrollmentResult, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return model.EnrollmentResult{}, fmt.Errorf("memulai transaction: %w", err)
	}
	defer tx.Rollback(ctx) // tidak berefek kalau sudah di-Commit

	// 1. Kunci baris mahasiswa milik user yang login. Dua request dari
	//    mahasiswa yang sama jadi diproses bergantian, sehingga tidak bisa
	//    sama-sama lolos batas SKS.
	var studentID int
	var ipk float64
	err = tx.QueryRow(ctx,
		`SELECT id, ipk_terakhir::float8
		 FROM students
		 WHERE user_id = $1 AND deleted_at IS NULL
		 FOR UPDATE`, userID,
	).Scan(&studentID, &ipk)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.EnrollmentResult{}, ErrStudentNotFound
		}
		return model.EnrollmentResult{}, fmt.Errorf("mengunci mahasiswa: %w", err)
	}

	// 2. Kunci baris mata kuliah (row locking). Dua mahasiswa yang berebut
	//    kursi terakhir diproses antri, bukan sama-sama lolos.
	var course model.EnrollmentCourse
	var kuota int
	err = tx.QueryRow(ctx,
		`SELECT kode_mk, nama_mk, sks, semester, kuota
		 FROM courses
		 WHERE id = $1
		 FOR UPDATE`, courseID,
	).Scan(&course.KodeMK, &course.NamaMK, &course.SKS, &course.Semester, &kuota)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.EnrollmentResult{}, ErrCourseNotFound
		}
		return model.EnrollmentResult{}, fmt.Errorf("mengunci mata kuliah: %w", err)
	}

	// 3. Duplikasi
	var sudahAda bool
	err = tx.QueryRow(ctx,
		`SELECT EXISTS (
		   SELECT 1 FROM enrollments
		   WHERE student_id = $1 AND course_id = $2 AND tahun_akademik = $3)`,
		studentID, courseID, tahun,
	).Scan(&sudahAda)
	if err != nil {
		return model.EnrollmentResult{}, fmt.Errorf("cek duplikasi: %w", err)
	}
	if sudahAda {
		return model.EnrollmentResult{}, ErrAlreadyEnrolled
	}

	// 4. Kuota (dihitung per tahun akademik)
	var terisi int
	err = tx.QueryRow(ctx,
		`SELECT COUNT(*) FROM enrollments WHERE course_id = $1 AND tahun_akademik = $2`,
		courseID, tahun,
	).Scan(&terisi)
	if err != nil {
		return model.EnrollmentResult{}, fmt.Errorf("cek kuota: %w", err)
	}
	if terisi >= kuota {
		return model.EnrollmentResult{}, ErrCourseFull
	}

	// 5. Batas SKS (dihitung per tahun akademik)
	var diambil int
	err = tx.QueryRow(ctx,
		`SELECT COALESCE(SUM(c.sks), 0)
		 FROM enrollments e
		 JOIN courses c ON c.id = e.course_id
		 WHERE e.student_id = $1 AND e.tahun_akademik = $2`,
		studentID, tahun,
	).Scan(&diambil)
	if err != nil {
		return model.EnrollmentResult{}, fmt.Errorf("cek total SKS: %w", err)
	}
	batas := batasSKS(ipk)
	if diambil+course.SKS > batas {
		sisa := batas - diambil
		if sisa < 0 {
			sisa = 0
		}
		return model.EnrollmentResult{}, &SKSExceededError{
			Batas: batas, Diambil: diambil, Sisa: sisa, SKSMatkul: course.SKS,
		}
	}

	// 6. Simpan
	var e model.Enrollment
	err = tx.QueryRow(ctx,
		`INSERT INTO enrollments (student_id, course_id, tahun_akademik)
		 VALUES ($1, $2, $3)
		 RETURNING id, student_id, course_id, tahun_akademik, created_at`,
		studentID, courseID, tahun,
	).Scan(&e.ID, &e.StudentID, &e.CourseID, &e.TahunAkademik, &e.CreatedAt)
	if err != nil {
		// Jaring pengaman: unique constraint tetap menjaga duplikasi di database.
		if errors.Is(mapWriteError(err), ErrDuplicate) {
			return model.EnrollmentResult{}, ErrAlreadyEnrolled
		}
		return model.EnrollmentResult{}, fmt.Errorf("menyimpan KRS: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return model.EnrollmentResult{}, fmt.Errorf("commit transaction: %w", err)
	}

	total := diambil + course.SKS
	return model.EnrollmentResult{
		Enrollment: e,
		Course:     course,
		TotalSKS:   total,
		BatasSKS:   batas,
		SisaSKS:    batas - total,
	}, nil
}

// FindOwnerUserID mengembalikan user_id pemilik catatan KRS (dipakai endpoint 10).
func (r *enrollmentPostgresRepository) FindOwnerUserID(ctx context.Context, enrollmentID int) (int, error) {
	var userID int
	err := r.pool.QueryRow(ctx,
		`SELECT s.user_id
		 FROM enrollments e
		 JOIN students s ON s.id = e.student_id
		 WHERE e.id = $1`, enrollmentID,
	).Scan(&userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrNotFound
		}
		return 0, fmt.Errorf("mencari pemilik KRS: %w", err)
	}
	return userID, nil
}

// Delete menghapus catatan KRS (dipakai endpoint 10). Kuota otomatis kembali
// karena jumlah peserta dihitung dari isi tabel enrollments.
func (r *enrollmentPostgresRepository) Delete(ctx context.Context, enrollmentID int) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM enrollments WHERE id = $1`, enrollmentID)
	if err != nil {
		return fmt.Errorf("menghapus KRS: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}