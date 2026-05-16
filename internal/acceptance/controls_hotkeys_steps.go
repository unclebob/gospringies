package acceptance

import (
	"fmt"
	"strings"

	"springs/internal/app"
	"springs/internal/sim"
)

const (
	controlCurrentValue = "current"
	controlLoadedValue  = "loaded"
	controlCustomValue  = "custom"
)

func createRunningApplication(w *world, _ map[string]string) error {
	w.appGame = app.NewGame()
	return nil
}

func pressShortcut(w *world, example map[string]string) error {
	shortcut, err := stringValue(example, "shortcut")
	if err != nil {
		return err
	}
	command, err := stringValue(example, "command")
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
	w.appCommand = command
	return nil
}

func createControlWorldState(w *world, _ map[string]string) error {
	game := app.NewGame()
	_ = game.World().AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 1, Y: 2}, Mass: 1})
	game.World().Parameters.Set("current mass", controlCurrentValue)
	w.appGame = game
	return nil
}

func runFileCommand(w *world, example map[string]string) error {
	command, err := stringValue(example, "command")
	if err != nil {
		return err
	}
	return withConcreteGame(w, func(game *app.Game) error { return runNamedFileCommand(w, game, command) })
}

func assertControlWorldState(w *world, example map[string]string) error {
	state, err := stringValue(example, "expected_state")
	if err != nil {
		return err
	}
	assertState, ok := controlWorldStateAssertions(w)[state]
	if !ok {
		return fmt.Errorf("unsupported expected state %q", state)
	}
	return withConcreteGame(w, assertState)
}

func assertControlParameterResult(w *world, example map[string]string) error {
	result, err := stringValue(example, "parameter_result")
	if err != nil {
		return err
	}
	game, err := concreteGame(w)
	if err != nil {
		return err
	}
	expected := map[string]string{
		"unchanged":                 controlCurrentValue,
		"replaced by XSP file":      controlLoadedValue,
		"existing values preserved": controlCurrentValue,
	}[result]
	if expected == "" {
		return fmt.Errorf("unsupported parameter result %q", result)
	}
	return assertControlParameter(game, "current mass", expected)
}

func createWorldObjects(w *world, _ map[string]string) error {
	return createControlWorldState(w, nil)
}

func setCustomSystemParameters(w *world, _ map[string]string) error {
	game, err := concreteGame(w)
	if err != nil {
		return err
	}
	game.SetParameter("current mass", controlCustomValue)
	return nil
}

func runResetCommand(w *world, _ map[string]string) error {
	game, err := concreteGame(w)
	if err != nil {
		return err
	}
	game.RunCommand("reset")
	return nil
}

func assertControlMassCountZero(w *world, _ map[string]string) error {
	return assertControlObjectCount(w, "mass", func(game *app.Game) int { return len(game.World().Masses) })
}

func assertControlSpringCountZero(w *world, _ map[string]string) error {
	return assertControlObjectCount(w, "spring", func(game *app.Game) int { return len(game.World().Springs) })
}

func assertControlParametersDefault(w *world, _ map[string]string) error {
	game, err := concreteGame(w)
	if err != nil {
		return err
	}
	if game.World().Parameters.Value("current mass") != sim.DefaultParameters().Value("current mass") {
		return fmt.Errorf("parameters = %#v", game.World().Parameters)
	}
	return nil
}

func setControlParameterValue(w *world, example map[string]string) error {
	parameter, value, err := controlParameterAndValue(example, "old_value")
	if err != nil {
		return err
	}
	game := app.NewGame()
	if value != "default" {
		game.SetParameter(parameter, value)
	}
	w.appGame = game
	return nil
}

func changeControlParameterValue(w *world, example map[string]string) error {
	parameter, value, err := controlParameterAndValue(example, "new_value")
	if err != nil {
		return err
	}
	game, err := concreteGame(w)
	if err != nil {
		return err
	}
	game.SetParameter(parameter, value)
	return nil
}

func assertControlParameterValue(w *world, example map[string]string) error {
	parameter, value, err := controlParameterAndValue(example, "new_value")
	if err != nil {
		return err
	}
	game, err := concreteGame(w)
	if err != nil {
		return err
	}
	return assertControlParameter(game, parameter, value)
}

func runNamedFileCommand(w *world, game *app.Game, command string) error {
	switch command {
	case "save":
		w.xspSavedFirst = game.SaveXSP()
	case "load":
		return game.LoadXSP(controlFileXSP())
	case "insert":
		return game.InsertXSP(controlFileXSP())
	default:
		return fmt.Errorf("unsupported file command %q", command)
	}
	return nil
}

func controlWorldStateAssertions(w *world) map[string]func(*app.Game) error {
	return map[string]func(*app.Game) error{
		"written to XSP file": func(*app.Game) error {
			return requirePrerequisite(strings.HasPrefix(w.xspSavedFirst, "#1.0\n"), "world was not saved")
		},
		"replaced by XSP file":       assertLoadedControlWorld,
		"current plus inserted file": assertInsertedControlWorld,
	}
}

func assertControlObjectCount(w *world, objectType string, count func(*app.Game) int) error {
	return withConcreteGame(w, func(game *app.Game) error {
		if actual := count(game); actual != 0 {
			return fmt.Errorf("%s count = %d", objectType, actual)
		}
		return nil
	})
}

func withConcreteGame(w *world, action func(*app.Game) error) error {
	game, err := concreteGame(w)
	if err != nil {
		return err
	}
	return action(game)
}

func concreteGame(w *world) (*app.Game, error) {
	game, ok := w.appGame.(*app.Game)
	if !ok || game == nil {
		return nil, fmt.Errorf("application is not running")
	}
	return game, nil
}

func controlFileXSP() string {
	return "#1.0\ncmas " + controlLoadedValue + "\nmass 9 10 20 1 0\n"
}

func assertLoadedControlWorld(game *app.Game) error {
	if err := assertControlMassPresence(game, 9, true, "loaded mass missing"); err != nil {
		return err
	}
	if err := assertControlMassPresence(game, 1, false, "current mass was not replaced"); err != nil {
		return err
	}
	return nil
}

func assertInsertedControlWorld(game *app.Game) error {
	if err := assertControlMassPresence(game, 1, true, "current mass missing"); err != nil {
		return err
	}
	return assertControlMassPresence(game, 9, true, "inserted mass missing")
}

func assertControlMassPresence(game *app.Game, id int, expected bool, message string) error {
	_, ok := game.World().MassByID(id)
	if ok != expected {
		return fmt.Errorf("%s", message)
	}
	return nil
}

func assertControlParameter(game *app.Game, parameter string, expected string) error {
	if got := game.World().Parameters.Value(parameter); got != expected {
		return fmt.Errorf("parameter %q = %q, expected %q", parameter, got, expected)
	}
	return nil
}

func controlParameterAndValue(example map[string]string, valueKey string) (string, string, error) {
	parameter, value, err := stringPair(example, "parameter", valueKey)
	if err != nil {
		return "", "", err
	}
	if value == "custom" {
		value = controlCustomValue
	}
	return parameter, value, nil
}
