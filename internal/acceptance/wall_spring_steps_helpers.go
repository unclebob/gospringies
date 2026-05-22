package acceptance

import (
	"fmt"
	"math"
	"slices"
	"strings"

	"springs/internal/app"
	"springs/internal/sim"
)

func addBarrierSpring(w *world, example map[string]string) error {
	if err := requireWallSpringExampleValues(example, map[string]string{"spring_id": "1", "mass_a": "1", "mass_b": "2"}); err != nil {
		return err
	}
	springID, err := intValue(example, "spring_id")
	if err != nil {
		return err
	}
	massA, err := intValue(example, "mass_a")
	if err != nil {
		return err
	}
	massB, err := intValue(example, "mass_b")
	if err != nil {
		return err
	}
	world := ensureDomainWorld(w)
	ensureWallSpringMass(world, massA, sim.Vec2{})
	ensureWallSpringMass(world, massB, sim.Vec2{X: 30})
	return world.AddSpring(sim.Spring{ID: springID, MassA: massA, MassB: massB})
}

func setBarrierSpringWall(w *world, example map[string]string) error {
	if err := requireWallSpringExampleValues(example, map[string]string{"spring_id": "1"}); err != nil {
		return err
	}
	wall, err := boolValue(example, "wall")
	if err != nil {
		return err
	}
	springID, err := intValue(example, "spring_id")
	if err != nil {
		return err
	}
	if err := ensureBarrierSpring(w, springID); err != nil {
		return err
	}
	return updateBarrierSpring(w, example, func(spring *sim.Spring) { spring.Wall = wall })
}

func setBarrierSpringWallFalse(w *world, example map[string]string) error {
	return setBarrierSpringWall(w, withExampleValue(example, "wall", "false"))
}

func setBarrierSpringParameters(w *world, example map[string]string) error {
	if err := requireWallSpringExampleValues(example, map[string]string{"kspring": "10", "kdamp": "0.5", "rest_len": "20"}); err != nil {
		return err
	}
	kspring, err := floatValue(example, "kspring")
	if err != nil {
		return err
	}
	kdamp, err := floatValue(example, "kdamp")
	if err != nil {
		return err
	}
	restLen, err := floatValue(example, "rest_len")
	if err != nil {
		return err
	}
	return updateBarrierSpring(w, example, func(spring *sim.Spring) {
		spring.SpringConstant = kspring
		spring.Stiffness = kspring
		spring.Damping = kdamp
		spring.RestLength = restLen
	})
}

func evaluateBarrierSpringForces(w *world, _ map[string]string) error {
	world := ensureDomainWorld(w)
	if len(world.Masses) >= 2 {
		world.Masses[1].Velocity = sim.Vec2{X: 10}
	}
	w.forceEvaluation = world.EvaluateForces()
	return nil
}

func assertBarrierSpringForceState(w *world, example map[string]string) error {
	return assertBarrierSpringForceStateKey(w, example, "spring_force_state")
}

func assertBarrierSpringDampingState(w *world, example map[string]string) error {
	return assertBarrierSpringForceStateKey(w, example, "damping_force_state")
}

func assertBarrierSpringForceStateKey(w *world, example map[string]string, key string) error {
	state, err := stringValue(example, key)
	if err != nil {
		return err
	}
	if !validWallSpringForceState(state) {
		return fmt.Errorf("invalid force state %q", state)
	}
	enabled := forceEvaluationHasForce(w.forceEvaluation)
	if (state == "enabled") != enabled {
		return fmt.Errorf("%s force enabled = %t, forces = %#v", key, enabled, w.forceEvaluation.ByMassID)
	}
	return nil
}

func validWallSpringForceState(state string) bool {
	return state == "enabled" || state == "disabled"
}

func createWallSpringLengthConstraint(w *world, example map[string]string) error {
	if err := requireWallSpringExampleValues(example, map[string]string{"spring_id": "1"}); err != nil {
		return err
	}
	if err := requireWallSpringLengthExample(example); err != nil {
		return err
	}
	springID, err := intValue(example, "spring_id")
	if err != nil {
		return err
	}
	values, err := floatValues(example, "initial_length", "rest_len")
	if err != nil {
		return err
	}
	world := ensureDomainWorld(w)
	ensureWallSpringMass(world, 1, sim.Vec2{})
	ensureWallSpringMass(world, 2, sim.Vec2{X: values[0]})
	return world.AddSpring(sim.Spring{ID: springID, MassA: 1, MassB: 2, RestLength: values[1], Wall: true})
}

func requireWallSpringLengthExample(example map[string]string) error {
	for _, expected := range wallSpringLengthExamples {
		if wallSpringExampleMatches(example, expected) {
			return nil
		}
	}
	return fmt.Errorf("unsupported wall spring length example")
}

var wallSpringLengthExamples = []map[string]string{
	{"initial_length": "120", "rest_len": "100", "endpoint_a": "1", "endpoint_b": "2", "fixed_a": "false", "fixed_b": "false", "expected_length": "100", "correction_direction": "along segment"},
	{"initial_length": "80", "rest_len": "100", "endpoint_a": "1", "endpoint_b": "2", "fixed_a": "false", "fixed_b": "false", "expected_length": "100", "correction_direction": "along segment"},
	{"initial_length": "120", "rest_len": "100", "endpoint_a": "1", "endpoint_b": "2", "fixed_a": "true", "fixed_b": "false", "expected_length": "100", "correction_direction": "along segment"},
}

func wallSpringExampleMatches(example map[string]string, expected map[string]string) bool {
	for key, want := range expected {
		if example[key] != want {
			return false
		}
	}
	return true
}

func advanceWallSpringLengthConstraint(w *world, _ map[string]string) error {
	return advanceWallSpringWorld(w)
}

func assertWallSpringEndpointDistance(w *world, example map[string]string) error {
	springID, err := intValue(example, "spring_id")
	if err != nil {
		return err
	}
	expected, err := floatValue(example, "expected_length")
	if err != nil {
		return err
	}
	distance, err := wallSpringEndpointDistance(ensureDomainWorld(w), springID)
	if err != nil {
		return err
	}
	if !closeWallSpringEndpointDistance(distance, expected) {
		return fmt.Errorf("expected wall spring endpoint distance %f got %f", expected, distance)
	}
	return nil
}

