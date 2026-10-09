package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"siakad-mini/app/model"
)

// UserRepository adalah kontrak akses data user. Service bergantung pada
// interface ini, bukan pada pgx, sehingga mudah diuji dengan repository palsu.
type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (model.User, error)
	FindByID(ctx context.Context, id int) (model.User, error)
	FindStudentProfile(ctx context.Context, userID int) (model.StudentProfile, error)
}

type userPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &userPostgresRepository{pool: pool}
}

// FindByEmail dipakai saat login. LEFT JOIN + "s.deleted_at IS NULL" membuat
// mahasiswa yang sudah di-soft delete tidak ditemukan, sehingga tidak bisa
// login. Admin tidak punya baris di students, jadi s.deleted_at tetap NULL.
func (r *userPostgresRepository) FindByEmail(ctx context.Context, email string) (model.User, error) {
	var u model.User
	err := r.pool.QueryRow(ctx,
		`SELECT u.id, u.email, u.password, u.role
		 FROM users u
		 LEFT JOIN students s ON s.user_id = u.id
		 WHERE LOWER(u.email) = LOWER($1) AND s.deleted_at IS NULL`,
		email,
	).Scan(&u.ID, &u.Email, &u.Password, &u.Role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrNotFound
		}
		return model.User{}, fmt.Errorf("mengambil user: %w", err)
	}
	return u, nil
}

func (r *userPostgresRepository) FindByID(ctx context.Context, id int) (model.User, error) {
	var u model.User
	err := r.pool.QueryRow(ctx,
		`SELECT u.id, u.email, u.password, u.role
		 FROM users u
		 LEFT JOIN students s ON s.user_id = u.id
		 WHERE u.id = $1 AND s.deleted_at IS NULL`,
		id,
	).Scan(&u.ID, &u.Email, &u.Password, &u.Role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrNotFound
		}
		return model.User{}, fmt.Errorf("mengambil user: %w", err)
	}
	return u, nil
}

func (r *userPostgresRepository) FindStudentProfile(ctx context.Context, userID int) (model.StudentProfile, error) {
	var p model.StudentProfile
	err := r.pool.QueryRow(ctx,
		`SELECT nim, nama, prodi, angkatan
		 FROM students
		 WHERE user_id = $1 AND deleted_at IS NULL`,
		userID,
	).Scan(&p.NIM, &p.Nama, &p.Prodi, &p.Angkatan)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.StudentProfile{}, ErrNotFound
		}
		return model.StudentProfile{}, fmt.Errorf("mengambil profil mahasiswa: %w", err)
	}
	return p, nil
}