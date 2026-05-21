package acceptance

import (
	"fmt"
	"strings"

	"springs/internal/app"
	"springs/internal/sim"
)

func init() {
	for step, handler := range map[string]stepHandler{
		"the wall spring barriers task is accepted":                                                          acceptStep,
		"spring <spring_id> connects mass <mass_a> to mass <mass_b>":                                         addBarrierSpring,
		"spring <spring_id> has Wall value <wall>":                                                           setBarrierSpringWall,
		"spring <spring_id> has Kspring <kspring> Kdamp <kdamp> RestLen <rest_len>":                          setBarrierSpringParameters,
		"the coder evaluates spring <spring_id> forces":                                                      evaluateBarrierSpringForces,
		"spring <spring_id> should apply spring force state <spring_force_state>":                            assertBarrierSpringForceState,
		"spring <spring_id> should apply damping force state <damping_force_state>":                          assertBarrierSpringDampingState,
		"wall spring <spring_id> spans from <wall_x1>, <wall_y1> to <wall_x2>, <wall_y2>":                    createWallSpringByCoordinates,
		"moving mass <mass_id> starts at <mass_x>, <mass_y> with velocity <mass_vx>, <mass_vy>":              createBarrierMovingMass,
		"the coder advances through wall spring collision":                                                   advanceThroughWallSpringCollision,
		"mass <mass_id> should remain on the starting side of wall spring <spring_id>":                       assertMassOnStartingWallSpringSide,
		"mass <mass_id> velocity should be resolved away from wall spring <spring_id>":                       assertMassVelocityResolvedAwayFromWallSpring,
		"wall spring <spring_id> spans from mass <endpoint_a> to mass <endpoint_b>":                          createWallSpringByEndpointIDs,
		"wall spring endpoint <endpoint_a> fixed state is <fixed_a>":                                         setWallSpringEndpointFixed,
		"wall spring endpoint <endpoint_b> fixed state is <fixed_b>":                                         setWallSpringEndpointBFixed,
		"moving mass <mass_id> collides with wall spring <spring_id> at contact fraction <contact_fraction>": createMassCollidingWithWallSpring,
		"the coder resolves the wall spring collision":                                                       resolveWallSpringCollision,
		"wall spring endpoint <endpoint_a> should receive impulse share <impulse_share_a>":                   assertWallSpringEndpointImpulseShare,
		"wall spring endpoint <endpoint_b> should receive impulse share <impulse_share_b>":                   assertWallSpringEndpointBImpulseShare,
		"XSP input contains spring <spring_id> with Wall value <input_wall>":                                 createWallSpringXSPInput,
		"loaded spring <spring_id> should have Wall value <loaded_wall>":                                     assertLoadedWallSpringXSP,
		"saved spring <spring_id> should include Wall value <saved_wall>":                                    assertSavedWallSpringXSP,
		"selected spring <spring_id> has Wall value <old_wall>":                                              createSelectedSpringWithWall,
		"the coder changes spring control Wall to <new_wall>":                                                changeSpringWallControl,
		"spring <spring_id> should have Wall value <new_wall>":                                               assertSpringWallValue,
		"spring <spring_id> has Wall value <old_wall>":                                                       createMenuSpringWithWall,
		"spring <spring_id> right-click menu includes item <menu_item>":                                      assertSpringMenuIncludesItem,
		"the coder selects spring menu item Wall for spring <spring_id>":                                     selectSpringMenuWallItem,
		"the coder renders spring <spring_id>":                                                               renderWallSpring,
		"spring <spring_id> should use spring rendering style <rendering_style>":                             assertWallSpringRenderingStyle,
	} {
		stepHandlers[step] = handler
	}
}

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
	world := ensureDomainWorld(w)
	if _, ok := world.SpringByID(springID); !ok {
		ensureWallSpringMass(world, 1, sim.Vec2{})
		ensureWallSpringMass(world, 2, sim.Vec2{X: 20})
		if err := world.AddSpring(sim.Spring{ID: springID, MassA: 1, MassB: 2}); err != nil {
			return err
		}
	}
	return updateBarrierSpring(w, example, func(spring *sim.Spring) { spring.Wall = wall })
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
	if state != "enabled" && state != "disabled" {
		return fmt.Errorf("invalid force state %q", state)
	}
	enabled := false
	for _, forces := range w.forceEvaluation.ByMassID {
		if forces.Force != (sim.Vec2{}) {
			enabled = true
		}
	}
	if (state == "enabled") != enabled {
		return fmt.Errorf("%s force enabled = %t, forces = %#v", key, enabled, w.forceEvaluation.ByMassID)
	}
	return nil
}