func closeWallSpringEndpointDistance(got, expected float64) bool {
	return math.Abs(got-expected) <= 0.00001
}

func assertWallSpringEndpointCorrection(w *world, example map[string]string) error {
	direction, err := stringValue(example, "correction_direction")
	if err != nil {
		return err
	}
	if direction != "along segment" {
		return fmt.Errorf("unsupported correction direction %q", direction)
	}
	world := ensureDomainWorld(w)
	for _, massID := range wallSpringEndpointMassIDs {
		if err := assertWallSpringEndpointOnSegment(world, massID); err != nil {
			return err
		}
	}
	return nil
}

var wallSpringEndpointMassIDs = []int{1, 2}

func assertWallSpringEndpointOnSegment(world *sim.Simulation, massID int) error {
	mass, ok := world.MassByID(massID)
	if !ok {
		return fmt.Errorf("wall spring endpoint %d not found", massID)
	}
	if !sameFloat(mass.Position.Y, 0) {
		return fmt.Errorf("wall spring endpoint %d correction left segment: %#v", massID, mass.Position)
	}
	return nil
}

func wallSpringEndpointDistance(world *sim.Simulation, springID int) (float64, error) {
	spring, ok := world.SpringByID(springID)
	if !ok {
		return 0, fmt.Errorf("spring %d not found", springID)
	}
	a, okA := world.MassByID(spring.MassA)
	b, okB := world.MassByID(spring.MassB)
	if !okA || !okB {
		return 0, fmt.Errorf("spring %d endpoints not found", springID)
	}
	delta := b.Position.Sub(a.Position)
	return math.Sqrt(delta.X*delta.X + delta.Y*delta.Y), nil
}

func forceEvaluationHasForce(evaluation sim.ForceEvaluation) bool {
	for _, forces := range evaluation.ByMassID {
		if forces.Force != (sim.Vec2{}) {
			return true
		}
	}
	return false
}

func createWallSpringByCoordinates(w *world, example map[string]string) error {
	return createDefaultWallSpringByCoordinateKeys(w, example, "spring_id", []string{"wall_x1", "wall_y1", "wall_x2", "wall_y2"})
}

func createMovingWallSpringByCoordinates(w *world, example map[string]string) error {
	if err := requireWallSpringExampleValues(example, map[string]string{
		"spring_id": "1",
		"wall_x1":   "-5",
		"wall_y1":   "0",
		"wall_x2":   "-5",
		"wall_y2":   "100",
		"wall_vx":   "10",
		"wall_vy":   "0",
	}); err != nil {
		return err
	}
	springID, err := intValue(example, "spring_id")
	if err != nil {
		return err
	}
	values, err := floatValues(example, "wall_x1", "wall_y1", "wall_x2", "wall_y2", "wall_vx", "wall_vy")
	if err != nil {
		return err
	}
	world := ensureDomainWorld(w)
	ensureMovingWallSpringMass(world, 1, sim.Vec2{X: values[0], Y: values[1]}, sim.Vec2{X: values[4], Y: values[5]})
	ensureMovingWallSpringMass(world, 2, sim.Vec2{X: values[2], Y: values[3]}, sim.Vec2{X: values[4], Y: values[5]})
	return world.AddSpring(sim.Spring{ID: springID, MassA: 1, MassB: 2, Wall: true})
}

func createBarrierWallSpringByCoordinates(w *world, example map[string]string) error {
	return createDefaultWallSpringByCoordinateKeys(w, example, "barrier_spring", []string{"barrier_x1", "barrier_y1", "barrier_x2", "barrier_y2"})
}

func createDefaultWallSpringByCoordinateKeys(w *world, example map[string]string, springKey string, coordinateKeys []string) error {
	if err := requireWallSpringExampleValues(example, defaultWallSpringCoordinateExample(springKey, coordinateKeys)); err != nil {
		return err
	}
	return createWallSpringByCoordinateKeys(w, example, springKey, coordinateKeys...)
}

func defaultWallSpringCoordinateExample(springKey string, coordinateKeys []string) map[string]string {
	return map[string]string{
		springKey:         "1",
		coordinateKeys[0]: "0",
		coordinateKeys[1]: "0",
		coordinateKeys[2]: "0",
		coordinateKeys[3]: "100",
	}
}

func createWallSpringByCoordinateKeys(w *world, example map[string]string, springKey string, coordinateKeys ...string) error {
	springID, err := intValue(example, springKey)
	if err != nil {
		return err
	}
	values, err := floatValues(example, coordinateKeys...)
	if err != nil {
		return err
	}
	world := ensureDomainWorld(w)
	ensureWallSpringMass(world, 1, sim.Vec2{X: values[0], Y: values[1]})
	ensureWallSpringMass(world, 2, sim.Vec2{X: values[2], Y: values[3]})
	return world.AddSpring(sim.Spring{ID: springID, MassA: 1, MassB: 2, Wall: true})
}

func createConstrainedWallSpringEndpointA(w *world, example map[string]string) error {
	return createConstrainedWallSpringEndpoint(w, example, "endpoint_a", "endpoint_a_x", "endpoint_a_y")
}

func createConstrainedWallSpringEndpointB(w *world, example map[string]string) error {
	return createConstrainedWallSpringEndpoint(w, example, "endpoint_b", "endpoint_b_x", "endpoint_b_y")
}

func createConstrainedWallSpringEndpoint(w *world, example map[string]string, endpointKey, xKey, yKey string) error {
	id, err := intValue(example, endpointKey)
	if err != nil {
		return err
	}
	values, err := floatValues(example, xKey, yKey)
	if err != nil {
		return err
	}
	ensureWallSpringMass(ensureDomainWorld(w), id, sim.Vec2{X: values[0], Y: values[1]})
	return nil
}

func createConstrainedWallSpring(w *world, example map[string]string) error {
	springID, err := intValue(example, "moving_spring")
	if err != nil {
		return err
	}
	endpointA, endpointB, err := intPair(example, "endpoint_a", "endpoint_b")
	if err != nil {
		return err
	}
	restLength, err := floatValue(example, "rest_len")
	if err != nil {
		return err
	}
	return ensureDomainWorld(w).AddSpring(sim.Spring{ID: springID, MassA: endpointA, MassB: endpointB, RestLength: restLength, Wall: true})
}

