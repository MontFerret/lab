package sources

type Stream struct {
	Files  <-chan File
	Errors <-chan error
}
