//go:build property

package sim

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
	"testing/quick"
)

func TestPropertyMassCrossingWallSpringStaysOnStartingSide(t *testing.T) {
	checkProperty(t, 1, 500, massCrossingWallSpringStaysOnStartingSide)
}

func TestPropertyVec2AlgebraIdentities(t *testing.T) {
	checkProperty(t, 2, 500, vec2AlgebraIdentities)
}

func TestPropertyBoundsCenterMatchesConfiguredExtents(t *testing.T) {
	checkProperty(t, 3, 500, boundsCenterMatchesConfiguredExtents)
}

func TestPropertyCloneAndLoadFromDoNotAliasSlicesOrMaps(t *testing.T) {
	checkProperty(t, 4, 300, cloneAndLoadFromDoNotAliasSlicesOrMaps)
}

func TestPropertyAddMassAndSpringLookupConsistency(t *testing.T) {
	checkProperty(t, 5, 300, addMassAndSpringLookupConsistency)
}

func TestPropertySpringForcesAreEqualAndOpposite(t *testing.T) {
	checkProperty(t, 6, 500, springForcesAreEqualAndOpposite)
}

func TestPropertyFixedMassesHaveZeroAcceleration(t *testing.T) {
	checkProperty(t, 7, 500, fixedMassesHaveZeroAcceleration)
}

func TestPropertyScreenWallCollisionReturnsMassInside(t *testing.T) {
	checkProperty(t, 8, 400, screenWallCollisionReturnsMassInside)
}

func TestPropertyWallSpringLengthConstraintRestoresRestLength(t *testing.T) {
	checkProperty(t, 9, 500, wallSpringLengthConstraintRestoresRestLength)
}

func TestPropertyClosestPointAndFractionStayOnSegment(t *testing.T) {
	checkProperty(t, 10, 500, closestPointAndFractionStayOnSegment)
}

func TestPropertyOffCanvasCleanupKeepsValidSpringEndpoints(t *testing.T) {
	checkProperty(t, 11, 300, offCanvasCleanupKeepsValidSpringEndpoints)
}

func TestPropertyParametersRoundTripWithoutAliasing(t *testing.T) {
	checkProperty(t, 12, 300, parametersRoundTripWithoutAliasing)
}

func TestPropertyAdvanceDurationUsesPositiveBoundedSteps(t *testing.T) {
	checkProperty(t, 13, 300, advanceDurationUsesPositiveBoundedSteps)
}

func TestPropertyStepWithNoForcesMovesLinearlyAndKeepsFixedMasses(t *testing.T) {
	checkProperty(t, 14, 300, stepWithNoForcesMovesLinearlyAndKeepsFixedMasses)
}

func TestPropertyGravityForceScalesWithMass(t *testing.T) {
	checkProperty(t, 15, 300, gravityForceScalesWithMass)
}

func TestPropertyViscosityForceOpposesVelocity(t *testing.T) {
	checkProperty(t, 16, 300, viscosityForceOpposesVelocity)
}

func TestPropertyCenterOfMassTranslatesWithMasses(t *testing.T) {
	checkProperty(t, 17, 300, centerOfMassTranslatesWithMasses)
}

func TestPropertyForceCenterTracksSingleSelectedMass(t *testing.T) {
	checkProperty(t, 18, 300, forceCenterTracksSingleSelectedMass)
}

func TestPropertyMassRadiusIsBoundedAndMonotonic(t *testing.T) {
	checkProperty(t, 19, 300, massRadiusIsBoundedAndMonotonic)
}

func TestPropertyWallSpringContactFractionFindsSegmentCrossing(t *testing.T) {
	checkProperty(t, 20, 300, wallSpringContactFractionFindsSegmentCrossing)
}

func TestPropertyResolveWallSpringVelocitySeparatesOrKeepsSeparating(t *testing.T) {
	checkProperty(t, 21, 300, resolveWallSpringVelocitySeparatesOrKeepsSeparating)
}

func TestPropertyResolveWallSpringVelocityUsesPositiveElasticity(t *testing.T) {
	checkProperty(t, 22, 300, resolveWallSpringVelocityUsesPositiveElasticity)
}

func TestPropertyResetClearsWorldAndInsertFromAppendsObjects(t *testing.T) {
	checkProperty(t, 23, 300, resetClearsWorldAndInsertFromAppendsObjects)
}

func TestPropertyAddMassAtAndAddSpringBetweenGenerateConsistentIDs(t *testing.T) {
	checkProperty(t, 24, 300, addMassAtAndAddSpringBetweenGenerateConsistentIDs)
}

func TestPropertyAdaptiveStepHelpersStayPositiveAndBounded(t *testing.T) {
	checkProperty(t, 24, 300, adaptiveStepHelpersStayPositiveAndBounded)
}

func TestPropertyCenterAndWallForcesPointTowardTheirTargets(t *testing.T) {
	checkProperty(t, 25, 300, centerAndWallForcesPointTowardTheirTargets)
}

func TestPropertyEnabledForceMatchesParameterState(t *testing.T) {
	checkProperty(t, 26, 300, enabledForceMatchesParameterState)
}

func TestPropertyStuckMassStaysOnWallUntilReleaseForceWins(t *testing.T) {
	checkProperty(t, 27, 300, stuckMassStaysOnWallUntilReleaseForceWins)
}

func TestPropertyStepDurationFollowsConfiguredTimestep(t *testing.T) {
	checkProperty(t, 28, 300, stepDurationFollowsConfiguredTimestep)
}

func TestPropertyMassCollisionConservesMovableMomentum(t *testing.T) {
	checkProperty(t, 29, 300, massCollisionConservesMovableMomentum)
}

func TestPropertyFiniteStepOutputsRemainFinite(t *testing.T) {
	checkProperty(t, 30, 300, finiteStepOutputsRemainFinite)
}

func TestPropertyForceEvaluationSkipsInvalidSpringsAndScalesAcceleration(t *testing.T) {
	checkProperty(t, 31, 300, forceEvaluationSkipsInvalidSpringsAndScalesAcceleration)
}

func TestPropertyWallSpringLengthConstraintCollisionKeepsEndpointOnBarrierSide(t *testing.T) {
	checkProperty(t, 32, 300, wallSpringLengthConstraintCollisionKeepsEndpointOnBarrierSide)
}

func TestPropertyMovingWallSpringFixedEndpointCollisionSeparatesContact(t *testing.T) {
	checkProperty(t, 33, 300, movingWallSpringFixedEndpointCollisionSeparatesContact)
}

func TestPropertyWallSpringCollisionConservesMomentumWithVaryingMasses(t *testing.T) {
	checkProperty(t, 34, 300, wallSpringCollisionConservesMomentumWithVaryingMasses)
}

func checkProperty(t *testing.T, seed int64, maxCount int, property any) {
	t.Helper()
	config := &quick.Config{
		MaxCount: maxCount,
		Rand:     rand.New(rand.NewSource(seed)),
	}
	if err := quick.Check(property, config); err != nil {
		t.Fatal(err)
	}
}

