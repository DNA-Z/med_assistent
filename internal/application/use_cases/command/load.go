package command

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/DNA-Z/med_assistent/internal/application/ports"
	"github.com/DNA-Z/med_assistent/internal/domain/aggregate"
	"github.com/google/uuid"
)

func (s *Service) Load(ctx context.Context, cmd ports.LoadExaminationCommand) (uuid.UUID, error) {
	if cmd.DoctorID == 0 {
		return uuid.Nil, ports.ErrDoctorNotFound
	}
	if cmd.File == nil && cmd.Transcript == nil {
		return uuid.Nil, ports.ErrFileRequired
	}
	now := time.Now().UTC()
	if cmd.PatientID == uuid.Nil {
		cmd.PatientID = uuid.New()
	}
	examinationID, jobID := uuid.New(), uuid.New()
	objectKey := ""
	if cmd.File != nil {
		extension := strings.ToLower(filepath.Ext(cmd.FileName))
		objectKey = fmt.Sprintf("examinations/%s/source%s", examinationID, extension)
		if err := s.storage.Put(ctx, ports.StoredObject{Key: objectKey, Reader: cmd.File, Size: cmd.FileSize, ContentType: cmd.ContentType}); err != nil {
			_ = cmd.File.Close()
			return uuid.Nil, err
		}
		_ = cmd.File.Close()
	}
	examination, err := aggregate.NewExamination(examinationID, now, uuid.NewSHA1(uuid.Nil, []byte(doctorIdentity(cmd.DoctorID))), cmd.PatientID)
	if err != nil {
		return uuid.Nil, err
	}
	job, err := aggregate.NewProcessingJob(jobID, examinationID)
	if err != nil {
		return uuid.Nil, err
	}
	err = s.writeRepo.CreateExamination(ctx,
		ports.ExaminationWriteModel{ID: examination.ID(), DoctorID: cmd.DoctorID, PatientID: cmd.PatientID, ExaminationDate: examination.ExaminationDate(), Status: examination.Status().String(), CreatedAt: examination.CreatedAt(), UpdatedAt: examination.UpdatedAt(), AudioObjectKey: objectKey, AudioFileName: cmd.FileName, AudioContentType: cmd.ContentType, AudioSize: cmd.FileSize},
		ports.ProcessingJobWriteModel{ID: job.ID(), ExaminationID: job.ExaminationID(), Status: job.Status().String(), Attempt: job.Attempt(), CreatedAt: job.CreatedAt(), UpdatedAt: job.UpdatedAt()},
	)
	if err != nil {
		if objectKey != "" {
			cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = s.storage.Delete(cleanupCtx, objectKey)
		}
		return uuid.Nil, err
	}
	s.logger.Info(
		"обследование и задача обработки созданы",
		"doctor_id", cmd.DoctorID,
		"examination_id", examinationID,
		"job_id", jobID,
		"file_name", cmd.FileName,
	)
	transcript := cmd.Transcript
	s.processing.Go(func() error {
		s.process(examinationID, jobID, objectKey, cmd.FileName, transcript)
		return nil
	})
	return examinationID, nil
}

func doctorIdentity(id int64) string { return time.Unix(id, 0).UTC().Format(time.RFC3339Nano) }
