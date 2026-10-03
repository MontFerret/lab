package runtime

import (
	"fmt"
	"os"
	"sync"

	"github.com/MontFerret/api"
	"github.com/MontFerret/ferret/v2"
	ferretnet "github.com/MontFerret/ferret/v2/pkg/net"
	ferrethttp "github.com/MontFerret/ferret/v2/pkg/net/http"
	"github.com/MontFerret/ferret/v2/uapi"
)

var version = "unknown"

type Builtin struct {
	api.Runtime
	network ferretnet.Network
	once    sync.Once
	err     error
}

var _ api.Runtime = (*Builtin)(nil)

// NewBuiltin owns a native UAPI runtime and any network configured by its policies.
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
		engine, err := uapi.New(api.Version(version), engineOptions...)
		if err != nil {
			if fsPolicy != nil {
				return nil, fmt.Errorf("filesystem policy: %w", err)
			}

			return nil, err
		}

		return &Builtin{Runtime: engine}, nil
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
	engine, err := uapi.New(api.Version(version), engineOptions...)

	if err != nil {
		ferretnet.CloseIdleNetworkConnections(network)
		if fsPolicy != nil {
			return nil, fmt.Errorf("filesystem policy: %w", err)
		}

		return nil, err
	}

	return &Builtin{
		Runtime: engine,
		network: network,
	}, nil
}

func newDefaultBuiltin(params map[string]any) (*Builtin, error) {
	root, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	engine, err := uapi.New(api.Version(version),
		ferret.WithFSRoot(root),
		ferret.WithParams(params),
	)
	if err != nil {
		return nil, err
	}

	return &Builtin{Runtime: engine}, nil
}

// Close releases the native runtime before its owned network, retaining cleanup
// results for repeated and concurrent calls. Callers settle work first.
func (r *Builtin) Close() error {
	r.once.Do(func() {
		r.err = r.Runtime.Close()

		if r.network != nil {
			ferretnet.CloseIdleNetworkConnections(r.network)
		}
	})

	return r.err
}