func massCrossingWallSpringStaysOnStartingSide(startYInput float64, distanceInput float64, speedInput float64, dtInput float64) bool {
	startY := propertyFloat(startYInput, 1, 99)
	distance := propertyFloat(distanceInput, 0.1, 100)
	dt := propertyFloat(dtInput, 0.001, 1)
	extraDistance := propertyFloat(speedInput, 0.1, 100)
	speed := (distance + extraDistance) / dt
	world := NewWorld()
	_ = world.AddMass(Mass{ID: 1, Position: Vec2{}, Mass: 1})
	_ = world.AddMass(Mass{ID: 2, Position: Vec2{Y: 100}, Mass: 1})
	_ = world.AddMass(Mass{ID: 3, Position: Vec2{X: -distance, Y: startY}, Velocity: Vec2{X: speed}, Mass: 1})
	_ = world.AddSpring(Spring{ID: 1, MassA: 1, MassB: 2, Wall: true})

	before, err := propertyWallSpringSide(world, 1, world.Masses[2].Position)
	if err != nil {
		panic(err)
	}
	world.Step(dt)
	mass, _ := world.MassByID(3)
	after, err := propertyWallSpringSide(world, 1, mass.Position)
	if err != nil {
		panic(err)
	}
	if before*after < -1e-9 {
		panic(fmt.Sprintf("mass crossed wall spring: y=%f distance=%f speed=%f dt=%f before=%f after=%f mass=%#v", startY, distance, speed, dt, before, after, mass))
	}
	return true
}

func vec2AlgebraIdentities(axInput, ayInput, bxInput, byInput, scaleInput float64) bool {
	a := propertyVec(axInput, ayInput, 10)
	b := propertyVec(bxInput, byInput, 10)
	scale := propertySignedFloat(scaleInput, 100)
	sum := a.Add(b)
	assertVecClose("add commutative", sum, b.Add(a), 1e-9)
	assertVecClose("sub inverse", sum.Sub(b), a, 1e-9)
	assertVecClose("scale distributes", sum.Scale(scale), a.Scale(scale).Add(b.Scale(scale)), 1e-6)
	if !propertyClose(dot(a, b), dot(b, a), 1e-9) {
		panic(fmt.Sprintf("dot is not symmetric: a=%#v b=%#v", a, b))
	}
	normalized := a.Normalize()
	normalizedLength := length(normalized)
	if a == (Vec2{}) {
		assertVecClose("zero normalize", normalized, Vec2{}, 0)
	} else if !propertyClose(normalizedLength, 1, 1e-6) {
		panic(fmt.Sprintf("normalized vector length = %f for %#v", normalizedLength, a))
	}
	return true
}

func boundsCenterMatchesConfiguredExtents(widthInput, heightInput, leftInput, rightInput, bottomInput, topInput float64) bool {
	width := propertyFloat(widthInput, 1, 2000)
	height := propertyFloat(heightInput, 1, 2000)
	left := propertySignedFloat(leftInput, 1000)
	right := propertySignedFloat(rightInput, 1000)
	bottom := propertySignedFloat(bottomInput, 1000)
	top := propertySignedFloat(topInput, 1000)
	bounds := Bounds{Width: width, Height: height, Left: left, Right: right, Bottom: bottom, Top: top}
	expectedRight := right
	if expectedRight == 0 {
		expectedRight = width
	}
	expectedTop := top
	if expectedTop == 0 {
		expectedTop = height
	}
	assertClose("min x", bounds.MinX(), left, 0)
	assertClose("max x", bounds.MaxX(), expectedRight, 0)
	assertClose("min y", bounds.MinY(), bottom, 0)
	assertClose("max y", bounds.MaxY(), expectedTop, 0)
	assertVecClose("center", bounds.Center(), Vec2{X: (left + expectedRight) / 2, Y: (bottom + expectedTop) / 2}, 0)
	return true
}

func cloneAndLoadFromDoNotAliasSlicesOrMaps(xInput, yInput, massInput float64) bool {
	world := propertySampleWorld(xInput, yInput, massInput)
	clone := world.Clone()
	loaded := NewWorld()
	loaded.LoadFrom(world)

	world.Masses[0].Position.X += 100
	world.Springs[0].RestLength += 100
	world.Parameters.Set("timestep", "42")
	world.Parameters.Forces["gravity"] = ForceConfig{Enabled: "true", Values: map[string]string{"magnitude": "99", "direction": "0"}}
	world.Parameters.Walls["left"] = true

	assertIndependentWorldCopy("clone", clone)
	assertIndependentWorldCopy("loaded", loaded)
	return true
}

func addMassAndSpringLookupConsistency(axInput, ayInput, bxInput, byInput float64) bool {
	world := NewWorld()
	a := Mass{ID: 11, Position: propertyVec(axInput, ayInput, 100), Mass: 1}
	b := Mass{ID: 17, Position: propertyVec(bxInput, byInput, 100), Mass: 2}
	if err := world.AddMass(a); err != nil {
		panic(err)
	}
	if err := world.AddMass(b); err != nil {
		panic(err)
	}
	if err := world.AddMass(Mass{ID: a.ID}); err == nil {
		panic("duplicate mass ID accepted")
	}
	spring := Spring{ID: 23, MassA: a.ID, MassB: b.ID, RestLength: 12, Stiffness: 3}
	if err := world.AddSpring(spring); err != nil {
		panic(err)
	}
	if err := world.AddSpring(spring); err == nil {
		panic("duplicate spring ID accepted")
	}
	if err := world.AddSpring(Spring{ID: 24, MassA: a.ID, MassB: 999}); err == nil {
		panic("missing spring endpoint accepted")
	}
	foundA, okA := world.MassByID(a.ID)
	foundB, okB := world.MassByID(b.ID)
	foundSpring, okSpring := world.SpringByID(spring.ID)
	if !okA || foundA.ID != a.ID || !okB || foundB.ID != b.ID || !okSpring || foundSpring.MassA != a.ID || foundSpring.MassB != b.ID {
		panic(fmt.Sprintf("lookup mismatch: a=%#v/%v b=%#v/%v spring=%#v/%v", foundA, okA, foundB, okB, foundSpring, okSpring))
	}
	if !world.validSpringMassIndexes(foundSpring) {
		panic(fmt.Sprintf("spring indexes not valid: %#v", foundSpring))
	}
	return true
}

func springForcesAreEqualAndOpposite(axInput, ayInput, bxInput, byInput, restInput, stiffnessInput, dampingInput float64) bool {
	aPosition := propertyVec(axInput, ayInput, 100)
	bPosition := propertyVec(bxInput, byInput, 100)
	if length(bPosition.Sub(aPosition)) < 0.001 {
		bPosition = bPosition.Add(Vec2{X: 1})
	}
	world := NewWorld()
	_ = world.AddMass(Mass{ID: 1, Position: aPosition, Velocity: Vec2{X: 2, Y: -1}, Mass: 1})
	_ = world.AddMass(Mass{ID: 2, Position: bPosition, Velocity: Vec2{X: -1, Y: 3}, Mass: 2})
	_ = world.AddSpring(Spring{
		ID:             1,
		MassA:          1,
		MassB:          2,
		RestLength:     propertyFloat(restInput, 0.1, 100),
		SpringConstant: propertyFloat(stiffnessInput, 0.1, 50),
		Damping:        propertyFloat(dampingInput, 0, 10),
	})

	evaluation := world.EvaluateForces()
	total := evaluation.ByMassID[1].Force.Add(evaluation.ByMassID[2].Force)
	assertVecClose("spring forces", total, Vec2{}, 1e-9)
	return true
}

