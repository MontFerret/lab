package runtime

import (
	"context"
	"fmt"
	"os"

	"github.com/MontFerret/ferret/v2"
	ferretnet "github.com/MontFerret/ferret/v2/pkg/net"
	ferrethttp "github.com/MontFerret/ferret/v2/pkg/net/http"
	"github.com/MontFerret/ferret/v2/pkg/source"
)

var version = "unknown"

type Builtin struct {
	engine  *ferret.Engine
	network ferretnet.Network
}

func NewBuiltin(params map[string]any, policyOptions ...ferrethttp.PolicyOption) (*Builtin, error) {
	return newBuiltin(params, nil, policyOptions...)
}

func newBuiltin(params map[string]any, fsPolicy *FileSystemPolicy, policyOptions ...ferrethttp.PolicyOption) (*Builtin, error) {
	if fsPolicy == nil && len(policyOptions) == 0 {
		return newDefaultBuiltin(params)
	}

	root := ""
	if fsPolicy != nil {
		root = fsPolicy.Root
	}

	if root == "" {
		var err error
		root, err = os.Getwd()
		if err != nil {
			return nil, err
		}
	}

	engineOptions := []ferret.Option{
		ferret.WithFSRoot(root),
		ferret.WithParams(params),
	}

	if fsPolicy != nil && fsPolicy.ReadOnly != nil && *fsPolicy.ReadOnly {
		engineOptions = append(engineOptions, ferret.WithFSReadOnly())
	}

	if len(policyOptions) == 0 {
		engine, err := ferret.New(engineOptions...)
		if err != nil {
			if fsPolicy != nil {
				return nil, fmt.Errorf("filesystem policy: %w", err)
			}

			return nil, err
		}

		return &Builtin{engine: engine}, nil
	}

	client, err := ferrethttp.New(policyOptions...)
	if err != nil {
		return nil, fmt.Errorf("HTTP policy: %w", err)
	}

	network, err := ferretnet.New(ferretnet.WithHTTPClient(client))
	if err != nil {
		if closer, ok := client.(ferrethttp.IdleConnectionCloser); ok {
			closer.CloseIdleConnections()
		}

		return nil, fmt.Errorf("network: %w", err)
	}

	engineOptions = append(engineOptions, ferret.WithNetwork(network))
	engine, err := ferret.New(engineOptions...)

	if err != nil {
		ferretnet.CloseIdleNetworkConnections(network)
		if fsPolicy != nil {
			return nil, fmt.Errorf("filesystem policy: %w", err)
		}

		return nil, err
	}

	return &Builtin{
		engine:  engine,
		network: network,
	}, nil
}

func newDefaultBuiltin(params map[string]any) (*Builtin, error) {
	root, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	engine, err := ferret.New(
		ferret.WithFSRoot(root),
		ferret.WithParams(params),
	)
	if err != nil {
		return nil, err
	}

	return &Builtin{engine: engine}, nil
}

func (r *Builtin) Version(_ context.Context) (string, error) {
	return version, nil
}

func (r *Builtin) Run(ctx context.Context, query *source.Source, params map[string]any) ([]byte, error) {
	out, err := r.engine.Run(ctx, query, ferret.WithSessionParams(params))

	if err != nil {
		return nil, err
	}

	return out.Content, nil
}

// Close shuts down the embedded engine and its configured HTTP network.
func (r *Builtin) Close() error {
	err := r.engine.Close()

	if r.network != nil {
		ferretnet.CloseIdleNetworkConnections(r.network)
	}

	return err
}
