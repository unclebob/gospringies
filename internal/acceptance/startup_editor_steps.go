package acceptance

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"springs/internal/app"
	xspfmt "springs/internal/format"
	"springs/internal/sim"
)

func assertStartupEditorChrome(w *world, _ map[string]string) error {
	game, err := concreteStartupGame(w)
	if err != nil {
		return err
	}
	screen := game.EditorScreen()
	if err := assertStartupRegions(screen); err != nil {
		return err
	}
	return requirePrerequisite(screen.Editor && screen.CanvasVisible && screen.ControlsUsable, "startup editor chrome was not visible")
}

func assertStartupRegions(screen editorScreen) error {
	for _, region := range startupRegions() {
		if _, ok := screen.RegionPurpose(region); !ok {
			return fmt.Errorf("startup region %q was not visible", region)
		}
	}
	return nil
}

func assertStartupWorldContent(w *world, _ map[string]string) error {
	game, err := concreteStartupGame(w)
	if err != nil {
		return err
	}
	result := game.RenderWorld()
	if !result.MassesVisible || !result.SpringLinesVisible {
		return fmt.Errorf("startup world content was not visible: %#v", result)
	}
	return nil
}

func assertDebugTextNotOnlyContent(w *world, example map[string]string) error {
	if err := assertStartupEditorChrome(w, example); err != nil {
		return err
	}
	return assertStartupWorldContent(w, example)
}

func assertStartupRegionVisible(w *world, example map[string]string) error {
	region, err := stringValue(example, "region")
	if err != nil {
		return err
	}
	game, err := concreteStartupGame(w)
	if err != nil {
		return err
	}
	if _, ok := game.EditorScreen().RegionPurpose(region); !ok {
		return fmt.Errorf("startup region %q was not visible", region)
	}
	return nil
}

func assertStartupObjectCount(w *world, example map[string]string) error {
	count, objectType, err := stringPair(example, "object_count", "object_type")
	if err != nil {
		return err
	}
	if count != "at least 1" {
		return fmt.Errorf("unsupported object count %q", count)
	}
	game, err := concreteStartupGame(w)
	if err != nil {
		return err
	}
	if startupObjectCount(game.World(), objectType) < 1 {
		return fmt.Errorf("startup world has no %s", objectType)
	}
	return nil
}

func assertStartupWorldLoadedFromDemo(w *world, example map[string]string) error {
	defaultDemo, err := stringValue(example, "default_demo")
	if err != nil {
		return err
	}
	if defaultDemo != app.DefaultStartupScenePath() {
		return fmt.Errorf("startup demo = %q, want %q", app.DefaultStartupScenePath(), defaultDemo)
	}
	_, err = startupDemoWorld(defaultDemo)
	return err
}

func assertStartupWorldMatchesDemo(w *world, example map[string]string) error {
	defaultDemo, err := stringValue(example, "default_demo")
	if err != nil {
		return err
	}
	game, err := concreteStartupGame(w)
	if err != nil {
		return err
	}
	expected, err := startupDemoWorld(defaultDemo)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(game.World(), expected) {
		return fmt.Errorf("startup world did not match %s", defaultDemo)
	}
	return nil
}

func startupDemoWorld(path string) (*sim.Simulation, error) {
	content, err := readStartupDemo(path)
	if err != nil {
		return nil, err
	}
	world, err := xspfmt.LoadXSP(string(content))
	if err != nil {
		return nil, fmt.Errorf("load startup demo %s: %w", path, err)
	}
	world.Bounds = newApplicationDriverWorld().Bounds
	return world, nil
}

func readStartupDemo(path string) ([]byte, error) {
	var lastErr error
	for _, candidate := range startupDemoCandidates(path) {
		content, err := os.ReadFile(candidate)
		if err == nil {
			return content, nil
		}
		lastErr = err
	}
	return nil, fmt.Errorf("read startup demo %s: %w", path, lastErr)
}

func startupDemoCandidates(path string) []string {
	return []string{
		path,
		filepath.Join("..", "..", path),
	}
}

func startupObjectCount(world *sim.Simulation, objectType string) int {
	switch objectType {
	case "fixed mass":
		return countMasses(world, true)
	case "movable mass":
		return countMasses(world, false)
	case "spring":
		return len(world.Springs)
	default:
		return 0
	}
}

func countMasses(world *sim.Simulation, fixed bool) int {
	count := 0
	for _, mass := range world.Masses {
		if mass.Fixed == fixed {
			count++
		}
	}
	return count
}

func startDesktopApplicationTwice(w *world, _ map[string]string) error {
	first := newApplicationDriverGame()
	second := newApplicationDriverGame()
	w.domainWorld = first.World().Clone()
	w.resultingWorld = second.World().Clone()
	w.editorScreen = first.EditorScreen()
	w.startupSecondScreen = second.EditorScreen()
	return nil
}

func assertStartupWorldsEquivalent(w *world, _ map[string]string) error {
	if !sameWorldState(w.domainWorld, w.resultingWorld) || len(w.domainWorld.Springs) != len(w.resultingWorld.Springs) {
		return fmt.Errorf("startup worlds differed")
	}
	return nil
}

func assertStartupScreensEquivalent(w *world, _ map[string]string) error {
	if !sameStringSlices(w.editorScreen.CommandControls, w.startupSecondScreen.CommandControls) ||
		len(w.editorScreen.Regions) != len(w.startupSecondScreen.Regions) {
		return fmt.Errorf("startup screens differed")
	}
	return nil
}

func sameStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func startupRegions() []string {
	return []string{"canvas", "left toolbar", "top bar", "right inspector"}
}

func concreteStartupGame(w *world) (*app.Game, error) {
	return concreteApplicationDriverWithMessage(w, "startup application was not started")
}