func createBarrierMovingMass(w *world, example map[string]string) error {
	if err := requireWallSpringExampleValues(example, map[string]string{
		"mass_id": "3",
		"mass_x":  "-5",
		"mass_y":  "50",
		"mass_vx": "10",
		"mass_vy": "0",
	}); err != nil {
		return err
	}
	id, err := intValue(example, "mass_id")
	if err != nil {
		return err
	}
	values, err := floatValues(example, "mass_x", "mass_y", "mass_vx", "mass_vy")
	if err != nil {
		return err
	}
	world := ensureDomainWorld(w)
	if err := world.AddMass(sim.Mass{ID: id, Position: sim.Vec2{X: values[0], Y: values[1]}, Velocity: sim.Vec2{X: values[2], Y: values[3]}, Mass: 1}); err != nil {
		return err
	}
	return rememberWallSpringStartingSide(w, example, id)
}

func createFastBarrierMovingMass(w *world, example map[string]string) error {
	if err := requireWallSpringExampleValues(example, map[string]string{
		"mass_id":  "3",
		"mass_x":   "-50",
		"mass_y":   "50",
		"mass_vx":  "1000",
		"mass_vy":  "0",
		"duration": "1 step",
	}); err != nil {
		return err
	}
	return createBarrierMovingMassWithoutExampleRestriction(w, example)
}

func createBarrierMovingMassWithoutExampleRestriction(w *world, example map[string]string) error {
	id, err := intValue(example, "mass_id")
	if err != nil {
		return err
	}
	values, err := floatValues(example, "mass_x", "mass_y", "mass_vx", "mass_vy")
	if err != nil {
		return err
	}
	world := ensureDomainWorld(w)
	if err := world.AddMass(sim.Mass{ID: id, Position: sim.Vec2{X: values[0], Y: values[1]}, Velocity: sim.Vec2{X: values[2], Y: values[3]}, Mass: 1}); err != nil {
		return err
	}
	return rememberWallSpringStartingSide(w, example, id)
}

func createBarrierStationaryMass(w *world, example map[string]string) error {
	if err := requireWallSpringExampleValues(example, map[string]string{
		"mass_id": "3",
		"mass_x":  "0",
		"mass_y":  "50",
	}); err != nil {
		return err
	}
	id, err := intValue(example, "mass_id")
	if err != nil {
		return err
	}
	values, err := floatValues(example, "mass_x", "mass_y")
	if err != nil {
		return err
	}
	world := ensureDomainWorld(w)
	if err := world.AddMass(sim.Mass{ID: id, Position: sim.Vec2{X: values[0], Y: values[1]}, Mass: 1}); err != nil {
		return err
	}
	return rememberWallSpringStartingSide(w, example, id)
}

func advanceThroughWallSpringCollision(w *world, _ map[string]string) error {
	return advanceWallSpringWorld(w)
}

func advanceThroughWallSpringCollisionByDuration(w *world, example map[string]string) error {
	duration, err := stringValue(example, "duration")
	if err != nil {
		return err
	}
	if duration != "1 step" {
		return fmt.Errorf("unsupported wall spring collision duration %q", duration)
	}
	return advanceWallSpringWorld(w)
}

func advanceWallSpringLengthConstraintsAndCollisions(w *world, example map[string]string) error {
	if err := rememberWallSpringEndpointStartingSide(w, example, "endpoint_a"); err != nil {
		return err
	}
	if err := rememberWallSpringEndpointStartingSide(w, example, "endpoint_b"); err != nil {
		return err
	}
	return advanceWallSpringWorld(w)
}

func rememberWallSpringEndpointStartingSide(w *world, example map[string]string, endpointKey string) error {
	endpointID, err := intValue(example, endpointKey)
	if err != nil {
		return err
	}
	return rememberWallSpringStartingSide(w, withBarrierSpringID(example), endpointID)
}

func withBarrierSpringID(example map[string]string) map[string]string {
	return withExampleValue(example, "spring_id", example["barrier_spring"])
}

func assertMassOnStartingWallSpringSide(w *world, example map[string]string) error {
	state, err := wallSpringMassState(w, example)
	if err != nil {
		return err
	}
	current, err := wallSpringSide(state.world, state.springID, state.mass.Position)
	if err != nil {
		return err
	}
	if current*state.side < 0 {
		return fmt.Errorf("mass %d crossed wall spring %d: side %f started %f", state.massID, state.springID, current, state.side)
	}
	return nil
}

func assertWallSpringEndpointAOnStartingBarrierSide(w *world, example map[string]string) error {
	return assertWallSpringEndpointOnStartingBarrierSide(w, example, "endpoint_a")
}

func assertWallSpringEndpointBOnStartingBarrierSide(w *world, example map[string]string) error {
	return assertWallSpringEndpointOnStartingBarrierSide(w, example, "endpoint_b")
}

func assertWallSpringEndpointOnStartingBarrierSide(w *world, example map[string]string, endpointKey string) error {
	endpointID, err := intValue(example, endpointKey)
	if err != nil {
		return err
	}
	return assertMassOnStartingWallSpringSide(w, withExampleValue(withBarrierSpringID(example), "mass_id", fmt.Sprintf("%d", endpointID)))
}

func assertMassVelocityResolvedAwayFromWallSpring(w *world, example map[string]string) error {
	state, err := wallSpringMassState(w, example)
	if err != nil {
		return err
	}
	normal, err := wallSpringNormal(state.world, state.springID)
	if err != nil {
		return err
	}
	if dotAcceptance(state.mass.Velocity, normal)*state.side < 0 {
		return fmt.Errorf("mass %d velocity penetrates wall spring %d: %#v", state.massID, state.springID, state.mass.Velocity)
	}
	return nil
}

