package runtime

import (
	"context"
	"os"

	"github.com/MontFerret/ferret"
	"github.com/MontFerret/ferret/pkg/drivers"
	"github.com/MontFerret/ferret/pkg/drivers/cdp"
	"github.com/MontFerret/ferret/pkg/drivers/http"
	"github.com/MontFerret/ferret/pkg/runtime"
	"github.com/rs/zerolog"
)

var version = "unknown"

type Builtin struct {
	compiler *ferret.Instance
	cdp      string
}

func NewBuiltin(cdp string, _ map[string]interface{}) (*Builtin, error) {
	c := ferret.New()

	return &Builtin{c, cdp}, nil
}

func (r *Builtin) Version(_ context.Context) (string, error) {
	return version, nil
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
