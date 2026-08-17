package query

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"github.com/DNA-Z/med_assistent/internal/application/ports"
)

type Service struct {
	readRepo ports.ExaminationReadRepository
	llm      ports.LLMClient
}

func NewService(
	readRepo ports.ExaminationReadRepository,
	llm ports.LLMClient,
) *Service {
	return &Service{
		readRepo: readRepo,
		llm:      llm,
	}
}

func (s *Service) List(
	ctx context.Context,
	q ports.ListExaminationsQuery,
) ([]ports.ExaminationDTO, error) {
	return s.readRepo.List(ctx, q.DoctorID)
}

func (s *Service) Status(
	ctx context.Context,
	q ports.GetExaminationStatusQuery,
) (*ports.ExaminationStatusDTO, error) {
	return s.readRepo.Status(
		ctx,
		q.DoctorID,
		q.ExaminationID,
	)
}

func (s *Service) Get(
	ctx context.Context,
	q ports.GetExaminationQuery,
) (*ports.ExaminationDTO, error) {
	return s.readRepo.Get(
		ctx,
		q.DoctorID,
		q.ExaminationID,
	)
}

func (s *Service) Find(
	ctx context.Context,
	q ports.FindExaminationsQuery,
) ([]ports.ExaminationDTO, error) {
	if strings.TrimSpace(q.Keyword) == "" {
		return nil, ports.ErrEmptyKeyword
	}

	return s.readRepo.Find(
		ctx,
		q.DoctorID,
		q.Keyword,
	)
}

func (s *Service) Chat(
	ctx context.Context,
	q ports.ChatQuery,
) (string, error) {
	if strings.TrimSpace(q.Question) == "" {
		return "", ports.ErrEmptyQuestion
	}

	items, err := s.readRepo.ChatContext(
		ctx,
		q.DoctorID,
		q.ExaminationID,
	)
	if err != nil {
		return "", err
	}

	var contextText strings.Builder

	for _, item := range items {
		contextText.WriteString(
			"Обследование: ",
		)
		contextText.WriteString(
			item.ExaminationID.String(),
		)
		contextText.WriteString("\n")

		if item.Summary != "" {
			contextText.WriteString(
				"Выжимка:\n",
			)
			contextText.WriteString(item.Summary)
			contextText.WriteString("\n")
		}

		if item.Transcript != "" {
			contextText.WriteString(
				"Транскрипция:\n",
			)
			contextText.WriteString(item.Transcript)
			contextText.WriteString("\n")
		}
	}

	return s.llm.Answer(
		ctx,
		contextText.String(),
		q.Question,
	)
}

var _ ports.ExaminationQueryHandler = (*Service)(nil)

var _ = uuid.Nil
