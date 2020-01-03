package assertions

import (
	"context"
	"fmt"
	"github.com/MontFerret/ferret/pkg/runtime/values"

	"github.com/MontFerret/ferret/pkg/runtime/core"
)

func Assertions(ns core.Namespace) error {
	return ns.RegisterFunctions(
		core.Functions{
			"EXPECT": func(ctx context.Context, args ...core.Value) (core.Value, error) {
				err := core.ValidateArgs(args, 2, 2)

				if err != nil {
					return values.None, err
				}

				if args[0].Compare(args[1]) == 0 {
					return values.EmptyString, nil
				}

				return values.NewString(fmt.Sprintf(`expected "%s", but got "%s"`, args[0], args[1])), nil
			},
		},
	)
}
