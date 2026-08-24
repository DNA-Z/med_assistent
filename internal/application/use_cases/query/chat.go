package query

import (
	"context"
	"github.com/DNA-Z/med_assistent/internal/application/ports"
	"strings"
)

func (s *Service) Chat(ctx context.Context, q ports.ChatQuery) (string, error) {
	if strings.TrimSpace(q.Question) == "" {
		return "", ports.ErrEmptyQuestion
	}
	items, err := s.readRepo.ChatContext(ctx, q.DoctorID, q.ExaminationID)
	if err != nil {
		return "", err
	}
	var contextText strings.Builder
	for _, item := range items {
		contextText.WriteString("Обследование: ")
		contextText.WriteString(item.ExaminationID.String())
		contextText.WriteString("\n")
		if item.Summary != "" {
			contextText.WriteString("Выжимка:\n")
			contextText.WriteString(item.Summary)
			contextText.WriteString("\n")
		}
		if item.Transcript != "" {
			contextText.WriteString("Транскрипция:\n")
			contextText.WriteString(item.Transcript)
			contextText.WriteString("\n")
		}
	}
	s.logger.Info("вызов LLM-клиента для ответа на вопрос", "doctor_id", q.DoctorID, "examinations_count", len(items))
	answer, err := s.llm.Answer(ctx, contextText.String(), q.Question)
	if err != nil {
		return "", err
	}
	s.logger.Info("LLM-клиент сформировал ответ", "doctor_id", q.DoctorID)
	return answer, nil
}
