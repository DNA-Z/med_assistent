package aggregate

import (
	"time"

	"github.com/DNA-Z/med_assistent/internal/domain/entity"
	"github.com/DNA-Z/med_assistent/internal/domain/value_object"
)

type ExaminationSheet struct {
	ExaminationData time.Time
	Doctor          entity.Doctor
	Patient         entity.Patient
	Transcript      value_object.Transcript
	BriefSummary    value_object.BriefSummary
	Diagnosis       value_object.Diagnosis
}
