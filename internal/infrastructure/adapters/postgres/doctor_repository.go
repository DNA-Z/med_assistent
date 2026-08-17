package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/DNA-Z/med_assistent/internal/application/ports"
)

type DoctorRepository struct {
	store *Store
}

func NewDoctorRepository(store *Store) *DoctorRepository {
	return &DoctorRepository{store: store}
}

func (r *DoctorRepository) Exists(
	ctx context.Context,
	doctorID int64,
) (bool, error) {
	var exists bool

	err := r.store.pool.QueryRow(
		ctx,
		`SELECT EXISTS (
			SELECT 1
			FROM doctors
			WHERE telegram_id = $1
		)`,
		doctorID,
	).Scan(&exists)

	return exists, err
}

func (r *DoctorRepository) Create(
	ctx context.Context,
	doctorID int64,
) error {
	_, err := r.store.pool.Exec(
		ctx,
		`INSERT INTO doctors (
			telegram_id,
			created_at
		)
		VALUES ($1, now())
		ON CONFLICT (telegram_id) DO NOTHING`,
		doctorID,
	)

	return err
}

var _ ports.DoctorWriteRepository = (*DoctorRepository)(nil)

var _ = pgx.ErrNoRows
