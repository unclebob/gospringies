package acceptance

import (
	"fmt"
	"math"
	"strings"

	"springs/internal/sim"
)

var supportedForceNames = map[string]struct{}{
	"gravity":                   {},
	"center of mass attraction": {},
	"center attraction":         {},
	"wall repulsion":            {},
}

const forceDirectionTolerance = 0.000001

func selectForce(w *world, example map[string]string) error {
	force, err := stringValue(example, "force")
	if err != nil {
		return err
	}
	if !supportedForceName(force) {
		return fmt.Errorf("unsupported force %q", force)
	}
	world := ensureDomainWorld(w)
	world.Parameters.SelectForce(force)
	return nil
}

func assertForceExposesParameter(_ *world, example map[string]string) error {
	force, parameter, err := stringPair(example, "force", "parameter_one")
	if err != nil {
		return err
	}
	if !hasForceParameter(force, parameter) {
		return fmt.Errorf("%s does not expose %s", force, parameter)
	}
	second, err := stringValue(example, "parameter_two")
	if err != nil {
		return err
	}
	if !hasForceParameter(force, second) {
		return fmt.Errorf("%s does not expose %s", force, second)
	}
	return nil
}

func hasForceParameter(force, parameter string) bool {
	for _, candidate := range sim.ForceParameterNames(force) {
		if candidate == parameter {
			return true
		}
	}
	return false
}

func supportedForceName(force string) bool {
	_, ok := supportedForceNames[force]
	return ok
}

func setGravityDirection(w *world, example map[string]string) error {
	direction, err := stringValue(example, "direction_degrees")
	if err != nil {
		return err
	}
	world := ensureDomainWorld(w)
	world.Parameters.EnableForce("gravity", map[string]string{"magnitude": "1", "direction": direction})
	_ = world.AddMass(sim.Mass{ID: 1, Mass: 1})
	return nil
}

func evaluateGravity(w *world, _ map[string]string) error {
	return evaluateCurrentForces(w)
}

func assertGravityDirection(w *world, example map[string]string) error {
	expected, err := stringValue(example, "expected_direction")
	if err != nil {
		return err
	}
	force := w.forceEvaluation.ByMassID[1].Force
	if !matchesExpectedDirection(force, expected) {
		return fmt.Errorf("gravity force = %#v, want %s", force, expected)
	}
	return nil
}

func matchesExpectedDirection(force sim.Vec2, expected string) bool {
	directions := map[string]sim.Vec2{
		"down":  {Y: -1},
		"right": {X: 1},
		"up":    {Y: 1},
		"left":  {X: -1},
	}
	want, ok := directions[expected]
	return ok && matchesForceDirectionComponent(force.X, want.X) && matchesForceDirectionComponent(force.Y, want.Y)
}

func matchesForceDirectionComponent(actual, expected float64) bool {
	return math.Abs(actual-expected) < forceDirectionTolerance
}

func createSelectedMasses(w *world, example map[string]string) error {
	selected, err := stringValue(example, "selected_masses")
	if err != nil {
		return err
	}
	world := ensureDomainWorld(w)
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 10, Y: 20}, Mass: 1})
	ids, err := selectedMassIDs(selected)
	if err != nil {
		return err
	}
	w.originalMassIDs = ids
	return nil
}

func selectedMassIDs(selected string) ([]int, error) {
	if selected == "none" {
		return nil, nil
	}
	if selected == "1" {
		return []int{1}, nil
	}
	return nil, fmt.Errorf("unsupported selected masses %q", selected)
}

func setForceCenter(w *world, _ map[string]string) error {
	ensureDomainWorld(w).SetForceCenter(w.originalMassIDs)
	return nil
}

func assertForceCenter(w *world, example map[string]string) error {
	expected, err := stringValue(example, "expected_center")
	if err != nil {
		return err
	}
	actual := "screen center"
	if ensureDomainWorld(w).CenterMassID() > 0 {
		actual = fmt.Sprintf("mass %d", ensureDomainWorld(w).CenterMassID())
	}
	if actual != expected {
		return fmt.Errorf("force center = %s, want %s", actual, expected)
	}
	return nil
}

