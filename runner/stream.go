package runner

type Stream struct {
	Summary  <-chan Summary
	Progress <-chan Result
}