func fixedMassesHaveZeroAcceleration(xInput, yInput, vxInput, vyInput, massInput float64) bool {
	mass := propertyFloat(massInput, 0.1, 100)
	world := NewWorld()
	world.Parameters.EnableForce("gravity", map[string]string{"magnitude": "50", "direction": "180"})
	world.Parameters.Set("viscosity", "3")
	_ = world.AddMass(Mass{
		ID:       1,
		Position: propertyVec(xInput, yInput, 100),
		Velocity: propertyVec(vxInput, vyInput, 100),
		Mass:     mass,
		Fixed:    true,
	})
	evaluation := world.EvaluateForces()
	assertVecClose("fixed acceleration", evaluation.ByMassID[1].Acceleration, Vec2{}, 0)
	return true
}

func screenWallCollisionReturnsMassInside(axisInput, sideInput, distanceInput, velocityInput, elasticityInput float64) bool {
	world := NewWorld()
	world.Bounds = Bounds{Width: 100, Height: 80}
	for _, wall := range []string{"left", "right", "bottom", "top"} {
		world.Parameters.EnableWall(wall)
	}
	distance := propertyFloat(distanceInput, 0.1, 50)
	speed := propertyFloat(velocityInput, 0.1, 200)
	elasticity := propertyFloat(elasticityInput, 0.1, 2)
	mass := Mass{ID: 1, Position: Vec2{X: 50, Y: 40}, Velocity: Vec2{}, Mass: 1, Elasticity: elasticity}
	wall := int(propertyFloat(axisInput+sideInput, 0, 4))
	switch wall {
	case 0:
		mass.Position.X = world.Bounds.MinX() - distance
		mass.Velocity.X = -speed
	case 1:
		mass.Position.X = world.Bounds.MaxX() + distance
		mass.Velocity.X = speed
	case 2:
		mass.Position.Y = world.Bounds.MinY() - distance
		mass.Velocity.Y = -speed
	default:
		mass.Position.Y = world.Bounds.MaxY() + distance
		mass.Velocity.Y = speed
	}
	world.applyWallCollision(&mass)
	if mass.Position.X < world.Bounds.MinX() || mass.Position.X > world.Bounds.MaxX() ||
		mass.Position.Y < world.Bounds.MinY() || mass.Position.Y > world.Bounds.MaxY() {
		panic(fmt.Sprintf("mass outside after wall collision: %#v", mass))
	}
	for _, collisionWall := range world.collisionWalls(&mass) {
		if collisionWall.outside(*collisionWall.position) && collisionWall.movingOutward(*collisionWall.velocity) {
			panic(fmt.Sprintf("mass still moving outward after wall collision: %#v wall=%s", mass, collisionWall.name))
		}
	}
	return true
}

func wallSpringLengthConstraintRestoresRestLength(axInput, ayInput, bxInput, byInput, restInput float64) bool {
	endpointA := Mass{ID: 1, Position: propertyVec(axInput, ayInput, 100), Mass: 1}
	endpointB := Mass{ID: 2, Position: propertyVec(bxInput, byInput, 100), Mass: 1}
	if length(endpointB.Position.Sub(endpointA.Position)) < 0.001 {
		endpointB.Position = endpointA.Position.Add(Vec2{X: 10})
	}
	restLength := propertyFloat(restInput, 0.1, 50)
	spring := Spring{ID: 1, Wall: true, RestLength: restLength}
	beforeError := math.Abs(length(endpointB.Position.Sub(endpointA.Position)) - restLength)
	world := NewWorld()
	world.applyWallSpringLengthConstraint(&spring, &endpointA, &endpointB)
	afterError := math.Abs(length(endpointB.Position.Sub(endpointA.Position)) - restLength)
	if afterError > beforeError && afterError > 1e-9 {
		panic(fmt.Sprintf("wall spring length constraint increased error: before=%f after=%f rest=%f", beforeError, afterError, restLength))
	}
	return true
}

func closestPointAndFractionStayOnSegment(pxInput, pyInput, axInput, ayInput, bxInput, byInput float64) bool {
	point := propertyVec(pxInput, pyInput, 1000)
	start := propertyVec(axInput, ayInput, 1000)
	end := propertyVec(bxInput, byInput, 1000)
	segment := end.Sub(start)
	lengthSquared := dot(segment, segment)
	if lengthSquared == 0 {
		segment = Vec2{X: 1, Y: 0}
		lengthSquared = 1
	}
	fraction := closestFractionOnSegment(point, start, segment, lengthSquared)
	if fraction < 0 || fraction > 1 {
		panic(fmt.Sprintf("fraction outside segment: %f", fraction))
	}
	closest := closestPointOnSegment(point, start, segment, lengthSquared)
	assertVecClose("closest point matches fraction", closest, start.Add(segment.Scale(fraction)), 1e-9)
	reversedFraction := closestFractionOnSegment(point, start.Add(segment), segment.Scale(-1), lengthSquared)
	reversedClosest := closestPointOnSegment(point, start.Add(segment), segment.Scale(-1), lengthSquared)
	assertClose("reversed fraction", fraction, 1-reversedFraction, 1e-9)
	assertVecClose("reversed closest point", closest, reversedClosest, 1e-9)
	return true
}

func offCanvasCleanupKeepsValidSpringEndpoints(xInput, yInput float64) bool {
	world := NewWorld()
	world.Bounds = Bounds{Width: 100, Height: 80}
	inside := Mass{ID: 10, Position: Vec2{X: propertyFloat(xInput, 1, 99), Y: propertyFloat(yInput, 1, 79)}, Mass: 1}
	alsoInside := Mass{ID: 20, Position: Vec2{X: 50, Y: 40}, Mass: 1}
	outside := Mass{ID: 30, Position: Vec2{X: world.Bounds.MaxX() + world.Bounds.Height + 1, Y: 40}, Mass: 1}
	_ = world.AddMass(inside)
	_ = world.AddMass(alsoInside)
	_ = world.AddMass(outside)
	_ = world.AddSpring(Spring{ID: 1, A: 0, B: 1})
	_ = world.AddSpring(Spring{ID: 2, A: 1, B: 2})
	world.cleanupOffCanvasObjects()
	if _, ok := world.MassByID(outside.ID); ok {
		panic("off-canvas mass survived cleanup")
	}
	if _, ok := world.MassByID(inside.ID); !ok {
		panic("inside mass removed by cleanup")
	}
	for _, spring := range world.Springs {
		if !world.validMassIndex(spring.A) || !world.validMassIndex(spring.B) {
			panic(fmt.Sprintf("spring has invalid indexes after cleanup: %#v masses=%#v", spring, world.Masses))
		}
		if _, ok := world.MassByID(spring.MassA); !ok {
			panic(fmt.Sprintf("spring has missing MassA after cleanup: %#v", spring))
		}
		if _, ok := world.MassByID(spring.MassB); !ok {
			panic(fmt.Sprintf("spring has missing MassB after cleanup: %#v", spring))
		}
	}
	return true
}

