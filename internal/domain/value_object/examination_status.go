package value_object

type ExaminationStatus string

const (
	ExaminationCreated     ExaminationStatus = "created"
	ExaminationProcessing  ExaminationStatus = "processing"
	ExaminationTranscribed ExaminationStatus = "transcribed"
	ExaminationSummarized  ExaminationStatus = "summarized"
	ExaminationCompleted   ExaminationStatus = "completed"
	ExaminationFailed      ExaminationStatus = "failed"
)

func (s ExaminationStatus) String() string {
	return string(s)
}
