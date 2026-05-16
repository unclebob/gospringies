package acceptance

import (
	"fmt"
	"math"

	"springs/internal/sim"
)

func createMovableMassAtStart(w *world, example map[string]string) error {
	if err := requireMarker(example, "start_position", "initial"); err != nil {
		return err
	}
	return ensureDomainWorld(w).AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 0, Y: 0}, Mass: 1})
}

func enableGravity(w *world, _ map[string]string) error {
	ensureDomainWorld(w).Parameters.EnableForce("gravity", map[string]string{"magnitude": "10", "direction": "90"})
	return nil
}

func advanceByDuration(w *world, example map[string]string) error {
	world, err := domainWorld(w)
	if err != nil {
		return err
	}
	duration, err := durationValue(example, "duration")
	if err != nil {
		return err
	}
	world.AdvanceDuration(duration)
	w.resultingWorld = world
	return nil
}

func assertMassPositionDiffers(w *world, example map[string]string) error {
	return assertMassVectorDiffersFromMarker(w, example, "start_position", "initial", "position", func(mass sim.Mass) sim.Vec2 { return mass.Position })
}

func assertMassVelocityDiffers(w *world, example map[string]string) error {
	return assertMassVectorDiffersFromMarker(w, example, "start_velocity", "zero", "velocity", func(mass sim.Mass) sim.Vec2 { return mass.Velocity })
}

func assertMassVectorDiffersFromMarker(w *world, example map[string]string, key, expected, name string, value func(sim.Mass) sim.Vec2) error {
	if err := requireMarker(example, key, expected); err != nil {
		return err
	}
	return assertMassVectorDiffers(w, name, value)
}

func assertMassVectorDiffers(w *world, name string, value func(sim.Mass) sim.Vec2) error {
	mass, ok := w.resultingWorld.MassByID(1)
	if !ok {
		return fmt.Errorf("mass 1 not found")
	}
	if value(mass) == (sim.Vec2{}) {
		return fmt.Errorf("mass %s did not differ from initial", name)
	}
	return nil
}

func createMassStartPosition(w *world, example map[string]string) error {
	if err := requireMarker(example, "start_position", "initial"); err != nil {
		return err
	}
	world := ensureDomainWorld(w)
	id, err := intValue(example, "mass_id")
	if err != nil {
		return err
	}
	if _, ok := world.MassByID(id); ok {
		for i := range world.Masses {
			if world.Masses[i].ID == id {
				world.Masses[i].Position = sim.Vec2{X: 5, Y: 6}
			}
		}
		return nil
	}
	return world.AddMass(sim.Mass{ID: id, Position: sim.Vec2{X: 5, Y: 6}, Mass: 1})
}

func assertMassPositionRemains(w *world, example map[string]string) error {
	id, err := intValue(example, "mass_id")
	if err != nil {
		return err
	}
	mass, ok := w.resultingWorld.MassByID(id)
	if !ok {
		return fmt.Errorf("mass %d not found", id)
	}
	if mass.Position != (sim.Vec2{X: 5, Y: 6}) {
		return fmt.Errorf("mass %d position changed to %#v", id, mass.Position)
	}
	return nil
}

func assertMassVelocityRemains(w *world, example map[string]string) error {
	if err := requireMarker(example, "start_velocity", "zero"); err != nil {
		return err
	}
	id, err := intValue(example, "mass_id")
	if err != nil {
		return err
	}
	mass, ok := w.resultingWorld.MassByID(id)
	if !ok {
		return fmt.Errorf("mass %d not found", id)
	}
	if mass.Velocity != (sim.Vec2{}) {
		return fmt.Errorf("mass %d velocity changed to %#v", id, mass.Velocity)
	}
	return nil
}

func createWorldInState(w *world, example map[string]string) error {
	state, err := stringValue(example, "initial_state")
	if err != nil {
		return err
	}
	world, err := worldForState(state)
	if err != nil {
		return err
	}
	w.domainWorld = world
	return nil
}

