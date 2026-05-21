package acceptance

import (
	"fmt"
	"math"

	"springs/internal/sim"
)

const wallCollisionSpeed = 10.0

var wallReleaseForces = map[string]float64{"insufficient": 5, "sufficient": 20}

type collisionWallSpec struct {
	inside        sim.Vec2
	outside       sim.Vec2
	outward       sim.Vec2
	passedThrough func(sim.Vec2) bool
}

var wallCollisionSpecs = map[string]collisionWallSpec{
	"left": {
		inside:        sim.Vec2{X: 1, Y: 50},
		outside:       sim.Vec2{X: -5, Y: 50},
		outward:       sim.Vec2{X: -wallCollisionSpeed},
		passedThrough: func(position sim.Vec2) bool { return position.X > 0 },
	},
	"right": {
		inside:        sim.Vec2{X: 99, Y: 50},
		outside:       sim.Vec2{X: 105, Y: 50},
		outward:       sim.Vec2{X: wallCollisionSpeed},
		passedThrough: func(position sim.Vec2) bool { return position.X < 100 },
	},
	"top": {
		inside:        sim.Vec2{X: 50, Y: 99},
		outside:       sim.Vec2{X: 50, Y: 105},
		outward:       sim.Vec2{Y: wallCollisionSpeed},
		passedThrough: func(position sim.Vec2) bool { return position.Y < 100 },
	},
	"bottom": {
		inside:        sim.Vec2{X: 50, Y: 1},
		outside:       sim.Vec2{X: 50, Y: -5},
		outward:       sim.Vec2{Y: -wallCollisionSpeed},
		passedThrough: func(position sim.Vec2) bool { return position.Y > 0 },
	},
}

func setCollisionMassElasticity(w *world, example map[string]string) error {
	id, elasticity, err := intAndFloat(example, "mass_id", "elasticity")
	if err != nil {
		return err
	}
	ensureCollisionMass(w, id).Elasticity = elasticity
	return nil
}

func moveMassFromInsideTowardWall(w *world, example map[string]string) error {
	return setCollisionMassMotion(w, example, insideCollisionPosition, outwardVelocity)
}

func advanceThroughWallCollision(w *world, _ map[string]string) error {
	return stepCollisionWorld(w)
}

func assertWallNormalVelocityReversed(w *world, example map[string]string) error {
	mass, wall, err := collisionMassByExample(w, example)
	if err != nil {
		return err
	}
	if normalVelocityTowardInside(mass, wall) <= 0 {
		return fmt.Errorf("velocity was not reversed for %s: %#v", wall, mass.Velocity)
	}
	return nil
}

func assertWallNormalVelocityScaled(w *world, example map[string]string) error {
	mass, wall, err := collisionMassByExample(w, example)
	if err != nil {
		return err
	}
	elasticity, err := floatValue(example, "elasticity")
	if err != nil {
		return err
	}
	return assertFloat("normal velocity magnitude", math.Abs(normalVelocity(mass, wall)), wallCollisionSpeed*elasticity)
}

func moveMassFromOutsideThroughWall(w *world, example map[string]string) error {
	return setCollisionMassMotion(w, example, outsideCollisionPosition, inwardVelocity)
}

func setCollisionMassMotion(
	w *world,
	example map[string]string,
	position func(string) sim.Vec2,
	velocity func(string) sim.Vec2,
) error {
	id, wall, err := collisionMassAndWall(example)
	if err != nil {
		return err
	}
	mass := ensureCollisionMass(w, id)
	mass.Position = position(wall)
	mass.Velocity = velocity(wall)
	return nil
}

func advanceThroughWallBoundary(w *world, _ map[string]string) error {
	return advanceThroughWallCollision(w, nil)
}

func startMassAtPositionWithVelocity(w *world, example map[string]string) error {
	id, wall, err := collisionMassAndWall(example)
	if err != nil {
		return err
	}
	if err := requireSweptWallExample(wall, example); err != nil {
		return err
	}
	values, err := floatValues(example, "start_x", "start_y", "velocity_x", "velocity_y")
	if err != nil {
		return err
	}
	world := collisionWorld(w)
	world.Bounds = sim.Bounds{Width: 800, Height: 600}
	mass := ensureCollisionMass(w, id)
	world.Bounds = sim.Bounds{Width: 800, Height: 600}
	mass.Position = sim.Vec2{X: values[0], Y: values[1]}
	mass.Velocity = sim.Vec2{X: values[2], Y: values[3]}
	rememberScreenWallStartingSide(w, id, wall, mass.Position)
	return nil
}

