package acceptance

import (
	"fmt"

	"springs/internal/app"
)

type editorScreen = app.EditorScreen

var applicationStateChanges = map[string]func(appGame){
	"select mode":     func(game appGame) { game.SetMode("select") },
	"paused":          func(game appGame) { game.SetPaused(true) },
	"running":         func(game appGame) { game.SetPaused(false) },
	"object selected": func(game appGame) { game.SetSelected(true) },
	"unsaved changes": func(game appGame) { game.SetDirty(true) },
}

func assertFirstScreenEditor(w *world, _ map[string]string) error {
	return assertCurrentScreen(w, func(screen editorScreen) bool { return screen.Editor }, "first screen was not the simulation editor")
}

func assertNoLandingPage(w *world, _ map[string]string) error {
	return assertCurrentScreen(w, func(screen editorScreen) bool { return !screen.LandingPage }, "first screen showed a landing page")
}

func layoutEditorScreen(w *world, _ map[string]string) error {
	return refreshEditorScreen(w)
}

func assertScreenRegionVisible(w *world, example map[string]string) error {
	region, err := stringValue(example, "region")
	if err != nil {
		return err
	}
	_, ok := w.editorScreen.RegionPurpose(region)
	return requirePrerequisite(ok, fmt.Sprintf("screen region %q was not visible", region))
}

func assertScreenRegionPurpose(w *world, example map[string]string) error {
	region, err := stringValue(example, "region")
	if err != nil {
		return err
	}
	expected, err := stringValue(example, "purpose")
	if err != nil {
		return err
	}
	actual, ok := w.editorScreen.RegionPurpose(region)
	if !ok || actual != expected {
		return fmt.Errorf("screen region %q purpose = %q, expected %q", region, actual, expected)
	}
	return nil
}

func assertModeVisible(w *world, example map[string]string) error {
	return assertVisibleControl(w, example, "mode", "mode", editorScreen.HasModeControl)
}

func assertCommandVisible(w *world, example map[string]string) error {
	return assertVisibleControl(w, example, "command", "command", editorScreen.HasCommandControl)
}

func setApplicationState(w *world, example map[string]string) error {
	state, err := stringValue(example, "state")
	if err != nil {
		return err
	}
	change, ok := applicationStateChanges[state]
	if !ok {
		return fmt.Errorf("unsupported application state %q", state)
	}
	return updateApplicationGame(w, change)
}

func assertVisibleIndicator(w *world, example map[string]string) error {
	indicator, err := stringValue(example, "indicator")
	if err != nil {
		return err
	}
	state, err := stringValue(example, "state")
	if err != nil {
		return err
	}
	if actual := w.editorScreen.Indicators[indicator]; actual != state {
		return fmt.Errorf("indicator %q = %q, expected %q", indicator, actual, state)
	}
	return nil
}

func setVisibleCommandControl(w *world, example map[string]string) error {
	command, err := stringValue(example, "command")
	if err != nil {
		return err
	}
	control, err := stringValue(example, "control")
	if err != nil {
		return err
	}
	game, err := ensureApplicationGame(w)
	if err != nil {
		return err
	}
	if !game.EditorScreen().HasCommandControl(control) {
		return fmt.Errorf("control %q was not visible", control)
	}
	w.appCommand = command
	return nil
}

func pressKeyboardShortcut(w *world, example map[string]string) error {
	shortcut, err := stringValue(example, "shortcut")
	if err != nil {
		return err
	}
	game, err := ensureApplicationGame(w)
	if err != nil {
		return err
	}
	if !game.HandleShortcut(shortcut) {
		return fmt.Errorf("shortcut %q was not handled", shortcut)
	}
	return nil
}

func assertCommandRan(w *world, example map[string]string) error {
	command, err := stringValue(example, "command")
	if err != nil {
		return err
	}
	game, err := applicationGame(w)
	if err != nil {
		return err
	}
	if game.LastCommand() != command || w.appCommand != command {
		return fmt.Errorf("command ran = %q, queued = %q, expected %q", game.LastCommand(), w.appCommand, command)
	}
	return nil
}

func setSimulationState(w *world, example map[string]string) error {
	state, err := stringValue(example, "simulation_state")
	if err != nil {
		return err
	}
	paused, ok := simulationPausedState(state)
	if !ok {
		return fmt.Errorf("unsupported simulation state %q", state)
	}
	return updateApplicationGame(w, func(game appGame) { game.SetPaused(paused) })
}

func simulationPausedState(state string) (bool, bool) {
	return booleanState(state, map[string]bool{"paused": true, "running": false})
}

func assertCanvasVisible(w *world, _ map[string]string) error {
	return requirePrerequisite(w.editorScreen.CanvasVisible, "canvas was not visible")
}

func assertControlsUsable(w *world, _ map[string]string) error {
	return requirePrerequisite(w.editorScreen.ControlsUsable, "controls were not usable")
}

func currentEditorScreen(w *world) (editorScreen, error) {
	game, err := applicationGame(w)
	if err != nil {
		return editorScreen{}, err
	}
	screen := game.EditorScreen()
	w.editorScreen = screen
	return screen, nil
}

func ensureApplicationGame(w *world) (appGame, error) {
	if w.appGame == nil {
		w.appGame = app.NewGame()
	}
	return applicationGame(w)
}

func refreshEditorScreen(w *world) error {
	game, err := ensureApplicationGame(w)
	if err != nil {
		return err
	}
	w.editorScreen = game.EditorScreen()
	return nil
}

func updateApplicationGame(w *world, update func(appGame)) error {
	game, err := ensureApplicationGame(w)
	if err != nil {
		return err
	}
	update(game)
	return nil
}

func assertCurrentScreen(w *world, matches func(editorScreen) bool, failure string) error {
	screen, err := currentEditorScreen(w)
	if err != nil {
		return err
	}
	return requirePrerequisite(matches(screen), failure)
}

func assertVisibleControl(
	w *world,
	example map[string]string,
	key string,
	controlType string,
	hasControl func(editorScreen, string) bool,
) error {
	control, err := stringValue(example, key)
	if err != nil {
		return err
	}
	message := fmt.Sprintf("%s %q was not visible", controlType, control)
	return requirePrerequisite(hasControl(w.editorScreen, control), message)
}
