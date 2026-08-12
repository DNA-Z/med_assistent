package value_object

// Category - тип для категории врача
type Category int

const (
	FirstCategory Category = iota + 1
	SecondCategory
	HighestCategory
)

func (c Category) String() string {
	names := [...]string{
		"первая",
		"вторая",
		"высшая",
	}
	if c < FirstCategory || c > HighestCategory {
		return "неизвестная"
	}
	return names[c-1]
}
