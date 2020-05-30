package testing

type Params struct {
	system map[string]interface{}
	user   map[string]interface{}
}

func NewParams() Params {
	return Params{
		system: make(map[string]interface{}),
		user:   make(map[string]interface{}),
	}
}

func (p *Params) SetSystemValue(name string, value interface{}) {
	p.system[name] = value
}

func (p *Params) SetUserValue(name string, value interface{}) {
	p.user[name] = value
}

func (p *Params) SetUserValues(values map[string]interface{}) {
	for k, v := range values {
		p.SetUserValue(k, v)
	}
}

func (p *Params) ToMap() map[string]interface{} {
	out := p.copyMap(p.user)

	out["lab"] = p.copyMap(p.system)

	return out
}

func (p *Params) Clone() Params {
	return Params{
		system: p.copyMap(p.system),
		user:   p.copyMap(p.user),
	}
}

func (p *Params) copyMap(m map[string]interface{}) map[string]interface{} {
	cp := make(map[string]interface{})

	for k, v := range m {
		vm, ok := v.(map[string]interface{})
		if ok {
			cp[k] = p.copyMap(vm)
		} else {
			cp[k] = v
		}
	}

	return cp
}
