package acceptance

import (
	"fmt"
	"math"

	"springs/internal/sim"
)

func createSpringForceWorld(w *world, example map[string]string) error {
	world := ensureDomainWorld(w)
	massA, massB, err := springForceMassIDs(example)
	if err != nil {
		return err
	}
	if err := ensureForceMass(world, massA, sim.Vec2{X: 0, Y: 0}); err != nil {
		return err
	}
	if err := ensureForceMass(world, massB, sim.Vec2{X: 12, Y: 0}); err != nil {
		return err
	}
	return ensureForceSpring(world, massA, massB)
}

func springForceMassIDs(example map[string]string) (int, int, error) {
	massA, err := intValue(example, "mass_a")
	if err != nil {
		return 0, 0, err
	}
	massB, err := intValue(example, "mass_b")
	if err != nil {
		return 0, 0, err
	}
	return massA, massB, nil
}

func ensureForceMass(world *sim.Simulation, id int, position sim.Vec2) error {
	if _, ok := world.MassByID(id); ok {
		return nil
	}
	return world.AddMass(sim.Mass{ID: id, Position: position, Mass: 1})
}

func ensureForceSpring(world *sim.Simulation, massA, massB int) error {
	if len(world.Springs) != 0 {
		return nil
	}
	return world.AddSpring(sim.Spring{ID: 1, MassA: massA, MassB: massB, RestLength: 10, SpringConstant: 1})
}

func setOnlySpringRestLength(w *world, example map[string]string) error {
	return setSpringFloat(w, example, "rest_length", setSpringRestLength)
}

func setOnlySpringConstant(w *world, example map[string]string) error {
	return setSpringFloat(w, example, "spring_constant", setSpringConstant)
}

func setMassAVelocity(w *world, example map[string]string) error {
	return setMassNamedVelocity(w, example, "mass_a", "velocity_a")
}

func setMassBVelocity(w *world, example map[string]string) error {
	return setMassNamedVelocity(w, example, "mass_b", "velocity_b")
}

func setMassNamedVelocity(w *world, example map[string]string, massKey, velocityKey string) error {
	world, err := domainWorld(w)
	if err != nil {
		return err
	}
	massID, velocity, err := massVelocityFromExample(example, massKey, velocityKey)
	if err != nil {
		return err
	}
	return setMassVelocityByID(world, massID, velocity)
}

func massVelocityFromExample(example map[string]string, massKey, velocityKey string) (int, sim.Vec2, error) {
	massID, err := intValue(example, massKey)
	if err != nil {
		return 0, sim.Vec2{}, err
	}
	value, err := stringValue(example, velocityKey)
	if err != nil {
		return 0, sim.Vec2{}, err
	}
	velocity, err := namedVelocity(value)
	if err != nil {
		return 0, sim.Vec2{}, err
	}
	return massID, velocity, nil
}

func setMassVelocityByID(world *sim.Simulation, massID int, velocity sim.Vec2) error {
	for i := range world.Masses {
		if world.Masses[i].ID == massID {
			world.Masses[i].Velocity = velocity
			return nil
		}
	}
	return fmt.Errorf("mass %d not found", massID)
}

func setOnlySpringDamping(w *world, example map[string]string) error {
	return setSpringFloat(w, example, "damping_constant", setSpringDamping)
}

func setSpringFloat(w *world, example map[string]string, key string, apply func(*sim.Spring, float64)) error {
	return updateFirstSpring(w, func(spring *sim.Spring) error {
		value, err := floatValue(example, key)
		if err != nil {
			return err
		}
		apply(spring, value)
		return nil
	})
}

func evaluateForces(w *world, _ map[string]string) error {
	world, err := domainWorld(w)
	if err != nil {
		return err
	}
	w.forceEvaluation = world.EvaluateForces()
	return nil
}

func assertSpringForcesEqualOpposite(w *world, example map[string]string) error {
	massA, err := intValue(example, "mass_a")
	if err != nil {
		return err
	}
	massB, err := intValue(example, "mass_b")
	if err != nil {
		return err
	}
	a := w.forceEvaluation.ByMassID[massA].Force
	b := w.forceEvaluation.ByMassID[massB].Force
	if !vecClose(a, b.Scale(-1)) {
		return fmt.Errorf("forces are not equal and opposite: %#v %#v", a, b)
	}
	return nil
}

func assertSpringDampingDirection(w *world, _ map[string]string) error {
	for _, force := range w.forceEvaluation.ByMassID {
		if force.Force.Y != 0 || force.Force.X == 0 {
			return fmt.Errorf("damping force not isolated to spring direction: %#v", force.Force)
		}
	}
	return nil
}

func enableEnvironmentalForce(w *world, example map[string]string) error {
	world := ensureDomainWorld(w)
	world.Bounds = sim.Bounds{Width: 100, Height: 100}
	force, err := stringValue(example, "force")
	if err != nil {
		return err
	}
	if force == "viscosity" {
		world.Parameters.Set("viscosity", "1")
		return nil
	}
	if force == "wall repulsion" {
		world.Parameters.EnableWall("left")
	}
	world.Parameters.EnableForce(force, map[string]string{"magnitude": "10", "direction": "90", "exponent": "1", "damping": "1"})
	return nil
}

