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
	_ = game.World().AddMass(sim.Mass{})
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
			return requirePrerequisite(strings.HasPrefix(w.xspSavedFirst, "#1.0"), "world was not saved")
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
	return parameter, value, nil
}

// mutate4go-manifest-begin
// {"version":1,"tested_at":"2026-05-18T23:50:06-05:00","module_hash":"43695c965e7d69b6a15afddba2f42f9828bb3d5407af9f6082457afd2c9db7e0","functions":[{"id":"func/createRunningApplication","name":"createRunningApplication","line":17,"end_line":20,"hash":"d500dddad1cff2802e03e457677f1cc9784c4356820d1713f39f548d65bff06f"},{"id":"func/pressShortcut","name":"pressShortcut","line":22,"end_line":40,"hash":"5d258154687c7dd40311e6062870454d2a3380f32c5d72a4385f038b39d355ab"},{"id":"func/createControlWorldState","name":"createControlWorldState","line":42,"end_line":48,"hash":"b95eda9f0becdff674490539e6b15f33bf353699a6e63b3e99b3e13e6b049ab0"},{"id":"func/runFileCommand","name":"runFileCommand","line":50,"end_line":56,"hash":"06d4a0d030cbba550770e5a8381fa7b7d8f491c84717c6e8f42465c7ba9c7ab3"},{"id":"func/assertControlWorldState","name":"assertControlWorldState","line":58,"end_line":68,"hash":"43a5c292f29e574ef0772fd183fbcc175306dd3d837cdbc6450e2489ac1ebbd9"},{"id":"func/assertControlParameterResult","name":"assertControlParameterResult","line":70,"end_line":88,"hash":"19ee2946975f232c24785e92360ed1ce1c08c3239660fbd2c7d2c0a91ef40826"},{"id":"func/createWorldObjects","name":"createWorldObjects","line":90,"end_line":92,"hash":"0b62d7ee3189174f9f16f3c81a7c7eacc5ab1416efe267111dee701e2cae3274"},{"id":"func/setCustomSystemParameters","name":"setCustomSystemParameters","line":94,"end_line":101,"hash":"90f5e98c7637c67528bc4a9cfa9bdc1d9724ae5794d1c378add15fad30bb838f"},{"id":"func/runResetCommand","name":"runResetCommand","line":103,"end_line":110,"hash":"75a87d27b5aad2320f437f7828ac2cf42eaff3fb383c02a12d0e25b5272effec"},{"id":"func/assertControlMassCountZero","name":"assertControlMassCountZero","line":112,"end_line":114,"hash":"69e57823728772a0959667503b44d7d095415a185798252a855c87742bcabfd0"},{"id":"func/assertControlSpringCountZero","name":"assertControlSpringCountZero","line":116,"end_line":118,"hash":"11db5030304f785ced91213246b9a617a91a5376ef5777e8241ce5ae1560c684"},{"id":"func/assertControlParametersDefault","name":"assertControlParametersDefault","line":120,"end_line":129,"hash":"0b54e9d8ec3b01241e4020807c555addd9c06daa2e98bc4f8fba591218025f6f"},{"id":"func/setControlParameterValue","name":"setControlParameterValue","line":131,"end_line":142,"hash":"e52902d6aad986792409620e3845f382978fafd64c4e348ee000861717bcc0f2"},{"id":"func/changeControlParameterValue","name":"changeControlParameterValue","line":144,"end_line":155,"hash":"b831e32016d5e0193d8ec79a365376530a02fd8212a238f6494ce053a506e9b1"},{"id":"func/assertControlParameterValue","name":"assertControlParameterValue","line":157,"end_line":167,"hash":"63024b90438e35a568f20e171e203aeb02313151ca7e07d0351c150359ee3cd8"},{"id":"func/runNamedFileCommand","name":"runNamedFileCommand","line":169,"end_line":181,"hash":"e1627589912033f5cae06bc0a6b0f8d5eaaf1d7897570f42650ce6286bb8e09c"},{"id":"func/controlWorldStateAssertions","name":"controlWorldStateAssertions","line":183,"end_line":191,"hash":"d4f3f3f91d29c177a696c0a32b66e83a1586470323ddf17e943bef9e9081d71e"},{"id":"func/assertControlObjectCount","name":"assertControlObjectCount","line":193,"end_line":200,"hash":"255aa5b96703fe17125d65227a5492b549084776e35bbe1ba0ad12765d79ff7c"},{"id":"func/withConcreteGame","name":"withConcreteGame","line":202,"end_line":208,"hash":"db0d077e3a1a828ba2f390c5b0fbea9d75612834b576c34d42c131d232c83c88"},{"id":"func/concreteGame","name":"concreteGame","line":210,"end_line":216,"hash":"7059ca021b33ce86db033542d17d1c4948926f82e3a97f53c210d184d1ae0191"},{"id":"func/controlFileXSP","name":"controlFileXSP","line":218,"end_line":220,"hash":"a1fb19c37582b4576daf5c0cb32559dc7bc28bbe2cb32443d6f5cff8f5daaf40"},{"id":"func/assertLoadedControlWorld","name":"assertLoadedControlWorld","line":222,"end_line":230,"hash":"d95f1439535732b1feeba937b7d7f067fb2d06578e8ad40a546cda4b5dad3465"},{"id":"func/assertInsertedControlWorld","name":"assertInsertedControlWorld","line":232,"end_line":237,"hash":"51a54541601294a926fa9614b090ec8fb0879bbad7e0f17ffc22b36c51a9872d"},{"id":"func/assertControlMassPresence","name":"assertControlMassPresence","line":239,"end_line":245,"hash":"cdde28987fadedd019a29a24c90daefbe64b833247513d429c958887e3848ef8"},{"id":"func/assertControlParameter","name":"assertControlParameter","line":247,"end_line":252,"hash":"36af1956f2d6b483be33246a7d2817580aaf102a0fc911d28ac81d76c7157a47"},{"id":"func/controlParameterAndValue","name":"controlParameterAndValue","line":254,"end_line":260,"hash":"65b484d25ce58efd805913a04cbe4cd177a453317bb60a44486eeb39d2dc3ecf"}]}
// mutate4go-manifest-end
