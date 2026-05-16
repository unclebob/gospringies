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
		}
	}
	return document
}
