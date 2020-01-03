package runner

import "github.com/MontFerret/lab/sources"

type Settings struct {
	CDPAddress string
	Source     sources.Source
	Params     map[string]interface{}
}