func createForceCenterMass(w *world, example map[string]string) error {
	id, err := intValue(example, "center_mass")
	if err != nil {
		return err
	}
	world := ensureDomainWorld(w)
	_ = world.AddMass(sim.Mass{ID: id, Position: sim.Vec2{X: 50, Y: 50}, Mass: 1})
	_ = world.AddMass(sim.Mass{ID: id + 1, Position: sim.Vec2{X: 0, Y: 50}, Mass: 1})
	world.SetForceCenter([]int{id})
	return nil
}

func enableNamedForce(w *world, example map[string]string) error {
	force, err := stringValue(example, "force")
	if err != nil {
		return err
	}
	if !supportedForceName(force) {
		return fmt.Errorf("unsupported force %q", force)
	}
	ensureDomainWorld(w).Parameters.EnableForce(force, map[string]string{"magnitude": "10", "exponent": "0", "damping": "1"})
	return nil
}

func evaluateCenterForces(w *world, _ map[string]string) error {
	return evaluateCurrentForces(w)
}

func evaluateCurrentForces(w *world) error {
	w.forceEvaluation = ensureDomainWorld(w).EvaluateForces()
	return nil
}

func assertCenterMassVisuallyMarked(w *world, example map[string]string) error {
	id, err := intValue(example, "center_mass")
	if err != nil {
		return err
	}
	game := newApplicationDriverGame()
	game.ReplaceWorld(ensureDomainWorld(w))
	if !game.RenderWorld().HasVisibleRepresentation("force center") || !game.World().IsCenterMass(id) {
		return fmt.Errorf("center mass %d was not visually marked", id)
	}
	return nil
}

func assertNoReciprocalCenterForce(w *world, example map[string]string) error {
	id, err := intValue(example, "center_mass")
	if err != nil {
		return err
	}
	forceName, err := stringValue(example, "force")
	if err != nil {
		return err
	}
	if w.forceEvaluation.ByMassID[id].Force != (sim.Vec2{}) {
		return fmt.Errorf("center mass received reciprocal response from %s: %#v", forceName, w.forceEvaluation.ByMassID[id].Force)
	}
	return nil
}

func enableForceForControls(w *world, example map[string]string) error {
	return enableNamedForce(w, example)
}

func assertForceControlsActive(w *world, example map[string]string) error {
	force, err := stringValue(example, "force")
	if err != nil {
		return err
	}
	if !supportedForceName(force) {
		return fmt.Errorf("unsupported force %q", force)
	}
	active := strings.TrimSpace(ensureDomainWorld(w).Parameters.ActiveForce)
	if active != force {
		return fmt.Errorf("active force controls = %q, want %q", active, force)
	}
	return nil
}

