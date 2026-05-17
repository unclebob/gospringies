package acceptance

import (
	"fmt"
	"math"

	"springs/internal/sim"
)

const wallCollisionSpeed = 10.0

var wallReleaseForces = map[string]float64{"insufficient": 5, "sufficient": 20}
var wallCollisionPositions = map[string]struct {
	inside  sim.Vec2
	outside sim.Vec2
}{
	"left":   {inside: sim.Vec2{X: 1, Y: 50}, outside: sim.Vec2{X: -5, Y: 50}},
	"right":  {inside: sim.Vec2{X: 99, Y: 50}, outside: sim.Vec2{X: 105, Y: 50}},
	"top":    {inside: sim.Vec2{X: 50, Y: 1}, outside: sim.Vec2{X: 50, Y: -5}},
	"bottom": {inside: sim.Vec2{X: 50, Y: 99}, outside: sim.Vec2{X: 50, Y: 105}},
}
var wallBoundaryChecks = map[string]func(sim.Vec2) bool{
	"left":   func(position sim.Vec2) bool { return position.X > 0 },
	"right":  func(position sim.Vec2) bool { return position.X < 100 },
	"top":    func(position sim.Vec2) bool { return position.Y > 0 },
	"bottom": func(position sim.Vec2) bool { return position.Y < 100 },
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
	if normalVelocity(mass, wall)*normalSignTowardInside(wall) <= 0 {
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
	if normalVelocity(mass, wall)*normalSignTowardInside(wall) > 0 {
		return fmt.Errorf("mass bounced from disabled %s: %#v", wall, mass.Velocity)
	}
	return nil
}

func collisionWorld(w *world) *sim.Simulation {
	world := ensureDomainWorld(w)
	world.Bounds = sim.Bounds{Width: 100, Height: 100}
	world.Damping = 1
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
	if !validCollisionWall(wall) {
		return 0, "", fmt.Errorf("unsupported wall %q", wall)
	}
	return id, wall, nil
}

func validCollisionWall(wall string) bool {
	return wall == "left" || wall == "right" || wall == "top" || wall == "bottom"
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
	return wallCollisionPositions[wall].inside
}

func outsideCollisionPosition(wall string) sim.Vec2 {
	return wallCollisionPositions[wall].outside
}

func outwardVelocity(wall string) sim.Vec2 {
	switch wall {
	case "left":
		return sim.Vec2{X: -wallCollisionSpeed}
	case "right":
		return sim.Vec2{X: wallCollisionSpeed}
	case "top":
		return sim.Vec2{Y: -wallCollisionSpeed}
	default:
		return sim.Vec2{Y: wallCollisionSpeed}
	}
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

func normalSignTowardInside(wall string) float64 {
	if wall == "left" || wall == "top" {
		return 1
	}
	return -1
}

func insideWallBoundary(position sim.Vec2, wall string) bool {
	return wallBoundaryChecks[wall](position)
}
