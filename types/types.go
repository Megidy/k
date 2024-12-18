package types

type Question struct {
	Id            string
	Question      string
	Answers       []string
	CorrectAnswer string
}
type Message struct {
	Headers map[string]any
	Answer  map[string]string
}
