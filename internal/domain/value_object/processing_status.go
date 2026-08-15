package value_object

type ProcessingStatus string

const (
	ProcessingCreated    ProcessingStatus = "created"
	ProcessingProcessing ProcessingStatus = "processing"
	ProcessingCompleted  ProcessingStatus = "completed"
	ProcessingFailed     ProcessingStatus = "failed"
)

func (s ProcessingStatus) String() string {
	return string(s)
}
