package command

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/DNA-Z/med_assistent/internal/application/ports"
	"github.com/google/uuid"
)

type repositoryStub struct {
	mu                sync.Mutex
	completed, failed chan struct{}
}

func (r *repositoryStub) CreateExamination(context.Context, ports.ExaminationWriteModel, ports.ProcessingJobWriteModel) error {
	return nil
}
func (r *repositoryStub) StartProcessing(context.Context, uuid.UUID, uuid.UUID, time.Time, int) error {
	return nil
}
func (r *repositoryStub) SaveTranscript(context.Context, uuid.UUID, uuid.UUID, string, time.Time) error {
	return nil
}
func (r *repositoryStub) SaveSummary(context.Context, uuid.UUID, string, time.Time) error { return nil }
func (r *repositoryStub) CompleteProcessing(context.Context, uuid.UUID, uuid.UUID, time.Time) error {
	r.completed <- struct{}{}
	return nil
}
func (r *repositoryStub) FailProcessing(context.Context, uuid.UUID, uuid.UUID, string, time.Time) error {
	r.failed <- struct{}{}
	return nil
}
func (r *repositoryStub) RetryProcessing(context.Context, int64, uuid.UUID, uuid.UUID, time.Time) (uuid.UUID, string, string, error) {
	return uuid.New(), "transcript", "", nil
}
func (r *repositoryStub) DeleteExamination(context.Context, int64, uuid.UUID) error {
	return nil
}

type storageStub struct{ data []byte }

func (s *storageStub) Put(_ context.Context, object ports.StoredObject) error {
	data, err := io.ReadAll(object.Reader)
	s.data = data
	return err
}
func (s *storageStub) Open(context.Context, string) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader(string(s.data))), nil
}
func (s *storageStub) Delete(context.Context, string) error { return nil }

type speechStub struct{ err error }

func (s speechStub) Transcribe(_ context.Context, reader io.Reader, _ string) (string, error) {
	if s.err != nil {
		return "", s.err
	}
	data, err := io.ReadAll(reader)
	return string(data), err
}

type llmStub struct{}

func (llmStub) Summarize(context.Context, string) (string, error)      { return "summary", nil }
func (llmStub) Answer(context.Context, string, string) (string, error) { return "answer", nil }

func TestLoadProcessesInBackground(t *testing.T) {
	repo := &repositoryStub{completed: make(chan struct{}, 1), failed: make(chan struct{}, 1)}
	service := NewService(context.Background(), repo, speechStub{}, llmStub{}, &storageStub{}, slog.Default(), 1)
	defer service.Close()
	id, err := service.Load(context.Background(), ports.LoadExaminationCommand{DoctorID: 42, FileName: "test.txt", File: io.NopCloser(strings.NewReader("patient transcript"))})
	if err != nil || id == uuid.Nil {
		t.Fatalf("Load() id=%s err=%v", id, err)
	}
	select {
	case <-repo.completed:
	case <-time.After(time.Second):
		t.Fatal("processing did not complete")
	}
}

func TestLoadPersistsExternalClientFailure(t *testing.T) {
	repo := &repositoryStub{completed: make(chan struct{}, 1), failed: make(chan struct{}, 1)}
	service := NewService(context.Background(), repo, speechStub{err: errors.New("speech unavailable")}, llmStub{}, &storageStub{}, slog.Default(), 1)
	defer service.Close()
	_, err := service.Load(context.Background(), ports.LoadExaminationCommand{DoctorID: 42, File: io.NopCloser(strings.NewReader("audio"))})
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-repo.failed:
	case <-time.After(time.Second):
		t.Fatal("failure was not persisted")
	}
}
