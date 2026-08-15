package create_examination

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/DNA-Z/med_assistent/internal/application/ports"
	"github.com/DNA-Z/med_assistent/internal/domain/aggregate"
	"github.com/google/uuid"
)

var (
	ErrDoctorNotFound = errors.New("doctor not found")
)

type Command struct {
	DoctorID uuid.UUID

	PatientID uuid.UUID

	Audio io.Reader

	ExaminationDate time.Time
}

type Result struct {
	ExaminationID uuid.UUID
	TaskID        uuid.UUID
}

type Handler struct {
	doctors      ports.DoctorRepository
	examinations ports.ExaminationRepository
	tasks        ports.ProcessingRepository
	storage      ports.FileStorage
	tx           ports.TransactionManager
}

func NewHandler(
	doctors ports.DoctorRepository,
	examinations ports.ExaminationRepository,
	tasks ports.ProcessingRepository,
	storage ports.FileStorage,
	tx ports.TransactionManager,
) *Handler {
	return &Handler{
		doctors:      doctors,
		examinations: examinations,
		tasks:        tasks,
		storage:      storage,
		tx:           tx,
	}
}

func (h *Handler) Handle(
	ctx context.Context,
	cmd Command,
) (Result, error) {

	doctor, err := h.doctors.GetByID(ctx, cmd.DoctorID)
	if err != nil {
		return Result{}, err
	}

	examinationID := uuid.New()
	taskID := uuid.New()

	examination := &aggregate.Examination{
		ID:              examinationID,
		ExaminationDate: cmd.ExaminationDate,
		Doctor:          *doctor,
		PatientID:       cmd.PatientID,
	}

	task := &ports.ProcessingTask{
		ID:            taskID,
		ExaminationID: examinationID,
		Status:        ports.ProcessingCreated,
	}

	err = h.tx.WithinTransaction(ctx, func(ctx context.Context) error {

		if err := h.examinations.Create(
			ctx,
			examination,
		); err != nil {
			return err
		}

		if err := h.tasks.Create(
			ctx,
			task,
		); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return Result{}, err
	}

	return Result{
		ExaminationID: examinationID,
		TaskID:        taskID,
	}, nil
}