func parametersRoundTripWithoutAliasing(valueInput float64, forceInput float64) bool {
	parameters := DefaultParameters()
	value := fmt.Sprintf("%.6f", propertyFloat(valueInput, 0.001, 10))
	forceValue := fmt.Sprintf("%.6f", propertyFloat(forceInput, 0.001, 100))
	parameters.Set("timestep", value)
	parameters.EnableForce("gravity", map[string]string{"magnitude": forceValue, "direction": "180"})
	parameters.SelectForce("gravity")
	parameters.EnableWall("left")

	clone := parameters.Clone()
	parameters.Set("timestep", "999")
	parameters.Forces["gravity"] = ForceConfig{Enabled: "false", Values: map[string]string{"magnitude": "999", "direction": "0"}}
	parameters.Walls["left"] = false

	if !clone.Has("timestep") || clone.Value("timestep") != value {
		panic(fmt.Sprintf("parameter value did not round trip: %#v", clone))
	}
	force, ok := clone.Force("gravity")
	if !ok || force.Enabled != "true" || force.Values["magnitude"] != forceValue || force.Values["direction"] != "180" {
		panic(fmt.Sprintf("force did not round trip: %#v ok=%v", force, ok))
	}
	wall, ok := clone.WallEnabled("left")
	if !ok || !wall {
		panic(fmt.Sprintf("wall did not round trip: wall=%v ok=%v", wall, ok))
	}
	if clone.ActiveForce != "gravity" {
		panic(fmt.Sprintf("active force did not round trip: %q", clone.ActiveForce))
	}
	return true
}

func advanceDurationUsesPositiveBoundedSteps(durationInput, timestepInput float64) bool {
	duration := propertyFloat(durationInput, 0.001, 1)
	timestep := propertyFloat(timestepInput, 0.001, 0.1)
	world := NewWorld()
	world.Parameters.Set("timestep", fmt.Sprintf("%.9f", timestep))
	world.AdvanceDuration(duration)
	if world.LastAdvanceSteps <= 0 {
		panic("advance duration used no steps")
	}
	assertClose("advanced time", world.Time, duration, 1e-9)
	if float64(world.LastAdvanceSteps) > math.Ceil(duration/timestep)+1 {
		panic(fmt.Sprintf("too many advance steps: duration=%f timestep=%f steps=%d", duration, timestep, world.LastAdvanceSteps))
	}
	return true
}

func stepWithNoForcesMovesLinearlyAndKeepsFixedMasses(xInput, yInput, vxInput, vyInput, dtInput float64) bool {
	position := propertyVec(xInput, yInput, 10)
	velocity := propertyVec(vxInput, vyInput, 10)
	dt := propertyFloat(dtInput, 0.001, 0.1)
	world := NewWorld()
	world.Bounds = Bounds{Width: 10000, Height: 10000, Left: -5000, Bottom: -5000}
	_ = world.AddMass(Mass{ID: 1, Position: position, Velocity: velocity, Mass: 1})
	_ = world.AddMass(Mass{ID: 2, Position: position.Add(Vec2{X: 100, Y: 100}), Velocity: velocity.Scale(-1), Mass: 1, Fixed: true})
	fixedStart := world.Masses[1]
	world.Step(dt)
	assertVecClose("free linear position", world.Masses[0].Position, position.Add(velocity.Scale(dt)), 1e-9)
	assertVecClose("free linear velocity", world.Masses[0].Velocity, velocity, 1e-9)
	assertVecClose("fixed position", world.Masses[1].Position, fixedStart.Position, 0)
	assertVecClose("fixed velocity", world.Masses[1].Velocity, fixedStart.Velocity, 0)
	return true
}

func gravityForceScalesWithMass(massAInput, massBInput, magnitudeInput, directionInput float64) bool {
	massA := propertyFloat(massAInput, 0.1, 100)
	massB := propertyFloat(massBInput, 0.1, 100)
	magnitude := propertyFloat(magnitudeInput, 0.1, 100)
	direction := propertyFloat(directionInput, 0, 360)
	world := NewWorld()
	world.Parameters.EnableForce("gravity", map[string]string{
		"magnitude": fmt.Sprintf("%.9f", magnitude),
		"direction": fmt.Sprintf("%.9f", direction),
	})
	forceA := world.gravityForce(Mass{ID: 1, Mass: massA})
	forceB := world.gravityForce(Mass{ID: 2, Mass: massB})
	assertVecClose("gravity scales with mass", forceA.Scale(massB), forceB.Scale(massA), 1e-6)
	assertClose("gravity magnitude", math.Hypot(forceA.X, forceA.Y), magnitude*massA, 1e-6)
	return true
}

func viscosityForceOpposesVelocity(vxInput, vyInput, viscosityInput float64) bool {
	velocity := propertyVec(vxInput, vyInput, 100)
	viscosity := propertyFloat(viscosityInput, 0, 10)
	world := NewWorld()
	world.Parameters.Set("viscosity", fmt.Sprintf("%.9f", viscosity))
	force := world.viscosityForce(Mass{Velocity: velocity})
	assertVecClose("viscosity force", force, velocity.Scale(-viscosity), 1e-9)
	if dot(force, velocity) > 1e-9 {
		panic(fmt.Sprintf("viscosity accelerates velocity: force=%#v velocity=%#v", force, velocity))
	}
	return true
}

func centerOfMassTranslatesWithMasses(axInput, ayInput, bxInput, byInput, dxInput, dyInput float64) bool {
	a := propertyVec(axInput, ayInput, 100)
	b := propertyVec(bxInput, byInput, 100)
	delta := propertyVec(dxInput, dyInput, 100)
	world := NewWorld()
	_ = world.AddMass(Mass{ID: 1, Position: a, Mass: 1})
	_ = world.AddMass(Mass{ID: 2, Position: b, Mass: 10})
	center := world.centerOfMass()
	world.Masses[0].Position = world.Masses[0].Position.Add(delta)
	world.Masses[1].Position = world.Masses[1].Position.Add(delta)
	assertVecClose("translated center of mass", world.centerOfMass(), center.Add(delta), 1e-9)
	return true
}

func forceCenterTracksSingleSelectedMass(xInput, yInput float64) bool {
	position := propertyVec(xInput, yInput, 100)
	world := NewWorld()
	_ = world.AddMass(Mass{ID: 101, Position: position, Mass: 1})
	_ = world.AddMass(Mass{ID: 202, Position: position.Add(Vec2{X: 50}), Mass: 1})

	world.SetForceCenter([]int{101})
	if !world.IsCenterMass(101) || world.IsCenterMass(202) || world.CenterMassID() != 101 {
		panic(fmt.Sprintf("single force center not selected: center=%d", world.CenterMassID()))
	}
	assertVecClose("force center", world.forceCenter(), position, 0)

	world.SetForceCenter([]int{101, 202})
	if world.CenterMassID() != -1 {
		panic(fmt.Sprintf("multi-selection should clear center mass: %d", world.CenterMassID()))
	}
	assertVecClose("screen force center", world.forceCenter(), world.screenCenter(), 0)
	return true
}

func massRadiusIsBoundedAndMonotonic(massAInput, massBInput float64) bool {
	massA := propertyFloat(massAInput, 0, 10000)
	massB := propertyFloat(massBInput, 0, 10000)
	if massA > massB {
		massA, massB = massB, massA
	}
	radiusA := MassRadius(Mass{Mass: massA})
	radiusB := MassRadius(Mass{Mass: massB})
	if radiusA < 1 || radiusA > 64 || radiusB < 1 || radiusB > 64 {
		panic(fmt.Sprintf("radius outside bounds: massA=%f radiusA=%f massB=%f radiusB=%f", massA, radiusA, massB, radiusB))
	}
	if radiusA > radiusB {
		panic(fmt.Sprintf("radius not monotonic: massA=%f radiusA=%f massB=%f radiusB=%f", massA, radiusA, massB, radiusB))
	}
	if MassRadius(Mass{Mass: massB, Fixed: true}) != fixedMassCollisionRadius {
		panic("fixed mass radius changed")
	}
	return true
}

