package runtime

import "github.com/MontFerret/api"

func wireSessionSetters(opts []api.SessionOption) []api.SessionOption {
	setters := make([]api.SessionOption, len(opts))

	for index, option := range opts {
		if option == nil {
			continue
		}

		setters[index] = func(target api.SessionOptions) error {
			return option(&wireSessionOptions{SessionOptions: target})
		}
	}

	return setters
}
