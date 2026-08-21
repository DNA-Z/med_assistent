package query

import "github.com/DNA-Z/med_assistent/internal/application/ports"

// Service объединяет обработчики запросов к read model.
// Реализация каждого запроса вынесена в отдельный файл пакета query.
type Service struct {
	readRepo ports.ExaminationReadRepository
	llm      ports.LLMClient
}

// NewService создаёт фасад запросов поверх read model и LLM-клиента.
func NewService(readRepo ports.ExaminationReadRepository, llm ports.LLMClient) *Service {
	return &Service{readRepo: readRepo, llm: llm}
}

var _ ports.ExaminationQueryHandler = (*Service)(nil)
