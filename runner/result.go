package runner

import "time"

type (
	Result struct {
		Times    uint64
		Filename string
		Duration time.Duration
		Error    error
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
