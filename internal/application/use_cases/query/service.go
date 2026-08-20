package query

import "github.com/DNA-Z/med_assistent/internal/application/ports"

// Service is a query facade. Each query is implemented in its own file.
type Service struct {
	readRepo ports.ExaminationReadRepository
	llm      ports.LLMClient
}

func NewService(readRepo ports.ExaminationReadRepository, llm ports.LLMClient) *Service {
	return &Service{readRepo: readRepo, llm: llm}
}

var _ ports.ExaminationQueryHandler = (*Service)(nil)
