package sim

import (
	"errors"
	"testing"
)

func TestNewWorldIsEmpty(t *testing.T) {
	world := NewWorld()

	if len(world.Masses) != 0 {
		t.Fatalf("mass count = %d", len(world.Masses))
	}
	if len(world.Springs) != 0 {
		t.Fatalf("spring count = %d", len(world.Springs))
	}
}

func TestMassLookupReturnsModeledProperties(t *testing.T) {
	world := NewWorld()
	err := world.AddMass(Mass{
		ID:         2,
		Position:   Vec2{X: 30, Y: 40},
		Velocity:   Vec2{X: 1.5, Y: -2},
		Mass:       2.5,
		Elasticity: 0.4,
		Fixed:      true,
	})
	if err != nil {
		t.Fatal(err)
	}

	mass, ok := world.MassByID(2)
	if !ok {
		t.Fatal("mass not found")
	}
	if mass.Position.X != 30 || mass.Position.Y != 40 {
		t.Fatalf("position = %#v", mass.Position)
	}
	if mass.Velocity.X != 1.5 || mass.Velocity.Y != -2 {
		t.Fatalf("velocity = %#v", mass.Velocity)
	}
	if mass.Mass != 2.5 || mass.Elasticity != 0.4 || !mass.Fixed {
		t.Fatalf("mass properties = %#v", mass)
	}
}

func TestSpringLookupReturnsModeledProperties(t *testing.T) {
	world := NewWorld()
	_ = world.AddMass(Mass{ID: 1, Position: Vec2{X: 0, Y: 0}, Mass: 1})
	_ = world.AddMass(Mass{ID: 2, Position: Vec2{X: 10, Y: 0}, Mass: 1})
	err := world.AddSpring(Spring{
		ID:             7,
		MassA:          1,
		MassB:          2,
		SpringConstant: 12.5,
		Damping:        0.7,
		RestLength:     10,
	})
	if err != nil {
		t.Fatal(err)
	}

	spring, ok := world.SpringByID(7)
	if !ok {
		t.Fatal("spring not found")
	}
	if spring.MassA != 1 || spring.MassB != 2 {
		t.Fatalf("spring endpoints = %#v", spring)
	}
	if spring.SpringConstant != 12.5 || spring.Damping != 0.7 || spring.RestLength != 10 {
		t.Fatalf("spring properties = %#v", spring)
	}
}

func TestDuplicateIDsAreInvalid(t *testing.T) {
	world := NewWorld()
	if err := world.AddMass(Mass{ID: 1, Mass: 1}); err != nil {
		t.Fatal(err)
	}
	if err := world.AddMass(Mass{ID: 1, Mass: 1}); !errors.Is(err, ErrDuplicateID) {
		t.Fatalf("duplicate mass error = %v", err)
	}

	if err := world.AddMass(Mass{ID: 2, Mass: 1}); err != nil {
		t.Fatal(err)
	}
	if err := world.AddSpring(Spring{ID: 5, MassA: 1, MassB: 2}); err != nil {
		t.Fatal(err)
	}
	if err := world.AddSpring(Spring{ID: 5, MassA: 1, MassB: 2}); !errors.Is(err, ErrDuplicateID) {
		t.Fatalf("duplicate spring error = %v", err)
	}
}

func TestSpringsRequireExistingEndpoints(t *testing.T) {
	world := NewWorld()
	if err := world.AddMass(Mass{ID: 1, Mass: 1}); err != nil {
		t.Fatal(err)
	}

	if err := world.AddSpring(Spring{ID: 2, MassA: 1, MassB: 9}); !errors.Is(err, ErrMissingSpringEndpoint) {
		t.Fatalf("missing endpoint error = %v", err)
	}
}

func TestFixedMassStateIsIndependentFromMassValue(t *testing.T) {
	world := NewWorld()
	if err := world.AddMass(Mass{ID: 4, Position: Vec2{X: 5, Y: 6}, Mass: 3, Fixed: true}); err != nil {
		t.Fatal(err)
	}

	mass, _ := world.MassByID(4)
	if !mass.Fixed {
		t.Fatal("fixed state was not explicit")
	}
	if mass.Mass != 3 {
		t.Fatalf("mass value = %f", mass.Mass)
	}
}