func createWallSpringByCoordinates(w *world, example map[string]string) error {
	if err := requireWallSpringExampleValues(example, map[string]string{
		"spring_id": "1",
		"wall_x1":   "0",
		"wall_y1":   "0",
		"wall_x2":   "0",
		"wall_y2":   "100",
	}); err != nil {
		return err
	}
	springID, err := intValue(example, "spring_id")
	if err != nil {
		return err
	}
	x1, err := floatValue(example, "wall_x1")
	if err != nil {
		return err
	}
	y1, err := floatValue(example, "wall_y1")
	if err != nil {
		return err
	}
	x2, err := floatValue(example, "wall_x2")
	if err != nil {
		return err
	}
	y2, err := floatValue(example, "wall_y2")
	if err != nil {
		return err
	}
	world := ensureDomainWorld(w)
	ensureWallSpringMass(world, 1, sim.Vec2{X: x1, Y: y1})
	ensureWallSpringMass(world, 2, sim.Vec2{X: x2, Y: y2})
	return world.AddSpring(sim.Spring{ID: springID, MassA: 1, MassB: 2, Wall: true})
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
	x, err := floatValue(example, "mass_x")
	if err != nil {
		return err
	}
	y, err := floatValue(example, "mass_y")
	if err != nil {
		return err
	}
	vx, err := floatValue(example, "mass_vx")
	if err != nil {
		return err
	}
	vy, err := floatValue(example, "mass_vy")
	if err != nil {
		return err
	}
	world := ensureDomainWorld(w)
	if err := world.AddMass(sim.Mass{ID: id, Position: sim.Vec2{X: x, Y: y}, Velocity: sim.Vec2{X: vx, Y: vy}, Mass: 1}); err != nil {
		return err
	}
	return rememberWallSpringStartingSide(w, example, id)
}

func advanceThroughWallSpringCollision(w *world, _ map[string]string) error {
	ensureDomainWorld(w).Step(1)
	return nil
}

func assertMassOnStartingWallSpringSide(w *world, example map[string]string) error {
	massID, err := intValue(example, "mass_id")
	if err != nil {
		return err
	}
	springID, err := intValue(example, "spring_id")
	if err != nil {
		return err
	}
	world := ensureDomainWorld(w)
	mass, ok := world.MassByID(massID)
	if !ok {
		return fmt.Errorf("mass %d not found", massID)
	}
	side, ok := w.wallSpringSides[massID]
	if !ok {
		return fmt.Errorf("starting side for mass %d was not recorded", massID)
	}
	current, err := wallSpringSide(world, springID, mass.Position)
	if err != nil {
		return err
	}
	if current*side < 0 {
		return fmt.Errorf("mass %d crossed wall spring %d: side %f started %f", massID, springID, current, side)
	}
	return nil
}

