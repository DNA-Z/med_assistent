package query

import (
	"context"
	"errors"
	"github.com/DNA-Z/med_assistent/internal/application/ports"
	"github.com/google/uuid"
	"log/slog"
	"testing"
)

type readStub struct{ doctorID int64 }

func (r *readStub) List(context.Context, int64) ([]ports.ExaminationDTO, error) { return nil, nil }
func (r *readStub) Get(_ context.Context, doctorID int64, _ uuid.UUID) (*ports.ExaminationDTO, error) {
	r.doctorID = doctorID
	return nil, ports.ErrExaminationNotFound
}
func (r *readStub) Status(context.Context, int64, uuid.UUID) (*ports.ExaminationStatusDTO, error) {
	return nil, nil
}
func (r *readStub) Find(context.Context, int64, string) ([]ports.ExaminationDTO, error) {
	return nil, nil
}
func (r *readStub) ChatContext(context.Context, int64, *uuid.UUID) ([]ports.ChatContextItem, error) {
	return nil, nil
}

type llmStub struct{}

func (llmStub) Summarize(context.Context, string) (string, error)      { return "", nil }
func (llmStub) Answer(context.Context, string, string) (string, error) { return "", nil }

func TestGetPassesCurrentDoctorToRepository(t *testing.T) {
	repo := &readStub{}
	_, err := NewService(repo, llmStub{}, slog.Default()).Get(context.Background(), ports.GetExaminationQuery{DoctorID: 77, ExaminationID: uuid.New()})
	if !errors.Is(err, ports.ErrExaminationNotFound) || repo.doctorID != 77 {
		t.Fatalf("врач=%d, ошибка=%v", repo.doctorID, err)
	}
}

func TestFindRejectsEmptyKeyword(t *testing.T) {
	_, err := NewService(&readStub{}, llmStub{}, slog.Default()).Find(context.Background(), ports.FindExaminationsQuery{DoctorID: 1, Keyword: "  "})
	if !errors.Is(err, ports.ErrEmptyKeyword) {
		t.Fatalf("ошибка=%v", err)
	}
}