func wallSpringContactFractionFindsSegmentCrossing(yInput, xInput float64) bool {
	y := propertyFloat(yInput, 0.1, 100)
	x := propertyFloat(xInput, 0, 100)
	segment := Vec2{X: 100}
	lengthSquared := dot(segment, segment)
	previous := Vec2{X: x, Y: -y}
	current := Vec2{X: x, Y: y}
	fraction, ok := wallSpringContactFraction(previous, current, segment, lengthSquared, previous.Y, current.Y, false)
	if !ok {
		panic(fmt.Sprintf("crossing not found: previous=%#v current=%#v", previous, current))
	}
	assertClose("contact projection", fraction, x/100, 1e-9)
	rejectedFraction, rejected := wallSpringContactFraction(previous, Vec2{X: x, Y: y * -0.5}, segment, lengthSquared, previous.Y, -y*0.5, false)
	if rejected {
		panic(fmt.Sprintf("same-side crossing accepted: fraction=%f", rejectedFraction))
	}
	return true
}

func resolveWallSpringVelocitySeparatesOrKeepsSeparating(vxInput, vyInput, wallVxInput, wallVyInput, sideInput float64) bool {
	normal := Vec2{X: 1}
	startingSide := sideSign(propertySignedFloat(sideInput, 10))
	wallVelocity := propertyVec(wallVxInput, wallVyInput, 10)
	mass := Mass{
		Velocity:   propertyVec(vxInput, vyInput, 10),
		Elasticity: propertyFloat(sideInput+vxInput, 0, 2),
	}
	resolveWallSpringVelocity(&mass, wallVelocity, normal, startingSide)
	normalVelocity := dot(mass.Velocity.Sub(wallVelocity), normal)
	if !wallSpringVelocitySeparating(normalVelocity, startingSide) {
		panic(fmt.Sprintf("wall spring velocity still penetrating: normalVelocity=%f startingSide=%f mass=%#v wall=%#v", normalVelocity, startingSide, mass, wallVelocity))
	}
	return true
}

func resolveWallSpringVelocityUsesPositiveElasticity(speedInput, elasticityInput, sideInput float64) bool {
	normal := Vec2{X: 1}
	startingSide := sideSign(propertySignedFloat(sideInput, 10))
	speed := propertyFloat(speedInput, 0.1, 100)
	elasticity := propertyFloat(elasticityInput, 0.1, 2)
	mass := Mass{Velocity: normal.Scale(-startingSide * speed), Elasticity: elasticity}
	resolveWallSpringVelocity(&mass, Vec2{}, normal, startingSide)
	normalVelocity := dot(mass.Velocity, normal)
	assertClose("configured elasticity rebound", normalVelocity, startingSide*speed*elasticity, 1e-9)
	return true
}

func resetClearsWorldAndInsertFromAppendsObjects(xInput, yInput, massInput float64) bool {
	source := propertySampleWorld(xInput, yInput, massInput)
	world := NewWorld()
	world.InsertFrom(source)
	world.InsertFrom(source)
	if len(world.Masses) != len(source.Masses)*2 || len(world.Springs) != len(source.Springs)*2 {
		panic(fmt.Sprintf("insert did not append objects: masses=%d springs=%d", len(world.Masses), len(world.Springs)))
	}
	world.Reset()
	if len(world.Masses) != 0 || len(world.Springs) != 0 || world.Time != 0 {
		panic(fmt.Sprintf("reset left world state: %#v", world))
	}
	if world.Parameters.Value("timestep") != DefaultParameters().Value("timestep") {
		panic("reset did not restore default parameters")
	}
	return true
}

func addMassAtAndAddSpringBetweenGenerateConsistentIDs(axInput, ayInput, bxInput, byInput float64) bool {
	world := NewWorld()
	aIndex := world.AddMassAt(propertyVec(axInput, ayInput, 100), 1, false)
	bIndex := world.AddMassAt(propertyVec(bxInput, byInput, 100), 2, true)
	if aIndex != 0 || bIndex != 1 {
		panic(fmt.Sprintf("unexpected mass indexes: %d %d", aIndex, bIndex))
	}
	if world.Masses[aIndex].ID != 1 || world.Masses[bIndex].ID != 2 {
		panic(fmt.Sprintf("unexpected generated mass IDs: %#v", world.Masses))
	}
	world.AddSpringBetween(aIndex, bIndex, 10, 3)
	spring := world.Springs[0]
	if spring.ID != 1 || spring.A != aIndex || spring.B != bIndex || spring.MassA != world.Masses[aIndex].ID || spring.MassB != world.Masses[bIndex].ID {
		panic(fmt.Sprintf("generated spring inconsistent: %#v masses=%#v", spring, world.Masses))
	}
	if spring.Stiffness != 3 || spring.SpringConstant != 3 || spring.RestLength != 10 {
		panic(fmt.Sprintf("generated spring values inconsistent: %#v", spring))
	}
	return true
}

func adaptiveStepHelpersStayPositiveAndBounded(dtInput, precisionInput, badInput float64) bool {
	dt := propertyFloat(dtInput, 0.001, 1)
	precision := propertyFloat(precisionInput, 0.000001, 1)
	step := adaptiveStepDuration(dt, precision)
	if step <= 0 || step > dt {
		panic(fmt.Sprintf("adaptive step outside bounds: dt=%f precision=%f step=%f", dt, precision, step))
	}
	if positiveAdvanceStep(-propertyFloat(badInput, 0.001, 1), dt) <= 0 {
		panic("positiveAdvanceStep returned non-positive fallback")
	}
	parameters := DefaultParameters()
	parameters.Set("precision", fmt.Sprintf("%.9f", precision))
	parameters.Set("timestep", fmt.Sprintf("%.9f", dt))
	configuredDT := parameterFloat(parameters, "timestep")
	assertClose("configured precision", positiveParameterOrDefault(parameters, "precision", defaultPrecision), precision, 1e-9)
	assertClose("configured timestep", positiveParameterOrDefault(parameters, "timestep", defaultStepDuration), dt, 1e-9)
	world := NewWorld()
	world.Parameters = parameters
	world.Parameters.Set("adaptive timestep", "true")
	if world.advanceStepDuration() <= 0 || world.advanceStepDuration() > configuredDT {
		panic(fmt.Sprintf("world adaptive step outside bounds: %f", world.advanceStepDuration()))
	}
	world.Advance(3, configuredDT)
	assertClose("advance time", world.Time, 3*configuredDT, 1e-9)
	return true
}