func createMovableMassAffectedByForce(w *world, example map[string]string) error {
	world := ensureDomainWorld(w)
	force, err := stringValue(example, "force")
	if err != nil {
		return err
	}
	if err := world.AddMass(sim.Mass{ID: 1, Position: forceMassPosition(force), Velocity: sim.Vec2{X: 2, Y: 0}, Mass: 1}); err != nil {
		return err
	}
	return addCenterOfMassPartner(world, force)
}

func forceMassPosition(force string) sim.Vec2 {
	if force == "center attraction" || force == "center of mass attraction" {
		return sim.Vec2{X: 0, Y: 0}
	}
	return sim.Vec2{X: -1, Y: 50}
}

func addCenterOfMassPartner(world *sim.Simulation, force string) error {
	if force == "center of mass attraction" {
		return world.AddMass(sim.Mass{ID: 2, Position: sim.Vec2{X: 50, Y: 50}, Mass: 1})
	}
	return nil
}

func assertMassReceivesForce(w *world, example map[string]string) error {
	force, err := stringValue(example, "force")
	if err != nil {
		return err
	}
	if w.forceEvaluation.ByMassID[1].Force == (sim.Vec2{}) {
		return fmt.Errorf("mass received no force from %s", force)
	}
	return nil
}

func createMassFixedState(w *world, example map[string]string) error {
	world := ensureDomainWorld(w)
	id, err := intValue(example, "mass_id")
	if err != nil {
		return err
	}
	fixed, err := boolValue(example, "fixed")
	if err != nil {
		return err
	}
	return world.AddMass(sim.Mass{ID: id, Position: sim.Vec2{X: 10, Y: 10}, Mass: 1, Fixed: fixed})
}

func affectMassByForce(w *world, example map[string]string) error {
	force, err := stringValue(example, "force")
	if err != nil {
		return err
	}
	if force != "gravity" {
		return fmt.Errorf("unsupported force %q", force)
	}
	ensureDomainWorld(w).Parameters.EnableForce("gravity", map[string]string{"magnitude": "10", "direction": "90"})
	return nil
}

func assertMassAcceleration(w *world, example map[string]string) error {
	id, err := intValue(example, "mass_id")
	if err != nil {
		return err
	}
	expected, err := stringValue(example, "acceleration")
	if err != nil {
		return err
	}
	acceleration := w.forceEvaluation.ByMassID[id].Acceleration
	if expected != "zero" {
		return fmt.Errorf("unsupported acceleration expectation %q", expected)
	}
	if acceleration != (sim.Vec2{}) {
		return fmt.Errorf("expected zero acceleration, got %#v", acceleration)
	}
	return nil
}

func enableWall(w *world, example map[string]string) error {
	wall, err := stringValue(example, "wall")
	if err != nil {
		return err
	}
	world := ensureDomainWorld(w)
	world.Bounds = sim.Bounds{Width: 100, Height: 100}
	world.Parameters.EnableWall(wall)
	world.Parameters.EnableForce("wall repulsion", map[string]string{"magnitude": "10", "exponent": "1"})
	return nil
}

func createMassOutsideWall(w *world, example map[string]string) error {
	world := ensureDomainWorld(w)
	id, err := intValue(example, "mass_id")
	if err != nil {
		return err
	}
	wall, err := stringValue(example, "wall")
	if err != nil {
		return err
	}
	return world.AddMass(sim.Mass{ID: id, Position: outsideWallPosition(wall), Mass: 1})
}

func assertWallForceTowardInside(w *world, example map[string]string) error {
	id, err := intValue(example, "mass_id")
	if err != nil {
		return err
	}
	wall, err := stringValue(example, "wall")
	if err != nil {
		return err
	}
	force := w.forceEvaluation.ByMassID[id].Force
	if simDot(force, insideDirection(wall)) <= 0 {
		return fmt.Errorf("force %#v is not toward inside for %s", force, wall)
	}
	return nil
}

func updateFirstSpring(w *world, update func(*sim.Spring) error) error {
	world, err := domainWorld(w)
	if err != nil {
		return err
	}
	if len(world.Springs) == 0 {
		return fmt.Errorf("no spring exists")
	}
	return update(&world.Springs[0])
}

func namedVelocity(value string) (sim.Vec2, error) {
	if value == "moving" {
		return sim.Vec2{X: 1, Y: 5}, nil
	}
	if value == "still" {
		return sim.Vec2{}, nil
	}
	return sim.Vec2{}, fmt.Errorf("unsupported velocity %q", value)
}

func outsideWallPosition(wall string) sim.Vec2 {
	switch wall {
	case "top":
		return sim.Vec2{X: 50, Y: 105}
	case "left":
		return sim.Vec2{X: -5, Y: 50}
	case "right":
		return sim.Vec2{X: 105, Y: 50}
	default:
		return sim.Vec2{X: 50, Y: -5}
	}
}

func insideDirection(wall string) sim.Vec2 {
	switch wall {
	case "top":
		return sim.Vec2{Y: -1}
	case "left":
		return sim.Vec2{X: 1}
	case "right":
		return sim.Vec2{X: -1}
	default:
		return sim.Vec2{Y: 1}
	}
}

func vecClose(a, b sim.Vec2) bool {
	return math.Abs(a.X-b.X) <= 0.000001 && math.Abs(a.Y-b.Y) <= 0.000001
}

func simDot(a, b sim.Vec2) float64 {
	return a.X*b.X + a.Y*b.Y
}
