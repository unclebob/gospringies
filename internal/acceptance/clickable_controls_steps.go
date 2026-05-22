package acceptance

import "fmt"

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

func clickableEditorMode(game *driverGame) string { return "" }

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

func clickablePathEntryCommand(game *driverGame) string { return game.PathEntryCommand() }

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

func assertClickableGameValue(w *world, example map[string]string, key string, name string, actual func(*driverGame) string) error {
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

func clickableApplicationState(game *driverGame) string {
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

func appControlWithLabel(label string) (drawFrameReport, bool) {
	game := newApplicationDriverGame()
	report, ok := drawFrameWithControlLabel(game, label)
	if ok {
		return report, true
	}
	game.SetPaused(true)
	return drawFrameWithControlLabel(game, label)
}

func drawFrameWithControlLabel(game *driverGame, label string) (drawFrameReport, bool) {
	report := game.DrawFrameReport()
	for _, controlLabel := range report.Controls {
		if controlLabel == label {
			return report, true
		}
	}
	return report, false
}

// mutate4go-manifest-begin
// {"version":1,"tested_at":"2026-05-22T10:58:20-05:00","module_hash":"2b45fd137636ad427c837191c47830a2442f8cf2052a1db254df705fbc10499b","functions":[{"id":"func/setClickableEditorMode","name":"setClickableEditorMode","line":5,"end_line":7,"hash":"64ab03d88b9be5441f90856c5aa045b4e42d9c5b532db6f2fb7c93d1d7e0cbc0"},{"id":"func/supportedClickableMode","name":"supportedClickableMode","line":9,"end_line":16,"hash":"d02af774212b5b5b5d4c96b87808e8d3031dc3ef0cffea0b6966b00613a4153d"},{"id":"func/clickInsideRenderedVisibleControlBounds","name":"clickInsideRenderedVisibleControlBounds","line":18,"end_line":24,"hash":"5841041e428eeb645741c2b43099d69babeb2ef0ad13b018f2652c4a6573f47b"},{"id":"func/clickInsideRenderedControl","name":"clickInsideRenderedControl","line":26,"end_line":30,"hash":"31b9620cfbc6e98519ec409ae11cfd215d19c3caf86151ad2bc912baacc3b958"},{"id":"func/clickInsideRenderedBoundsOfControl","name":"clickInsideRenderedBoundsOfControl","line":32,"end_line":48,"hash":"9b363cdd70da665a8950d2ba0a4cea1c41b290a06ab04e6cab76c0f0948f0b53"},{"id":"func/assertClickableEditorMode","name":"assertClickableEditorMode","line":50,"end_line":52,"hash":"cf160cd862f834f60c4c2154af8d39785a763fae56ed4aa3f9e1c723bda3379f"},{"id":"func/clickableEditorMode","name":"clickableEditorMode","line":54,"end_line":54,"hash":"34272e67fe36dbb08d00005c185ae495c13848435c9d7d46d649f2e2cc2b64ce"},{"id":"func/assertVisibleControlActive","name":"assertVisibleControlActive","line":56,"end_line":69,"hash":"ec14c6cb9c18509370e653b0cccf22e749c6581f00a10e965f6ef30d20b7ef94"},{"id":"func/assertKeyboardPathEntryOpen","name":"assertKeyboardPathEntryOpen","line":71,"end_line":73,"hash":"7da9dd50dfb7bb4a57e3495ef819d4b1ca06fa764dde57df652b65cf7cbb2e57"},{"id":"func/clickablePathEntryCommand","name":"clickablePathEntryCommand","line":75,"end_line":75,"hash":"e1050ed685ca0bfde8e520e8c07be8d51a79aa5b3bf32fd8aba290169838569f"},{"id":"func/assertDemoPickerOpen","name":"assertDemoPickerOpen","line":77,"end_line":86,"hash":"75229396f81086a009063088e5c4d2a158272d832e94875b181def06abea6167"},{"id":"func/assertClickableGameValue","name":"assertClickableGameValue","line":88,"end_line":101,"hash":"df2d7a7fa0c81642f86be0abd49f711471035718b687f293b936f838af0c17ba"},{"id":"func/recordVisibleControlShortcut","name":"recordVisibleControlShortcut","line":103,"end_line":114,"hash":"4b149be2dce354e0887b12513862d9da484639923218db77603e808b61224f1c"},{"id":"func/assertClickMatchesShortcut","name":"assertClickMatchesShortcut","line":116,"end_line":134,"hash":"473a2c127e5d822f3c5848404551879b917bf5729b2a29b2ffe9618d3f4c3131"},{"id":"func/assertRecordedShortcut","name":"assertRecordedShortcut","line":136,"end_line":141,"hash":"dc13736ef090d3557d94bd3cf90ef5f710970d945088ffb6e49525bd1733ff2f"},{"id":"func/shortcutApplicationState","name":"shortcutApplicationState","line":143,"end_line":149,"hash":"074687f144fd8e2be37333d298f5ee8d0a7787cffc7b11f98bea9c5e6ff1859e"},{"id":"func/assertSameClickableApplicationState","name":"assertSameClickableApplicationState","line":151,"end_line":156,"hash":"8b4b3fbe1a2d14120589d5576fc90d1c423e432643a47ca0948d54257dea8dd6"},{"id":"func/recordClickableApplicationState","name":"recordClickableApplicationState","line":158,"end_line":162,"hash":"c1167de8e313cab6f5c460bb059149932346b99575cf5485c3e1f1c7c54afc9d"},{"id":"func/clickOutsideVisibleControls","name":"clickOutsideVisibleControls","line":164,"end_line":173,"hash":"feea6d46c86ceee4cf2adef38d3ba0b738166e8256934a33d5c1665eced0449a"},{"id":"func/assertClickableApplicationStateUnchanged","name":"assertClickableApplicationStateUnchanged","line":175,"end_line":184,"hash":"ce924e9028d197fffd54edc7da498939a8f32f48a0a6d27eee746d0746373fc5"},{"id":"func/setClickableSimulationState","name":"setClickableSimulationState","line":186,"end_line":201,"hash":"414a6af95021ecb8dc68bf034c00f718abd30846a9b71b97a6b5849b33a10b1f"},{"id":"func/assertClickableSimulationState","name":"assertClickableSimulationState","line":203,"end_line":220,"hash":"ba0595d42821e46f9f1c58ee535c8abab9a9b75660112c9a39cc0ca45f0827d0"},{"id":"func/clickableApplicationState","name":"clickableApplicationState","line":222,"end_line":234,"hash":"7909299c5a58a367f974e18f629a67974cfdeb58939a89ed17e7c09335532e3f"},{"id":"func/appControlWithLabel","name":"appControlWithLabel","line":236,"end_line":244,"hash":"0e5ee22fa45cb0a7e9c6d5709bffb8863f6472f287eecae63c363b535160c9dd"}]}
// mutate4go-manifest-end
