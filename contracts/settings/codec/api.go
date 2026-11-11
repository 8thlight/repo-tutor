package codec

type Values struct {
	TabStyle              string
	MarkdownOpensRendered bool
	Highlight             string
	QuizLength            int
	ExamLength            int
}

type Status struct {
	Unreadable bool
	ReadOnly   bool
}

type Choices struct {
	TabStyles   []string
	Highlights  []string
	QuizLengths []int
	ExamLengths []int
}

type Codec interface {
	Defaults() Values
	Choices() Choices
	Decode(data []byte) (Values, Status)
	Encode(values Values) []byte
}
