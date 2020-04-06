package runtime

import (
	"context"
	"os"

	"github.com/MontFerret/ferret/pkg/compiler"
	"github.com/MontFerret/ferret/pkg/drivers"
	"github.com/MontFerret/ferret/pkg/drivers/cdp"
	"github.com/MontFerret/ferret/pkg/drivers/http"
	"github.com/MontFerret/ferret/pkg/runtime"
	"github.com/rs/zerolog"

	"github.com/MontFerret/lab/assertions"
)

type Builtin struct {
	compiler *compiler.Compiler
	cdp      string
}

func NewBuiltin(cdp string, _ map[string]interface{}) (*Builtin, error) {
	c := compiler.New()

	assertions.Assertions(c.Namespace("T"))

	return &Builtin{c, cdp}, nil
}

func (r *Builtin) Run(ctx context.Context, query string, params map[string]interface{}) ([]byte, error) {
	p, err := r.compiler.Compile(query)

	if err != nil {
		return nil, err
	}

	ctx = drivers.WithContext(
		ctx,
		cdp.NewDriver(cdp.WithAddress(r.cdp)),
	)

	ctx = drivers.WithContext(
		ctx,
		http.NewDriver(),
		drivers.AsDefault(),
	)

	return p.Run(
		ctx,
		runtime.WithLog(zerolog.ConsoleWriter{Out: os.Stdout}),
		runtime.WithParams(params),
	)
}
