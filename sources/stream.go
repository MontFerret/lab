package sources

type Stream interface {
	OnNext() <-chan File
	OnError() <-chan error
}

type streamImpl struct {
	onNext  <-chan File
	onError <-chan error
}

func NewStream(onNext <-chan File, onError <-chan error) Stream {
	return &streamImpl{
		onNext,
		onError,
	}
}

func (stream *streamImpl) OnNext() <-chan File {
	return stream.onNext
}

func (stream *streamImpl) OnError() <-chan error {
	return stream.onError
}
