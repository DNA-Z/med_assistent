// Package sqlqueries предоставляет SQL-запросы, встроенные в бинарный файл.
package sqlqueries

import (
	"embed"
	"fmt"
)

//go:embed *.sql
var files embed.FS

func mustRead(name string) string {
	query, err := files.ReadFile(name)
	if err != nil {
		panic(fmt.Sprintf("не удалось загрузить SQL-запрос %s: %v", name, err))
	}
	return string(query)
}

var (
	DoctorExists               = mustRead("doctor_exists.sql")
	DoctorCreate               = mustRead("doctor_create.sql")
	PatientCreate              = mustRead("patient_create.sql")
	ExaminationCreate          = mustRead("examination_create.sql")
	ProcessingJobCreate        = mustRead("processing_job_create.sql")
	OutboxCreate               = mustRead("outbox_create.sql")
	ExaminationStart           = mustRead("examination_start.sql")
	ProcessingJobStart         = mustRead("processing_job_start.sql")
	TranscriptUpsert           = mustRead("transcript_upsert.sql")
	ExaminationMarkTranscribed = mustRead("examination_mark_transcribed.sql")
	SummaryUpsert              = mustRead("summary_upsert.sql")
	ExaminationMarkSummarized  = mustRead("examination_mark_summarized.sql")
	ExaminationComplete        = mustRead("examination_complete.sql")
	ProcessingJobComplete      = mustRead("processing_job_complete.sql")
	ExaminationFail            = mustRead("examination_fail.sql")
	ProcessingJobFail          = mustRead("processing_job_fail.sql")
	ProcessingErrorCreate      = mustRead("processing_error_create.sql")
	RetryJobGet                = mustRead("retry_job_get.sql")
	ProcessingJobReset         = mustRead("processing_job_reset.sql")
	ExaminationReset           = mustRead("examination_reset.sql")
	ExaminationDelete          = mustRead("examination_delete.sql")
	PendingProcessingGet       = mustRead("pending_processing_get.sql")
	OutboxBatch                = mustRead("outbox_batch.sql")
	OutboxMarkProcessed        = mustRead("outbox_mark_processed.sql")
	ExaminationProjectionGet   = mustRead("examination_projection_get.sql")
)
