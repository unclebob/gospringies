package acceptance

import (
	"fmt"
	"strings"

	xspfmt "springs/internal/format"
	"springs/internal/sim"
)

func createXSPInputWithMarker(w *world, example map[string]string) error {
	marker, err := stringValue(example, "marker")
	if err != nil {
		return err
	}
	if marker == "#1.0" {
		w.xspInput = "#1.0\n"
		return nil
	}
	if marker == "none" {
		w.xspInput = "mass 1 0 0 1 0.8\n"
		return nil
	}
	return fmt.Errorf("unsupported marker %q", marker)
}

func loadXSPInput(w *world, _ map[string]string) error {
	w.xspWorld, w.xspLoadErr = xspfmt.LoadXSP(w.xspInput)
	return nil
}

func assertXSPLoadResult(w *world, example map[string]string) error {
	result, err := stringValue(example, "result")
	if err != nil {
		return err
	}
	checks := map[string]func() error{
		"pass": func() error {
			if w.xspLoadErr != nil {
				return fmt.Errorf("expected load pass, got %v", w.xspLoadErr)
			}
			return nil
		},
		"fail": func() error {
			if w.xspLoadErr == nil {
				return fmt.Errorf("expected load failure")
			}
			return nil
		},
	}
	check, ok := checks[result]
	if !ok {
		return fmt.Errorf("unsupported load result %q", result)
	}
	return check()
}

func createXSPInputWithCommand(w *world, example map[string]string) error {
	input, err := xspInputForCommand(example["command"])
	if err != nil {
		return err
	}
	w.xspInput = input
	return nil
}

func xspInputForCommand(command string) (string, error) {
	lines := map[string]string{
		"cmas": "#1.0\ncmas 3.0\n",
		"elas": "#1.0\nelas 0.4\n",
		"kspr": "#1.0\nkspr 12.5\n",
		"kdmp": "#1.0\nkdmp 0.7\n",
		"frce": "#1.0\nfrce gravity true magnitude=10 direction=90\n",
		"wall": "#1.0\nwall left true\n",
		"mass": "#1.0\nmass 1 10 20 3.0 0.8\n",
		"spng": "#1.0\nmass 1 0 0 1 0.8\nmass 2 10 0 1 0.8\nspng 1 1 2 10 12 0.7\n",
	}
	if input, ok := lines[command]; ok {
		return input, nil
	}
	return "", fmt.Errorf("unsupported command %q", command)
}

func assertXSPLoadedState(w *world, example map[string]string) error {
	state, err := stringValue(example, "loaded_state")
	if err != nil {
		return err
	}
	if w.xspLoadErr != nil {
		return w.xspLoadErr
	}
	return assertXSPState(w.xspWorld, state)
}

func assertXSPState(world *sim.Simulation, state string) error {
	checks := map[string]func(*sim.Simulation) error{
		"current mass":        func(w *sim.Simulation) error { return assertParameterValue(w, "current mass", "3.0") },
		"current elasticity":  func(w *sim.Simulation) error { return assertParameterValue(w, "elasticity", "0.4") },
		"current spring k":    func(w *sim.Simulation) error { return assertParameterValue(w, "spring constant", "12.5") },
		"current damping":     func(w *sim.Simulation) error { return assertParameterValue(w, "damping", "0.7") },
		"force configuration": assertForceLoaded,
		"wall configuration":  assertWallLoaded,
		"mass":                assertMassLoaded,
		"spring":              assertSpringLoaded,
	}
	check, ok := checks[state]
	if !ok {
		return fmt.Errorf("unsupported loaded state %q", state)
	}
	return check(world)
}

func assertForceLoaded(world *sim.Simulation) error {
	force, _ := world.Parameters.Force("gravity")
	if force.Enabled != "true" || force.Values["magnitude"] != "10" {
		return fmt.Errorf("gravity force = %#v", force)
	}
	return nil
}

func assertWallLoaded(world *sim.Simulation) error {
	if enabled, _ := world.Parameters.WallEnabled("left"); !enabled {
		return fmt.Errorf("left wall was not enabled")
	}
	return nil
}

func assertMassLoaded(world *sim.Simulation) error {
	if mass, ok := world.MassByID(1); !ok || mass.Position != (sim.Vec2{X: 10, Y: 20}) {
		return fmt.Errorf("mass 1 = %#v, %t", mass, ok)
	}
	return nil
}

func assertSpringLoaded(world *sim.Simulation) error {
	if spring, ok := world.SpringByID(1); !ok || spring.MassA != 1 || spring.MassB != 2 {
		return fmt.Errorf("spring 1 = %#v, %t", spring, ok)
	}
	return nil
}

