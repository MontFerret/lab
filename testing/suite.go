package testing

import (
	"context"
	"encoding/json"
	"net/url"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"

	"github.com/MontFerret/lab/runtime"
	"github.com/MontFerret/lab/sources"
)

type (
	Suite struct {
		file     sources.File
		manifest SuiteManifest
	}

	SuiteManifest struct {
		Query  ScriptManifest `yaml:"query"`
		Assert ScriptManifest `yaml:"assert"`
	}

	ScriptManifest struct {
		Text   string                 `yaml:"text"`
		Ref    string                 `yaml:"ref"`
		Params map[string]interface{} `yaml:"params"`
	}

	DataContext struct {
		Query DataContextValues `json:"query"`
	}

	DataContextValues struct {
		Result interface{}            `json:"result"`
		Params map[string]interface{} `json:"params"`
	}
)

func NewSuite(file sources.File) (*Suite, error) {
	manifest := SuiteManifest{}

	if err := yaml.Unmarshal(file.Content, &manifest); err != nil {
		return nil, errors.Wrap(err, "failed to parse file")
	}

	if err := validateScriptManifest(manifest.Query); err != nil {
		return nil, errors.Wrap(err, "query")
	}

	if err := validateScriptManifest(manifest.Assert); err != nil {
		return nil, errors.Wrap(err, "assert")
	}

	return &Suite{
		file:     file,
		manifest: manifest,
	}, nil
}

func (suite *Suite) Run(ctx context.Context, rt runtime.Runtime, params Params) error {
	query, err := suite.resolveScript(ctx, suite.manifest.Query)

	if err != nil {
		return errors.Wrap(err, "resolve query script")
	}

	assertion, err := suite.resolveScript(ctx, suite.manifest.Assert)

	if err != nil {
		return errors.Wrap(err, "resolve assertion script")
	}

	queryParams := resolveRuntimeParams(params.Clone(), suite.manifest.Query)

	out, err := rt.Run(ctx, query, queryParams)

	if err != nil {
		return errors.Wrap(err, "failed to execute query script")
	}

	outVal, err := toAny(out)

	if err != nil {
		return errors.Wrap(err, "deserialize query output")
	}

	params.SetSystemValue("context", DataContext{
		Query: DataContextValues{
			Result: outVal,
			Params: queryParams,
		},
	})

	_, err = rt.Run(ctx, assertion, resolveRuntimeParams(params, suite.manifest.Query))

	return err
}

func (suite *Suite) resolveScript(ctx context.Context, manifest ScriptManifest) (string, error) {
	if manifest.Text != "" {
		return manifest.Text, nil
	}

	u, err := url.Parse(manifest.Ref)

	if err != nil {
		return "", errors.Wrap(err, "parse 'src'")
	}

	// set a query param that indicates from what relative location to resolve a given script
	q := u.Query()
	q.Add("from", suite.file.Name)
	u.RawQuery = q.Encode()

	src, err := sources.New(u.String())

	if err != nil {
		return "", errors.Wrap(err, "new src source")
	}

	out := src.Read(ctx)

	select {
	case e := <-out.Errors:
		return "", e
	case f := <-out.Files:
		if f.Error != nil {
			return "", f.Error
		}

		return string(f.Content), nil
	}
}

func resolveRuntimeParams(params Params, manifest ScriptManifest) map[string]interface{} {
	params.SetUserValues(manifest.Params)

	return params.ToMap()
}

func toAny(values []byte) (interface{}, error) {
	if len(values) == 0 {
		return nil, nil
	}

	var o interface{}

	if err := json.Unmarshal(values, &o); err != nil {
		return nil, err
	}

	return o, nil
}

func validateScriptManifest(manifest ScriptManifest) error {
	if manifest.Ref == "" && manifest.Text == "" {
		return errors.New("ref or text must have value")
	}

	if manifest.Ref != "" && manifest.Text != "" {
		return errors.New("only either ref or text must have value")
	}

	return nil
}
