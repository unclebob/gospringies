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
	SetSelected(bool)
	SetDirty(bool)
	HandleShortcut(string) bool
	LastCommand() string
	DrawFrameReport() app.DrawFrameReport
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
	game := newSteppingGame()
	game.SetPaused(paused)
	w.appGame = game
	w.appBeforeTime = game.World().Time
	return nil
}

func newSteppingGame() appGame {
	game := app.NewGame()
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

// mutate4go-manifest-begin
// {"version":1,"tested_at":"2026-05-18T22:17:35-05:00","module_hash":"ff168ddb2ef7a212027151b10f5247f47f44b33c7175a62ff148219249f83c6f","functions":[{"id":"func/startDesktopApplication","name":"startDesktopApplication","line":28,"end_line":32,"hash":"0590129dfd0aa8cdd4aa671a415b566db1e3427f31c7e717bd8435f33c5def6b"},{"id":"func/assertApplicationWindowOpened","name":"assertApplicationWindowOpened","line":34,"end_line":42,"hash":"9b64d821c5b495c9e71a3a174c91fc7cebbffe8a38e61bf2e45abcf990fc850b"},{"id":"func/assertApplicationWorldEmpty","name":"assertApplicationWorldEmpty","line":44,"end_line":53,"hash":"451814f242ac954e0f5c16ca7f6cb8436036070a85b11b3c355c2f8bf911d6b6"},{"id":"func/springOnlyWorld","name":"springOnlyWorld","line":55,"end_line":57,"hash":"a09a3c8ea21822ba56ed373a1e116dad87ff11ce6f7bd22c47b3f3411b16b5ad"},{"id":"func/resizeApplicationWindow","name":"resizeApplicationWindow","line":59,"end_line":73,"hash":"91d935ceaaf775007a77e9a47fb58c718465a5d71644bf55167c9922406c9ba2"},{"id":"func/supportedWindowSize","name":"supportedWindowSize","line":75,"end_line":82,"hash":"00573cc1c859de5098f39acb820272613a49be6b9eb02e1b3967d499d653ced0"},{"id":"func/assertApplicationContinuesRunning","name":"assertApplicationContinuesRunning","line":84,"end_line":89,"hash":"976d0e76ac0e3b9bc296ad230789208023e7e5ed9ad57b2975f7fdf824edce77"},{"id":"func/setApplicationPauseState","name":"setApplicationPauseState","line":91,"end_line":101,"hash":"b24be17e764100619b8404006cfe32b75b88abd498913fe5f2354fa44911cbdf"},{"id":"func/newSteppingGame","name":"newSteppingGame","line":103,"end_line":109,"hash":"0b243ad057c9e46fd186bce04765685b42f0f2d018dcbc305b8d96bfc25e65c4"},{"id":"func/updateApplicationLoop","name":"updateApplicationLoop","line":111,"end_line":122,"hash":"be4c4464550175eefdc11a37b2861410c94c55b15aab61c35def3f61d4a6c17e"},{"id":"func/assertApplicationStepping","name":"assertApplicationStepping","line":124,"end_line":141,"hash":"6d61617491003978f59916c74e2e531f8fdc4d357770d4d3a120ceed3ded9c53"},{"id":"func/expectedSteppingState","name":"expectedSteppingState","line":143,"end_line":145,"hash":"6cc9066eb3be622e189d442b4c9865e21be058dbab947b88a6b6992a8cee52f6"},{"id":"func/assertApplicationInputActive","name":"assertApplicationInputActive","line":147,"end_line":149,"hash":"85846d88e9335d8b127a4ddb7c855a95824e6d0d0d56142706ce4d20161eadfa"},{"id":"func/assertApplicationRenderingActive","name":"assertApplicationRenderingActive","line":151,"end_line":153,"hash":"339316815d458055f44b42899155b8940ba9435b3ee84007ec940a9388687564"},{"id":"func/assertApplicationActive","name":"assertApplicationActive","line":155,"end_line":161,"hash":"4a7f2f9bfddc67e9256d5acd6503ad6db6da03111bc2eac99e3051cd72337d97"},{"id":"func/closeApplicationWindow","name":"closeApplicationWindow","line":163,"end_line":168,"hash":"080ad7ee9bd809d060fea29d3c50dd324bf4b34acbcc4d0d8eb944957847cb55"},{"id":"func/assertApplicationExitClean","name":"assertApplicationExitClean","line":170,"end_line":182,"hash":"764438644b4b06d16722cc4f00266dc82598cf4479d52584d50c28cb65c932d9"},{"id":"func/applicationGame","name":"applicationGame","line":184,"end_line":189,"hash":"f041c8267d4373beea37ab7f534bc0e911e9ddc990f419f7e94a7d8410d2634f"}]}
// mutate4go-manifest-end
