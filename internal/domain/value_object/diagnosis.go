package value_object

// Diagnosis - тип для хранения кода диагноза по МКБ-11
type Diagnosis int

const (
	// Раздел 6A20: Шизофрения
	SchizophreniaFirstEpisodeCurrent     Diagnosis = 6_20_00 // 6A20.0
	SchizophreniaFirstEpisodePartialRem  Diagnosis = 6_20_01 // 6A20.1
	SchizophreniaMultipleEpisodesCurrent Diagnosis = 6_20_10 // 6A20.10

	// Раздел 6A60: Биполярное расстройство I типа
	BipolarITypeCurrentManicNoPsycho   Diagnosis = 6_60_00 // 6A60.0
	BipolarITypeCurrentManicWithPsycho Diagnosis = 6_60_01 // 6A60.1
	BipolarITypeCurrentDepressionMild  Diagnosis = 6_60_03 // 6A60.3

	// Раздел 6B00: Тревожные расстройства
	GeneralizedAnxietyDisorder Diagnosis = 6_00_00 // 6B00
	PanicDisorder              Diagnosis = 6_01_00 // 6B01

	// Раздел 6D70: Нейрокогнитивные расстройства
	DementiaFrontotemporal         Diagnosis = 6_83_00 // 6D83
	DementiaDueToParkinsonDisease  Diagnosis = 6_85_00 // 6D85.0
	DementiaDueToMultipleSclerosis Diagnosis = 6_85_04 // 6D85.4
)

// String возвращает название диагноза на русском
func (d Diagnosis) String() string {
	names := map[Diagnosis]string{
		SchizophreniaFirstEpisodeCurrent:     "Шизофрения, первый эпизод, текущее состояние (6A20.0)",
		SchizophreniaFirstEpisodePartialRem:  "Шизофрения, первый эпизод, неполная ремиссия (6A20.1)",
		SchizophreniaMultipleEpisodesCurrent: "Шизофрения, множественные эпизоды, текущее состояние (6A20.10)",
		BipolarITypeCurrentManicNoPsycho:     "Биполярное расстройство I типа, текущий маниакальный эпизод, без психотических симптомов (6A60.0)",
		BipolarITypeCurrentManicWithPsycho:   "Биполярное расстройство I типа, текущий маниакальный эпизод, с психотическими симптомами (6A60.1)",
		BipolarITypeCurrentDepressionMild:    "Биполярное расстройство I типа, текущий депрессивный эпизод, легкий (6A60.3)",
		GeneralizedAnxietyDisorder:           "Генерализованное тревожное расстройство (6B00)",
		PanicDisorder:                        "Паническое расстройство (6B01)",
		DementiaFrontotemporal:               "Лобно-височная деменция (6D83)",
		DementiaDueToParkinsonDisease:        "Деменция вследствие болезни Паркинсона (6D85.0)",
		DementiaDueToMultipleSclerosis:       "Деменция вследствие рассеянного склероза (6D85.4)",
	}
	if name, ok := names[d]; ok {
		return name
	}
	return "Неизвестный диагноз по МКБ-11"
}
