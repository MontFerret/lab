package testing

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
