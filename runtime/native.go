package runtime

import (
	"context"
	"github.com/MontFerret/lab/assertions"
	"os"

	"github.com/MontFerret/ferret/pkg/compiler"
	"github.com/MontFerret/ferret/pkg/drivers"
	"github.com/MontFerret/ferret/pkg/drivers/cdp"
	"github.com/MontFerret/ferret/pkg/drivers/http"
	"github.com/MontFerret/ferret/pkg/runtime"
	"github.com/rs/zerolog"
)

type Native struct {
	compiler *compiler.Compiler
	cdp      string
}

func NewNative(cdp string) *Native {
	c := compiler.New()

	assertions.Assertions(c.Namespace("T"))

	return &Native{c, cdp}
}

func (n *Native) Run(ctx context.Context, query string, params map[string]interface{}) ([]byte, error) {
	p, err := n.compiler.Compile(query)

	if err != nil {
		return nil, err
	}

	ctx = drivers.WithContext(
		ctx,
		cdp.NewDriver(cdp.WithAddress(n.cdp)),
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
