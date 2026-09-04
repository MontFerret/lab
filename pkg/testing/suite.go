package testing

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/MontFerret/ferret/v2"

	"github.com/MontFerret/lab/v2/pkg/runtime"
	"github.com/MontFerret/lab/v2/pkg/sources"
)

type (
	Suite struct {
		file     sources.File
		timeout  time.Duration
		manifest SuiteManifest
	}

	DataContext struct {
		Query DataContextValues `json:"query"`
	}

	DataContextValues struct {
		Result any            `json:"result"`
		Params map[string]any `json:"params"`
	}
)

func NewSuite(opts Options) (*Suite, error) {
	manifest := SuiteManifest{}

	if err := yaml.Unmarshal(opts.File.Content, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	if err := manifest.validate(); err != nil {
		return nil, err
	}

	timeout := opts.Timeout

	if manifest.Timeout > 0 {
		timeout = time.Duration(manifest.Timeout) * time.Second
	}

	return &Suite{
		file:     opts.File,
		timeout:  timeout,
		manifest: manifest,
	}, nil
}

func (suite *Suite) Run(ctx context.Context, rt runtime.Runtime, params Params) error {
	ctx, cancel := context.WithTimeout(ctx, suite.timeout)
	defer cancel()

	query, err := suite.resolveScript(ctx, "query", suite.manifest.Query)
	if err != nil {
		return fmt.Errorf("resolve query script: %w", err)
	}

	if expectedError := suite.manifest.Expect.Error; expectedError != nil {
		_, err := rt.Run(ctx, query, suite.manifest.Query.runtimeParams(params.Clone()))

		return expectedError.evaluate(err)
	}

	assertion, err := suite.resolveScript(ctx, "assert", *suite.manifest.Assert)
	if err != nil {
		return fmt.Errorf("resolve assertion script: %w", err)
	}

	queryParams := suite.manifest.Query.runtimeParams(params.Clone())

	out, err := rt.Run(ctx, query, queryParams)
	if err != nil {
		return fmt.Errorf("failed to execute query script: %w", err)
	}

	outVal, err := suite.deserializeQueryOutput(out)
	if err != nil {
		return fmt.Errorf("deserialize query output: %w", err)
	}

	params.SetSystemValue("data", DataContext{
		Query: DataContextValues{
			Result: outVal,
			Params: queryParams,
		},
	})

	_, err = rt.Run(ctx, assertion, suite.manifest.Assert.runtimeParams(params))

	return err
}

func (suite *Suite) resolveScript(ctx context.Context, scriptType string, manifest ScriptManifest) (ferret.Source, error) {
	if manifest.Text != "" {
		return ferret.NewSource(fmt.Sprintf("%s -> %s", suite.file.Name, scriptType), manifest.Text), nil
	}

	u, err := url.Parse(manifest.Ref)
	if err != nil {
		return ferret.Source{}, fmt.Errorf("parse 'ref': %w", err)
	}

	onNext, onError := suite.file.Resolve(ctx, u)

	select {
	case e := <-onError:
		return ferret.Source{}, fmt.Errorf("resolve 'ref': %w", e)
	case f := <-onNext:
		return ferret.NewSource(f.Name, string(f.Content)), nil
	}
}

func (suite *Suite) deserializeQueryOutput(values []byte) (any, error) {
	if len(values) == 0 {
		return nil, nil
	}

	var o any

	if err := json.Unmarshal(values, &o); err != nil {
		return nil, err
	}

	return o, nil
}
