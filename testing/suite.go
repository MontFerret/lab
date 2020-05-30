package testing

import (
	"context"
	"encoding/json"
	"github.com/MontFerret/lab/runtime"
	"github.com/MontFerret/lab/sources"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"net/url"
)

type Suite struct {
	file       sources.File
	beforeHook string
	afterHook  string
	query      interface{}
	assertion  interface{}
}

func NewSuite(file sources.File) (*Suite, error) {
	config := make(map[interface{}]interface{})

	if err := yaml.Unmarshal(file.Content, &config); err != nil {
		return nil, errors.Wrap(err, "failed to parse file")
	}

	query, ok := config["query"]

	if !ok {
		return nil, errors.New("'query' property is required")
	}

	assertion, ok := config["assert"]

	if !ok {
		return nil, errors.New("'assert' property is required")
	}

	return &Suite{
		file:       file,
		beforeHook: "", // TODO: implement
		afterHook:  "", // TODO: implement
		query:      query,
		assertion:  assertion,
	}, nil
}

func (suite *Suite) Run(ctx context.Context, rt runtime.Runtime, params Params) error {
	query, err := suite.resolveScript(ctx, suite.query)

	if err != nil {
		return errors.Wrap(err, "resolve query script")
	}

	assertion, err := suite.resolveScript(ctx, suite.assertion)

	if err != nil {
		return errors.Wrap(err, "resolve assertion script")
	}

	out, err := rt.Run(ctx, query, params.ToMap())

	if err != nil {
		return errors.Wrap(err, "failed to execute query script")
	}

	outVal, err := toAny(out)

	if err != nil {
		return errors.Wrap(err, "deserialize query output")
	}

	params.SetSystemValue("data", map[string]interface{}{
		"query": outVal,
	})

	_, err = rt.Run(ctx, assertion, params.ToMap())

	return err
}

func (suite *Suite) resolveScript(ctx context.Context, config interface{}) (string, error) {
	switch value := config.(type) {
	case string:
		return value, nil
	case map[interface{}]interface{}:
		location, ok := value["src"]

		if !ok {
			return "", errors.New("missed 'src' keyword")
		}

		locationStr, ok := location.(string)

		if !ok {
			return "", errors.New("'src' must be a string")
		}

		u, err := url.Parse(locationStr)

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
	default:
		return "", errors.New("invalid script definition")
	}
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
