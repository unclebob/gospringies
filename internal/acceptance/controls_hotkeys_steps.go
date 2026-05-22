package acceptance

import (
	"fmt"
	"strings"

	"springs/internal/sim"
)

const (
	controlCurrentValue = "current"
	controlLoadedValue  = "loaded"
	controlCustomValue  = "custom"
)

func createRunningApplication(w *world, _ map[string]string) error {
	startApplicationDriver(w)
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
	game := newApplicationDriverGame()
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
	return withConcreteGame(w, func(game *driverGame) error { return runNamedFileCommand(w, game, command) })
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
	return assertControlObjectCount(w, "mass", func(game *driverGame) int { return len(game.World().Masses) })
}

func assertControlSpringCountZero(w *world, _ map[string]string) error {
	return assertControlObjectCount(w, "spring", func(game *driverGame) int { return len(game.World().Springs) })
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
	game := newApplicationDriverGame()
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

func runNamedFileCommand(w *world, game *driverGame, command string) error {
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

func controlWorldStateAssertions(w *world) map[string]func(*driverGame) error {
	return map[string]func(*driverGame) error{
		"written to XSP file": func(*driverGame) error {
			return requirePrerequisite(strings.HasPrefix(w.xspSavedFirst, "#1.0"), "world was not saved")
		},
		"replaced by XSP file":       assertLoadedControlWorld,
		"current plus inserted file": assertInsertedControlWorld,
	}
}

func assertControlObjectCount(w *world, objectType string, count func(*driverGame) int) error {
	return withConcreteGame(w, func(game *driverGame) error {
		if actual := count(game); actual != 0 {
			return fmt.Errorf("%s count = %d", objectType, actual)
		}
		return nil
	})
}

func withConcreteGame(w *world, action func(*driverGame) error) error {
	game, err := concreteGame(w)
	if err != nil {
		return err
	}
	return action(game)
}

func concreteGame(w *world) (*driverGame, error) {
	return concreteApplicationDriver(w)
}

func controlFileXSP() string {
	return "#1.0\ncmas " + controlLoadedValue + "\nmass 9 10 20 1 0\n"
}

func assertLoadedControlWorld(game *driverGame) error {
	if err := assertControlMassPresence(game, 9, true, "loaded mass missing"); err != nil {
		return err
	}
	if err := assertControlMassPresence(game, 1, false, "current mass was not replaced"); err != nil {
		return err
	}
	return nil
}

func assertInsertedControlWorld(game *driverGame) error {
	if err := assertControlMassPresence(game, 1, true, "current mass missing"); err != nil {
		return err
	}
	return assertControlMassPresence(game, 9, true, "inserted mass missing")
}

func assertControlMassPresence(game *driverGame, id int, expected bool, message string) error {
	_, ok := game.World().MassByID(id)
	if ok != expected {
		return fmt.Errorf("%s", message)
	}
	return nil
}

func assertControlParameter(game *driverGame, parameter string, expected string) error {
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
// {"version":1,"tested_at":"2026-05-22T10:56:50-05:00","module_hash":"5170377bef9c96fcbd808249a7780710c6c9708d2566a84662f14b511e9f5b59","functions":[{"id":"func/createRunningApplication","name":"createRunningApplication","line":16,"end_line":19,"hash":"d3c22d5a44fe73dc97d98828222ceadd34420b6cb0619cf73cdf11d3e2990e35"},{"id":"func/pressShortcut","name":"pressShortcut","line":21,"end_line":39,"hash":"5d258154687c7dd40311e6062870454d2a3380f32c5d72a4385f038b39d355ab"},{"id":"func/createControlWorldState","name":"createControlWorldState","line":41,"end_line":47,"hash":"0cdff387fae2c4ea7ad7e3f00403e1903b7e7613d164677900e08c097f5701f8"},{"id":"func/runFileCommand","name":"runFileCommand","line":49,"end_line":55,"hash":"7df69534108431fd1b7f51f67fd97999759d6518d3e138cd56d2d2db09313308"},{"id":"func/assertControlWorldState","name":"assertControlWorldState","line":57,"end_line":67,"hash":"43a5c292f29e574ef0772fd183fbcc175306dd3d837cdbc6450e2489ac1ebbd9"},{"id":"func/assertControlParameterResult","name":"assertControlParameterResult","line":69,"end_line":87,"hash":"19ee2946975f232c24785e92360ed1ce1c08c3239660fbd2c7d2c0a91ef40826"},{"id":"func/createWorldObjects","name":"createWorldObjects","line":89,"end_line":91,"hash":"0b62d7ee3189174f9f16f3c81a7c7eacc5ab1416efe267111dee701e2cae3274"},{"id":"func/setCustomSystemParameters","name":"setCustomSystemParameters","line":93,"end_line":100,"hash":"90f5e98c7637c67528bc4a9cfa9bdc1d9724ae5794d1c378add15fad30bb838f"},{"id":"func/runResetCommand","name":"runResetCommand","line":102,"end_line":109,"hash":"75a87d27b5aad2320f437f7828ac2cf42eaff3fb383c02a12d0e25b5272effec"},{"id":"func/assertControlMassCountZero","name":"assertControlMassCountZero","line":111,"end_line":113,"hash":"2c61eab2bb7a86de8379bc1f3b1452666a656a305bc5fc7dcf098d12f55ab52d"},{"id":"func/assertControlSpringCountZero","name":"assertControlSpringCountZero","line":115,"end_line":117,"hash":"97fae711d140180322e8ffd142d160fecd6c3fca70fd2930fa54bff9afe56dca"},{"id":"func/assertControlParametersDefault","name":"assertControlParametersDefault","line":119,"end_line":128,"hash":"0b54e9d8ec3b01241e4020807c555addd9c06daa2e98bc4f8fba591218025f6f"},{"id":"func/setControlParameterValue","name":"setControlParameterValue","line":130,"end_line":141,"hash":"e31fc31d62da4a646c7654f508ac802dfc0ad64543719277f16ab042b3fe9a7d"},{"id":"func/changeControlParameterValue","name":"changeControlParameterValue","line":143,"end_line":154,"hash":"b831e32016d5e0193d8ec79a365376530a02fd8212a238f6494ce053a506e9b1"},{"id":"func/assertControlParameterValue","name":"assertControlParameterValue","line":156,"end_line":166,"hash":"63024b90438e35a568f20e171e203aeb02313151ca7e07d0351c150359ee3cd8"},{"id":"func/runNamedFileCommand","name":"runNamedFileCommand","line":168,"end_line":180,"hash":"856035642bf5d78fca03d54f320bfabfa6ffb99bdbd5e29f996f2d00b23d7eb3"},{"id":"func/controlWorldStateAssertions","name":"controlWorldStateAssertions","line":182,"end_line":190,"hash":"1d7dd8bf39fe6e533b7bb0a48fa16175b3c21eb56ef34734cae8b4e0f2964ff0"},{"id":"func/assertControlObjectCount","name":"assertControlObjectCount","line":192,"end_line":199,"hash":"c20b2533a1bdfc72f934abe6e39d13a993a41484d42cba13274c0865107a7bd4"},{"id":"func/withConcreteGame","name":"withConcreteGame","line":201,"end_line":207,"hash":"9bbebebf462310ea21e942116d9b4269ca990ca0eb256fe91ce4a6fe8dbdbc2c"},{"id":"func/concreteGame","name":"concreteGame","line":209,"end_line":211,"hash":"49d960a45d816efcd706879ede55f1c5cdb82eab03e65b48228d24a91670ced4"},{"id":"func/controlFileXSP","name":"controlFileXSP","line":213,"end_line":215,"hash":"a1fb19c37582b4576daf5c0cb32559dc7bc28bbe2cb32443d6f5cff8f5daaf40"},{"id":"func/assertLoadedControlWorld","name":"assertLoadedControlWorld","line":217,"end_line":225,"hash":"fc4c3413499ab073df1b106e5ba2398725b83c1d618a3133ce49d1a568debd0b"},{"id":"func/assertInsertedControlWorld","name":"assertInsertedControlWorld","line":227,"end_line":232,"hash":"c29824abbca54d0d00520dbd8f2a877ae2bfd16e70b149f6031e7e1e398e4b48"},{"id":"func/assertControlMassPresence","name":"assertControlMassPresence","line":234,"end_line":240,"hash":"a2b6deceed370342015e6c7731c0dec5b0e7a3485d15140ce5a8d1a876de8fe7"},{"id":"func/assertControlParameter","name":"assertControlParameter","line":242,"end_line":247,"hash":"ac5854d1daa18fc38e288a356207c9823f2fe3fe121fe73d3cbce8a59cb2baaf"},{"id":"func/controlParameterAndValue","name":"controlParameterAndValue","line":249,"end_line":255,"hash":"65b484d25ce58efd805913a04cbe4cd177a453317bb60a44486eeb39d2dc3ecf"}]}
// mutate4go-manifest-end
