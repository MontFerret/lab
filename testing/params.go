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
