package domain

type Event struct {
	ID          string
	SourceID    string
	Title       string
	Description string
	Embedding   []float32
}
