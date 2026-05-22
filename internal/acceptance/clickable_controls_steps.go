package acceptance

import (
	"fmt"

	"springs/internal/app"
)

func setClickableEditorMode(w *world, example map[string]string) error {
	return fmt.Errorf("editor modes were removed from the app")
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
	return fmt.Errorf("editor modes were removed from the app")
}

func clickableEditorMode(game *app.Game) string { return "" }

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
	startApplicationDriver(w)
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
	game := newApplicationDriverGame()
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
	game := startApplicationDriver(w)
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
		"paused=%t command=%s path=%s closed=%t file=%s counts=%d/%d",
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
	report := newApplicationDriverGame().DrawFrameReport()
	for _, controlLabel := range report.Controls {
		if controlLabel == label {
			return report, true
		}
	}
	return report, false
}

// mutate4go-manifest-begin
// {"version":1,"tested_at":"2026-05-18T22:24:12-05:00","module_hash":"157d64bc9ac2b6c4996202f6ed40757507811313fb070906bc2d68b0b2ae9cb8","functions":[{"id":"func/setClickableEditorMode","name":"setClickableEditorMode","line":9,"end_line":11,"hash":"64ab03d88b9be5441f90856c5aa045b4e42d9c5b532db6f2fb7c93d1d7e0cbc0"},{"id":"func/supportedClickableMode","name":"supportedClickableMode","line":13,"end_line":20,"hash":"d02af774212b5b5b5d4c96b87808e8d3031dc3ef0cffea0b6966b00613a4153d"},{"id":"func/clickInsideRenderedVisibleControlBounds","name":"clickInsideRenderedVisibleControlBounds","line":22,"end_line":28,"hash":"5841041e428eeb645741c2b43099d69babeb2ef0ad13b018f2652c4a6573f47b"},{"id":"func/clickInsideRenderedControl","name":"clickInsideRenderedControl","line":30,"end_line":34,"hash":"31b9620cfbc6e98519ec409ae11cfd215d19c3caf86151ad2bc912baacc3b958"},{"id":"func/clickInsideRenderedBoundsOfControl","name":"clickInsideRenderedBoundsOfControl","line":36,"end_line":52,"hash":"9b363cdd70da665a8950d2ba0a4cea1c41b290a06ab04e6cab76c0f0948f0b53"},{"id":"func/assertClickableEditorMode","name":"assertClickableEditorMode","line":54,"end_line":56,"hash":"cf160cd862f834f60c4c2154af8d39785a763fae56ed4aa3f9e1c723bda3379f"},{"id":"func/clickableEditorMode","name":"clickableEditorMode","line":58,"end_line":58,"hash":"6ce6cd872963176553b96623503e14722baf5a26e7e91312a6149ac0d321c71f"},{"id":"func/assertVisibleControlActive","name":"assertVisibleControlActive","line":60,"end_line":73,"hash":"ec14c6cb9c18509370e653b0cccf22e749c6581f00a10e965f6ef30d20b7ef94"},{"id":"func/assertKeyboardPathEntryOpen","name":"assertKeyboardPathEntryOpen","line":75,"end_line":77,"hash":"7da9dd50dfb7bb4a57e3495ef819d4b1ca06fa764dde57df652b65cf7cbb2e57"},{"id":"func/clickablePathEntryCommand","name":"clickablePathEntryCommand","line":79,"end_line":79,"hash":"b627760a2beffeca0333c94ad4e72e6b7bec833e34f7b3e46de3e3d09b9b5878"},{"id":"func/assertDemoPickerOpen","name":"assertDemoPickerOpen","line":81,"end_line":90,"hash":"75229396f81086a009063088e5c4d2a158272d832e94875b181def06abea6167"},{"id":"func/assertClickableGameValue","name":"assertClickableGameValue","line":92,"end_line":105,"hash":"de0df05af33fb68dadbf519848f293dd259703ba6fb3a1bbebc8adda315530f3"},{"id":"func/recordVisibleControlShortcut","name":"recordVisibleControlShortcut","line":107,"end_line":118,"hash":"73dd0f004c414692ecb4751110fcee04e3a5e7d50722ec6aca7bb0efe8b1b274"},{"id":"func/assertClickMatchesShortcut","name":"assertClickMatchesShortcut","line":120,"end_line":138,"hash":"473a2c127e5d822f3c5848404551879b917bf5729b2a29b2ffe9618d3f4c3131"},{"id":"func/assertRecordedShortcut","name":"assertRecordedShortcut","line":140,"end_line":145,"hash":"dc13736ef090d3557d94bd3cf90ef5f710970d945088ffb6e49525bd1733ff2f"},{"id":"func/shortcutApplicationState","name":"shortcutApplicationState","line":147,"end_line":153,"hash":"f2d7a058b27cb3dae128bed9a74549f5a01c66153ebee90d1d097a29e6e08c2f"},{"id":"func/assertSameClickableApplicationState","name":"assertSameClickableApplicationState","line":155,"end_line":160,"hash":"8b4b3fbe1a2d14120589d5576fc90d1c423e432643a47ca0948d54257dea8dd6"},{"id":"func/recordClickableApplicationState","name":"recordClickableApplicationState","line":162,"end_line":167,"hash":"484263bef27cfda52b36539726268ab3578092e13bfab07c2877bb7336b2c931"},{"id":"func/clickOutsideVisibleControls","name":"clickOutsideVisibleControls","line":169,"end_line":178,"hash":"feea6d46c86ceee4cf2adef38d3ba0b738166e8256934a33d5c1665eced0449a"},{"id":"func/assertClickableApplicationStateUnchanged","name":"assertClickableApplicationStateUnchanged","line":180,"end_line":189,"hash":"ce924e9028d197fffd54edc7da498939a8f32f48a0a6d27eee746d0746373fc5"},{"id":"func/setClickableSimulationState","name":"setClickableSimulationState","line":191,"end_line":206,"hash":"414a6af95021ecb8dc68bf034c00f718abd30846a9b71b97a6b5849b33a10b1f"},{"id":"func/assertClickableSimulationState","name":"assertClickableSimulationState","line":208,"end_line":225,"hash":"ba0595d42821e46f9f1c58ee535c8abab9a9b75660112c9a39cc0ca45f0827d0"},{"id":"func/clickableApplicationState","name":"clickableApplicationState","line":227,"end_line":239,"hash":"027d33c71049a1d792daf2fb4fb96cea86174599635828531bafd5ef688b2425"},{"id":"func/appControlWithLabel","name":"appControlWithLabel","line":241,"end_line":249,"hash":"b950cb81287e13c6d18bed9c501a2d0accf9e9fff0d8b1593d4b66dc06a38170"}]}
// mutate4go-manifest-end