func createWorldLoadedFromFile(w *world, example map[string]string) error {
	name, err := stringValue(example, "input_file")
	if err != nil {
		return err
	}
	if name != "simple scene" {
		return fmt.Errorf("unsupported input file %q", name)
	}
	w.xspWorld, w.xspLoadErr = xspfmt.LoadXSP(simpleSceneXSP())
	return w.xspLoadErr
}

func saveXSPWorldTwice(w *world, _ map[string]string) error {
	w.xspSavedFirst = xspfmt.SaveXSP(w.xspWorld)
	w.xspSavedSecond = xspfmt.SaveXSP(w.xspWorld)
	return nil
}

func assertXSPSavesIdentical(w *world, _ map[string]string) error {
	if w.xspSavedFirst != w.xspSavedSecond {
		return fmt.Errorf("saved outputs differ")
	}
	return nil
}

func assertXSPSaveEndsWithNewline(w *world, _ map[string]string) error {
	if !strings.HasSuffix(w.xspSavedFirst, "\n") {
		return fmt.Errorf("saved output missing final newline")
	}
	return nil
}

func createXSPInputWithFileMass(w *world, example map[string]string) error {
	id, value, err := stringPair(example, "mass_id", "file_mass_value")
	if err != nil {
		return err
	}
	w.xspInput = fmt.Sprintf("#1.0\nmass %s 10 20 %s 0.8\n", id, value)
	return nil
}

func loadAndSaveXSPInput(w *world, _ map[string]string) error {
	if err := loadXSPInput(w, nil); err != nil {
		return err
	}
	if w.xspLoadErr != nil {
		return w.xspLoadErr
	}
	w.xspSavedFirst = xspfmt.SaveXSP(w.xspWorld)
	return nil
}

func assertXSPMassFixedState(w *world, example map[string]string) error {
	id, err := intValue(example, "mass_id")
	if err != nil {
		return err
	}
	expected, err := boolValue(example, "fixed")
	if err != nil {
		return err
	}
	mass, ok := w.xspWorld.MassByID(id)
	if !ok {
		return fmt.Errorf("mass %d not found", id)
	}
	if mass.Fixed != expected {
		return fmt.Errorf("mass %d fixed = %t, expected %t", id, mass.Fixed, expected)
	}
	return nil
}

func assertSavedMassSign(w *world, example map[string]string) error {
	id, sign, err := stringPair(example, "mass_id", "file_mass_sign")
	if err != nil {
		return err
	}
	for _, line := range strings.Split(w.xspSavedFirst, "\n") {
		if strings.HasPrefix(line, "mass "+id+" ") {
			return assertFileMassSign(line, sign)
		}
	}
	return fmt.Errorf("saved mass %s not found", id)
}

func assertFileMassSign(line, sign string) error {
	fields := strings.Fields(line)
	if len(fields) < 5 {
		return fmt.Errorf("malformed saved mass line %q", line)
	}
	expectedNegative, ok := map[string]bool{"negative": true, "positive": false}[sign]
	if !ok {
		return fmt.Errorf("unsupported file mass sign %q", sign)
	}
	isNegative := strings.HasPrefix(fields[4], "-")
	if isNegative != expectedNegative {
		return fmt.Errorf("file mass sign was not %s: %q", sign, line)
	}
	return nil
}

func createMalformedXSPInput(w *world, example map[string]string) error {
	problem, err := stringValue(example, "problem")
	if err != nil {
		return err
	}
	inputs := map[string]string{
		"duplicate mass id":       "#1.0\nmass 1 0 0 1 0.8\nmass 1 1 1 1 0.8\n",
		"duplicate spring id":     "#1.0\nmass 1 0 0 1 0.8\nmass 2 1 1 1 0.8\nspng 1 1 2 1 1 0\nspng 1 1 2 1 1 0\n",
		"missing spring endpoint": "#1.0\nmass 1 0 0 1 0.8\nspng 1 1 2 1 1 0\n",
		"missing final newline":   "#1.0",
	}
	if input, ok := inputs[problem]; ok {
		w.xspInput = input
		return nil
	}
	return fmt.Errorf("unsupported problem %q", problem)
}

func assertXSPLoadErrorReason(w *world, example map[string]string) error {
	reason, err := stringValue(example, "reason")
	if err != nil {
		return err
	}
	if w.xspLoadErr == nil {
		return fmt.Errorf("expected load error")
	}
	if !strings.Contains(w.xspLoadErr.Error(), reason) {
		return fmt.Errorf("load error %q does not contain %q", w.xspLoadErr, reason)
	}
	return nil
}

func simpleSceneXSP() string {
	return "#1.0\ncmas 1.5\nelas 0.8\nkspr 12\nkdmp 0.7\nmass 1 0 0 1 0.8\nmass 2 10 0 1 0.8\nspng 1 1 2 10 12 0.7\n"
}
