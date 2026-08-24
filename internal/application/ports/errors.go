package ports

import "errors"

var (
	ErrExaminationNotFound = errors.New("обследование не найдено")
	ErrDoctorNotFound      = errors.New("врач не найден")
	ErrInvalidCommand      = errors.New("некорректная команда")
	ErrEmptyQuestion       = errors.New("вопрос не может быть пустым")
	ErrEmptyKeyword        = errors.New("поисковая строка не может быть пустой")
	ErrFileRequired        = errors.New("требуется файл или транскрипция")
	ErrPatientRequired     = errors.New("не указан пациент")
	ErrProcessingQueueFull = errors.New("очередь обработки заполнена")
)