func assertMassVelocityResolvedAwayFromWallSpring(w *world, example map[string]string) error {
	massID, err := intValue(example, "mass_id")
	if err != nil {
		return err
	}
	springID, err := intValue(example, "spring_id")
	if err != nil {
		return err
	}
	world := ensureDomainWorld(w)
	mass, ok := world.MassByID(massID)
	if !ok {
		return fmt.Errorf("mass %d not found", massID)
	}
	side, ok := w.wallSpringSides[massID]
	if !ok {
		return fmt.Errorf("starting side for mass %d was not recorded", massID)
	}
	normal, err := wallSpringNormal(world, springID)
	if err != nil {
		return err
	}
	if dotAcceptance(mass.Velocity, normal)*side < 0 {
		return fmt.Errorf("mass %d velocity penetrates wall spring %d: %#v", massID, springID, mass.Velocity)
	}
	return nil
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
	if err := requireWallSpringExampleValues(example, map[string]string{"spring_id": "1", "mass_id": "3"}); err != nil {
		return err
	}
	massID, err := intValue(example, "mass_id")
	if err != nil {
		return err
	}
	contactFraction, err := floatValue(example, "contact_fraction")
	if err != nil {
		return err
	}
	world := ensureDomainWorld(w)
	if err := world.AddMass(sim.Mass{ID: massID, Position: sim.Vec2{X: -5, Y: 100 * contactFraction}, Velocity: sim.Vec2{X: 10}, Mass: 1}); err != nil {
		return err
	}
	w.wallSpringImpulses = map[int]sim.Vec2{}
	for _, mass := range world.Masses {
		w.wallSpringImpulses[mass.ID] = mass.Velocity
	}
	return rememberWallSpringStartingSide(w, example, massID)
}

func resolveWallSpringCollision(w *world, _ map[string]string) error {
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
	endpoint, err := intValue(example, endpointKey)
	if err != nil {
		return err
	}
	expected, err := stringValue(example, shareKey)
	if err != nil {
		return err
	}
	if expected == "absorbed" {
		expected = "0"
	}
	expectedShare, err := parseExpectedImpulseShare(expected)
	if err != nil {
		return fmt.Errorf("invalid impulse share %q", expected)
	}
	world := ensureDomainWorld(w)
	mass, ok := world.MassByID(endpoint)
	if !ok {
		return fmt.Errorf("endpoint %d not found", endpoint)
	}
	before := w.wallSpringImpulses[endpoint]
	actualShare := (mass.Velocity.X - before.X) / 20
	if actualShare < expectedShare-0.000001 || actualShare > expectedShare+0.000001 {
		return fmt.Errorf("endpoint %d impulse share = %f, expected %f", endpoint, actualShare, expectedShare)
	}
	return nil
}

func parseExpectedImpulseShare(value string) (float64, error) {
	switch value {
	case "0.75":
		return 0.75, nil
	case "0.50":
		return 0.50, nil
	case "0.25":
		return 0.25, nil
	case "0":
		return 0, nil
	default:
		return 0, fmt.Errorf("unsupported share")
	}
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
	springID, err := intValue(example, "spring_id")
	if err != nil {
		return err
	}
	expected, err := stringValue(example, "saved_wall")
	if err != nil {
		return err
	}
	needle := fmt.Sprintf("spng %d 1 2 12 0.7 10 %s\n", springID, expected)
	if !strings.Contains(w.xspSavedFirst, needle) {
		return fmt.Errorf("saved XSP missing %q in:\n%s", needle, w.xspSavedFirst)
	}
	return nil
}

func createSelectedSpringWithWall(w *world, example map[string]string) error {
	if err := requireWallSpringExampleValues(example, map[string]string{"spring_id": "1"}); err != nil {
		return err
	}
	springID, oldWall, err := springIDAndWall(example, "old_wall")
	if err != nil {
		return err
	}
	game := app.NewGame()
	world := sim.NewWorld()
	ensureWallSpringMass(world, 1, sim.Vec2{})
	ensureWallSpringMass(world, 2, sim.Vec2{X: 20})
	if err := world.AddSpring(sim.Spring{ID: springID, MassA: 1, MassB: 2, Wall: oldWall}); err != nil {
		return err
	}
	game.ReplaceWorld(world)
	if err := game.SelectSpring(springID); err != nil {
		return err
	}
	w.appGame = game
	return nil
}

