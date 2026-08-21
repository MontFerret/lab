package runner

import "time"

type (
	Result struct {
		Times    uint64
		Attempts uint64
		Filename string
		Duration time.Duration
		Error    error
		Warning  string
	}

	Summary struct {
		Passed   int
		Failed   int
		Duration time.Duration
	}
)

func (sum Summary) HasErrors() bool {
	return sum.Failed > 0
}
