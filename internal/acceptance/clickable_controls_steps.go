package acceptance

import (
	"fmt"

	"springs/internal/app"
)

func setClickableEditorMode(w *world, example map[string]string) error {
	mode, err := stringValue(example, "old_mode")
	if err != nil {
		return err
	}
	if !supportedClickableMode(mode) {
		return fmt.Errorf("unsupported editor mode %q", mode)
	}
	game, err := visibleControlsGame(w)
	if err != nil {
		return err
	}
	game.SetMode(mode)
	return nil
}

func supportedClickableMode(mode string) bool {
	switch mode {
	case "select", "add mass", "add spring", "drag":
		return true
	default:
		return false
	}
}

func clickInsideRenderedVisibleControlBounds(w *world, example map[string]string) error {
	control, err := stringValue(example, "control")
	if err != nil {
		return err
	}
	return clickInsideRenderedBoundsOfControl(w, control)
}

func clickInsideRenderedControl(control string) stepHandler {
	return func(w *world, _ map[string]string) error {
		return clickInsideRenderedBoundsOfControl(w, control)
	}
}

func clickInsideRenderedBoundsOfControl(w *world, control string) error {
	game, err := visibleControlsGame(w)
	if err != nil {
		return err
	}
	rect, ok := game.VisibleControlBounds(control)
	if !ok {
		return fmt.Errorf("visible control %q does not have rendered bounds", control)
	}
	x := rect.Min.X + rect.Dx()/2
	y := rect.Min.Y + rect.Dy()/2
	if !game.ClickAt(x, y) {
		return fmt.Errorf("click inside rendered bounds of visible control %q was not handled", control)
	}
	w.appCommand = game.LastCommand()
	return nil
}

func assertClickableEditorMode(w *world, example map[string]string) error {
	return assertClickableGameValue(w, example, "new_mode", "editor mode", clickableEditorMode)
}

func clickableEditorMode(game *app.Game) string { return game.Mode() }

func assertVisibleControlActive(w *world, example map[string]string) error {
	control, err := stringValue(example, "control")
	if err != nil {
		return err
	}
	game, err := visibleControlsGame(w)
	if err != nil {
		return err
	}
	if !game.DrawFrameReport().ActiveControls[control] {
		return fmt.Errorf("visible control %q did not show active state", control)
	}
	return nil
}

func assertKeyboardPathEntryOpen(w *world, example map[string]string) error {
	return assertClickableGameValue(w, example, "command", "path entry", clickablePathEntryCommand)
}

func clickablePathEntryCommand(game *app.Game) string { return game.PathEntryCommand() }

func assertDemoPickerOpen(w *world, _ map[string]string) error {
	game, err := visibleControlsGame(w)
	if err != nil {
		return err
	}
	if !game.DemoPickerOpen() {
		return fmt.Errorf("demo picker was not open")
	}
	return nil
}

func assertClickableGameValue(w *world, example map[string]string, key string, name string, actual func(*app.Game) string) error {
	expected, err := stringValue(example, key)
	if err != nil {
		return err
	}
	game, err := visibleControlsGame(w)
	if err != nil {
		return err
	}
	if got := actual(game); got != expected {
		return fmt.Errorf("%s = %q, want %q", name, got, expected)
	}
	return nil
}

func recordVisibleControlShortcut(w *world, example map[string]string) error {
	control, shortcut, err := stringPair(example, "control", "shortcut")
	if err != nil {
		return err
	}
	if _, ok := appControlWithLabel(control); !ok {
		return fmt.Errorf("visible control %q does not exist", control)
	}
	w.clickShortcut = shortcut
	w.appGame = app.NewGame()
	return nil
}

func assertClickMatchesShortcut(w *world, example map[string]string) error {
	shortcut, err := stringValue(example, "shortcut")
	if err != nil {
		return err
	}
	if err := assertRecordedShortcut(w, shortcut); err != nil {
		return err
	}
	clicked, err := visibleControlsGame(w)
	if err != nil {
		return err
	}
	clickedState := clickableApplicationState(clicked)
	shortcutState, err := shortcutApplicationState(shortcut)
	if err != nil {
		return err
	}
	return assertSameClickableApplicationState(clickedState, shortcutState)
}

func assertRecordedShortcut(w *world, shortcut string) error {
	if w.clickShortcut != shortcut {
		return fmt.Errorf("recorded shortcut = %q, want %q", w.clickShortcut, shortcut)
	}
	return nil
}

func shortcutApplicationState(shortcut string) (string, error) {
	game := app.NewGame()
	if !game.HandleShortcut(shortcut) {
		return "", fmt.Errorf("shortcut %q was not handled", shortcut)
	}
	return clickableApplicationState(game), nil
}

func assertSameClickableApplicationState(clickedState string, shortcutState string) error {
	if clickedState != shortcutState {
		return fmt.Errorf("click state = %q, shortcut state = %q", clickedState, shortcutState)
	}
	return nil
}

func recordClickableApplicationState(w *world, _ map[string]string) error {
	game := app.NewGame()
	w.appGame = game
	w.recordedAppState = clickableApplicationState(game)
	return nil
}

func clickOutsideVisibleControls(w *world, _ map[string]string) error {
	game, err := visibleControlsGame(w)
	if err != nil {
		return err
	}
	if game.ClickAt(500, 300) {
		return fmt.Errorf("outside click was handled")
	}
	return nil
}

func assertClickableApplicationStateUnchanged(w *world, _ map[string]string) error {
	game, err := visibleControlsGame(w)
	if err != nil {
		return err
	}
	if got := clickableApplicationState(game); got != w.recordedAppState {
		return fmt.Errorf("application state = %q, want %q", got, w.recordedAppState)
	}
	return nil
}

func setClickableSimulationState(w *world, example map[string]string) error {
	state, err := stringValue(example, "old_state")
	if err != nil {
		return err
	}
	paused, ok := simulationPausedState(state)
	if !ok {
		return fmt.Errorf("unsupported simulation state %q", state)
	}
	game, err := visibleControlsGame(w)
	if err != nil {
		return err
	}
	game.SetPaused(paused)
	return nil
}

func assertClickableSimulationState(w *world, example map[string]string) error {
	state, err := stringValue(example, "new_state")
	if err != nil {
		return err
	}
	paused, ok := simulationPausedState(state)
	if !ok {
		return fmt.Errorf("unsupported simulation state %q", state)
	}
	game, err := visibleControlsGame(w)
	if err != nil {
		return err
	}
	if game.Paused() != paused {
		return fmt.Errorf("paused = %t, want %t for state %q", game.Paused(), paused, state)
	}
	return nil
}

func clickableApplicationState(game *app.Game) string {
	screen := game.EditorScreen()
	return fmt.Sprintf(
		"mode=%s paused=%t command=%s path=%s closed=%t file=%s counts=%d/%d",
		game.Mode(),
		game.Paused(),
		game.LastCommand(),
		game.PathEntryCommand(),
		game.Closed(),
		screen.Indicators["file state"],
		len(game.World().Masses),
		len(game.World().Springs),
	)
}

func appControlWithLabel(label string) (app.DrawFrameReport, bool) {
	report := app.NewGame().DrawFrameReport()
	for _, controlLabel := range report.Controls {
		if controlLabel == label {
			return report, true
		}
	}
	return report, false
}