func changeSpringWallControl(w *world, _ map[string]string) error {
	game, ok := w.appGame.(*app.Game)
	if !ok {
		return fmt.Errorf("application game is not concrete")
	}
	if !game.ClickVisibleControl("Wall") {
		return fmt.Errorf("Wall control click was not handled")
	}
	return nil
}

func assertSpringWallValue(w *world, example map[string]string) error {
	springID, expected, err := springIDAndWall(example, "new_wall")
	if err != nil {
		return err
	}
	var world *sim.Simulation
	if w.appGame != nil {
		game, _ := w.appGame.(*app.Game)
		world = game.World()
	} else {
		world = ensureDomainWorld(w)
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

func createMenuSpringWithWall(w *world, example map[string]string) error {
	if err := requireWallSpringExampleValues(example, map[string]string{"spring_id": "1"}); err != nil {
		return err
	}
	springID, oldWall, err := springIDAndWall(example, "old_wall")
	if err != nil {
		return err
	}
	game := app.NewGame()
	world := sim.NewWorld()
	ensureWallSpringMass(world, 1, sim.Vec2{})
	ensureWallSpringMass(world, 2, sim.Vec2{X: 20})
	if err := world.AddSpring(sim.Spring{ID: springID, MassA: 1, MassB: 2, Wall: oldWall}); err != nil {
		return err
	}
	game.ReplaceWorld(world)
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
	game, ok := w.appGame.(*app.Game)
	if !ok {
		return fmt.Errorf("application game is not concrete")
	}
	for _, label := range game.SpringContextMenuLabelsForSpring(springID) {
		if label == item {
			return nil
		}
	}
	return fmt.Errorf("spring menu did not include %q", item)
}

func selectSpringMenuWallItem(w *world, example map[string]string) error {
	springID, err := intValue(example, "spring_id")
	if err != nil {
		return err
	}
	game, ok := w.appGame.(*app.Game)
	if !ok {
		return fmt.Errorf("application game is not concrete")
	}
	if !game.SelectSpringContextMenuItem(springID, "Wall") {
		return fmt.Errorf("Wall spring menu item was not handled")
	}
	return nil
}

func createRenderableWallSpring(w *world, example map[string]string) error {
	if err := requireWallSpringExampleValues(example, map[string]string{"spring_id": "1"}); err != nil {
		return err
	}
	springID, wall, err := springIDAndWall(example, "wall")
	if err != nil {
		return err
	}
	game := app.NewGame()
	world := sim.NewWorld()
	ensureWallSpringMass(world, 1, sim.Vec2{})
	ensureWallSpringMass(world, 2, sim.Vec2{X: 20})
	if err := world.AddSpring(sim.Spring{ID: springID, MassA: 1, MassB: 2, Wall: wall}); err != nil {
		return err
	}
	game.ReplaceWorld(world)
	w.appGame = game
	return nil
}

func renderWallSpring(w *world, _ map[string]string) error {
	game, ok := w.appGame.(*app.Game)
	if !ok {
		game = app.NewGame()
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
	switch style {
	case "normal":
		if got := w.renderResult.Representations["spring"]; got != "cyan line" {
			return fmt.Errorf("normal spring representation = %q", got)
		}
		if got := w.renderResult.Representations["wall spring"]; got != "" {
			return fmt.Errorf("normal spring should not have wall representation %q", got)
		}
	case "wall":
		if got := w.renderResult.Representations["wall spring"]; got != "heavy orange line" {
			return fmt.Errorf("wall spring representation = %q", got)
		}
	default:
		return fmt.Errorf("unsupported rendering style %q", style)
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

func ensureWallSpringMass(world *sim.Simulation, id int, position sim.Vec2) {
	for i := range world.Masses {
		if world.Masses[i].ID == id {
			world.Masses[i].Position = position
			return
		}
	}
	_ = world.AddMass(sim.Mass{ID: id, Position: position, Mass: 1})
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
