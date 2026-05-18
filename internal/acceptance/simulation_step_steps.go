package acceptance

import (
	"fmt"
	"math"
	"strconv"

	"springs/internal/sim"
)

func createMovableMassAtStart(w *world, example map[string]string) error {
	if err := requireMarker(example, "start_position", "initial"); err != nil {
		return err
	}
	return ensureDomainWorld(w).AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 0, Y: 0}, Mass: 1})
}

func setAdaptiveTimestep(w *world, example map[string]string) error {
	adaptive, err := boolValue(example, "adaptive")
	if err != nil {
		return err
	}
	ensureDomainWorld(w).Parameters.Set("adaptive timestep", fmt.Sprintf("%t", adaptive))
	return nil
}

func setTimeStep(w *world, example map[string]string) error {
	timeStep, err := stringValue(example, "time_step")
	if err != nil {
		return err
	}
	ensureDomainWorld(w).Parameters.Set("timestep", timeStep)
	return nil
}

func assertRK4DeterministicAdvance(_ *world, example map[string]string) error {
	duration, err := durationValue(example, "duration")
	if err != nil {
		return err
	}
	first, second, err := matchingRK4Worlds(example)
	if err != nil {
		return err
	}
	first.AdvanceDuration(duration)
	second.AdvanceDuration(duration)
	return assertDeterministicAdvance(first, second, duration)
}

func applyNumericsSettings(world *sim.Simulation, example map[string]string) error {
	if adaptive, ok := example["adaptive"]; ok {
		world.Parameters.Set("adaptive timestep", adaptive)
	}
	if timeStep, ok := example["time_step"]; ok {
		world.Parameters.Set("timestep", timeStep)
	}
	return nil
}

func rk4AcceptanceWorld() *sim.Simulation {
	world := sim.NewWorld()
	world.Parameters.EnableForce("gravity", map[string]string{"magnitude": "10", "direction": "0"})
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{}, Mass: 1})
	return world
}

func matchingRK4Worlds(example map[string]string) (*sim.Simulation, *sim.Simulation, error) {
	first := rk4AcceptanceWorld()
	second := rk4AcceptanceWorld()
	if err := applyNumericsSettings(first, example); err != nil {
		return nil, nil, err
	}
	if err := applyNumericsSettings(second, example); err != nil {
		return nil, nil, err
	}
	return first, second, nil
}

func assertDeterministicAdvance(first *sim.Simulation, second *sim.Simulation, duration float64) error {
	if !sameWorldState(first, second) {
		return fmt.Errorf("RK4 advancement differed between runs")
	}
	if first.LastAdvanceSteps <= 1 {
		return fmt.Errorf("fixed timestep used %d steps", first.LastAdvanceSteps)
	}
	if math.Abs(first.Time-duration) > 0.000001 {
		return fmt.Errorf("time = %f, want %f", first.Time, duration)
	}
	return nil
}

func setPrecision(w *world, example map[string]string) error {
	precision, err := precisionValue(example)
	if err != nil {
		return err
	}
	world := ensureDomainWorld(w)
	world.Parameters.Set("precision", precision)
	world.Parameters.Set("timestep", "0.1")
	return nil
}

func precisionValue(example map[string]string) (string, error) {
	value, err := stringValue(example, "precision")
	if err != nil {
		return "", err
	}
	precisions := map[string]string{"low": "0.0001", "high": "0.1"}
	precision, ok := precisions[value]
	if !ok {
		return "", fmt.Errorf("unsupported precision %q", value)
	}
	return precision, nil
}

func advanceUnstableSimulation(w *world, example map[string]string) error {
	world := ensureDomainWorld(w)
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 0, Y: 0}, Mass: 1, Fixed: true})
	_ = world.AddMass(sim.Mass{ID: 2, Position: sim.Vec2{X: 20, Y: 0}, Mass: 1})
	_ = world.AddSpring(sim.Spring{ID: 1, MassA: 1, MassB: 2, RestLength: 5, SpringConstant: 50})
	return advanceByDuration(w, example)
}

func assertAdaptiveStepBehavior(w *world, example map[string]string) error {
	if w.resultingWorld.Parameters.Value("adaptive timestep") != "true" {
		return fmt.Errorf("adaptive timestep was not enabled")
	}
	behavior, err := stringValue(example, "step_behavior")
	if err != nil {
		return err
	}
	check, ok := adaptiveStepChecks[behavior]
	if !ok {
		return fmt.Errorf("unsupported step behavior %q", behavior)
	}
	return requireAdaptiveSteps(w.resultingWorld.LastAdvanceSteps, behavior, check(w.resultingWorld.LastAdvanceSteps))
}

func requireAdaptiveSteps(steps int, want string, matches bool) error {
	if !matches {
		return fmt.Errorf("adaptive step count = %d, want %s", steps, want)
	}
	return nil
}

var adaptiveStepChecks = map[string]func(int) bool{
	"smaller steps": func(steps int) bool { return steps > 10 },
	"larger steps":  func(steps int) bool { return steps <= 10 },
}

func assertSimulationTimeAdvanced(w *world, example map[string]string) error {
	return assertSimulationTime(w, example)
}

func setRenderFrameRate(w *world, example map[string]string) error {
	frameRate, err := frameRateValue(example)
	w.appBeforeTime = frameRate
	return err
}

