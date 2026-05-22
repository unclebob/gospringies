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
	A           int     `json:"a"`
	B           int     `json:"b"`
	RestLength  float64 `json:"rest_length"`
	Stiffness   float64 `json:"stiffness"`
	Wall        bool    `json:"wall"`
	Temperature float64 `json:"temperature"`
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
			A:           spring.A,
			B:           spring.B,
			RestLength:  spring.RestLength,
			Stiffness:   spring.Stiffness,
			Wall:        spring.Wall,
			Temperature: spring.Temperature,
		}
	}
	return document
}

// mutate4go-manifest-begin
// {"version":1,"tested_at":"2026-05-22T10:50:34-05:00","module_hash":"cb4e77e5be944e0c120c992e36e58921d0e4f1aa0574d6fe14d936f3e8aa126d","functions":[{"id":"func/FromSimulation","name":"FromSimulation","line":25,"end_line":44,"hash":"5bcfd8db1bf642204285e2c760912ebe446b1a2f4ceedb6da20b23c896f1c5bb"}]}
// mutate4go-manifest-end
