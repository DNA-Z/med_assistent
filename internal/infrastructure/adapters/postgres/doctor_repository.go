package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/DNA-Z/med_assistent/internal/application/ports"
	"github.com/DNA-Z/med_assistent/internal/infrastructure/adapters/postgres/sqlqueries"
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
		sqlqueries.DoctorExists,
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
		sqlqueries.DoctorCreate,
		doctorID,
	)

	return err
}

var _ ports.DoctorWriteRepository = (*DoctorRepository)(nil)

var _ = pgx.ErrNoRows
