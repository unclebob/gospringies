package format

import "springs/internal/sim"

type Document struct {
	Masses  []Mass   `json:"masses"`
	Springs []Spring `json:"springs"`
}

type Mass struct {
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	Fixed bool    `json:"fixed"`
}

type Spring struct {
	A          int     `json:"a"`
	B          int     `json:"b"`
	RestLength float64 `json:"rest_length"`
	Stiffness  float64 `json:"stiffness"`
	Wall       bool    `json:"wall"`
}

func FromSimulation(s *sim.Simulation) Document {
	document := Document{
		Masses:  make([]Mass, len(s.Masses)),
		Springs: make([]Spring, len(s.Springs)),
	}
	for i, mass := range s.Masses {
		document.Masses[i] = Mass{X: mass.Position.X, Y: mass.Position.Y, Fixed: mass.Fixed}
	}
	for i, spring := range s.Springs {
		document.Springs[i] = Spring{
			A:          spring.A,
			B:          spring.B,
			RestLength: spring.RestLength,
			Stiffness:  spring.Stiffness,
			Wall:       spring.Wall,
		}
	}
	return document
}

// mutate4go-manifest-begin
// {"version":1,"tested_at":"2026-05-19T10:14:03-05:00","module_hash":"73b2505868e8218c4384ccc2ed5ff529c9df18dbe9d0dd9d2de1733a5e1ffab2","functions":[{"id":"func/FromSimulation","name":"FromSimulation","line":23,"end_line":40,"hash":"bdeb8ace35512f1a1ed273e489d778b331d24ce34b5a8a7315c5eda6c8ff916b"}]}
// mutate4go-manifest-end