func assertStateIndependentOfFrameRate(_ *world, example map[string]string) error {
	frameRate, err := frameRateValue(example)
	if err != nil {
		return err
	}
	duration, err := durationValue(example, "duration")
	if err != nil {
		return err
	}
	byDuration, byFrames, err := matchingRK4Worlds(example)
	if err != nil {
		return err
	}
	byDuration.AdvanceDuration(duration)
	advanceInFrames(byFrames, duration, frameRate)
	if !sameWorldState(byDuration, byFrames) {
		return fmt.Errorf("simulation state depended on frame rate %s", example["frame_rate"])
	}
	return nil
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
	return assertMassVectorDiffers(w, example, positionDifference())
}

func assertMassVelocityDiffers(w *world, example map[string]string) error {
	return assertMassVectorDiffers(w, example, velocityDifference())
}

type massVectorDifference struct {
	markerKey   string
	markerValue string
	label       string
	vector      func(sim.Mass) sim.Vec2
}

func positionDifference() massVectorDifference {
	return massVectorDifference{"start_position", "initial", "position", func(mass sim.Mass) sim.Vec2 { return mass.Position }}
}

func velocityDifference() massVectorDifference {
	return massVectorDifference{"start_velocity", "zero", "velocity", func(mass sim.Mass) sim.Vec2 { return mass.Velocity }}
}

func assertMassVectorDiffers(w *world, example map[string]string, difference massVectorDifference) error {
	if err := requireMarker(example, difference.markerKey, difference.markerValue); err != nil {
		return err
	}
	mass, err := resultingMass(w, 1)
	if err != nil {
		return err
	}
	if difference.vector(mass) == (sim.Vec2{}) {
		return fmt.Errorf("mass %s did not differ from %s", difference.label, difference.markerValue)
	}
	return nil
}

func resultingMass(w *world, id int) (sim.Mass, error) {
	mass, ok := w.resultingWorld.MassByID(id)
	if !ok {
		return sim.Mass{}, fmt.Errorf("mass %d not found", id)
	}
	return mass, nil
}

func createMassStartPosition(w *world, example map[string]string) error {
	world := ensureDomainWorld(w)
	id, err := intValue(example, "mass_id")
	if err != nil {
		return err
	}
	position, err := startPosition(example)
	if err != nil {
		return err
	}
	if setMassStartPosition(world, id, position) {
		return nil
	}
	return world.AddMass(sim.Mass{ID: id, Position: position, Mass: 1})
}

func setMassStartPosition(world *sim.Simulation, id int, position sim.Vec2) bool {
	for i := range world.Masses {
		if world.Masses[i].ID == id {
			world.Masses[i].Position = position
			return true
		}
	}
	return false
}

func startPosition(example map[string]string) (sim.Vec2, error) {
	value, err := stringValue(example, "start_position")
	if err != nil {
		return sim.Vec2{}, err
	}
	if value == "initial" {
		return sim.Vec2{X: 5, Y: 6}, nil
	}
	return positionValue(example, "start_position")
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
	return assertZeroVelocity(id, mass.Velocity)
}

func assertZeroVelocity(id int, velocity sim.Vec2) error {
	if velocity != (sim.Vec2{}) {
		return fmt.Errorf("mass %d velocity changed to %#v", id, velocity)
	}
	return nil
}

func createWorldInState(w *world, example map[string]string) error {
	world, err := worldFromExampleState(example)
	if err != nil {
		return err
	}
	w.domainWorld = world
	return nil
}

func assertResultDeterministic(_ *world, example map[string]string) error {
	first, second, duration, err := deterministicWorlds(example)
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

func deterministicWorlds(example map[string]string) (*sim.Simulation, *sim.Simulation, float64, error) {
	duration, err := durationValue(example, "duration")
	if err != nil {
		return nil, nil, 0, err
	}
	first, err := worldFromExampleState(example)
	if err != nil {
		return nil, nil, 0, err
	}
	return first, first.Clone(), duration, nil
}

func worldFromExampleState(example map[string]string) (*sim.Simulation, error) {
	state, err := stringValue(example, "initial_state")
	if err != nil {
		return nil, err
	}
	return worldForState(state)
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
	advanceInFrames(world, duration, frameRate)
	w.resultingWorld = world
	return nil
}

func advanceInFrames(world *sim.Simulation, duration, frameRate float64) {
	_ = frameRate
	world.AdvanceDuration(duration)
}

func nextFrameStep(remaining, frameDuration float64) float64 {
	if remaining < frameDuration {
		return remaining
	}
	return frameDuration
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
	if len(a.Masses) != len(b.Masses) || !sameFloat(a.Time, b.Time) {
		return false
	}
	for i := range a.Masses {
		if !sameMassState(a.Masses[i], b.Masses[i]) {
			return false
		}
	}
	return true
}

func sameMassState(a, b sim.Mass) bool {
	return a.Position == b.Position && a.Velocity == b.Velocity
}

func sameFloat(a, b float64) bool {
	return math.Abs(a-b) <= 0.000001
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
	if duration, ok := durations[value]; ok {
		return duration, nil
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("unsupported duration %q", value)
	}
	if parsed <= 0 {
		return 0, fmt.Errorf("duration must be positive: %q", value)
	}
	return parsed, nil
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