func assertMovingWallSpringVelocityResolvedAwayFromMass(w *world, example map[string]string) error {
	state, err := wallSpringMassState(w, example)
	if err != nil {
		return err
	}
	normal, err := wallSpringNormal(state.world, state.springID)
	if err != nil {
		return err
	}
	wallVelocity, err := wallSpringVelocity(state.world, state.springID)
	if err != nil {
		return err
	}
	if dotAcceptance(wallVelocity.Sub(state.mass.Velocity), normal)*state.side > 0.000001 {
		return fmt.Errorf("wall spring %d velocity penetrates mass %d: wall=%#v mass=%#v", state.springID, state.massID, wallVelocity, state.mass.Velocity)
	}
	return nil
}

func wallSpringVelocity(world *sim.Simulation, springID int) (sim.Vec2, error) {
	spring, ok := world.SpringByID(springID)
	if !ok {
		return sim.Vec2{}, fmt.Errorf("spring %d not found", springID)
	}
	a, okA := world.MassByID(spring.MassA)
	b, okB := world.MassByID(spring.MassB)
	if !okA || !okB {
		return sim.Vec2{}, fmt.Errorf("spring %d endpoints not found", springID)
	}
	return a.Velocity.Add(b.Velocity).Scale(0.5), nil
}

func createWallSpringByEndpointIDs(w *world, example map[string]string) error {
	if err := requireWallSpringExampleValues(example, map[string]string{"spring_id": "1", "endpoint_a": "1", "endpoint_b": "2"}); err != nil {
		return err
	}
	springID, err := intValue(example, "spring_id")
	if err != nil {
		return err
	}
	endpointA, err := intValue(example, "endpoint_a")
	if err != nil {
		return err
	}
	endpointB, err := intValue(example, "endpoint_b")
	if err != nil {
		return err
	}
	world := ensureDomainWorld(w)
	ensureWallSpringMass(world, endpointA, sim.Vec2{})
	ensureWallSpringMass(world, endpointB, sim.Vec2{Y: 100})
	return world.AddSpring(sim.Spring{ID: springID, MassA: endpointA, MassB: endpointB, Wall: true})
}

func setWallSpringEndpointFixed(w *world, example map[string]string) error {
	return setWallSpringNamedEndpointFixed(w, example, "endpoint_a", "fixed_a")
}

func setWallSpringEndpointBFixed(w *world, example map[string]string) error {
	return setWallSpringNamedEndpointFixed(w, example, "endpoint_b", "fixed_b")
}

func setWallSpringNamedEndpointFixed(w *world, example map[string]string, endpointKey, fixedKey string) error {
	endpoint, err := intValue(example, endpointKey)
	if err != nil {
		return err
	}
	fixed, err := boolValue(example, fixedKey)
	if err != nil {
		return err
	}
	world := ensureDomainWorld(w)
	for i := range world.Masses {
		if world.Masses[i].ID == endpoint {
			world.Masses[i].Fixed = fixed
			return nil
		}
	}
	return fmt.Errorf("endpoint %d not found", endpoint)
}

func createMassCollidingWithWallSpring(w *world, example map[string]string) error {
	massID, mass, err := wallSpringCollisionMass(example)
	if err != nil {
		return err
	}
	world := ensureDomainWorld(w)
	if err := world.AddMass(mass); err != nil {
		return err
	}
	w.wallSpringImpulses = map[int]sim.Vec2{}
	for _, mass := range world.Masses {
		w.wallSpringImpulses[mass.ID] = mass.Velocity
	}
	return rememberWallSpringStartingSide(w, example, massID)
}

func wallSpringCollisionMass(example map[string]string) (int, sim.Mass, error) {
	if err := requireWallSpringExampleValues(example, map[string]string{"spring_id": "1", "mass_id": "3"}); err != nil {
		return 0, sim.Mass{}, err
	}
	massID, contactFraction, err := intAndFloat(example, "mass_id", "contact_fraction")
	if err != nil {
		return 0, sim.Mass{}, err
	}
	mass := sim.Mass{ID: massID, Position: sim.Vec2{X: -5, Y: 100 * contactFraction}, Velocity: sim.Vec2{X: 10}, Mass: 1}
	return massID, mass, nil
}

func resolveWallSpringCollision(w *world, _ map[string]string) error {
	return advanceWallSpringWorld(w)
}

func advanceWallSpringWorld(w *world) error {
	ensureDomainWorld(w).Step(1)
	return nil
}

func assertWallSpringEndpointImpulseShare(w *world, example map[string]string) error {
	return assertWallSpringNamedEndpointImpulseShare(w, example, "endpoint_a", "impulse_share_a")
}

func assertWallSpringEndpointBImpulseShare(w *world, example map[string]string) error {
	return assertWallSpringNamedEndpointImpulseShare(w, example, "endpoint_b", "impulse_share_b")
}

func assertWallSpringNamedEndpointImpulseShare(w *world, example map[string]string, endpointKey, shareKey string) error {
	endpoint, expectedShare, err := endpointImpulseExpectation(example, endpointKey, shareKey)
	if err != nil {
		return err
	}
	actualShare, err := actualWallSpringEndpointImpulseShare(w, example, endpoint)
	if err != nil {
		return err
	}
	if actualShare < expectedShare-0.000001 || actualShare > expectedShare+0.000001 {
		return fmt.Errorf("endpoint %d impulse share = %f, expected %f", endpoint, actualShare, expectedShare)
	}
	return nil
}

func createWallSpringWithTemperature(w *world, example map[string]string) error {
	if err := createWallSpringByEndpointIDs(w, map[string]string{"spring_id": example["spring_id"], "endpoint_a": "1", "endpoint_b": "2"}); err != nil {
		return err
	}
	return setSpringTemperature(w, example)
}

func setTemperatureRandomSeed(w *world, example map[string]string) error {
	if err := requireWallSpringExampleValues(example, map[string]string{"seed": "11"}); err != nil {
		return err
	}
	seed, err := intValue(example, "seed")
	if err != nil {
		return err
	}
	ensureDomainWorld(w).SetTemperatureSeed(int64(seed))
	return nil
}

func setSpringTemperature(w *world, example map[string]string) error {
	if err := requireSupportedTemperatureExample(example); err != nil {
		return err
	}
	temperature, err := floatValue(example, "temperature")
	if err != nil {
		return err
	}
	return updateBarrierSpring(w, example, func(spring *sim.Spring) {
		spring.Temperature = temperature
	})
}

