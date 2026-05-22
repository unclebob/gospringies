package acceptance

import (
	"fmt"

	"springs/internal/sim"
)

func startDesktopApplication(w *world, _ map[string]string) error {
	startApplicationDriver(w)
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
	if springOnlyWorld(game.World()) {
		return fmt.Errorf("world has springs without masses")
	}
	return nil
}

func springOnlyWorld(world *sim.Simulation) bool {
	return len(world.Masses) == 0 && len(world.Springs) != 0
}

func resizeApplicationWindow(w *world, example map[string]string) error {
	size, err := stringValue(example, "window_size")
	if err != nil {
		return err
	}
	if !supportedWindowSize(size) {
		return fmt.Errorf("unsupported window size %q", size)
	}
	if !applicationWindowResizable() {
		return fmt.Errorf("window is not resizable")
	}
	w.appWindowSize = size
	startApplicationDriver(w)
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
	game := newSteppingGame()
	game.SetPaused(paused)
	w.appGame = game
	w.appBeforeTime = game.World().Time
	return nil
}

func newSteppingGame() appGame {
	game := newApplicationDriverGame()
	game.World().Reset()
	_ = game.World().AddMass(sim.Mass{ID: 1, Mass: 1})
	game.World().Parameters.EnableForce("gravity", map[string]string{"magnitude": "10", "direction": "90"})
	return game
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
	game := startApplicationDriver(w)
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
	return applicationDriverGame(w)
}

// mutate4go-manifest-begin
// {"version":1,"tested_at":"2026-05-22T10:56:16-05:00","module_hash":"a18f9b5af4c64da3ffa2c8894e7197000baaaa522529d00b7431c51e493b30e7","functions":[{"id":"func/startDesktopApplication","name":"startDesktopApplication","line":9,"end_line":13,"hash":"ec40b9529be0af86266c7d5737873891c8d39d0c38ad02002529104e7f938f2f"},{"id":"func/assertApplicationWindowOpened","name":"assertApplicationWindowOpened","line":15,"end_line":23,"hash":"9b64d821c5b495c9e71a3a174c91fc7cebbffe8a38e61bf2e45abcf990fc850b"},{"id":"func/assertApplicationWorldEmpty","name":"assertApplicationWorldEmpty","line":25,"end_line":34,"hash":"451814f242ac954e0f5c16ca7f6cb8436036070a85b11b3c355c2f8bf911d6b6"},{"id":"func/springOnlyWorld","name":"springOnlyWorld","line":36,"end_line":38,"hash":"a09a3c8ea21822ba56ed373a1e116dad87ff11ce6f7bd22c47b3f3411b16b5ad"},{"id":"func/resizeApplicationWindow","name":"resizeApplicationWindow","line":40,"end_line":54,"hash":"273f988f0e08961cd389272d604467177a71e49abdf3738f951ff601ebc5f751"},{"id":"func/supportedWindowSize","name":"supportedWindowSize","line":56,"end_line":63,"hash":"00573cc1c859de5098f39acb820272613a49be6b9eb02e1b3967d499d653ced0"},{"id":"func/assertApplicationContinuesRunning","name":"assertApplicationContinuesRunning","line":65,"end_line":70,"hash":"976d0e76ac0e3b9bc296ad230789208023e7e5ed9ad57b2975f7fdf824edce77"},{"id":"func/setApplicationPauseState","name":"setApplicationPauseState","line":72,"end_line":82,"hash":"b24be17e764100619b8404006cfe32b75b88abd498913fe5f2354fa44911cbdf"},{"id":"func/newSteppingGame","name":"newSteppingGame","line":84,"end_line":90,"hash":"d798ecf59ac55d28240f7970964bc520c84f17f928834b0b2208e6f64b552c5e"},{"id":"func/updateApplicationLoop","name":"updateApplicationLoop","line":92,"end_line":103,"hash":"be4c4464550175eefdc11a37b2861410c94c55b15aab61c35def3f61d4a6c17e"},{"id":"func/assertApplicationStepping","name":"assertApplicationStepping","line":105,"end_line":122,"hash":"6d61617491003978f59916c74e2e531f8fdc4d357770d4d3a120ceed3ded9c53"},{"id":"func/expectedSteppingState","name":"expectedSteppingState","line":124,"end_line":126,"hash":"6cc9066eb3be622e189d442b4c9865e21be058dbab947b88a6b6992a8cee52f6"},{"id":"func/assertApplicationInputActive","name":"assertApplicationInputActive","line":128,"end_line":130,"hash":"85846d88e9335d8b127a4ddb7c855a95824e6d0d0d56142706ce4d20161eadfa"},{"id":"func/assertApplicationRenderingActive","name":"assertApplicationRenderingActive","line":132,"end_line":134,"hash":"339316815d458055f44b42899155b8940ba9435b3ee84007ec940a9388687564"},{"id":"func/assertApplicationActive","name":"assertApplicationActive","line":136,"end_line":142,"hash":"4a7f2f9bfddc67e9256d5acd6503ad6db6da03111bc2eac99e3051cd72337d97"},{"id":"func/closeApplicationWindow","name":"closeApplicationWindow","line":144,"end_line":148,"hash":"c0299a2802635761956953de932627b6c2f4f55f68b33b84799e90cd3cd63abd"},{"id":"func/assertApplicationExitClean","name":"assertApplicationExitClean","line":150,"end_line":162,"hash":"764438644b4b06d16722cc4f00266dc82598cf4479d52584d50c28cb65c932d9"},{"id":"func/applicationGame","name":"applicationGame","line":164,"end_line":166,"hash":"785b221154b1934f19ce3fc88de568fe566cde71b4041a1a67e4b309dc88ff74"}]}
// mutate4go-manifest-end
