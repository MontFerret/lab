package runner

import (
	"github.com/MontFerret/lab/v2/pkg/sources"
)

type Settings struct {
	CDPAddress string
	Source     sources.Source
	Params     map[string]interface{}
}
