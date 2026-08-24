package redis

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"

	"github.com/DNA-Z/med_assistent/internal/application/ports"
)

func ProjectExamination(
	ctx context.Context,
	client *Client,
	item ports.ExaminationDTO,
) error {
	data, err := json.Marshal(item)
	if err != nil {
		return err
	}

	key := examinationKey(item.ID)
	doctorKey := doctorExaminationsKey(item.DoctorID)

	pipe := client.client.TxPipeline()

	pipe.Set(
		ctx,
		key,
		data,
		0,
	)

	pipe.ZAdd(
		ctx,
		doctorKey,
		struct {
			Score  float64
			Member any
		}{
			Score:  float64(item.CreatedAt.Unix()),
			Member: item.ID.String(),
		},
	)

	_, err = pipe.Exec(ctx)

	return err
}

func RemoveExamination(
	ctx context.Context,
	client *Client,
	doctorID int64,
	id uuid.UUID,
) error {
	pipe := client.client.TxPipeline()

	pipe.Del(ctx, examinationKey(id))
	pipe.ZRem(
		ctx,
		doctorExaminationsKey(doctorID),
		id.String(),
	)

	_, err := pipe.Exec(ctx)

	return err
}