func assertResultDeterministic(_ *world, example map[string]string) error {
	duration, err := durationValue(example, "duration")
	if err != nil {
		return err
	}
	first, second, err := deterministicWorlds(example)
	if err != nil {
		return err
	}
	first.AdvanceDuration(duration)
	second.AdvanceDuration(duration)
	if !sameWorldState(first, second) {
		return fmt.Errorf("state differs between runs")
	}
	return nil
}

func deterministicWorlds(example map[string]string) (*sim.Simulation, *sim.Simulation, error) {
	state, err := stringValue(example, "initial_state")
	if err != nil {
		return nil, nil, err
	}
	first, err := worldForState(state)
	if err != nil {
		return nil, nil, err
	}
	second, err := worldForState(state)
	return first, second, err
}

func advanceByDurationAtFrameRate(w *world, example map[string]string) error {
	world, err := domainWorld(w)
	if err != nil {
		return err
	}
	duration, err := durationValue(example, "duration")
	if err != nil {
		return err
	}
	frameRate, err := frameRateValue(example)
	if err != nil {
		return err
	}
	frameDuration := 1 / frameRate
	for remaining := duration; remaining > 0; {
		step := frameDuration
		if remaining < step {
			step = remaining
		}
		world.AdvanceDuration(step)
		remaining -= step
	}
	w.resultingWorld = world
	return nil
}

func assertSimulationTime(w *world, example map[string]string) error {
	expected, err := durationValue(example, "duration")
	if err != nil {
		return err
	}
	if math.Abs(w.resultingWorld.Time-expected) > 0.000001 {
		return fmt.Errorf("time = %f, expected %f", w.resultingWorld.Time, expected)
	}
	return nil
}

func worldForState(state string) (*sim.Simulation, error) {
	world := sim.NewWorld()
	switch state {
	case "simple spring":
		_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 0, Y: 0}, Mass: 1, Fixed: true})
		_ = world.AddMass(sim.Mass{ID: 2, Position: sim.Vec2{X: 12, Y: 0}, Mass: 1})
		_ = world.AddSpring(sim.Spring{ID: 1, MassA: 1, MassB: 2, RestLength: 10, SpringConstant: 12})
	case "gravity only":
		world.Parameters.EnableForce("gravity", map[string]string{"magnitude": "10", "direction": "90"})
		_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 0, Y: 0}, Mass: 1})
	default:
		return nil, fmt.Errorf("unsupported initial state %q", state)
	}
	return world, nil
}

func sameWorldState(a, b *sim.Simulation) bool {
	if len(a.Masses) != len(b.Masses) || math.Abs(a.Time-b.Time) > 0.000001 {
		return false
	}
	for i := range a.Masses {
		if a.Masses[i].Position != b.Masses[i].Position || a.Masses[i].Velocity != b.Masses[i].Velocity {
			return false
		}
	}
	return true
}

func durationValue(example map[string]string, key string) (float64, error) {
	value, err := stringValue(example, key)
	if err != nil {
		return 0, err
	}
	durations := map[string]float64{
		"1 step":   sim.DefaultParameters().StepDuration(),
		"10 steps": sim.DefaultParameters().StepDuration() * 10,
		"1 second": 1,
	}
	duration, ok := durations[value]
	if !ok {
		return 0, fmt.Errorf("unsupported duration %q", value)
	}
	return duration, nil
}

func frameRateValue(example map[string]string) (float64, error) {
	value, err := stringValue(example, "frame_rate")
	if err != nil {
		return 0, err
	}
	switch value {
	case "30 fps":
		return 30, nil
	case "60 fps":
		return 60, nil
	default:
		return 0, fmt.Errorf("unsupported frame rate %q", value)
	}
}

func requireMarker(example map[string]string, key, expected string) error {
	value, err := stringValue(example, key)
	if err != nil {
		return err
	}
	if value != expected {
		return fmt.Errorf("expected %s marker %q, got %q", key, expected, value)
	}
	return nil
}
