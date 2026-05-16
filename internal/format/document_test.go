package format

import (
	"testing"

	"springs/internal/sim"
)

func TestFromSimulationCopiesMassesAndSprings(t *testing.T) {
	s := sim.NewSimulation()
	left := s.AddMass(sim.Vec2{X: 1, Y: 2}, 3, true)
	right := s.AddMass(sim.Vec2{X: 4, Y: 5}, 6, false)
	s.AddSpring(left, right, 7, 8)

	document := FromSimulation(s)

	if len(document.Masses) != 2 || len(document.Springs) != 1 {
		t.Fatalf("document = %#v", document)
	}
	if document.Masses[0] != (Mass{X: 1, Y: 2, Fixed: true}) {
		t.Fatalf("first mass = %#v", document.Masses[0])
	}
	if document.Springs[0] != (Spring{A: left, B: right, RestLength: 7, Stiffness: 8}) {
		t.Fatalf("spring = %#v", document.Springs[0])
	}
}
