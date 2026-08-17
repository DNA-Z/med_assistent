package redis

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/DNA-Z/med_assistent/internal/application/ports"
)

type ExaminationReadRepository struct {
	client *Client
}

func NewExaminationReadRepository(
	client *Client,
) *ExaminationReadRepository {
	return &ExaminationReadRepository{
		client: client,
	}
}

func examinationKey(id uuid.UUID) string {
	return "examination:" + id.String()
}

func doctorExaminationsKey(doctorID int64) string {
	return "doctor:" + strconv.FormatInt(doctorID, 10) + ":examinations"
}

func (r *ExaminationReadRepository) List(
	ctx context.Context,
	doctorID int64,
) ([]ports.ExaminationDTO, error) {
	ids, err := r.client.client.ZRevRange(
		ctx,
		doctorExaminationsKey(doctorID),
		0,
		-1,
	).Result()
	if err != nil {
		return nil, err
	}

	result := make([]ports.ExaminationDTO, 0, len(ids))

	for _, id := range ids {
		data, err := r.client.client.Get(
			ctx,
			examinationKey(uuid.MustParse(id)),
		).Bytes()
		if err != nil {
			continue
		}

		var item ports.ExaminationDTO

		if err := json.Unmarshal(data, &item); err != nil {
			return nil, err
		}

		if item.DoctorID == doctorID {
			result = append(result, item)
		}
	}

	return result, nil
}

func (r *ExaminationReadRepository) Get(
	ctx context.Context,
	doctorID int64,
	examinationID uuid.UUID,
) (*ports.ExaminationDTO, error) {
	data, err := r.client.client.Get(
		ctx,
		examinationKey(examinationID),
	).Bytes()
	if err != nil {
		return nil, ports.ErrExaminationNotFound
	}

	var result ports.ExaminationDTO

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	if result.DoctorID != doctorID {
		return nil, ports.ErrExaminationNotFound
	}

	return &result, nil
}

func (r *ExaminationReadRepository) Status(
	ctx context.Context,
	doctorID int64,
	examinationID uuid.UUID,
) (*ports.ExaminationStatusDTO, error) {
	item, err := r.Get(ctx, doctorID, examinationID)
	if err != nil {
		return nil, err
	}

	return &ports.ExaminationStatusDTO{
		ID:        item.ID,
		Status:    item.Status,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
		Error:     item.ErrorReason,
	}, nil
}

func (r *ExaminationReadRepository) Find(
	ctx context.Context,
	doctorID int64,
	keyword string,
) ([]ports.ExaminationDTO, error) {
	items, err := r.List(ctx, doctorID)
	if err != nil {
		return nil, err
	}

	keyword = strings.ToLower(
		strings.TrimSpace(keyword),
	)

	result := make([]ports.ExaminationDTO, 0)

	for _, item := range items {
		text := strings.ToLower(
			item.Transcript + " " + item.Summary,
		)

		if strings.Contains(text, keyword) {
			result = append(result, item)
		}
	}

	return result, nil
}

func (r *ExaminationReadRepository) ChatContext(
	ctx context.Context,
	doctorID int64,
	examinationID *uuid.UUID,
) ([]ports.ChatContextItem, error) {
	items, err := r.List(ctx, doctorID)
	if err != nil {
		return nil, err
	}

	result := make([]ports.ChatContextItem, 0)

	for _, item := range items {
		if examinationID != nil &&
			item.ID != *examinationID {
			continue
		}

		result = append(
			result,
			ports.ChatContextItem{
				ExaminationID: item.ID,
				Transcript:    item.Transcript,
				Summary:       item.Summary,
			},
		)
	}

	return result, nil
}
