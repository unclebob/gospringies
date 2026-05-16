package sim

type Parameters struct {
	Values map[string]string
	Forces map[string]ForceConfig
	Walls  map[string]bool
}

type ForceConfig struct {
	Enabled string
	Values  map[string]string
}

func DefaultParameters() Parameters {
	return Parameters{
		Values: map[string]string{
			"current mass":    "1.0",
			"elasticity":      "0.8",
			"spring constant": "12.0",
			"damping":         "0.7",
			"viscosity":       "0.0",
			"stickiness":      "0.0",
			"timestep":        "0.016",
			"precision":       "0.001",
			"grid snap":       "10",
			"show springs":    "true",
		},
		Forces: map[string]ForceConfig{
			"gravity":                   {Enabled: "false", Values: map[string]string{"magnitude": "0", "direction": "90"}},
			"center attraction":         {Enabled: "false", Values: map[string]string{"magnitude": "0", "exponent": "2"}},
			"center of mass attraction": {Enabled: "false", Values: map[string]string{"magnitude": "0", "damping": "0"}},
			"wall repulsion":            {Enabled: "false", Values: map[string]string{"magnitude": "0", "exponent": "2"}},
		},
		Walls: map[string]bool{
			"top":    false,
			"left":   false,
			"right":  false,
			"bottom": false,
		},
	}
}

func (p Parameters) Clone() Parameters {
	clone := Parameters{
		Values: map[string]string{},
		Forces: map[string]ForceConfig{},
		Walls:  map[string]bool{},
	}
	for key, value := range p.Values {
		clone.Values[key] = value
	}
	for key, force := range p.Forces {
		values := map[string]string{}
		for valueKey, value := range force.Values {
			values[valueKey] = value
		}
		clone.Forces[key] = ForceConfig{Enabled: force.Enabled, Values: values}
	}
	for key, value := range p.Walls {
		clone.Walls[key] = value
	}
	return clone
}

func (p Parameters) Has(name string) bool {
	_, ok := p.Values[name]
	return ok
}

func (p Parameters) Value(name string) string {
	return p.Values[name]
}

func (p Parameters) Set(name, value string) {
	p.Values[name] = value
}

func (p Parameters) Force(name string) (ForceConfig, bool) {
	return mapValue(p.Forces, name)
}

func (p Parameters) WallEnabled(name string) (bool, bool) {
	return mapValue(p.Walls, name)
}

func mapValue[T any](values map[string]T, name string) (T, bool) {
	value, ok := values[name]
	return value, ok
}
