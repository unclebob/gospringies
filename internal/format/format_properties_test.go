//go:build property

package format

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
	"testing/quick"

	"springs/internal/sim"
)

func TestPropertyFromSimulationAndXSPRoundTripPreserveNormalizedWorld(t *testing.T) {
	checkProperty(t, 1, 300, fromSimulationAndXSPRoundTripPreserveNormalizedWorld)
}

func checkProperty(t *testing.T, seed int64, maxCount int, property any) {
	t.Helper()
	config := &quick.Config{MaxCount: maxCount, Rand: rand.New(rand.NewSource(seed))}
	if err := quick.Check(property, config); err != nil {
		t.Fatal(err)
	}
}

func fromSimulationAndXSPRoundTripPreserveNormalizedWorld(axInput, ayInput, bxInput, byInput, massInput, stiffnessInput, restInput float64) bool {
	world := sim.NewWorld()
	a := sim.Mass{ID: 1, Position: propertyVec(axInput, ayInput), Mass: propertyFloat(massInput, 0.1, 100), Elasticity: 0.8}
	b := sim.Mass{ID: 2, Position: propertyVec(bxInput, byInput), Mass: propertyFloat(massInput+1, 0.1, 100), Elasticity: 0.9, Fixed: true}
	_ = world.AddMass(a)
	_ = world.AddMass(b)
	_ = world.AddSpring(sim.Spring{
		ID:             1,
		MassA:          1,
		MassB:          2,
		RestLength:     propertyFloat(restInput, 0.1, 100),
		SpringConstant: propertyFloat(stiffnessInput, 0.1, 100),
		Stiffness:      propertyFloat(stiffnessInput, 0.1, 100),
		Damping:        0.7,
		Wall:           true,
		Temperature:    2,
	})
	world.Parameters.Set("current mass", "3")
	world.Parameters.EnableWall("left")
	world.Parameters.EnableForce("gravity", map[string]string{"magnitude": "10", "direction": "90"})

	document := FromSimulation(world)
	if len(document.Masses) != len(world.Masses) || len(document.Springs) != len(world.Springs) {
		panic("document length mismatch")
	}
	if document.Masses[1].Fixed != world.Masses[1].Fixed || document.Springs[0].Wall != world.Springs[0].Wall {
		panic(fmt.Sprintf("document did not preserve flags: %#v", document))
	}

	loaded, err := LoadXSP(SaveXSP(world))
	if err != nil {
		panic(err)
	}
	assertWorldEquivalent("xsp round trip", loaded, world)
	resaved := SaveXSP(loaded)
	reloaded, err := LoadXSP(resaved)
	if err != nil {
		panic(err)
	}
	if SaveXSP(reloaded) != resaved {
		panic("SaveXSP(LoadXSP(SaveXSP(world))) is not stable")
	}
	return true
}

func assertWorldEquivalent(label string, actual, expected *sim.Simulation) {
	if len(actual.Masses) != len(expected.Masses) || len(actual.Springs) != len(expected.Springs) {
		panic(label + ": object count mismatch")
	}
	for i := range expected.Masses {
		assertClose(label+" mass x", actual.Masses[i].Position.X, expected.Masses[i].Position.X, 0)
		assertClose(label+" mass y", actual.Masses[i].Position.Y, expected.Masses[i].Position.Y, 0)
		assertClose(label+" mass", actual.Masses[i].Mass, expected.Masses[i].Mass, 0)
		if actual.Masses[i].Fixed != expected.Masses[i].Fixed {
			panic(label + ": fixed flag mismatch")
		}
	}
	for i := range expected.Springs {
		if actual.Springs[i].MassA != expected.Springs[i].MassA || actual.Springs[i].MassB != expected.Springs[i].MassB {
			panic(label + ": spring endpoint mismatch")
		}
		if actual.Springs[i].Wall != expected.Springs[i].Wall {
			panic(label + ": spring wall flag mismatch")
		}
		assertClose(label+" spring constant", actual.Springs[i].SpringConstant, expected.Springs[i].SpringConstant, 0)
		assertClose(label+" rest length", actual.Springs[i].RestLength, expected.Springs[i].RestLength, 0)
	}
}

func propertyVec(xInput, yInput float64) sim.Vec2 {
	return sim.Vec2{X: propertySignedFloat(xInput, 100), Y: propertySignedFloat(yInput, 100)}
}

func propertySignedFloat(input float64, magnitude float64) float64 {
	return propertyFloat(input, 0, magnitude*2) - magnitude
}

func propertyFloat(input float64, minimum float64, maximum float64) float64 {
	if math.IsNaN(input) || math.IsInf(input, 0) {
		return minimum
	}
	return minimum + math.Mod(math.Abs(input), maximum-minimum)
}

func assertClose(label string, actual, expected, tolerance float64) {
	if math.Abs(actual-expected) > tolerance {
		panic(fmt.Sprintf("%s: got %f, want %f +/- %f", label, actual, expected, tolerance))
	}
}