func requireSupportedTemperatureExample(example map[string]string) error {
	temperature, err := stringValue(example, "temperature")
	if err != nil {
		return err
	}
	if temperature == "0" || temperature == "10" {
		return nil
	}
	return fmt.Errorf("unsupported temperature example %q", temperature)
}

func createMassCollidingWithSpring(w *world, example map[string]string) error {
	return createMassCollidingWithWallSpring(w, withContactFraction(example, "0.50"))
}

func resolveSpringCollision(w *world, _ map[string]string) error {
	return advanceWallSpringWorld(w)
}

func assertMassTemperatureKick(w *world, example map[string]string) error {
	assertion, err := temperatureKickAssertion(w, example)
	if err != nil {
		return err
	}
	return assertTemperatureKickBehavior(assertion.behavior, assertion.kick, ensureDomainWorld(w).Bounds.Height)
}

type temperatureKickExpectation struct {
	behavior string
	kick     sim.Vec2
}

func temperatureKickAssertion(w *world, example map[string]string) (temperatureKickExpectation, error) {
	massID, err := intValue(example, "mass_id")
	if err != nil {
		return temperatureKickExpectation{}, err
	}
	behavior, err := stringValue(example, "kick_behavior")
	if err != nil {
		return temperatureKickExpectation{}, err
	}
	mass, ok := ensureDomainWorld(w).MassByID(massID)
	if !ok {
		return temperatureKickExpectation{}, fmt.Errorf("mass %d not found", massID)
	}
	return temperatureKickExpectation{
		behavior: behavior,
		kick:     mass.Velocity.Sub(expectedWallSpringCollisionVelocity(example)),
	}, nil
}

func assertTemperatureKickBehavior(behavior string, kick sim.Vec2, worldHeight float64) error {
	switch behavior {
	case "none":
		return assertVec("temperature kick", kick, 0, 0)
	case "full screen height against gravity 10":
		return assertFloat("temperature kick magnitude", math.Sqrt(kick.X*kick.X+kick.Y*kick.Y), math.Sqrt(2*10*worldHeight))
	default:
		return fmt.Errorf("unsupported temperature kick behavior %q", behavior)
	}
}

func expectedWallSpringCollisionVelocity(example map[string]string) sim.Vec2 {
	if _, ok := example["contact_fraction"]; ok {
		return sim.Vec2{X: -10}
	}
	return sim.Vec2{X: 10}
}

func withContactFraction(example map[string]string, contactFraction string) map[string]string {
	return withExampleValue(example, "contact_fraction", contactFraction)
}

func withExampleValue(example map[string]string, key string, value string) map[string]string {
	copy := map[string]string{}
	for existingKey, existingValue := range example {
		copy[existingKey] = existingValue
	}
	copy[key] = value
	return copy
}

func actualWallSpringEndpointImpulseShare(w *world, example map[string]string, endpoint int) (float64, error) {
	endpointDelta, err := wallSpringVelocityDelta(w, endpoint, "endpoint")
	if err != nil {
		return 0, err
	}
	movingMassID, err := intValue(example, "mass_id")
	if err != nil {
		return 0, err
	}
	movingDelta, err := wallSpringVelocityDelta(w, movingMassID, "moving mass")
	if err != nil {
		return 0, err
	}
	impulse := -movingDelta
	if impulse == 0 {
		return 0, nil
	}
	return endpointDelta / impulse, nil
}

func wallSpringVelocityDelta(w *world, massID int, label string) (float64, error) {
	mass, ok := ensureDomainWorld(w).MassByID(massID)
	if !ok {
		return 0, fmt.Errorf("%s %d not found", label, massID)
	}
	return mass.Velocity.X - w.wallSpringImpulses[massID].X, nil
}

func parseExpectedImpulseShare(value string) (float64, error) {
	share, ok := expectedImpulseShares[value]
	if !ok {
		return 0, fmt.Errorf("unsupported share")
	}
	return share, nil
}

var expectedImpulseShares = map[string]float64{
	"0.75":     0.75,
	"0.50":     0.50,
	"0.25":     0.25,
	"0":        0,
	"absorbed": 0,
}

func createWallSpringXSPInput(w *world, example map[string]string) error {
	wall, err := stringValue(example, "input_wall")
	if err != nil {
		return err
	}
	suffix := ""
	if wall != "absent" {
		suffix = " " + wall
	}
	w.xspInput = "#1.0\nmass 1 0 0 1 0.8\nmass 2 10 0 1 0.8\nspng 1 1 2 12 0.7 10" + suffix + "\n"
	return nil
}

func assertLoadedWallSpringXSP(w *world, example map[string]string) error {
	springID, err := intValue(example, "spring_id")
	if err != nil {
		return err
	}
	expected, err := boolValue(example, "loaded_wall")
	if err != nil {
		return err
	}
	spring, ok := w.xspWorld.SpringByID(springID)
	if !ok {
		return fmt.Errorf("spring %d not found", springID)
	}
	if spring.Wall != expected {
		return fmt.Errorf("loaded spring wall = %t, expected %t", spring.Wall, expected)
	}
	return nil
}

func assertSavedWallSpringXSP(w *world, example map[string]string) error {
	return assertSavedSpringXSPValue(w, example, "saved_wall", savedWallSpringXSPNeedle)
}

func createTemperatureSpringXSPInput(w *world, example map[string]string) error {
	temperature, err := stringValue(example, "input_temperature")
	if err != nil {
		return err
	}
	suffix := " true"
	if temperature != "absent" {
		suffix += " " + temperature
	}
	w.xspInput = "#1.0\nmass 1 0 0 1 0.8\nmass 2 10 0 1 0.8\nspng 1 1 2 12 0.7 10" + suffix + "\n"
	return nil
}

func assertLoadedSpringTemperatureXSP(w *world, example map[string]string) error {
	springID, err := intValue(example, "spring_id")
	if err != nil {
		return err
	}
	expected, err := floatValue(example, "loaded_temperature")
	if err != nil {
		return err
	}
	spring, ok := w.xspWorld.SpringByID(springID)
	if !ok {
		return fmt.Errorf("spring %d not found", springID)
	}
	return assertFloat("loaded spring temperature", spring.Temperature, expected)
}

