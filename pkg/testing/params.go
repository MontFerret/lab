package testing

import (
	"fmt"
	"strings"
)

type Params struct {
	system map[string]any
	user   map[string]any
}

func NewParams() Params {
	return Params{
		system: make(map[string]any),
		user:   make(map[string]any),
	}
}

func (p *Params) SetSystemValue(name string, value any) {
	p.system[name] = value
}

func (p *Params) SetUserValue(name string, value any) {
	p.user[name] = value
}

func (p *Params) SetUserValues(values map[string]any) {
	for k, v := range values {
		p.SetUserValue(k, v)
	}
}

// ApplyBindings resolves every source from one materialized snapshot and then
// atomically writes the bound values into the user parameter namespace.
func (p *Params) ApplyBindings(bindings []ParamBinding) error {
	if len(bindings) == 0 {
		return nil
	}

	snapshot := p.ToMap()
	resolved := make([]any, len(bindings))

	for i, binding := range bindings {
		target := binding.Target.String()
		if target == "" {
			return fmt.Errorf("parameter binding target cannot be empty")
		}

		source := binding.Source.String()
		if source == "" {
			return fmt.Errorf("parameter binding source for target %q cannot be empty", target)
		}

		if target == "lab" || strings.HasPrefix(target, "lab.") {
			return fmt.Errorf("parameter binding target %q uses reserved namespace \"lab\"", target)
		}

		value, exists := binding.Source.lookup(snapshot)
		if !exists {
			return fmt.Errorf("parameter binding target %q: source %q does not exist", target, "@"+source)
		}

		resolved[i] = value
	}

	user := ToMap(p.user)
	for i, binding := range bindings {
		if err := binding.Target.set(user, resolved[i]); err != nil {
			return err
		}
	}

	p.user = user

	return nil
}

func (p *Params) ToMap() map[string]any {
	out := ToMap(p.user)

	out["lab"] = ToMap(p.system)

	return out
}

func (p *Params) Clone() Params {
	return Params{
		system: ToMap(p.system),
		user:   ToMap(p.user),
	}
}
