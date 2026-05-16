package acceptance

import (
	"fmt"
	"math"

	"springs/internal/sim"
)

func createDemoSimulation(w *world, _ map[string]string) error {
	return setSimulation(&w.simulation, sim.NewDemoSimulation())
}

func advanceSimulation(w *world, example map[string]string) error {
	steps, err := intValue(example, "steps")
	if err != nil {
		return err
	}
	if w.simulation == nil {
		return fmt.Errorf("simulation is not ready")
	}
	w.simulation.Advance(steps, 0.016)
	return nil
}

func assertMassX(w *world, example map[string]string) error {
	massIndex, mass, err := exampleMass(w, example)
	if err != nil {
		return err
	}
	expected, err := floatValue(example, "x")
	if err != nil {
		return err
	}
	got := mass.Position.X
	if math.Abs(got-expected) > 0.00001 {
		return fmt.Errorf("expected mass %d x %f, got %f", massIndex, expected, got)
	}
	return nil
}

func exampleMass(w *world, example map[string]string) (int, sim.Mass, error) {
	massIndex, err := intValue(example, "mass")
	if err != nil {
		return 0, sim.Mass{}, err
	}
	if w.simulation == nil {
		return 0, sim.Mass{}, fmt.Errorf("simulation is not ready")
	}
	if massIndex < 0 || massIndex >= len(w.simulation.Masses) {
		return 0, sim.Mass{}, fmt.Errorf("mass index %d out of range", massIndex)
	}
	return massIndex, w.simulation.Masses[massIndex], nil
}
