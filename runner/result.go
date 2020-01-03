package runner

import "time"

type (
	Result struct {
		Filename string
		Duration time.Duration
		Error    error
	}

	Summary struct {
		Passed   int
		Failed   int
		Duration time.Duration
		Error    error
	}
)