func centerAndWallForcesPointTowardTheirTargets(xInput, yInput, magnitudeInput float64) bool {
	position := propertyVec(xInput, yInput, 50)
	if position == (Vec2{}) {
		position = Vec2{X: 1, Y: 1}
	}
	magnitude := propertyFloat(magnitudeInput, 1, 100)
	world := NewWorld()
	world.Bounds = Bounds{Width: 100, Height: 100}
	world.Parameters.EnableForce("center attraction", map[string]string{"magnitude": fmt.Sprintf("%.9f", magnitude), "exponent": "1"})
	center := Vec2{}
	centerForce := world.centerForce(Mass{ID: 1, Position: position, Mass: 1}, "center attraction", center)
	if dot(centerForce, center.Sub(position)) <= 0 {
		panic(fmt.Sprintf("center force does not point toward center: position=%#v force=%#v", position, centerForce))
	}
	world.Parameters.EnableForce("wall repulsion", map[string]string{"magnitude": fmt.Sprintf("%.9f", magnitude), "exponent": "1"})
	for _, wall := range []string{"left", "right", "bottom", "top"} {
		world.Parameters.EnableWall(wall)
	}
	nearLeftBottom := Mass{Position: Vec2{X: world.Bounds.MinX(), Y: world.Bounds.MinY()}}
	wallForce := world.wallForce(nearLeftBottom)
	if wallForce.X <= 0 || wallForce.Y <= 0 {
		panic(fmt.Sprintf("wall force does not point inward from left/bottom: %#v", wallForce))
	}
	checks := world.wallChecks(nearLeftBottom, magnitude)
	if len(checks) != 4 {
		panic(fmt.Sprintf("wallChecks count = %d", len(checks)))
	}
	return true
}

func enabledForceMatchesParameterState(forceInput float64) bool {
	world := NewWorld()
	if _, ok := world.enabledForce("gravity"); ok {
		panic("gravity unexpectedly enabled by default")
	}
	magnitude := fmt.Sprintf("%.9f", propertyFloat(forceInput, 1, 100))
	world.Parameters.EnableForce("gravity", map[string]string{"magnitude": magnitude})
	force, ok := world.enabledForce("gravity")
	if !ok || force.Enabled != "true" || force.Values["magnitude"] != magnitude {
		panic(fmt.Sprintf("enabled force mismatch: %#v ok=%v", force, ok))
	}
	return true
}

func stuckMassStaysOnWallUntilReleaseForceWins(forceInput float64) bool {
	world := NewWorld()
	world.Parameters.EnableWall("left")
	world.Parameters.Set("stickiness", "5")
	mass := Mass{ID: 1, Position: Vec2{X: -10, Y: 10}, Velocity: Vec2{X: -3, Y: 4}, StuckWall: "left", Mass: 1}
	if !world.keepStuck(&mass, Vec2{X: propertyFloat(forceInput, 0, 4.9)}) {
		panic("stuck mass released below stickiness")
	}
	if mass.Position.X != world.Bounds.MinX() || mass.Velocity.X != 0 || mass.Velocity.Y != 4 {
		panic(fmt.Sprintf("stuck mass not constrained to wall: %#v", mass))
	}
	wall, ok := world.stuckWall(&mass)
	if !ok || wall.name != "left" {
		panic(fmt.Sprintf("stuck wall lookup failed: %#v ok=%v", wall, ok))
	}
	if !world.wallReleasedBy(wall, Vec2{X: 6}) {
		panic("wall release force did not overcome stickiness")
	}
	if world.keepStuck(&mass, Vec2{X: 6}) {
		panic("stuck mass stayed stuck after release force")
	}
	return true
}

func stepDurationFollowsConfiguredTimestep(dtInput float64) bool {
	dt := propertyFloat(dtInput, 0.001, 1)
	parameters := DefaultParameters()
	parameters.Set("timestep", fmt.Sprintf("%.9f", dt))
	assertClose("step duration", parameters.StepDuration(), dt, 1e-9)
	return true
}

func massCollisionConservesMovableMomentum(massAInput, massBInput, speedAInput, speedBInput, overlapInput float64, fixedA bool, fixedB bool) bool {
	if fixedA && fixedB {
		fixedB = false
	}
	massAValue := propertyFloat(massAInput, 0.1, 100)
	massBValue := propertyFloat(massBInput, 0.1, 100)
	speedA := propertyFloat(speedAInput, 0.1, 100)
	speedB := -propertyFloat(speedBInput, 0.1, 100)
	massA := Mass{ID: 1, Position: Vec2{X: 0, Y: 0}, Velocity: Vec2{X: speedA, Y: 0}, Mass: massAValue, Fixed: fixedA}
	massBProbe := Mass{Mass: massBValue, Fixed: fixedB}
	massB := Mass{ID: 2, Position: Vec2{X: MassRadius(massA) + MassRadius(massBProbe) - propertyFloat(overlapInput, 0.001, 1), Y: 0}, Velocity: Vec2{X: speedB, Y: 0}, Mass: massBValue, Fixed: fixedB}
	geometry, ok := collisionGeometryFor(massA, massB)
	if !ok {
		panic(fmt.Sprintf("expected collision geometry for %#v %#v", massA, massB))
	}
	if firstCollisionPartnerIndex(4) != 5 {
		panic("firstCollisionPartnerIndex mismatch")
	}
	if !axisVelocitiesSeparating(-1, 1) || axisVelocitiesSeparating(1, 1) {
		panic("axisVelocitiesSeparating mismatch")
	}
	if collisionVelocitiesSeparating(massA.Velocity, massB.Velocity, geometry) {
		panic("approaching masses reported separating")
	}
	zeroDX := collisionGeometry{dy: 1, dyq: 1, sumxyq: 1}
	zeroDX.avoidVerticalDivision()
	if zeroDX.dx == 0 {
		panic("avoidVerticalDivision left dx zero")
	}
	if effectiveCollisionMass(Mass{}) != 1 {
		panic("zero mass effective mass should default to one")
	}
	if collisionRatio(massA, massB) <= 0 {
		panic("collisionRatio should be positive")
	}
	beforeMomentum := movableMomentum(massA).Add(movableMomentum(massB))
	beforeA := massA.Velocity
	beforeB := massB.Velocity
	world := NewWorld()
	world.Masses = []Mass{massA, massB}
	world.Parameters.EnableForce("mass collision", map[string]string{})
	world.applyMassCollisions()
	afterA := world.Masses[0]
	afterB := world.Masses[1]
	if fixedA && afterA.Velocity != beforeA {
		panic("fixed mass A velocity changed")
	}
	if fixedB && afterB.Velocity != beforeB {
		panic("fixed mass B velocity changed")
	}
	afterMomentum := movableMomentum(afterA).Add(movableMomentum(afterB))
	if !fixedA && !fixedB {
		assertClose("movable momentum x", afterMomentum.X, beforeMomentum.X, 1e-9)
		assertClose("movable momentum y", afterMomentum.Y, beforeMomentum.Y, 1e-9)
	}
	_, ok = collisionGeometryFor(afterA, afterB)
	if ok {
		relativeVelocity := afterA.Velocity.Sub(afterB.Velocity)
		displacement := afterB.Position.Sub(afterA.Position)
		if dot(relativeVelocity, displacement) > 1e-9 {
			panic(fmt.Sprintf("post-collision velocities still approaching: %#v %#v", afterA, afterB))
		}
	}
	return true
}

