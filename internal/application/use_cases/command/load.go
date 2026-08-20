package command

import (
	"context"
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
	examination, err := aggregate.NewExamination(examinationID, now, uuid.NewSHA1(uuid.Nil, []byte(doctorIdentity(cmd.DoctorID))), cmd.PatientID)
	if err != nil {
		return uuid.Nil, err
	}
	job, err := aggregate.NewProcessingJob(jobID, examinationID)
	if err != nil {
		return uuid.Nil, err
	}
	err = s.writeRepo.CreateExamination(ctx,
		ports.ExaminationWriteModel{ID: examination.ID(), DoctorID: cmd.DoctorID, PatientID: cmd.PatientID, ExaminationDate: examination.ExaminationDate(), Status: examination.Status().String(), CreatedAt: examination.CreatedAt(), UpdatedAt: examination.UpdatedAt()},
		ports.ProcessingJobWriteModel{ID: job.ID(), ExaminationID: job.ExaminationID(), Status: job.Status().String(), Attempt: job.Attempt(), CreatedAt: job.CreatedAt(), UpdatedAt: job.UpdatedAt()},
	)
	if err != nil {
		return uuid.Nil, err
	}
	file, transcript := cmd.File, cmd.Transcript
	s.processing.Go(func() error {
		if file != nil {
			defer file.Close()
		}
		s.process(examinationID, jobID, file, cmd.FileName, transcript)
		return nil
	})
	return examinationID, nil
}

func doctorIdentity(id int64) string { return time.Unix(id, 0).UTC().Format(time.RFC3339Nano) }
