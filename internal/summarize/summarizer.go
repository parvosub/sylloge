package summarize

// Summarizer interface defines the contract for summarizing student notes
type Summarizer interface {
	// Summarize takes raw notes and returns a coherent summary
	Summarize(notes string) (string, error)
}