func finiteStepOutputsRemainFinite(axInput, ayInput, bxInput, byInput, vxInput, vyInput, dtInput float64) bool {
	world := NewWorld()
	world.Bounds = Bounds{Width: 1000, Height: 1000, Left: -500, Bottom: -500}
	world.Parameters.Set("viscosity", "0.05")
	world.Parameters.EnableForce("gravity", map[string]string{"magnitude": "0.25", "direction": "180"})
	world.Parameters.EnableForce("center attraction", map[string]string{"magnitude": "0.25", "exponent": "1"})
	world.Parameters.EnableForce("wall repulsion", map[string]string{"magnitude": "0.25", "exponent": "1"})
	for _, wall := range []string{"left", "right", "bottom", "top"} {
		world.Parameters.EnableWall(wall)
	}
	a := propertyVec(axInput, ayInput, 100)
	b := propertyVec(bxInput, byInput, 100)
	if length(b.Sub(a)) < 1 {
		b = a.Add(Vec2{X: 10})
	}
	_ = world.AddMass(Mass{ID: 1, Position: a, Velocity: propertyVec(vxInput, vyInput, 10), Mass: propertyFloat(axInput+vxInput, 0.1, 10)})
	_ = world.AddMass(Mass{ID: 2, Position: b, Velocity: propertyVec(vyInput, vxInput, 10), Mass: propertyFloat(byInput+vyInput, 0.1, 10)})
	_ = world.AddMass(Mass{ID: 3, Position: a.Add(Vec2{X: 20, Y: 5}), Mass: 1, Fixed: true})
	_ = world.AddSpring(Spring{ID: 1, MassA: 1, MassB: 2, RestLength: propertyFloat(dtInput, 1, 50), SpringConstant: propertyFloat(axInput+byInput, 0.01, 5), Damping: propertyFloat(vxInput+vyInput, 0, 1)})

	world.Step(propertyFloat(dtInput, 0.001, 0.05))

	for _, mass := range world.Masses {
		assertFiniteVec("finite position", mass.Position)
		assertFiniteVec("finite velocity", mass.Velocity)
	}
	return true
}

func forceEvaluationSkipsInvalidSpringsAndScalesAcceleration(axInput, ayInput, bxInput, byInput, massAInput, massBInput float64) bool {
	massA := propertyFloat(massAInput, 0.1, 100)
	massB := propertyFloat(massBInput, 0.1, 100)
	a := propertyVec(axInput, ayInput, 100)
	b := propertyVec(bxInput, byInput, 100)
	if length(b.Sub(a)) < 1 {
		b = a.Add(Vec2{X: 10})
	}
	world := NewWorld()
	_ = world.AddMass(Mass{ID: 1, Position: a, Mass: massA})
	_ = world.AddMass(Mass{ID: 2, Position: b, Mass: massB})
	_ = world.AddMass(Mass{ID: 3, Position: b.Add(Vec2{X: 50}), Mass: 3})
	world.Springs = []Spring{
		{ID: 1, MassA: 1, MassB: 2, RestLength: propertyFloat(axInput+bxInput, 1, 50), SpringConstant: propertyFloat(ayInput+byInput, 0.1, 10)},
		{ID: 2, MassA: 1, MassB: 99, RestLength: 1, SpringConstant: 1000},
		{ID: 3, MassA: 2, MassB: 3, RestLength: 1, SpringConstant: 1000, Wall: true},
	}

	evaluation := world.EvaluateForces()
	force1 := evaluation.ByMassID[1].Force
	force2 := evaluation.ByMassID[2].Force
	force3 := evaluation.ByMassID[3].Force
	assertVecClose("valid spring forces", force1.Add(force2), Vec2{}, 1e-8)
	assertVecClose("invalid and wall springs skipped", force3, Vec2{}, 1e-9)
	assertVecClose("mass 1 acceleration", evaluation.ByMassID[1].Acceleration, force1.Scale(1/massA), 1e-9)
	assertVecClose("mass 2 acceleration", evaluation.ByMassID[2].Acceleration, force2.Scale(1/massB), 1e-9)
	assertVecClose("mass 3 acceleration", evaluation.ByMassID[3].Acceleration, Vec2{}, 1e-9)
	return true
}

func wallSpringLengthConstraintCollisionKeepsEndpointOnBarrierSide(yInput, previousDistanceInput, currentDistanceInput, velocityInput float64) bool {
	y := propertyFloat(yInput, 5, 95)
	previousDistance := propertyFloat(previousDistanceInput, 0.1, 50)
	currentDistance := propertyFloat(currentDistanceInput, 0.1, 50)
	speed := propertyFloat(velocityInput, 0, 50)
	world := NewWorld()
	_ = world.AddMass(Mass{ID: 1, Position: Vec2{}, Mass: 1, Fixed: true})
	_ = world.AddMass(Mass{ID: 2, Position: Vec2{Y: 100}, Mass: 1, Fixed: true})
	_ = world.AddMass(Mass{ID: 3, Position: Vec2{X: currentDistance, Y: y}, Velocity: Vec2{X: speed}, Mass: 1})
	_ = world.AddMass(Mass{ID: 4, Position: Vec2{X: currentDistance + 40, Y: y}, Mass: 1})
	_ = world.AddSpring(Spring{ID: 1, MassA: 1, MassB: 2, Wall: true})
	_ = world.AddSpring(Spring{ID: 2, MassA: 3, MassB: 4, Wall: true})
	before := []Vec2{{}, {Y: 100}, {X: -previousDistance, Y: y}, {X: currentDistance + 40, Y: y}}

	world.applyWallSpringLengthConstraintCollisions(1, before)

	endpoint, _ := world.MassByID(3)
	side, err := propertyWallSpringSide(world, 1, endpoint.Position)
	if err != nil {
		panic(err)
	}
	if side < MassRadius(endpoint)-1e-9 {
		panic(fmt.Sprintf("endpoint not kept on barrier side: side=%f endpoint=%#v", side, endpoint))
	}
	if endpoint.Velocity.X > 1e-9 {
		panic(fmt.Sprintf("endpoint velocity still penetrates barrier: %#v", endpoint.Velocity))
	}
	return true
}

func movingWallSpringFixedEndpointCollisionSeparatesContact(fractionInput, massAInput, massBInput, speedInput, penetrationInput float64) bool {
	fraction := propertyFloat(fractionInput, 0.05, 0.95)
	massA := propertyFloat(massAInput, 0.1, 100)
	massB := propertyFloat(massBInput, 0.1, 100)
	speed := propertyFloat(speedInput, 0.1, 100)
	penetration := propertyFloat(penetrationInput, 0.1, 20)
	left := -20.0 * fraction
	right := 20.0 * (1 - fraction)
	world := NewWorld()
	_ = world.AddMass(Mass{ID: 1, Position: Vec2{}, Mass: 1, Fixed: true})
	_ = world.AddMass(Mass{ID: 2, Position: Vec2{Y: 100}, Mass: 1, Fixed: true})
	_ = world.AddMass(Mass{ID: 3, Position: Vec2{X: left, Y: penetration}, Velocity: Vec2{Y: speed}, Mass: massA})
	_ = world.AddMass(Mass{ID: 4, Position: Vec2{X: right, Y: penetration}, Velocity: Vec2{Y: speed}, Mass: massB})
	_ = world.AddSpring(Spring{ID: 1, MassA: 1, MassB: 2, Wall: true})
	_ = world.AddSpring(Spring{ID: 2, MassA: 3, MassB: 4, RestLength: right - left, Wall: true})
	starts := []Vec2{{}, {Y: 100}, {X: left, Y: -penetration}, {X: right, Y: -penetration}}
	beforeFixed := world.Masses[0]

	aIndex, bIndex, ok := world.movingWallSpringEndpointIndexes(world.Springs[1])
	if !ok || aIndex != 2 || bIndex != 3 {
		panic(fmt.Sprintf("moving wall endpoint indexes mismatch: %d %d %t", aIndex, bIndex, ok))
	}
	normal, contactFraction, currentSide, ok := world.movingWallSpringFixedEndpointContact(aIndex, bIndex, world.Masses[0], starts)
	if !ok {
		panic("expected fixed endpoint contact")
	}
	if fixedEndpointContactOutside(currentSide) {
		panic(fmt.Sprintf("expected penetrating contact side, got %f", currentSide))
	}
	world.applyMovingWallSpringFixedEndpointCollisions(1, starts)

	fixed, _ := world.MassByID(1)
	if fixed != beforeFixed {
		panic(fmt.Sprintf("fixed endpoint changed: before=%#v after=%#v", beforeFixed, fixed))
	}
	endpointA, _ := world.MassByID(3)
	endpointB, _ := world.MassByID(4)
	afterFraction, afterSide, ok := currentFixedEndpointContact(fixed.Position, endpointA.Position, endpointB.Position, normal)
	if !ok {
		panic("current contact rejected after collision")
	}
	if afterSide < fixedMassCollisionRadius-1e-9 {
		panic(fmt.Sprintf("fixed endpoint contact still penetrating: side=%f fraction=%f", afterSide, afterFraction))
	}
	contactVelocity := wallSpringContactVelocity(&endpointA, &endpointB, contactFraction)
	if dot(contactVelocity, normal) < -1e-9 {
		panic(fmt.Sprintf("contact velocity still penetrates fixed endpoint: velocity=%#v normal=%#v", contactVelocity, normal))
	}
	if !propertyClose(afterFraction, contactFraction, 1e-9) {
		panic(fmt.Sprintf("contact fraction drifted: before=%f after=%f", contactFraction, afterFraction))
	}
	return true
}

