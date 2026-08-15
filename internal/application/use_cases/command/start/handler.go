package start

import (
	"context"
	"errors"

	"github.com/DNA-Z/med_assistent/internal/application/auth"
)

var (
	ErrDoctorNotRegistered = errors.New("doctor is not registered")
)

type Command struct {
	TelegramID int64
	Name       string
	LastName   string
}

type Result struct {
	DoctorID    string
	Registered  bool
	NeedProfile bool
}

type Handler struct {
	authService *auth.Service
}

func NewHandler(
	authService *auth.Service,
) *Handler {
	return &Handler{
		authService: authService,
	}
}

func (h *Handler) Handle(
	ctx context.Context,
	cmd Command,
) (Result, error) {

	doctor, err := h.authService.Authenticate(
		ctx,
		auth.Identity{
			TelegramID: cmd.TelegramID,
			Name:       cmd.Name,
			LastName:   cmd.LastName,
		},
	)

	if err != nil {
		return Result{}, err
	}

	return Result{
		DoctorID:   doctor.ID().String(),
		Registered: true,
	}, nil
}
