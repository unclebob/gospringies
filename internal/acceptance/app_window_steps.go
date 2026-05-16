package acceptance

import (
	"fmt"

	"springs/internal/app"
	"springs/internal/sim"
)

type appGame interface {
	Update() error
	RenderFrame()
	RenderWorld() renderResult
	World() *sim.Simulation
	SetPaused(bool)
	EditorScreen() editorScreen
	SetMode(string)
	SetSelected(bool)
	SetDirty(bool)
	HandleShortcut(string) bool
	LastCommand() string
	InputActive() bool
	RenderingActive() bool
	Close() error
	Closed() bool
}

func startDesktopApplication(w *world, _ map[string]string) error {
	w.appGame = app.NewGame()
	w.appErr = nil
	return nil
}

func assertApplicationWindowOpened(w *world, _ map[string]string) error {
	if w.appErr != nil {
		return w.appErr
	}
	if w.appGame == nil {
		return fmt.Errorf("application was not started")
	}
	return nil
}

func assertApplicationWorldEmpty(w *world, _ map[string]string) error {
	game, err := applicationGame(w)
	if err != nil {
		return err
	}
	if len(game.World().Masses) != 0 || len(game.World().Springs) != 0 {
		return fmt.Errorf("world is not empty")
	}
	return nil
}

func resizeApplicationWindow(w *world, example map[string]string) error {
	size, err := stringValue(example, "window_size")
	if err != nil {
		return err
	}
	if !supportedWindowSize(size) {
		return fmt.Errorf("unsupported window size %q", size)
	}
	if !app.DefaultWindowConfig().Resizable {
		return fmt.Errorf("window is not resizable")
	}
	w.appWindowSize = size
	w.appGame = app.NewGame()
	return nil
}

func supportedWindowSize(size string) bool {
	switch size {
	case "small", "large":
		return true
	default:
		return false
	}
}

func assertApplicationContinuesRunning(w *world, _ map[string]string) error {
	if w.appWindowSize == "" {
		return fmt.Errorf("window was not resized")
	}
	return assertApplicationWindowOpened(w, nil)
}

func setApplicationPauseState(w *world, example map[string]string) error {
	paused, err := boolValue(example, "paused")
	if err != nil {
		return err
	}
	game := app.NewGame()
	_ = game.World().AddMass(sim.Mass{ID: 1, Mass: 1})
	game.World().Parameters.EnableForce("gravity", map[string]string{"magnitude": "10", "direction": "90"})
	game.SetPaused(paused)
	w.appGame = game
	w.appBeforeTime = game.World().Time
	return nil
}

func updateApplicationLoop(w *world, _ map[string]string) error {
	game, err := applicationGame(w)
	if err != nil {
		return err
	}
	if err := game.Update(); err != nil {
		w.appErr = err
		return err
	}
	game.RenderFrame()
	return nil
}

func assertApplicationStepping(w *world, example map[string]string) error {
	game, err := applicationGame(w)
	if err != nil {
		return err
	}
	stepping, err := stringValue(example, "stepping")
	if err != nil {
		return err
	}
	expected, ok := expectedSteppingState(stepping)
	if !ok {
		return fmt.Errorf("unsupported stepping state %q", stepping)
	}
	if stepped := game.World().Time > w.appBeforeTime; stepped != expected {
		return fmt.Errorf("simulation stepping = %t, expected %t", stepped, expected)
	}
	return nil
}

func expectedSteppingState(stepping string) (bool, bool) {
	return booleanState(stepping, map[string]bool{"active": true, "stopped": false})
}

func assertApplicationInputActive(w *world, _ map[string]string) error {
	return assertApplicationActive(w, "input handling", appGame.InputActive)
}

func assertApplicationRenderingActive(w *world, _ map[string]string) error {
	return assertApplicationActive(w, "rendering", appGame.RenderingActive)
}

func assertApplicationActive(w *world, name string, active func(appGame) bool) error {
	game, err := applicationGame(w)
	if err != nil {
		return err
	}
	return requirePrerequisite(active(game), name+" was not active")
}

func closeApplicationWindow(w *world, _ map[string]string) error {
	game := app.NewGame()
	w.appGame = game
	w.appErr = game.Close()
	return nil
}

func assertApplicationExitClean(w *world, _ map[string]string) error {
	game, err := applicationGame(w)
	if err != nil {
		return err
	}
	if w.appErr != nil {
		return w.appErr
	}
	if !game.Closed() {
		return fmt.Errorf("application did not close")
	}
	return nil
}

func applicationGame(w *world) (appGame, error) {
	if w.appGame == nil {
		return nil, fmt.Errorf("application was not started")
	}
	return w.appGame, nil
}