func wallSpringCollisionConservesMomentumWithVaryingMasses(fractionInput, endpointAInput, endpointBInput, massInput, speedInput, distanceInput float64) bool {
	fraction := propertyFloat(fractionInput, 0.05, 0.95)
	endpointMassA := propertyFloat(endpointAInput, 0.1, 100)
	endpointMassB := propertyFloat(endpointBInput, 0.1, 100)
	collidingMass := propertyFloat(massInput, 0.1, 100)
	speed := propertyFloat(speedInput, 0.1, 100)
	distance := propertyFloat(distanceInput, 0.1, 100)
	y := fraction * 100
	world := NewWorld()
	_ = world.AddMass(Mass{ID: 1, Position: Vec2{}, Mass: endpointMassA})
	_ = world.AddMass(Mass{ID: 2, Position: Vec2{Y: 100}, Mass: endpointMassB})
	_ = world.AddMass(Mass{ID: 3, Position: Vec2{X: distance, Y: y}, Velocity: Vec2{X: speed}, Mass: collidingMass})
	_ = world.AddSpring(Spring{ID: 1, MassA: 1, MassB: 2, Wall: true})
	before := propertyMomentum(world, 1, 2, 3)

	world.applyWallSpringCollision(world.Springs[0], &world.Masses[2], &world.Masses[0], &world.Masses[1], Vec2{X: -distance, Y: y}, world.Masses[0].Position, world.Masses[1].Position, false)

	after := propertyMomentum(world, 1, 2, 3)
	assertVecClose("wall spring collision momentum", after, before, 1e-8)
	mass, _ := world.MassByID(3)
	side, err := propertyWallSpringSide(world, 1, mass.Position)
	if err != nil {
		panic(err)
	}
	if side < MassRadius(mass)-1e-9 {
		panic(fmt.Sprintf("mass not placed on starting side: side=%f mass=%#v", side, mass))
	}
	return true
}

func movableMomentum(mass Mass) Vec2 {
	if mass.Fixed {
		return Vec2{}
	}
	return mass.Velocity.Scale(effectiveCollisionMass(mass))
}

func propertyMomentum(world *Simulation, ids ...int) Vec2 {
	total := Vec2{}
	for _, id := range ids {
		mass, ok := world.MassByID(id)
		if !ok || mass.Fixed {
			continue
		}
		total = total.Add(movableMomentum(mass))
	}
	return total
}

func propertyFloat(input float64, minimum float64, maximum float64) float64 {
	if math.IsNaN(input) || math.IsInf(input, 0) {
		return minimum
	}
	value := math.Abs(input)
	return minimum + math.Mod(value, maximum-minimum)
}

func propertySignedFloat(input float64, magnitude float64) float64 {
	return propertyFloat(input, 0, magnitude*2) - magnitude
}

func propertyVec(xInput, yInput float64, magnitude float64) Vec2 {
	return Vec2{X: propertySignedFloat(xInput, magnitude), Y: propertySignedFloat(yInput, magnitude)}
}

func propertySampleWorld(xInput, yInput, massInput float64) *Simulation {
	world := NewWorld()
	world.Bounds = Bounds{Width: 200, Height: 100, Left: -10, Bottom: -20}
	world.Parameters.EnableForce("gravity", map[string]string{"magnitude": "1", "direction": "90"})
	world.Parameters.EnableWall("right")
	_ = world.AddMass(Mass{ID: 1, Position: propertyVec(xInput, yInput, 100), Mass: propertyFloat(massInput, 0.1, 100)})
	_ = world.AddMass(Mass{ID: 2, Position: Vec2{X: 10, Y: 10}, Mass: 2})
	_ = world.AddSpring(Spring{ID: 1, MassA: 1, MassB: 2, RestLength: 10, SpringConstant: 5})
	return world
}

func assertIndependentWorldCopy(label string, world *Simulation) {
	if world.Masses[0].Position.X == 100 {
		panic(label + " aliases masses")
	}
	if world.Springs[0].RestLength == 110 {
		panic(label + " aliases springs")
	}
	if world.Parameters.Value("timestep") == "42" {
		panic(label + " aliases parameter values")
	}
	if world.Parameters.Forces["gravity"].Values["magnitude"] == "99" {
		panic(label + " aliases force values")
	}
	if world.Parameters.Walls["left"] {
		panic(label + " aliases walls")
	}
}

func assertVecClose(label string, actual, expected Vec2, tolerance float64) {
	assertClose(label+" x", actual.X, expected.X, tolerance)
	assertClose(label+" y", actual.Y, expected.Y, tolerance)
}

func assertFiniteVec(label string, value Vec2) {
	if math.IsNaN(value.X) || math.IsInf(value.X, 0) || math.IsNaN(value.Y) || math.IsInf(value.Y, 0) {
		panic(fmt.Sprintf("%s is not finite: %#v", label, value))
	}
}

func assertClose(label string, actual, expected, tolerance float64) {
	if !propertyClose(actual, expected, tolerance) {
		panic(fmt.Sprintf("%s: got %f, want %f +/- %f", label, actual, expected, tolerance))
	}
}

func propertyClose(actual, expected, tolerance float64) bool {
	return math.Abs(actual-expected) <= tolerance
}

func propertyWallSpringSide(world *Simulation, springID int, point Vec2) (float64, error) {
	spring, ok := world.SpringByID(springID)
	if !ok {
		return 0, fmt.Errorf("spring %d not found", springID)
	}
	aIndex, bIndex, ok := world.wallSpringEndpointIndexes(spring)
	if !ok {
		return 0, fmt.Errorf("spring %d endpoints not found", springID)
	}
	start := world.Masses[aIndex].Position
	segment := world.Masses[bIndex].Position.Sub(start)
	normal := Vec2{X: -segment.Y, Y: segment.X}.Normalize()
	return dot(point.Sub(start), normal), nil
}