func requireSweptWallExample(wall string, example map[string]string) error {
	expected := map[string]map[string]string{
		"right": {"mass_id": "1", "start_x": "790", "start_y": "400", "velocity_x": "300", "velocity_y": "0", "duration": "1 step"},
		"top":   {"mass_id": "2", "start_x": "400", "start_y": "590", "velocity_x": "0", "velocity_y": "300", "duration": "1 step"},
	}
	for key, value := range expected[wall] {
		if example[key] != value {
			return fmt.Errorf("%s = %q, expected %q", key, example[key], value)
		}
	}
	return nil
}

func rememberScreenWallStartingSide(w *world, massID int, wall string, position sim.Vec2) {
	if w.wallSpringSides == nil {
		w.wallSpringSides = map[int]float64{}
	}
	w.wallSpringSides[massID] = screenWallSide(wall, position)
}

func screenWallSide(wall string, position sim.Vec2) float64 {
	switch wall {
	case "right":
		return 800 - position.X
	case "top":
		return 600 - position.Y
	case "left":
		return position.X
	case "bottom":
		return position.Y
	default:
		return 0
	}
}

func screenWallSideVelocity(wall string, velocity sim.Vec2) float64 {
	switch wall {
	case "right":
		return -velocity.X
	case "top":
		return -velocity.Y
	case "left":
		return velocity.X
	case "bottom":
		return velocity.Y
	default:
		return 0
	}
}

func advanceThroughWallBoundaryByDuration(w *world, example map[string]string) error {
	duration, err := stringValue(example, "duration")
	if err != nil {
		return err
	}
	if duration != "1 step" {
		return fmt.Errorf("unsupported wall boundary duration %q", duration)
	}
	world := collisionWorld(w)
	world.Bounds = sim.Bounds{Width: 800, Height: 600}
	world.Step(1)
	return nil
}

func assertMassOnStartingScreenWallSide(w *world, example map[string]string) error {
	mass, wall, err := collisionMassByExample(w, example)
	if err != nil {
		return err
	}
	startingSide, ok := w.wallSpringSides[mass.ID]
	if !ok {
		return fmt.Errorf("starting side for mass %d was not recorded", mass.ID)
	}
	if screenWallSide(wall, mass.Position)*startingSide < 0 {
		return fmt.Errorf("mass %d crossed wall %s: %#v", mass.ID, wall, mass.Position)
	}
	return nil
}

func assertWallNormalVelocityTowardStartingSide(w *world, example map[string]string) error {
	mass, wall, err := collisionMassByExample(w, example)
	if err != nil {
		return err
	}
	startingSide, ok := w.wallSpringSides[mass.ID]
	if !ok {
		return fmt.Errorf("starting side for mass %d was not recorded", mass.ID)
	}
	if screenWallSideVelocity(wall, mass.Velocity)*startingSide < 0 {
		return fmt.Errorf("mass %d velocity not resolved toward starting side of %s: %#v", mass.ID, wall, mass.Velocity)
	}
	return nil
}

func assertMassPassedThroughWall(w *world, example map[string]string) error {
	mass, wall, err := collisionMassByExample(w, example)
	if err != nil {
		return err
	}
	if !insideWallBoundary(mass.Position, wall) {
		return fmt.Errorf("mass did not pass through %s: %#v", wall, mass.Position)
	}
	return nil
}

func setCollisionStickiness(w *world, example map[string]string) error {
	stickiness, err := stringValue(example, "stickiness")
	if err != nil {
		return err
	}
	if stickiness != "high" {
		return fmt.Errorf("unsupported stickiness %q", stickiness)
	}
	collisionWorld(w).Parameters.Set("stickiness", "10")
	return nil
}

func collideMassWithWall(w *world, example map[string]string) error {
	if err := enableWall(w, example); err != nil {
		return err
	}
	if err := setCollisionMassElasticity(w, map[string]string{"mass_id": example["mass_id"], "elasticity": "1.0"}); err != nil {
		return err
	}
	return moveMassFromInsideTowardWall(w, example)
}

func removeWallNormalVelocity(w *world, _ map[string]string) error {
	return stepCollisionWorld(w)
}

func stepCollisionWorld(w *world) error {
	collisionWorld(w).Step(1)
	return nil
}

func assertMassStuckToWall(w *world, example map[string]string) error {
	mass, wall, err := collisionMassByExample(w, example)
	if err != nil {
		return err
	}
	if mass.StuckWall != wall {
		return fmt.Errorf("mass stuck wall = %q, want %q", mass.StuckWall, wall)
	}
	return nil
}