func assertSavedSpringTemperatureXSP(w *world, example map[string]string) error {
	return assertSavedSpringXSPValue(w, example, "saved_temperature", savedTemperatureSpringXSPNeedle)
}

func assertSavedSpringXSPValue(w *world, example map[string]string, valueKey string, needleFor func(int, string) string) error {
	springID, err := intValue(example, "spring_id")
	if err != nil {
		return err
	}
	expected, err := stringValue(example, valueKey)
	if err != nil {
		return err
	}
	return assertSavedSpringXSPContains(w, needleFor(springID, expected))
}

func savedWallSpringXSPNeedle(springID int, expected string) string {
	return fmt.Sprintf("spng %d 1 2 12 0.7 10 %s 0\n", springID, expected)
}

func savedTemperatureSpringXSPNeedle(springID int, expected string) string {
	return fmt.Sprintf("spng %d 1 2 12 0.7 10 true %s\n", springID, expected)
}

func assertSavedSpringXSPContains(w *world, needle string) error {
	if !strings.Contains(w.xspSavedFirst, needle) {
		return fmt.Errorf("saved XSP missing %q in:\n%s", needle, w.xspSavedFirst)
	}
	return nil
}

func createSelectedSpringWithWall(w *world, example map[string]string) error {
	return createAppSpringWithWall(w, example, "old_wall", true)
}

func createSelectedSpringsWithWalls(w *world, example map[string]string) error {
	if err := requireWallSpringExampleValues(example, map[string]string{
		"spring_ids": "1, 2, 3",
		"old_walls":  "false, false, true",
		"new_wall":   "true",
		"new_walls":  "true, true, true",
	}); err != nil {
		return err
	}
	springIDs, walls, err := springIDsAndWalls(example, "old_walls")
	if err != nil {
		return err
	}
	game, err := newAppGameWithSprings(springIDs, walls)
	if err != nil {
		return err
	}
	if err := game.SelectSprings(springIDs...); err != nil {
		return err
	}
	w.appGame = game
	return nil
}

func changeSpringWallControl(w *world, example map[string]string) error {
	game, err := concreteApplicationDriverWithMessage(w, "application game is not concrete")
	if err != nil {
		return err
	}
	if !game.ClickVisibleControl("Wall") {
		return fmt.Errorf("Wall control click was not handled")
	}
	return assertSelectedSpringsMatchRequestedWall(w, example)
}

func assertSelectedSpringsMatchRequestedWall(w *world, example map[string]string) error {
	requestedWall, err := boolValue(example, "new_wall")
	if err != nil {
		return err
	}
	game, err := concreteApplicationDriverWithMessage(w, "application game is not concrete")
	if err != nil {
		return err
	}
	for _, spring := range game.World().Springs {
		if spring.Wall != requestedWall {
			return fmt.Errorf("spring %d wall = %t, expected requested %t", spring.ID, spring.Wall, requestedWall)
		}
	}
	return nil
}

func assertSpringWallValue(w *world, example map[string]string) error {
	springID, expected, err := springIDAndWall(example, "new_wall")
	if err != nil {
		return err
	}
	world, err := wallSpringAssertionWorld(w)
	if err != nil {
		return err
	}
	spring, ok := world.SpringByID(springID)
	if !ok {
		return fmt.Errorf("spring %d not found", springID)
	}
	if spring.Wall != expected {
		return fmt.Errorf("spring %d wall = %t, expected %t", springID, spring.Wall, expected)
	}
	return nil
}

func wallSpringAssertionWorld(w *world) (*sim.Simulation, error) {
	if w.appGame == nil {
		return ensureDomainWorld(w), nil
	}
	game, err := concreteApplicationDriverWithMessage(w, "application game is not concrete")
	if err != nil {
		return nil, err
	}
	return game.World(), nil
}

func assertSelectedSpringsWallValues(w *world, example map[string]string) error {
	springIDs, expectedWalls, err := springIDsAndWalls(example, "new_walls")
	if err != nil {
		return err
	}
	game, err := concreteApplicationDriverWithMessage(w, "application game is not concrete")
	if err != nil {
		return err
	}
	for i, springID := range springIDs {
		if err := assertAppSpringWallValue(game, springID, expectedWalls[i]); err != nil {
			return err
		}
	}
	return nil
}

func assertAppSpringWallValue(game *app.Game, springID int, expected bool) error {
	spring, ok := game.World().SpringByID(springID)
	if !ok {
		return fmt.Errorf("spring %d not found", springID)
	}
	if spring.Wall != expected {
		return fmt.Errorf("spring %d wall = %t, expected %t", springID, spring.Wall, expected)
	}
	return nil
}

func createMenuSpringWithWall(w *world, example map[string]string) error {
	return createAppSpringWithWall(w, example, "old_wall", false)
}

func createMenuSpringWithTemperature(w *world, example map[string]string) error {
	springID, err := intValue(example, "spring_id")
	if err != nil {
		return err
	}
	temperature, err := floatValue(example, "old_temperature")
	if err != nil {
		return err
	}
	game, err := newAppGameWithSpring(springID, false)
	if err != nil {
		return err
	}
	game.World().Springs[0].Temperature = temperature
	w.appGame = game
	return nil
}

func assertSpringMenuIncludesItem(w *world, example map[string]string) error {
	springID, err := intValue(example, "spring_id")
	if err != nil {
		return err
	}
	item, err := stringValue(example, "menu_item")
	if err != nil {
		return err
	}
	labels, err := springContextMenuLabels(w, springID)
	if err != nil {
		return err
	}
	if slices.Contains(labels, item) {
		return nil
	}
	return fmt.Errorf("spring menu did not include %q", item)
}

func selectSpringMenuTemperatureItem(w *world, example map[string]string) error {
	return selectSpringMenuItem(w, example, "Temperature")
}

func assertSpringTemperatureDialogRange(w *world, example map[string]string) error {
	actual, err := springTemperatureDialogRange(w)
	if err != nil {
		return err
	}
	expected, err := expectedSpringTemperatureDialogRange(example)
	if err != nil {
		return err
	}
	if err := assertFloat("temperature dialog minimum", actual.minimum, expected.minimum); err != nil {
		return err
	}
	return assertFloat("temperature dialog maximum", actual.maximum, expected.maximum)
}

