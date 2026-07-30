package runtime

func pointerTo[T any](value T) *T {
	return &value
}
