package sim

type Parameters struct {
	Values      map[string]string
	Forces      map[string]ForceConfig
	Walls       map[string]bool
	ActiveForce string
}

type ForceConfig struct {
	Enabled string
	Values  map[string]string
}

func DefaultParameters() Parameters {
	return Parameters{
		Values: map[string]string{
			"current mass":      "1.0",
			"elasticity":        "0.8",
			"spring constant":   "12.0",
			"damping":           "0.7",
			"viscosity":         "0.0",
			"stickiness":        "0.0",
			"timestep":          "0.016",
			"precision":         "0.001",
			"grid snap":         "10",
			"show springs":      "true",
			"fixed mass":        "false",
			"center mass":       "-1",
			"adaptive timestep": "false",
		},
		Forces: map[string]ForceConfig{
			"gravity":                   {Enabled: "false", Values: map[string]string{"magnitude": "0", "direction": "90"}},
			"center attraction":         {Enabled: "false", Values: map[string]string{"magnitude": "0", "exponent": "2"}},
			"center of mass attraction": {Enabled: "false", Values: map[string]string{"magnitude": "0", "damping": "0"}},
			"wall repulsion":            {Enabled: "false", Values: map[string]string{"magnitude": "0", "exponent": "2"}},
			"mass collision":            {Enabled: "false", Values: map[string]string{}},
		},
		Walls: map[string]bool{
			"top":    false,
			"left":   false,
			"right":  false,
			"bottom": false,
		},
		ActiveForce: "gravity",
	}
}

func (p Parameters) Clone() Parameters {
	clone := Parameters{
		Values:      map[string]string{},
		Forces:      map[string]ForceConfig{},
		Walls:       map[string]bool{},
		ActiveForce: p.ActiveForce,
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

func (p *Parameters) EnableForce(name string, values map[string]string) {
	force := p.Forces[name]
	force.Enabled = "true"
	if force.Values == nil {
		force.Values = map[string]string{}
	}
	for key, value := range values {
		force.Values[key] = value
	}
	p.Forces[name] = force
	p.ActiveForce = name
}

func (p *Parameters) SelectForce(name string) {
	p.ActiveForce = name
}

func (p Parameters) EnableWall(name string) {
	p.Walls[name] = true
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

func (p Parameters) StepDuration() float64 {
	return parameterFloat(p, "timestep")
}

// mutate4go-manifest-begin
// {"version":1,"tested_at":"2026-05-19T09:37:38-05:00","module_hash":"90538624ca2f683d4d39f3eec3870c2fbf9012d1a3719776bbffe1207ed15b20","functions":[{"id":"func/DefaultParameters","name":"DefaultParameters","line":15,"end_line":47,"hash":"ba904154bfcf7dfc4ae4f96274279bdfb37dc91aa5b5a31ce195bebbd2160a12"},{"id":"func/Parameters.Clone","name":"Parameters.Clone","line":49,"end_line":70,"hash":"331f52c5a9964490473454b6790996c54637482425676f33505575882e2f1997"},{"id":"func/Parameters.Has","name":"Parameters.Has","line":72,"end_line":75,"hash":"4ce2d1433ad0ac530d4494fdc68a968bae4d2f1cec92e984143052b5455dc36d"},{"id":"func/Parameters.Value","name":"Parameters.Value","line":77,"end_line":79,"hash":"1a07bf4fb6f117a5adac270831d402eb68e6eaba6a3ebc06033780dba4b45a4e"},{"id":"func/Parameters.Set","name":"Parameters.Set","line":81,"end_line":83,"hash":"bfb83a17e33ba6641bf1662749a931d72659df7b1765e861f919c05c6fad95f7"},{"id":"func/Parameters.EnableForce","name":"Parameters.EnableForce","line":85,"end_line":96,"hash":"d336837e4bcc4c183d6c4401b47890deb41e4d4823571577011ea1f26fe91334"},{"id":"func/Parameters.SelectForce","name":"Parameters.SelectForce","line":98,"end_line":100,"hash":"b03e46287fa3e7326ccbd430fe28e6ec043c041dbc128f45db4ef72ebe5cbd3e"},{"id":"func/Parameters.EnableWall","name":"Parameters.EnableWall","line":102,"end_line":104,"hash":"0a9bb74291511e8940e7ac7c23ff74e351578ad60c7e57ac6c790f16ff110199"},{"id":"func/Parameters.Force","name":"Parameters.Force","line":106,"end_line":108,"hash":"6ce8bbc1b9e9a3505624b9c471f2f34b0eea5eb0f7f884961e725bb73a370094"},{"id":"func/Parameters.WallEnabled","name":"Parameters.WallEnabled","line":110,"end_line":112,"hash":"cf746ddf54fc9f3ec2b3f956bb9573f709eb02d6831540336408d9fd85018e72"},{"id":"func/mapValue","name":"mapValue","line":114,"end_line":117,"hash":"fb865dcc4fa40b00b87f94eddd37eb75d1cae76e4af3bdaf80cbe89d5694e2f0"},{"id":"func/Parameters.StepDuration","name":"Parameters.StepDuration","line":119,"end_line":121,"hash":"edca26dc3f7c1047f95bd00f051de0c3239248421eb712d7eb10cf4543e6c155"}]}
// mutate4go-manifest-end