type temperatureDialogRange struct {
	minimum float64
	maximum float64
}

func springTemperatureDialogRange(w *world) (temperatureDialogRange, error) {
	game, err := concreteApplicationDriverWithMessage(w, "application game is not concrete")
	if err != nil {
		return temperatureDialogRange{}, fmt.Errorf("application game is not concrete")
	}
	minimum, maximum, open := game.SpringTemperatureDialogRange()
	if !open {
		return temperatureDialogRange{}, fmt.Errorf("spring Temperature dialog was not open")
	}
	return temperatureDialogRange{minimum: minimum, maximum: maximum}, nil
}

func expectedSpringTemperatureDialogRange(example map[string]string) (temperatureDialogRange, error) {
	expectedMinimum, err := floatValue(example, "minimum")
	if err != nil {
		return temperatureDialogRange{}, err
	}
	expectedMaximum, err := floatValue(example, "maximum")
	if err != nil {
		return temperatureDialogRange{}, err
	}
	return temperatureDialogRange{minimum: expectedMinimum, maximum: expectedMaximum}, nil
}

func changeSpringTemperatureDialogValue(w *world, example map[string]string) error {
	game, err := concreteApplicationDriverWithMessage(w, "application game is not concrete")
	if err != nil {
		return err
	}
	value, err := stringValue(example, "new_temperature")
	if err != nil {
		return err
	}
	if !game.ApplyValueDialogText(value) {
		return fmt.Errorf("spring Temperature dialog value was not handled")
	}
	return nil
}

func assertSpringTemperatureValue(w *world, example map[string]string) error {
	springID, err := intValue(example, "spring_id")
	if err != nil {
		return err
	}
	expected, err := floatValue(example, "new_temperature")
	if err != nil {
		return err
	}
	game, err := concreteApplicationDriverWithMessage(w, "application game is not concrete")
	if err != nil {
		return err
	}
	spring, ok := game.World().SpringByID(springID)
	if !ok {
		return fmt.Errorf("spring %d not found", springID)
	}
	return assertFloat("spring temperature", spring.Temperature, expected)
}

func selectSpringMenuWallItem(w *world, example map[string]string) error {
	return selectSpringMenuItem(w, example, "Wall")
}

func selectSpringMenuItem(w *world, example map[string]string, item string) error {
	springID, err := intValue(example, "spring_id")
	if err != nil {
		return err
	}
	game, err := concreteApplicationDriverWithMessage(w, "application game is not concrete")
	if err != nil {
		return err
	}
	if !game.SelectSpringContextMenuItem(springID, item) {
		return fmt.Errorf("%s spring menu item was not handled", item)
	}
	return nil
}

func createRenderableWallSpring(w *world, example map[string]string) error {
	return createAppSpringWithWall(w, example, "wall", false)
}

func createAppSpringWithWall(w *world, example map[string]string, wallKey string, selectSpring bool) error {
	springID, wall, err := appSpringWallExample(example, wallKey)
	if err != nil {
		return err
	}
	game, err := newAppGameWithSpring(springID, wall)
	if err != nil {
		return err
	}
	if err := selectAppSpringIfRequested(game, springID, selectSpring); err != nil {
		return err
	}
	w.appGame = game
	return nil
}

func appSpringWallExample(example map[string]string, wallKey string) (int, bool, error) {
	if err := requireWallSpringExampleValues(example, map[string]string{"spring_id": "1"}); err != nil {
		return 0, false, err
	}
	return springIDAndWall(example, wallKey)
}

func newAppGameWithSpring(springID int, wall bool) (*app.Game, error) {
	return newAppGameWithSprings([]int{springID}, []bool{wall})
}

func newAppGameWithSprings(springIDs []int, walls []bool) (*app.Game, error) {
	game := newApplicationDriverGame()
	world := sim.NewWorld()
	ensureWallSpringMass(world, 1, sim.Vec2{})
	ensureWallSpringMass(world, 2, sim.Vec2{X: 20})
	for i, springID := range springIDs {
		if err := world.AddSpring(sim.Spring{ID: springID, MassA: 1, MassB: 2, Wall: walls[i]}); err != nil {
			return nil, err
		}
	}
	game.ReplaceWorld(world)
	return game, nil
}

func selectAppSpringIfRequested(game *app.Game, springID int, selectSpring bool) error {
	if !selectSpring {
		return nil
	}
	return game.SelectSpring(springID)
}

func renderWallSpring(w *world, _ map[string]string) error {
	game, err := concreteApplicationDriverWithMessage(w, "application game is not concrete")
	if err != nil {
		game = newApplicationDriverGame()
		game.ReplaceWorld(ensureDomainWorld(w))
		w.appGame = game
	}
	w.renderResult = game.RenderWorld()
	return nil
}

func assertWallSpringRenderingStyle(w *world, example map[string]string) error {
	style, err := stringValue(example, "rendering_style")
	if err != nil {
		return err
	}
	assert, ok := wallSpringRenderingAssertions[style]
	if !ok {
		return fmt.Errorf("unsupported rendering style %q", style)
	}
	return assert(w.renderResult.Representations)
}

var wallSpringRenderingAssertions = map[string]func(map[string]string) error{
	"normal": assertNormalSpringRendering,
	"wall":   assertWallSpringRendering,
}

func assertNormalSpringRendering(representations map[string]string) error {
	return assertSpringRenderRepresentations(representations, map[string]string{
		"spring":      "cyan line",
		"wall spring": "",
	})
}

func assertWallSpringRendering(representations map[string]string) error {
	return assertSpringRenderRepresentations(representations, map[string]string{
		"wall spring": "heavy orange line",
	})
}

func assertSpringRenderRepresentations(representations map[string]string, expected map[string]string) error {
	for key, want := range expected {
		if got := representations[key]; got != want {
			return fmt.Errorf("%s representation = %q, expected %q", key, got, want)
		}
	}
	return nil
}

func updateBarrierSpring(w *world, example map[string]string, update func(*sim.Spring)) error {
	springID, err := intValue(example, "spring_id")
	if err != nil {
		return err
	}
	world := ensureDomainWorld(w)
	for i := range world.Springs {
		if world.Springs[i].ID == springID {
			update(&world.Springs[i])
			return nil
		}
	}
	return fmt.Errorf("spring %d not found", springID)
}

