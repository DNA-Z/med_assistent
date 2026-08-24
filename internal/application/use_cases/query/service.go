package query

import (
	"log/slog"

	"github.com/DNA-Z/med_assistent/internal/application/ports"
)

// Service объединяет обработчики запросов к read model.
// Реализация каждого запроса вынесена в отдельный файл пакета query.
type Service struct {
	readRepo ports.ExaminationReadRepository
	llm      ports.LLMClient
	logger   *slog.Logger
}

// NewService создаёт фасад запросов поверх read model и LLM-клиента.
func NewService(readRepo ports.ExaminationReadRepository, llm ports.LLMClient, logger *slog.Logger) *Service {
	return &Service{readRepo: readRepo, llm: llm, logger: logger}
}

var _ ports.ExaminationQueryHandler = (*Service)(nil)
