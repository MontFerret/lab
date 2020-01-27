package runner

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

func (p Params) ToMap() map[string]interface{} {
	out := make(map[string]interface{})

	for k, v := range p.user {
		out[k] = v
	}

	out["lab"] = p.system

	return out
}