func ensureBarrierSpring(w *world, springID int) error {
	world := ensureDomainWorld(w)
	if _, ok := world.SpringByID(springID); ok {
		return nil
	}
	ensureWallSpringMass(world, 1, sim.Vec2{})
	ensureWallSpringMass(world, 2, sim.Vec2{X: 20})
	return world.AddSpring(sim.Spring{ID: springID, MassA: 1, MassB: 2})
}

func ensureWallSpringMass(world *sim.Simulation, id int, position sim.Vec2) {
	for i := range world.Masses {
		if world.Masses[i].ID == id {
			world.Masses[i].Position = position
			return
		}
	}
	_ = world.AddMass(sim.Mass{ID: id, Position: position, Mass: 1})
}

func ensureMovingWallSpringMass(world *sim.Simulation, id int, position sim.Vec2, velocity sim.Vec2) {
	ensureWallSpringMass(world, id, position)
	for i := range world.Masses {
		if world.Masses[i].ID == id {
			world.Masses[i].Velocity = velocity
			return
		}
	}
}

func springContextMenuLabels(w *world, springID int) ([]string, error) {
	game, err := concreteApplicationDriverWithMessage(w, "application game is not concrete")
	if err != nil {
		return nil, err
	}
	return game.SpringContextMenuLabelsForSpring(springID), nil
}

func rememberWallSpringStartingSide(w *world, example map[string]string, massID int) error {
	springID, err := intValue(example, "spring_id")
	if err != nil {
		return err
	}
	world := ensureDomainWorld(w)
	mass, ok := world.MassByID(massID)
	if !ok {
		return fmt.Errorf("mass %d not found", massID)
	}
	side, err := wallSpringSide(world, springID, mass.Position)
	if err != nil {
		return err
	}
	if w.wallSpringSides == nil {
		w.wallSpringSides = map[int]float64{}
	}
	w.wallSpringSides[massID] = side
	return nil
}

func wallSpringSide(world *sim.Simulation, springID int, point sim.Vec2) (float64, error) {
	normal, start, err := wallSpringNormalAndStart(world, springID)
	if err != nil {
		return 0, err
	}
	return dotAcceptance(point.Sub(start), normal), nil
}

func wallSpringNormal(world *sim.Simulation, springID int) (sim.Vec2, error) {
	normal, _, err := wallSpringNormalAndStart(world, springID)
	return normal, err
}

func wallSpringNormalAndStart(world *sim.Simulation, springID int) (sim.Vec2, sim.Vec2, error) {
	spring, ok := world.SpringByID(springID)
	if !ok {
		return sim.Vec2{}, sim.Vec2{}, fmt.Errorf("spring %d not found", springID)
	}
	a, okA := world.MassByID(spring.MassA)
	b, okB := world.MassByID(spring.MassB)
	if !okA || !okB {
		return sim.Vec2{}, sim.Vec2{}, fmt.Errorf("spring %d endpoints not found", springID)
	}
	segment := b.Position.Sub(a.Position)
	return sim.Vec2{X: -segment.Y, Y: segment.X}.Normalize(), a.Position, nil
}

func dotAcceptance(a, b sim.Vec2) float64 {
	return a.X*b.X + a.Y*b.Y
}

func springIDAndWall(example map[string]string, wallKey string) (int, bool, error) {
	springID, err := intValue(example, "spring_id")
	if err != nil {
		return 0, false, err
	}
	wall, err := boolValue(example, wallKey)
	return springID, wall, err
}

func springIDsAndWalls(example map[string]string, wallKey string) ([]int, []bool, error) {
	springIDs, err := editIDList(example, "spring_ids")
	if err != nil {
		return nil, nil, err
	}
	walls, err := boolValues(example, wallKey)
	if err != nil {
		return nil, nil, err
	}
	if len(springIDs) != len(walls) {
		return nil, nil, fmt.Errorf("spring_ids and %s lengths differ", wallKey)
	}
	return springIDs, walls, nil
}

func boolValues(example map[string]string, key string) ([]bool, error) {
	value, err := stringValue(example, key)
	if err != nil {
		return nil, err
	}
	parts := strings.Split(value, ",")
	values := make([]bool, 0, len(parts))
	for _, part := range parts {
		parsed, err := boolValue(map[string]string{key: strings.TrimSpace(part)}, key)
		if err != nil {
			return nil, err
		}
		values = append(values, parsed)
	}
	return values, nil
}

func requireWallSpringExampleValues(example map[string]string, expected map[string]string) error {
	for key, want := range expected {
		got, err := stringValue(example, key)
		if err != nil {
			return err
		}
		if got != want {
			return fmt.Errorf("%s = %q, expected %q", key, got, want)
		}
	}
	return nil
}

type wallSpringMassStateResult struct {
	world    *sim.Simulation
	mass     sim.Mass
	massID   int
	springID int
	side     float64
}

func wallSpringMassState(w *world, example map[string]string) (wallSpringMassStateResult, error) {
	massID, springID, err := intPair(example, "mass_id", "spring_id")
	if err != nil {
		return wallSpringMassStateResult{}, err
	}
	world := ensureDomainWorld(w)
	mass, ok := world.MassByID(massID)
	if !ok {
		return wallSpringMassStateResult{}, fmt.Errorf("mass %d not found", massID)
	}
	side, ok := w.wallSpringSides[massID]
	if !ok {
		return wallSpringMassStateResult{}, fmt.Errorf("starting side for mass %d was not recorded", massID)
	}
	return wallSpringMassStateResult{world: world, mass: mass, massID: massID, springID: springID, side: side}, nil
}

func endpointImpulseExpectation(example map[string]string, endpointKey string, shareKey string) (int, float64, error) {
	endpoint, err := intValue(example, endpointKey)
	if err != nil {
		return 0, 0, err
	}
	expected, err := stringValue(example, shareKey)
	if err != nil {
		return 0, 0, err
	}
	expectedShare, err := parseExpectedImpulseShare(expected)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid impulse share %q", expected)
	}
	return endpoint, expectedShare, nil
}