// mutate4go-manifest-begin
// {"version":1,"tested_at":"2026-05-22T10:55:55-05:00","module_hash":"f25eabe94b56af39a5b9c89f3abd35842d44ca5ad8cd6d1b4995a7e4019f2227","functions":[{"id":"func/selectForce","name":"selectForce","line":20,"end_line":31,"hash":"cf904d821050f877ed7d7b315b5584c5b0f403747a1f229c70bc4fdb73e8f298"},{"id":"func/assertForceExposesParameter","name":"assertForceExposesParameter","line":33,"end_line":49,"hash":"2290c41e99a84a9b4e9ef8395d00fadd244cdcc87dc391b1eee1f37c528bc6fd"},{"id":"func/hasForceParameter","name":"hasForceParameter","line":51,"end_line":58,"hash":"7f0267b924413baa51e088cba6c477f3e0c8ce4806d4eb879119cee1b7de9642"},{"id":"func/supportedForceName","name":"supportedForceName","line":60,"end_line":63,"hash":"c31e09553f1ef0a30cb802a77237ca65acc56a1caa62af3d68bbdb8572f9171f"},{"id":"func/setGravityDirection","name":"setGravityDirection","line":65,"end_line":74,"hash":"1cc1c4cb23ff5ae2f24438bc43f422ada4e10de408a3e09be9b1cfa2a391825c"},{"id":"func/evaluateGravity","name":"evaluateGravity","line":76,"end_line":78,"hash":"53fde89e8bbdbc64add3a9c4800980000b0aa30e7155316e4e66b45070effc13"},{"id":"func/assertGravityDirection","name":"assertGravityDirection","line":80,"end_line":90,"hash":"d26856d8a2fdc1a77f6727cc131d67ebaebc219f72888cb69df0891b47bf5ff0"},{"id":"func/matchesExpectedDirection","name":"matchesExpectedDirection","line":92,"end_line":101,"hash":"45c124e4c25be2fd2cd970cefdb217f31be49f4d5734ea37baa55a6fa8018dc7"},{"id":"func/matchesForceDirectionComponent","name":"matchesForceDirectionComponent","line":103,"end_line":105,"hash":"cae4add1fd3b45d7b5796fb17954a54e13d45c0cae94c8739d73c1293118942a"},{"id":"func/createSelectedMasses","name":"createSelectedMasses","line":107,"end_line":120,"hash":"2b24cce5e13922fb32f2edc94544648e6115ce426dd621725585420f821b96d3"},{"id":"func/selectedMassIDs","name":"selectedMassIDs","line":122,"end_line":130,"hash":"01c2ede77b1479ddd668dc61689fcf1c7499feaba305626aa580ac70d7d1470d"},{"id":"func/setForceCenter","name":"setForceCenter","line":132,"end_line":135,"hash":"59a29f74135e456479c7a60805dec26cca63a79705c3705bd2cdf0bd3a9af89f"},{"id":"func/assertForceCenter","name":"assertForceCenter","line":137,"end_line":150,"hash":"767c22dcedbab08e9cc77012d655db5e7725bd9803d665a68e661c4d40e01075"},{"id":"func/createForceCenterMass","name":"createForceCenterMass","line":152,"end_line":162,"hash":"25d0d93584c5cb09789eb6ef7566557bd5b1a61906c396eaf3f1ed6e56948a89"},{"id":"func/enableNamedForce","name":"enableNamedForce","line":164,"end_line":174,"hash":"912ffe9b212d5bed347ce9b8957e243848f5c213885341a69c343b68be30bfaf"},{"id":"func/evaluateCenterForces","name":"evaluateCenterForces","line":176,"end_line":178,"hash":"b000259e6aeba7a9c54a29755ea153df4af47e780062d6cc3e833fb73a5f94f4"},{"id":"func/evaluateCurrentForces","name":"evaluateCurrentForces","line":180,"end_line":183,"hash":"4516cf35b79b1cc7a9dcb5c31b2c7168400f6d7ed83b40cf7c817e1b42f1bdb3"},{"id":"func/assertCenterMassVisuallyMarked","name":"assertCenterMassVisuallyMarked","line":185,"end_line":196,"hash":"d3dc4bd6d4a182c7c9b848f6c1332732b3dd2be6cce4171457f90c1b5117fcfc"},{"id":"func/assertNoReciprocalCenterForce","name":"assertNoReciprocalCenterForce","line":198,"end_line":211,"hash":"fe9a6917c7bd9637f2fd9d64a01dc97d5d9dd6e91c020922fc2eff963fb9dcd4"},{"id":"func/enableForceForControls","name":"enableForceForControls","line":213,"end_line":215,"hash":"3411246cb783497a2f6be551f9417e9ac28072c07cc64ebf7b88ed88aac1521f"},{"id":"func/assertForceControlsActive","name":"assertForceControlsActive","line":217,"end_line":230,"hash":"2d795092f78aa152304256a8bab6c50808f03e54d0a53a494a7fc99f8b9d94d3"}]}
// mutate4go-manifest-end
