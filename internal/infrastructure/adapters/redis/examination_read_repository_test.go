package redis

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DNA-Z/med_assistent/internal/application/ports"
	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
)

// TestExaminationReadRepositoryIsolatesDoctors проверяет, что все операции
// чтения возвращают врачу только принадлежащие ему обследования.
func TestExaminationReadRepositoryIsolatesDoctors(t *testing.T) {
	t.Parallel()

	server := miniredis.RunT(t)
	redisClient := goredis.NewClient(&goredis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })
	client := &Client{client: redisClient}
	repository := NewExaminationReadRepository(client)
	ctx := context.Background()

	const doctorA int64 = 1001
	const doctorB int64 = 2002
	now := time.Now().UTC()
	examinationA := ports.ExaminationDTO{
		ID:         uuid.New(),
		DoctorID:   doctorA,
		PatientID:  uuid.New(),
		Status:     "completed",
		Transcript: "секретная запись врача А",
		Summary:    "конфиденциальная выжимка",
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	examinationB := ports.ExaminationDTO{
		ID:         uuid.New(),
		DoctorID:   doctorB,
		PatientID:  uuid.New(),
		Status:     "completed",
		Transcript: "запись врача Б",
		Summary:    "обычная выжимка",
		CreatedAt:  now.Add(time.Second),
		UpdatedAt:  now.Add(time.Second),
	}

	if err := ProjectExamination(ctx, client, examinationA); err != nil {
		t.Fatalf("не удалось создать проекцию врача А: %v", err)
	}
	if err := ProjectExamination(ctx, client, examinationB); err != nil {
		t.Fatalf("не удалось создать проекцию врача Б: %v", err)
	}

	// Даже при ошибочном попадании чужого ID в индекс врача дополнительная
	// проверка DoctorID не должна пропустить документ в результат.
	if err := redisClient.ZAdd(ctx, doctorExaminationsKey(doctorB), goredis.Z{
		Score:  float64(now.Unix()),
		Member: examinationA.ID.String(),
	}).Err(); err != nil {
		t.Fatalf("не удалось подготовить повреждённый индекс: %v", err)
	}

	items, err := repository.List(ctx, doctorB)
	if err != nil {
		t.Fatalf("ошибка List(): %v", err)
	}
	if len(items) != 1 || items[0].ID != examinationB.ID {
		t.Fatalf("метод List() раскрыл чужие данные: %+v", items)
	}

	if _, err := repository.Get(ctx, doctorB, examinationA.ID); !errors.Is(err, ports.ErrExaminationNotFound) {
		t.Fatalf("метод Get() должен скрывать чужую встречу, получено: %v", err)
	}
	if _, err := repository.Status(ctx, doctorB, examinationA.ID); !errors.Is(err, ports.ErrExaminationNotFound) {
		t.Fatalf("метод Status() должен скрывать чужую встречу, получено: %v", err)
	}

	found, err := repository.Find(ctx, doctorB, "секретная")
	if err != nil {
		t.Fatalf("ошибка Find(): %v", err)
	}
	if len(found) != 0 {
		t.Fatalf("метод Find() раскрыл чужие данные: %+v", found)
	}

	chatItems, err := repository.ChatContext(ctx, doctorB, nil)
	if err != nil {
		t.Fatalf("ошибка ChatContext(): %v", err)
	}
	if len(chatItems) != 1 || chatItems[0].ExaminationID != examinationB.ID {
		t.Fatalf("метод ChatContext() раскрыл чужие данные: %+v", chatItems)
	}

	chatItems, err = repository.ChatContext(ctx, doctorB, &examinationA.ID)
	if err != nil {
		t.Fatalf("ошибка ChatContext() для чужой встречи: %v", err)
	}
	if len(chatItems) != 0 {
		t.Fatalf("метод ChatContext() вернул чужую встречу: %+v", chatItems)
	}
}
