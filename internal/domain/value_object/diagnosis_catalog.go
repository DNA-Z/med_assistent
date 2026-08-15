package value_object

var (
	SchizophreniaFirstEpisodeCurrent = mustDiagnosis(
		"6A20.0",
		"Шизофрения, первый эпизод, текущее состояние",
	)

	SchizophreniaFirstEpisodePartialRemission = mustDiagnosis(
		"6A20.1",
		"Шизофрения, первый эпизод, неполная ремиссия",
	)

	SchizophreniaMultipleEpisodesCurrent = mustDiagnosis(
		"6A20.10",
		"Шизофрения, множественные эпизоды, текущее состояние",
	)

	BipolarITypeCurrentManicNoPsychosis = mustDiagnosis(
		"6A60.0",
		"Биполярное расстройство I типа, текущий маниакальный эпизод, без психотических симптомов",
	)

	BipolarITypeCurrentManicWithPsychosis = mustDiagnosis(
		"6A60.1",
		"Биполярное расстройство I типа, текущий маниакальный эпизод, с психотическими симптомами",
	)

	BipolarITypeCurrentDepressionMild = mustDiagnosis(
		"6A60.3",
		"Биполярное расстройство I типа, текущий депрессивный эпизод, легкий",
	)

	GeneralizedAnxietyDisorder = mustDiagnosis(
		"6B00",
		"Генерализованное тревожное расстройство",
	)

	PanicDisorder = mustDiagnosis(
		"6B01",
		"Паническое расстройство",
	)

	DementiaFrontotemporal = mustDiagnosis(
		"6D83",
		"Лобно-височная деменция",
	)

	DementiaDueToParkinsonDisease = mustDiagnosis(
		"6D85.0",
		"Деменция вследствие болезни Паркинсона",
	)

	DementiaDueToMultipleSclerosis = mustDiagnosis(
		"6D85.4",
		"Деменция вследствие рассеянного склероза",
	)
)

func mustDiagnosis(code string, description string) Diagnosis {
	icdCode, err := NewICDCode(code)
	if err != nil {
		panic(err)
	}

	diagnosis, err := NewDiagnosis(icdCode, description)
	if err != nil {
		panic(err)
	}

	return diagnosis
}
