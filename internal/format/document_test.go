package format

import (
	"testing"

	"springs/internal/sim"
)

func TestFromSimulationCopiesMassesAndSprings(t *testing.T) {
	s := sim.NewSimulation()
	if err := s.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 1, Y: 2}, Mass: 3, Fixed: true}); err != nil {
		t.Fatal(err)
	}
	if err := s.AddMass(sim.Mass{ID: 2, Position: sim.Vec2{X: 4, Y: 5}, Mass: 6}); err != nil {
		t.Fatal(err)
	}
	if err := s.AddSpring(sim.Spring{ID: 7, MassA: 1, MassB: 2, RestLength: 7, Stiffness: 8}); err != nil {
		t.Fatal(err)
	}

	document := FromSimulation(s)

	if len(document.Masses) != 2 || len(document.Springs) != 1 {
		t.Fatalf("document = %#v", document)
	}
	if document.Masses[0] != (Mass{X: 1, Y: 2, Fixed: true}) {
		t.Fatalf("first mass = %#v", document.Masses[0])
	}
	if document.Springs[0] != (Spring{A: 0, B: 1, RestLength: 7, Stiffness: 8}) {
		t.Fatalf("spring = %#v", document.Springs[0])
	}
}