func pullMassAwayFromWall(w *world, example map[string]string) error {
	forceName, err := stringValue(example, "release_force")
	if err != nil {
		return err
	}
	force, ok := wallReleaseForces[forceName]
	if !ok {
		return fmt.Errorf("unsupported release force %q", forceName)
	}
	collisionWorld(w).Parameters.EnableForce("center attraction", map[string]string{"magnitude": fmt.Sprintf("%f", force), "exponent": "0"})
	collisionWorld(w).Step(1)
	return nil
}

func assertMassReleaseResult(w *world, example map[string]string) error {
	released, err := expectedReleased(example)
	if err != nil {
		return err
	}
	mass, _, err := collisionMassByExample(w, example)
	if err != nil {
		return err
	}
	if (mass.StuckWall == "") != released {
		return fmt.Errorf("mass release result = %#v, want %s", mass, releaseResultName(released))
	}
	return nil
}

func expectedReleased(example map[string]string) (bool, error) {
	result, err := stringValue(example, "release_result")
	if err != nil {
		return false, err
	}
	switch result {
	case "released":
		return true, nil
	case "stuck":
		return false, nil
	default:
		return false, fmt.Errorf("unsupported release result %q", result)
	}
}

func releaseResultName(released bool) string {
	if released {
		return "released"
	}
	return "stuck"
}

func disableWall(w *world, example map[string]string) error {
	collisionWorld(w)
	_, err := stringValue(example, "wall")
	return err
}

func moveMassTowardWall(w *world, example map[string]string) error {
	return moveMassFromInsideTowardWall(w, example)
}

func assertMassDidNotBounce(w *world, example map[string]string) error {
	mass, wall, err := collisionMassByExample(w, example)
	if err != nil {
		return err
	}
	if normalVelocityTowardInside(mass, wall) > 0 {
		return fmt.Errorf("mass bounced from disabled %s: %#v", wall, mass.Velocity)
	}
	return nil
}

func collisionWorld(w *world) *sim.Simulation {
	world := ensureDomainWorld(w)
	world.Bounds = sim.Bounds{Width: 100, Height: 100}
	force := world.Parameters.Forces["wall repulsion"]
	force.Enabled = "false"
	world.Parameters.Forces["wall repulsion"] = force
	return world
}

func ensureCollisionMass(w *world, id int) *sim.Mass {
	world := collisionWorld(w)
	for i := range world.Masses {
		if world.Masses[i].ID == id {
			return &world.Masses[i]
		}
	}
	world.Masses = append(world.Masses, sim.Mass{ID: id, Position: sim.Vec2{X: 50, Y: 50}, Mass: 1, Elasticity: 1})
	return &world.Masses[len(world.Masses)-1]
}

func collisionMassAndWall(example map[string]string) (int, string, error) {
	id, err := intValue(example, "mass_id")
	if err != nil {
		return 0, "", err
	}
	wall, err := stringValue(example, "wall")
	if err != nil {
		return 0, "", err
	}
	if _, ok := wallCollisionSpecs[wall]; !ok {
		return 0, "", fmt.Errorf("unsupported wall %q", wall)
	}
	return id, wall, nil
}

func collisionMassByExample(w *world, example map[string]string) (sim.Mass, string, error) {
	id, wall, err := collisionMassAndWall(example)
	if err != nil {
		return sim.Mass{}, "", err
	}
	mass, ok := collisionWorld(w).MassByID(id)
	if !ok {
		return sim.Mass{}, "", fmt.Errorf("mass %d not found", id)
	}
	return mass, wall, nil
}

func insideCollisionPosition(wall string) sim.Vec2 {
	return wallCollisionSpecs[wall].inside
}

func outsideCollisionPosition(wall string) sim.Vec2 {
	return wallCollisionSpecs[wall].outside
}

func outwardVelocity(wall string) sim.Vec2 {
	return wallCollisionSpecs[wall].outward
}

func inwardVelocity(wall string) sim.Vec2 {
	return outwardVelocity(wall).Scale(-1)
}

func normalVelocity(mass sim.Mass, wall string) float64 {
	if wall == "left" || wall == "right" {
		return mass.Velocity.X
	}
	return mass.Velocity.Y
}

func normalVelocityTowardInside(mass sim.Mass, wall string) float64 {
	velocity := normalVelocity(mass, wall)
	switch wall {
	case "left", "bottom":
		return velocity
	default:
		return -velocity
	}
}

func normalSignTowardInside(wall string) float64 {
	if wall == "left" || wall == "bottom" {
		return 1
	}
	return -1
}

func insideWallBoundary(position sim.Vec2, wall string) bool {
	return wallCollisionSpecs[wall].passedThrough(position)
}
