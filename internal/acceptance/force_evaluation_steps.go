package acceptance

import (
	"fmt"
	"math"

	"springs/internal/sim"
)

func createSpringForceWorld(w *world, example map[string]string) error {
	world := ensureDomainWorld(w)
	massA, err := intValue(example, "mass_a")
	if err != nil {
		return err
	}
	massB, err := intValue(example, "mass_b")
	if err != nil {
		return err
	}
	masses := []sim.Mass{
		{ID: massA, Position: sim.Vec2{X: 0, Y: 0}, Mass: 1},
		{ID: massB, Position: sim.Vec2{X: 12, Y: 0}, Mass: 1},
	}
	return ensureSpringForceWorld(world, masses, sim.Spring{ID: 1, MassA: massA, MassB: massB, RestLength: 10, SpringConstant: 1})
}

func ensureSpringForceWorld(world *sim.Simulation, masses []sim.Mass, spring sim.Spring) error {
	for _, mass := range masses {
		if err := ensureMass(world, mass); err != nil {
			return err
		}
	}
	if len(world.Springs) > 0 {
		return nil
	}
	return world.AddSpring(spring)
}

func ensureMass(world *sim.Simulation, mass sim.Mass) error {
	if _, ok := world.MassByID(mass.ID); ok {
		return nil
	}
	return world.AddMass(mass)
}

func setOnlySpringRestLength(w *world, example map[string]string) error {
	return updateFirstSpringFloat(w, example, "rest_length", setSpringRestLength)
}

func setOnlySpringConstant(w *world, example map[string]string) error {
	return updateFirstSpringFloat(w, example, "spring_constant", setSpringConstant)
}

func setMassAVelocity(w *world, example map[string]string) error {
	return setMassNamedVelocity(w, example, "mass_a", "velocity_a")
}

func setMassBVelocity(w *world, example map[string]string) error {
	return setMassNamedVelocity(w, example, "mass_b", "velocity_b")
}

func setMassNamedVelocity(w *world, example map[string]string, massKey, velocityKey string) error {
	massID, err := intValue(example, massKey)
	if err != nil {
		return err
	}
	value, err := stringValue(example, velocityKey)
	if err != nil {
		return err
	}
	velocity, err := namedVelocity(value)
	if err != nil {
		return err
	}
	return updateMassByID(w, massID, func(mass *sim.Mass) error {
		mass.Velocity = velocity
		return nil
	})
}

func updateMassByID(w *world, massID int, update func(*sim.Mass) error) error {
	world, err := domainWorld(w)
	if err != nil {
		return err
	}
	return updateByID(world.Masses, massID, "mass", massIDValue, update)
}

func massIDValue(mass sim.Mass) int {
	return mass.ID
}

func setOnlySpringDamping(w *world, example map[string]string) error {
	return updateFirstSpringFloat(w, example, "damping_constant", setSpringDamping)
}

func updateFirstSpringFloat(w *world, example map[string]string, key string, assign func(*sim.Spring, float64)) error {
	return updateFirstSpring(w, func(spring *sim.Spring) error {
		value, err := floatValue(example, key)
		if err != nil {
			return err
		}
		assign(spring, value)
		return nil
	})
}

func setSpringConstant(spring *sim.Spring, value float64) {
	spring.SpringConstant = value
	spring.Stiffness = value
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
	if err := world.AddMass(affectedMass(force)); err != nil {
		return err
	}
	if needsReferenceMass(force) {
		return world.AddMass(sim.Mass{ID: 2, Position: sim.Vec2{X: 50, Y: 50}, Mass: 1})
	}
	return nil
}

func affectedMass(force string) sim.Mass {
	return sim.Mass{ID: 1, Position: affectedMassPosition(force), Velocity: sim.Vec2{X: 2, Y: 0}, Mass: 1}
}

func affectedMassPosition(force string) sim.Vec2 {
	positions := map[string]sim.Vec2{
		"center attraction":         {X: 0, Y: 0},
		"center of mass attraction": {X: 0, Y: 0},
	}
	if position, ok := positions[force]; ok {
		return position
	}
	return sim.Vec2{X: -1, Y: 50}
}

func needsReferenceMass(force string) bool {
	return force == "center of mass attraction"
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
	if !fixed {
		return fmt.Errorf("unsupported fixed state %t", fixed)
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
		return sim.Vec2{X: 50, Y: -5}
	case "left":
		return sim.Vec2{X: -5, Y: 50}
	case "right":
		return sim.Vec2{X: 105, Y: 50}
	default:
		return sim.Vec2{X: 50, Y: 105}
	}
}

func insideDirection(wall string) sim.Vec2 {
	switch wall {
	case "top":
		return sim.Vec2{Y: 1}
	case "left":
		return sim.Vec2{X: 1}
	case "right":
		return sim.Vec2{X: -1}
	default:
		return sim.Vec2{Y: -1}
	}
}

func vecClose(a, b sim.Vec2) bool {
	return math.Abs(a.X-b.X) <= 0.000001 && math.Abs(a.Y-b.Y) <= 0.000001
}

func simDot(a, b sim.Vec2) float64 {
	return a.X*b.X + a.Y*b.Y
}
