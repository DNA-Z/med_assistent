package redis

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/DNA-Z/med_assistent/internal/application/ports"
)

func ExaminationKey(id uuid.UUID) string {
	return "read:examination:" + id.String()
}

func DoctorExaminationsKey(doctorID int64) string {
	return "read:doctor:" +
		strconv.FormatInt(doctorID, 10) +
		":examinations"
}

func SaveExamination(
	ctx context.Context,
	client *Client,
	item ports.ExaminationDTO,
) error {
	data, err := json.Marshal(item)
	if err != nil {
		return err
	}

	pipe := client.client.TxPipeline()

	pipe.Set(
		ctx,
		ExaminationKey(item.ID),
		data,
		0,
	)

	pipe.ZAdd(
		ctx,
		DoctorExaminationsKey(item.DoctorID),
		redis.Z{
			Score:  float64(item.CreatedAt.Unix()),
			Member: item.ID.String(),
		},
	)

	_, err = pipe.Exec(ctx)

	return err
}
