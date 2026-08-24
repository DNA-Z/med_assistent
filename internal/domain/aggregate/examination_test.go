package aggregate

import (
	"errors"
	"github.com/DNA-Z/med_assistent/internal/domain/value_object"
	"github.com/google/uuid"
	"testing"
	"time"
)

func TestExaminationLifecycle(t *testing.T) {
	examination, err := NewExamination(uuid.New(), time.Now(), uuid.New(), uuid.New())
	if err != nil {
		t.Fatal(err)
	}
	if err := examination.StartProcessing(); err != nil {
		t.Fatal(err)
	}
	transcript, _ := value_object.NewTranscript("patient transcript")
	if err := examination.SetTranscript(transcript); err != nil {
		t.Fatal(err)
	}
	summary, _ := value_object.NewBriefSummary("summary")
	if err := examination.SetSummary(summary); err != nil {
		t.Fatal(err)
	}
	if examination.Status() != value_object.ExaminationSummarized {
		t.Fatalf("статус=%s", examination.Status())
	}
}

func TestExaminationRejectsTranscriptBeforeProcessing(t *testing.T) {
	examination, _ := NewExamination(uuid.New(), time.Now(), uuid.New(), uuid.New())
	transcript, _ := value_object.NewTranscript("patient transcript")
	if err := examination.SetTranscript(transcript); !errors.Is(err, ErrExaminationNotProcessing) {
		t.Fatalf("ошибка=%v", err)
	}
}
