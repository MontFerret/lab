package sources

type Stream struct {
	Files <-chan File
	Error <-chan error
}
