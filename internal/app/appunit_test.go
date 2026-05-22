//go:build appunit

package app

import (
	"image"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"springs/internal/sim"
)

func TestAppUnitVisibleControlsAndSliders(t *testing.T) {
	game := NewGame()
	game.World().Parameters.Forces["gravity"] = sim.ForceConfig{Enabled: "false", Values: map[string]string{"magnitude": "0", "direction": "90"}}
	game.dirty = false

	gravityCheckbox, ok := visibleControlWithName("gravity force")
	if !ok {
		t.Fatal("missing gravity force checkbox")
	}
	if !game.ClickAt(gravityCheckbox.Rect.Min.X+1, gravityCheckbox.Rect.Min.Y+1) {
		t.Fatal("Gravity checkbox click was not handled")
	}
	force, _ := game.World().Parameters.Force("gravity")
	if force.Enabled != "true" || force.Values["magnitude"] != "10" || force.Values["direction"] != "0" || !game.dirty {
		t.Fatalf("gravity force = %#v dirty=%t", force, game.dirty)
	}

	control, ok := visibleControlWithName("speed slider")
	if !ok {
		t.Fatal("missing speed slider")
	}
	track := sliderTrack(control)
	game.ClickAt(track.Max.X, track.Min.Y)
	if game.simulationSpeed != maxSpeed {
		t.Fatalf("simulation speed = %f, want %f", game.simulationSpeed, maxSpeed)
	}
	game.ClickAt(track.Min.X, track.Min.Y)
	if game.simulationSpeed != 0 {
		t.Fatalf("simulation speed = %f, want 0", game.simulationSpeed)
	}
	if game.VisibleControlActive("missing") {
		t.Fatal("missing control should not be active")
	}
}

func TestAppUnitSpringWallToggleSetsMixedSelectionTrue(t *testing.T) {
	game := NewGame()
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{}, Mass: 1})
	_ = world.AddMass(sim.Mass{ID: 2, Position: sim.Vec2{X: 10}, Mass: 1})
	_ = world.AddSpring(sim.Spring{ID: 1, MassA: 1, MassB: 2, Wall: false})
	_ = world.AddSpring(sim.Spring{ID: 2, MassA: 1, MassB: 2, Wall: false})
	_ = world.AddSpring(sim.Spring{ID: 3, MassA: 1, MassB: 2, Wall: true})
	game.ReplaceWorld(world)
	game.editing().SelectedSprings = map[int]bool{1: true, 2: true, 3: true}

	if !game.ClickVisibleControl("Wall") {
		t.Fatal("Wall control click was not handled")
	}

	for _, spring := range game.World().Springs {
		if !spring.Wall {
			t.Fatalf("spring %d wall = %t, expected true", spring.ID, spring.Wall)
		}
	}
}

func TestAppUnitNumericSettingTextInputBranches(t *testing.T) {
	game := NewGame()

	game.appendNumericSettingInput([]rune("7"))
	if game.EnterNumericSettingText("7") {
		t.Fatal("unfocused numeric entry should not be handled")
	}
	if !game.FocusNumericSettingTextField("Mass") {
		t.Fatal("mass text field focus should be handled")
	}
	if game.FocusNumericSettingTextField("Missing") {
		t.Fatal("missing text field focus should not be handled")
	}
	if !game.numericTextCursorVisible("Mass") {
		t.Fatal("focused mass cursor should start visible")
	}
	if !game.numericTextHighlighted("Mass") {
		t.Fatal("focused mass value should start highlighted")
	}

	game.appendNumericSettingInput([]rune("x2.5"))
	if got := game.controls.numericInputText; got != "2.5" {
		t.Fatalf("numeric input text = %q, want 2.5", got)
	}
	if got := game.World().Parameters.Value("current mass"); got != "1" {
		t.Fatalf("current mass before commit = %q, want 1", got)
	}
	if !game.CommitNumericSettingText() {
		t.Fatal("numeric input commit should be handled")
	}
	if got := game.World().Parameters.Value("current mass"); got != "2.5" {
		t.Fatalf("current mass after commit = %q, want 2.5", got)
	}
	if game.controls.focusedNumeric != "" || game.numericTextHighlighted("Mass") {
		t.Fatalf("numeric focus/highlight after commit = %q/%t", game.controls.focusedNumeric, game.numericTextHighlighted("Mass"))
	}
	game.FocusNumericSettingTextField("Mass")
	game.deleteNumericSettingCharacter()
	if got := game.controls.numericInputText; got != "2." {
		t.Fatalf("numeric input after delete = %q, want 2.", got)
	}
	game.controls.focusedNumeric = "Missing"
	game.appendNumericSettingInput([]rune("9"))
	game.deleteNumericSettingCharacter()

	if !game.SetNumericSettingValue("Speed", "99") || game.simulationSpeed != maxSpeed {
		t.Fatalf("speed = %f, want %f", game.simulationSpeed, maxSpeed)
	}
	if game.SetNumericSettingValue("Speed", "not numeric") {
		t.Fatal("invalid numeric value should not be handled")
	}
	if !game.SetNumericSettingValue("Gravity", "12.5") {
		t.Fatal("gravity numeric setting should be handled")
	}
	force, _ := game.World().Parameters.Force("gravity")
	if force.Enabled != "true" || force.Values["magnitude"] != "12.5" {
		t.Fatalf("gravity force = %#v", force)
	}
	if game.ChangeNumericSettingWithSlider("Missing", "1") {
		t.Fatal("missing slider setting should not be handled")
	}
	if game.ChangeNumericSettingWithSlider("Mass", "not numeric") {
		t.Fatal("invalid slider value should not be handled")
	}
	if !game.ChangeNumericSettingWithSlider("Mass", "5") {
		t.Fatal("mass slider change should be handled")
	}

	control, ok := visibleControlWithName("mass text field")
	if !ok {
		t.Fatal("missing mass text field")
	}
	if !game.ClickAt(control.Rect.Min.X+1, control.Rect.Min.Y+1) || game.controls.focusedNumeric != "Mass" {
		t.Fatalf("text field click focused %q", game.controls.focusedNumeric)
	}

	if !game.activateVisibleControl(controlBox{Name: "edit menu"}) || !game.controls.editMenuOpen {
		t.Fatal("edit menu activation should open menu")
	}
	if !game.ClickAt(10, 32) || game.lastCommand != "cut" {
		t.Fatalf("edit menu item click command = %q", game.lastCommand)
	}
	game.controls.editMenuOpen = true
	if !game.ClickAt(0, screenHeight-1) || game.controls.editMenuOpen {
		t.Fatal("outside click should close edit menu")
	}
	game.controls.focusedNumeric = "Mass"
	if game.ClickAt(0, screenHeight-1) || game.controls.focusedNumeric != "" {
		t.Fatalf("outside click handled=%t focused=%q", true, game.controls.focusedNumeric)
	}
	if !game.activateVisibleControl(controlBox{Name: "run pause toggle command"}) || game.lastCommand != "pause toggle" {
		t.Fatalf("run/pause activation command = %q", game.lastCommand)
	}
	if !game.activateVisibleControl(controlBox{Name: "fixed mass toggle"}) {
		t.Fatal("inspector toggle activation should be handled")
	}
	if game.activateVisibleControl(controlBox{Name: "missing"}) {
		t.Fatal("missing activation should not be handled")
	}
}

func TestAppUnitVisibleControlLayoutAndReport(t *testing.T) {
	game := appUnitGameWithMasses(
		sim.Mass{ID: 1, Position: sim.Vec2{X: 120, Y: 120}, Mass: 1},
		sim.Mass{ID: 2, Position: sim.Vec2{X: 180, Y: 120}, Mass: 1},
	)
	_ = game.World().AddSpring(sim.Spring{ID: 1, MassA: 1, MassB: 2})

	if inspectorLeft() != screenWidth-inspectorWidth {
		t.Fatalf("inspector left = %d", inspectorLeft())
	}
	for _, name := range []string{"mass slider", "gravity slider", "stick slider", "speed slider", "viscosity slider"} {
		if !isSliderControl(name) {
			t.Fatalf("%q should be a slider", name)
		}
	}
	if isSliderControl("gravity force") {
		t.Fatal("button reported as slider")
	}

	x := inspectorLeft() + 16
	right := screenWidth - 16
	half := (right - x - 8) / 2
	third := (right - x - 16) / 3
	second := x + third + 8
	thirdStart := second + third + 8
	massSetting, _ := numericSettingByName("Mass")
	_, label, decrement, slider, increment, text := numericSettingRects(massSetting)
	gravitySetting, _ := numericSettingByName("Gravity")
	gravityCheckbox, gravityLabel, gravityDecrement, gravitySlider, gravityIncrement, _ := numericSettingRects(gravitySetting)
	centerSetting, _ := numericSettingByName("Center Attraction")
	centerCheckbox, _, _, _, _, _ := numericSettingRects(centerSetting)
	centerMassSetting, _ := numericSettingByName("Center Of Mass Attraction")
	centerMassCheckbox, _, _, _, _, _ := numericSettingRects(centerMassSetting)
	wallSetting, _ := numericSettingByName("Wall Repulsion")
	wallCheckbox, wallLabel, _, _, _, _ := numericSettingRects(wallSetting)
	wallToggles := wallToggleControlsForSetting(wallSetting)
	precisionSetting, _ := numericSettingByName("Precision")
	precisionCheckbox, precisionLabel, precisionDecrement, precisionSlider, precisionIncrement, precisionText := numericSettingRects(precisionSetting)
	wantControls := map[string]image.Rectangle{
		"mass label":               label,
		"mass decrement":           decrement,
		"mass slider":              slider,
		"mass increment":           increment,
		"mass text field":          text,
		"gravity force":            gravityCheckbox,
		"gravity label":            gravityLabel,
		"gravity decrement":        gravityDecrement,
		"gravity slider":           gravitySlider,
		"gravity increment":        gravityIncrement,
		"center attraction force":  centerCheckbox,
		"center mass force":        centerMassCheckbox,
		"wall repulsion force":     wallCheckbox,
		"wall repulsion label":     wallLabel,
		"fixed mass toggle":        image.Rect(x, 120, x+third, 140),
		"set center command":       image.Rect(second, 120, second+third, 140),
		"mass collision force":     image.Rect(thirdStart, 120, right, 140),
		"set rest length command":  image.Rect(x, 229, x+half, 249),
		"spring wall toggle":       image.Rect(x+half+8, 229, right, 249),
		"top wall toggle":          wallToggles[0].Rect,
		"bottom wall toggle":       wallToggles[1].Rect,
		"left wall toggle":         wallToggles[2].Rect,
		"right wall toggle":        wallToggles[3].Rect,
		"adaptive timestep toggle": precisionCheckbox,
		"precision label":          precisionLabel,
		"precision decrement":      precisionDecrement,
		"precision slider":         precisionSlider,
		"precision increment":      precisionIncrement,
		"precision text field":     precisionText,
		"grid snap toggle":         image.Rect(x, 608, x+half, 628),
		"show springs toggle":      image.Rect(x+half+8, 608, right, 628),
	}
	controls := map[string]image.Rectangle{}
	for _, control := range inspectorControls() {
		controls[control.Name] = control.Rect
	}
	for name, want := range wantControls {
		if got := controls[name]; got != want {
			t.Fatalf("%s rect = %#v, want %#v", name, got, want)
		}
	}

	sections := inspectorSections()
	if len(sections) != 5 || sections[0].Rect != image.Rect(x, 44, right, 64) || sections[3].Rect != image.Rect(x, 397, right, 417) || sections[4].Rect != image.Rect(x, 584, right, 604) {
		t.Fatalf("inspector sections = %#v", sections)
	}

	if count := game.selectedObjectCount(); count != 0 {
		t.Fatalf("initial selected count = %d", count)
	}
	_ = game.editing().SelectMass(1)
	game.editing().SelectedSprings[1] = true
	if count := game.selectedObjectCount(); count != 2 {
		t.Fatalf("selected count = %d", count)
	}

	report := game.DrawFrameReport()
	if !report.ControlLabelsFit || report.CanvasWorldPixels != 100 {
		t.Fatalf("draw report = %#v", report)
	}
	if report.RegionPixels["canvas"] != rectPixels(visibleRegionRects()["canvas"]) {
		t.Fatalf("canvas pixels = %d", report.RegionPixels["canvas"])
	}
	if report.RegionControlCounts["right inspector"] != len(inspectorControls())+len(inspectorSections())+len(game.statusFields()) {
		t.Fatalf("right inspector count = %d", report.RegionControlCounts["right inspector"])
	}

	game.World().Parameters.EnableForce("gravity", map[string]string{"magnitude": "25"})
	game.World().Parameters.Set("fixed mass", "true")
	game.World().Parameters.EnableWall("top")
	game.simulationSpeed = maxSpeed / 2
	game.World().Parameters.Set("viscosity", "1")
	if !game.activeControl("gravity force") || !game.activeControl("fixed mass toggle") || !game.activeControl("top wall toggle") {
		t.Fatal("enabled controls should be active")
	}
	for _, name := range []string{"missing", "center attraction force", "adaptive timestep toggle", "left wall toggle"} {
		if game.activeControl(name) {
			t.Fatalf("%s should not be active", name)
		}
	}
	if game.activeRunControl("missing") || game.activeForceControl("missing") || game.activeParameterControl("missing") || game.activeWallControl("missing") {
		t.Fatal("missing active helper should return false")
	}
	if !game.forceEnabled("gravity") || game.forceEnabled("missing") || !game.parameterEnabled("fixed mass") || game.parameterEnabled("missing") || !game.wallEnabled("top") || game.wallEnabled("missing") {
		t.Fatal("enabled helper state mismatch")
	}
	game.World().Parameters.Set("grid snap", "0")
	if game.gridSnapEnabled() {
		t.Fatal("zero grid snap should be disabled")
	}
	game.World().Parameters.Set("grid snap", "10")
	if !game.gridSnapEnabled() {
		t.Fatal("positive grid snap should be enabled")
	}
	game.World().Parameters.Set("grid snap", "1")
	if !game.gridSnapEnabled() {
		t.Fatal("unit grid snap should be enabled")
	}
	if got := game.sliderFraction("gravity slider"); got != 0.5 {
		t.Fatalf("gravity slider fraction = %f", got)
	}
	if got := game.sliderFraction("speed slider"); got != 0.5 {
		t.Fatalf("speed slider fraction = %f", got)
	}
	if got := game.sliderFraction("viscosity slider"); got != 0.5 {
		t.Fatalf("viscosity slider fraction = %f", got)
	}
	if got := game.sliderFraction("missing"); got != 0 {
		t.Fatalf("missing slider fraction = %f", got)
	}

	active := game.visibleActiveControls()
	if !active["gravity force"] || active["Gravity"] || active["gravity slider"] {
		t.Fatalf("visible active controls = %#v", active)
	}
	sectionsMap := visibleInspectorSections()
	if !sectionsMap[sectionHeaderLabel("Selected Mass(es)")] ||
		!sectionsMap[sectionHeaderLabel("Selected Spring(s)")] ||
		!sectionsMap[sectionHeaderLabel("Forces")] ||
		sectionsMap[sectionHeaderLabel("Walls")] ||
		!sectionsMap[sectionHeaderLabel("Simulation")] ||
		!sectionsMap[sectionHeaderLabel("Display")] {
		t.Fatalf("visible inspector sections = %#v", sectionsMap)
	}

	status := game.statusFields()
	x = inspectorLeft() + 16
	right = screenWidth - 16
	wantStatus := []image.Rectangle{
		image.Rect(x, screenHeight-56, x+120, screenHeight-36),
		image.Rect(x+128, screenHeight-56, right, screenHeight-36),
		image.Rect(x, screenHeight-30, right, screenHeight-10),
	}
	for i, want := range wantStatus {
		if status[i].Rect != want {
			t.Fatalf("status rect %d = %#v, want %#v", i, status[i].Rect, want)
		}
	}

	regions := visibleRegionRects()
	wantRegions := map[string]image.Rectangle{
		"canvas":          image.Rect(toolbarWidth, topBarHeight, screenWidth-inspectorWidth, screenHeight),
		"left toolbar":    image.Rect(0, topBarHeight, toolbarWidth, screenHeight),
		"top command bar": image.Rect(toolbarWidth, 0, screenWidth-inspectorWidth, topBarHeight),
		"right inspector": image.Rect(screenWidth-inspectorWidth, topBarHeight, screenWidth, screenHeight),
	}
	for name, want := range wantRegions {
		if got := regions[name]; got != want {
			t.Fatalf("%s region = %#v, want %#v", name, got, want)
		}
	}
	if got := game.visibleRegionPixels("top command bar"); got != regionControlPixels(visibleControls(), "top command bar")+regionControlPixels(inspectorSections(), "top command bar") {
		t.Fatalf("top command pixels = %d", got)
	}
	if got := game.visibleRegionPixels("right inspector"); got != regionControlPixels(visibleControls(), "right inspector")+regionControlPixels(inspectorSections(), "right inspector")+game.regionStatusPixels("right inspector") {
		t.Fatalf("right inspector pixels = %d", got)
	}
	if got := game.regionStatusPixels("right inspector"); got != rectPixels(status[0].Rect)+rectPixels(status[1].Rect)+rectPixels(status[2].Rect) {
		t.Fatalf("status pixels = %d", got)
	}
	if game.regionStatusPixels("canvas") != 0 {
		t.Fatal("canvas should not have status pixels")
	}
	if rectPixels(image.Rect(0, 0, 3, 5)) != 15 {
		t.Fatal("rect pixel multiplication mismatch")
	}
	if !labelFits("abcd", image.Rect(0, 0, 32, debugGlyphHeight)) {
		t.Fatal("boundary label should fit")
	}
	if labelFits("abcde", image.Rect(0, 0, 32, debugGlyphHeight)) || labelFits("abcd", image.Rect(0, 0, 32, debugGlyphHeight-1)) {
		t.Fatal("oversized labels should not fit")
	}
	game.document.pathEntryCommand = "this/status/path/is/far/too/long/to/fit/even/with/a/full/inspector/row"
	if visibleLabelsFit(game) {
		t.Fatal("long status label should not fit")
	}
}

func TestAppUnitDragMassSnapsToGrid(t *testing.T) {
	game := appUnitGameWithMasses(
		sim.Mass{ID: 1, Position: sim.Vec2{X: 110, Y: 110}, Velocity: sim.Vec2{X: 1, Y: 1}, Mass: 1},
		sim.Mass{ID: 2, Position: sim.Vec2{X: 120, Y: 120}, Velocity: sim.Vec2{X: 2, Y: 2}, Mass: 1},
	)
	game.World().Parameters.Set("grid snap", "10")

	if !game.DragMass(1, sim.Vec2{X: 123, Y: 87}) {
		t.Fatal("single mass drag should succeed")
	}
	mass, _ := game.World().MassByID(1)
	if mass.Position != (sim.Vec2{X: 120, Y: 90}) || mass.Velocity != (sim.Vec2{}) || !game.dirty || !game.pointer.dragMoved {
		t.Fatalf("single dragged mass = %#v dirty=%t moved=%t", mass, game.dirty, game.pointer.dragMoved)
	}

	game.dirty = false
	game.pointer.dragMoved = false
	game.pointer.draggingOffsets = map[int]sim.Vec2{1: {X: 2, Y: 3}, 2: {X: -3, Y: -2}}
	_ = game.editing().SelectMass(1)
	_ = game.editing().AddMassSelection(2)
	if !game.DragMass(1, sim.Vec2{X: 113, Y: 114}) {
		t.Fatal("selected mass drag should succeed")
	}
	first, _ := game.World().MassByID(1)
	second, _ := game.World().MassByID(2)
	if first.Position != (sim.Vec2{X: 120, Y: 120}) || second.Position != (sim.Vec2{X: 110, Y: 110}) {
		t.Fatalf("selected dragged masses = %#v %#v", first, second)
	}
	if first.Velocity != (sim.Vec2{}) || second.Velocity != (sim.Vec2{}) || !game.dirty || !game.pointer.dragMoved {
		t.Fatalf("selected drag state first=%#v second=%#v dirty=%t moved=%t", first, second, game.dirty, game.pointer.dragMoved)
	}

	game.World().Parameters.Set("grid snap", "0")
	if got := game.snapToGrid(sim.Vec2{X: 12.5, Y: 17.5}); got != (sim.Vec2{X: 12.5, Y: 17.5}) {
		t.Fatalf("disabled snap = %#v", got)
	}
	game.World().Parameters.Set("grid snap", "1")
	if got := game.snapToGrid(sim.Vec2{X: 12.5, Y: 17.5}); got != (sim.Vec2{X: 13, Y: 18}) {
		t.Fatalf("unit snap = %#v", got)
	}
}

func TestAppUnitEditorScreenAndShortcuts(t *testing.T) {
	game := NewGame()
	game.SetPaused(true)
	game.SetSelected(true)
	game.SetDirty(true)

	screen := game.EditorScreen()
	if !screen.Editor || screen.LandingPage || !screen.CanvasVisible || !screen.ControlsUsable {
		t.Fatalf("screen flags = %#v", screen)
	}
	if screen.Indicators["simulation state"] != "paused" || screen.Indicators["selection"] != "object selected" || screen.Indicators["file state"] != "unsaved changes" {
		t.Fatalf("screen indicators = %#v", screen.Indicators)
	}
	if purpose, ok := screen.RegionPurpose("canvas"); !ok || purpose == "" {
		t.Fatalf("canvas purpose = %q, %t", purpose, ok)
	}
	if purpose, ok := screen.RegionPurpose("missing"); ok || purpose != "" {
		t.Fatalf("missing purpose = %q, %t", purpose, ok)
	}
	if !screen.HasCommandControl("pause toggle") || screen.HasCommandControl("missing") {
		t.Fatalf("command controls = %#v", screen.CommandControls)
	}

	if !game.HandleShortcut("Ctrl+A") || game.LastCommand() != "select all" {
		t.Fatalf("shortcut command = %q", game.LastCommand())
	}
	if game.HandleShortcut("Ctrl+Z") {
		t.Fatal("unknown shortcut should not be handled")
	}
}

func TestAppUnitSliderFractionBounds(t *testing.T) {
	track := image.Rect(10, 0, 30, 1)
	if got := sliderFractionAt(track, 10); got != 0 {
		t.Fatalf("min fraction = %f, want 0", got)
	}
	if got := sliderFractionAt(track, 30); got != 1 {
		t.Fatalf("max fraction = %f, want 1", got)
	}
	if got := sliderFractionAt(image.Rect(10, 0, 11, 1), 11); got != 1 {
		t.Fatalf("one-pixel fraction = %f, want 1", got)
	}
	if got := sliderFractionAt(image.Rect(10, 0, 10, 1), 20); got != 0 {
		t.Fatalf("zero-width fraction = %f, want 0", got)
	}
}

func TestAppUnitCommandsUpdateSelectionAndDirtyState(t *testing.T) {
	game := appUnitGameWithMasses(
		sim.Mass{ID: 1, Position: sim.Vec2{X: 10, Y: 20}, Mass: 2},
		sim.Mass{ID: 2, Position: sim.Vec2{X: 30, Y: 40}, Mass: 3},
	)
	_ = game.editing().SelectMass(1)
	game.editing().SelectedMasses[2] = true

	game.RunCommand("copy")
	game.pointer.lastCursor = sim.Vec2{X: 100, Y: 120}
	game.RunCommand("paste")
	if len(game.World().Masses) != 4 || !game.selected || !game.dirty {
		t.Fatalf("paste state masses=%d selected=%t dirty=%t", len(game.World().Masses), game.selected, game.dirty)
	}
	if !game.editing().MassSelected(3) || !game.editing().MassSelected(4) {
		t.Fatalf("pasted masses were not selected: %#v", game.editing().SelectedMasses)
	}

	game.RunCommand("cut")
	if game.selected || !game.dirty {
		t.Fatalf("cut state selected=%t dirty=%t", game.selected, game.dirty)
	}

	game.SaveState()
	game.dirty = false
	game.RunCommand("restore state")
	if !game.dirty {
		t.Fatal("restore command should mark game dirty")
	}
}

func TestAppUnitClipboardCopiesSpringsAndComputesOrigin(t *testing.T) {
	game := appUnitGameWithMasses(
		sim.Mass{ID: 7, Position: sim.Vec2{X: 30, Y: 50}, Mass: 1},
		sim.Mass{ID: 9, Position: sim.Vec2{X: 10, Y: 40}, Mass: 1},
		sim.Mass{ID: 11, Position: sim.Vec2{X: 20, Y: 20}, Mass: 1},
	)
	_ = game.World().AddSpring(sim.Spring{ID: 5, MassA: 7, MassB: 9, RestLength: 20})
	_ = game.World().AddSpring(sim.Spring{ID: 6, MassA: 9, MassB: 11, RestLength: 20})
	_ = game.editing().SelectMass(7)
	_ = game.editing().AddMassSelection(9)

	game.copySelection()

	if len(game.editClipboard.Masses) != 2 || len(game.editClipboard.Springs) != 1 {
		t.Fatalf("clipboard = %#v", game.editClipboard)
	}
	if got := game.editClipboard.origin(); got != (sim.Vec2{X: 10, Y: 40}) {
		t.Fatalf("origin = %#v", got)
	}
	if got := (editClipboard{}).origin(); got != (sim.Vec2{}) {
		t.Fatalf("empty origin = %#v", got)
	}
	if got := (editClipboard{Masses: []sim.Mass{{Position: sim.Vec2{X: 5, Y: 6}}}}).origin(); got != (sim.Vec2{X: 5, Y: 6}) {
		t.Fatalf("single origin = %#v", got)
	}
}

func TestAppUnitPasteSelectionCopiesSpringsAndIDs(t *testing.T) {
	game := appUnitGameWithMasses(
		sim.Mass{ID: 1, Position: sim.Vec2{X: 10, Y: 20}, Mass: 1},
		sim.Mass{ID: 2, Position: sim.Vec2{X: 30, Y: 40}, Mass: 1},
	)
	_ = game.World().AddSpring(sim.Spring{ID: 1, MassA: 1, MassB: 2, RestLength: 20})
	game.editClipboard = editClipboard{
		Masses:  []sim.Mass{{ID: 1, Position: sim.Vec2{X: 10, Y: 20}, Mass: 1}, {ID: 2, Position: sim.Vec2{X: 30, Y: 40}, Mass: 1}},
		Springs: []sim.Spring{{ID: 1, MassA: 1, MassB: 2, RestLength: 20}},
	}

	if !game.pasteSelectionAt(sim.Vec2{X: 100, Y: 120}) {
		t.Fatal("paste should report success")
	}
	if len(game.World().Masses) != 4 || len(game.World().Springs) != 2 {
		t.Fatalf("world after paste masses=%#v springs=%#v", game.World().Masses, game.World().Springs)
	}
	if !game.editing().MassSelected(3) || !game.editing().MassSelected(4) || !game.editing().SpringSelected(2) {
		t.Fatalf("pasted selection = masses %#v springs %#v", game.editing().SelectedMasses, game.editing().SelectedSprings)
	}

	game.editClipboard = editClipboard{}
	if game.pasteSelectionAt(sim.Vec2{}) {
		t.Fatal("empty clipboard should not paste")
	}
}

func TestAppUnitNextIDsUseOnePastMaximum(t *testing.T) {
	empty := NewGame()
	empty.ReplaceWorld(sim.NewWorld())
	if got := empty.nextMassID(); got != 1 {
		t.Fatalf("empty next mass id = %d", got)
	}
	if got := empty.nextSpringID(); got != 1 {
		t.Fatalf("empty next spring id = %d", got)
	}

	game := appUnitGameWithMasses(sim.Mass{ID: 1, Mass: 1}, sim.Mass{ID: 4, Mass: 1})
	_ = game.World().AddSpring(sim.Spring{ID: 1, MassA: 1, MassB: 4})
	_ = game.World().AddSpring(sim.Spring{ID: 4, MassA: 1, MassB: 4})
	if got := game.nextMassID(); got != 5 {
		t.Fatalf("next mass id = %d", got)
	}
	if got := game.nextSpringID(); got != 5 {
		t.Fatalf("next spring id = %d", got)
	}

	oneMass := appUnitGameWithMasses(sim.Mass{ID: 1, Mass: 1})
	if got := oneMass.nextMassID(); got != 2 {
		t.Fatalf("single next mass id = %d", got)
	}
}

func TestAppUnitLoadReplaceAndEditorAttachment(t *testing.T) {
	game := NewGame()
	_ = game.World().AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 1, Y: 2}, Mass: 1})
	_ = game.editing().SelectMass(1)
	game.syncSelectionState()
	game.editing().World = sim.NewWorld()

	if err := game.LoadXSP("#1.0\ncmas 7\nmass 9 10 20 1 0\n"); err != nil {
		t.Fatal(err)
	}
	if _, ok := game.World().MassByID(9); !ok || game.selected || game.editor.World != game.simulation || game.dirty {
		t.Fatalf("load state world=%#v selected=%t dirty=%t", game.World(), game.selected, game.dirty)
	}

	game.editing().World = sim.NewWorld()
	replacement := sim.NewWorld()
	_ = replacement.AddMass(sim.Mass{ID: 3, Position: sim.Vec2{X: 3, Y: 4}, Mass: 1})
	game.ReplaceWorld(replacement)
	if _, ok := game.World().MassByID(3); !ok || game.editor.World != game.simulation {
		t.Fatalf("replace state world=%#v", game.World())
	}
}

func TestAppUnitRenderWorldReportsRepresentations(t *testing.T) {
	game := appUnitGameWithMasses(
		sim.Mass{ID: 1, Position: sim.Vec2{X: 10, Y: 10}, Mass: 1, Fixed: true},
		sim.Mass{ID: 2, Position: sim.Vec2{X: 30, Y: 10}, Mass: 1},
	)
	_ = game.World().AddSpring(sim.Spring{ID: 1, MassA: 1, MassB: 2})
	game.World().Parameters.EnableWall("left")
	game.World().SetForceCenter([]int{1})
	game.SetSelected(true)

	result := game.RenderWorld()

	if !result.Completed || !result.SpringLinesVisible || !result.MassesVisible || !result.FixedMassDistinguishable {
		t.Fatalf("render result = %#v", result)
	}
	for _, object := range []string{"movable mass", "fixed mass", "spring", "enabled wall", "selection", "force center"} {
		if !result.HasVisibleRepresentation(object) {
			t.Fatalf("missing representation for %q: %#v", object, result.Representations)
		}
	}
	if result.FixedMassRepresentation != "red circle" || result.MovableMassRepresentation != "yellow circle" || result.SelectedMassRepresentation != "selection outline" {
		t.Fatalf("render labels = %#v", result)
	}
}

func TestAppUnitRenderWorldHidesMissingRepresentations(t *testing.T) {
	game := appUnitGameWithMasses(sim.Mass{ID: 1, Position: sim.Vec2{X: 10, Y: 10}, Mass: 1})
	game.World().Parameters.Set("show springs", "false")

	result := game.RenderWorld()

	if result.SpringLinesVisible || !result.MassesVisible || result.FixedMassDistinguishable {
		t.Fatalf("render result = %#v", result)
	}
	for _, object := range []string{"spring", "enabled wall", "selection", "force center"} {
		if result.HasVisibleRepresentation(object) {
			t.Fatalf("unexpected representation for %q: %#v", object, result.Representations)
		}
	}
	if (RenderResult{}).HasVisibleRepresentation("mass") {
		t.Fatal("empty render result reported visible representation")
	}

	empty := NewGame()
	empty.ReplaceWorld(sim.NewWorld())
	empty.World().Parameters.Set("center mass", "0")
	emptyResult := empty.RenderWorld()
	if emptyResult.HasVisibleRepresentation("spring") || emptyResult.HasVisibleRepresentation("force center") {
		t.Fatalf("empty render representations = %#v", emptyResult.Representations)
	}

	withHiddenSpring := appUnitGameWithMasses(
		sim.Mass{ID: 1, Position: sim.Vec2{X: 10, Y: 10}, Mass: 1},
		sim.Mass{ID: 2, Position: sim.Vec2{X: 30, Y: 10}, Mass: 1},
	)
	_ = withHiddenSpring.World().AddSpring(sim.Spring{ID: 1, MassA: 1, MassB: 2})
	withHiddenSpring.World().Parameters.Set("show springs", "false")
	hiddenSpringResult := withHiddenSpring.RenderWorld()
	if hiddenSpringResult.SpringLinesVisible || hiddenSpringResult.HasVisibleRepresentation("spring") {
		t.Fatalf("hidden spring render result = %#v", hiddenSpringResult)
	}
}

func TestAppUnitSpringEndpointResolution(t *testing.T) {
	game := appUnitGameWithMasses(
		sim.Mass{ID: 1, Position: sim.Vec2{X: 10, Y: 10}, Mass: 1},
		sim.Mass{ID: 2, Position: sim.Vec2{X: 30, Y: 10}, Mass: 1},
	)

	tests := []struct {
		name   string
		spring sim.Spring
		valid  bool
	}{
		{name: "index endpoints", spring: sim.Spring{A: 0, B: 1}, valid: true},
		{name: "id endpoints", spring: sim.Spring{MassA: 1, MassB: 2}, valid: true},
		{name: "partial id endpoints are not indexes", spring: sim.Spring{MassA: 0, MassB: 2}, valid: false},
		{name: "missing mass a", spring: sim.Spring{MassA: 9, MassB: 2}, valid: false},
		{name: "missing mass b", spring: sim.Spring{MassA: 1, MassB: 9}, valid: false},
		{name: "negative index a", spring: sim.Spring{A: -1, B: 1}, valid: false},
		{name: "negative index b", spring: sim.Spring{A: 0, B: -1}, valid: false},
		{name: "index a too high", spring: sim.Spring{A: 2, B: 1}, valid: false},
		{name: "index b too high", spring: sim.Spring{A: 0, B: 2}, valid: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := game.validSpring(test.spring); got != test.valid {
				t.Fatalf("validSpring(%#v) = %t, want %t", test.spring, got, test.valid)
			}
		})
	}

	if !validSpringIndex(0, game.World().Masses) || !validSpringIndex(1, game.World().Masses) {
		t.Fatal("valid spring indexes were rejected")
	}
	if validSpringIndex(-1, game.World().Masses) || validSpringIndex(2, game.World().Masses) {
		t.Fatal("invalid spring indexes were accepted")
	}
}

func TestAppUnitValueDialogBehavior(t *testing.T) {
	game := appUnitGameWithMasses(
		sim.Mass{ID: 1, Position: sim.Vec2{X: 10, Y: 10}, Mass: 5},
		sim.Mass{ID: 2, Position: sim.Vec2{X: 30, Y: 10}, Mass: 1},
	)
	_ = game.World().AddSpring(sim.Spring{ID: 3, MassA: 1, MassB: 2, SpringConstant: 12})

	game.openMassValueDialog(99)
	if game.overlays.value.Open {
		t.Fatal("missing mass opened value dialog")
	}
	game.openMassValueDialog(1)
	if !game.overlays.value.Open || game.overlays.value.Title != "Set Mass #1" || game.overlays.value.Text != "5" || game.overlays.value.Min != 0 || game.overlays.value.Max != 20 {
		t.Fatalf("mass value dialog = %#v", game.overlays.value)
	}

	game.overlays.value.Text = "1"
	game.appendValueDialogInput([]rune{'2', 'x', '.', '-', '3'})
	if game.overlays.value.Text != "12.-3" {
		t.Fatalf("dialog text = %q", game.overlays.value.Text)
	}
	game.deleteValueDialogCharacter()
	if game.overlays.value.Text != "12.-" {
		t.Fatalf("dialog text after delete = %q", game.overlays.value.Text)
	}
	game.overlays.value.Text = ""
	game.deleteValueDialogCharacter()
	if game.overlays.value.Text != "" {
		t.Fatalf("empty dialog text after delete = %q", game.overlays.value.Text)
	}
	game.overlays.value.Text = "a"
	game.deleteValueDialogCharacter()
	if game.overlays.value.Text != "" {
		t.Fatalf("single-character dialog text after delete = %q", game.overlays.value.Text)
	}

	game.overlays.value = valueDialog{Open: true, Text: "15", Min: 0, Max: 10}
	if got := game.valueDialogFraction(); got != 1 {
		t.Fatalf("high fraction = %f", got)
	}
	game.overlays.value.Text = "-5"
	if got := game.valueDialogFraction(); got != 0 {
		t.Fatalf("low fraction = %f", got)
	}
	game.overlays.value.Text = "5"
	if got := game.valueDialogFraction(); got != 0.5 {
		t.Fatalf("middle fraction = %f", got)
	}
	game.overlays.value = valueDialog{Text: "10", Min: 5, Max: 15}
	if got := game.valueDialogFraction(); got != 0.5 {
		t.Fatalf("nonzero range fraction = %f", got)
	}
	game.overlays.value.Text = "invalid"
	if got := game.valueDialogFraction(); got != 0 {
		t.Fatalf("invalid fraction = %f", got)
	}
	game.overlays.value = valueDialog{Text: "invalid", Min: -10, Max: 10}
	if got := game.valueDialogFraction(); got != 0 {
		t.Fatalf("invalid nonzero-range fraction = %f", got)
	}
	game.overlays.value = valueDialog{Text: "10", Min: 10, Max: 10}
	if got := game.valueDialogFraction(); got != 0 {
		t.Fatalf("flat range fraction = %f", got)
	}

	rect := valueDialogRect()
	if rect != image.Rect(screenWidth/2-valueDialogWidth/2, screenHeight/2-valueDialogHeight/2, screenWidth/2+valueDialogWidth/2, screenHeight/2+valueDialogHeight/2) {
		t.Fatalf("dialog rect = %#v", rect)
	}
	if got := game.valueDialogTextRect(); got != image.Rect(rect.Min.X+12, rect.Min.Y+42, rect.Max.X-12, rect.Min.Y+66) {
		t.Fatalf("text rect = %#v", got)
	}
	track := game.valueDialogSliderTrack()
	if got := game.valueDialogDecrementRect(); got != image.Rect(rect.Min.X+12, rect.Min.Y+86, rect.Min.X+12+numericStepButtonWidth, rect.Min.Y+106) {
		t.Fatalf("decrement rect = %#v", got)
	}
	if got := game.valueDialogIncrementRect(); got != image.Rect(rect.Max.X-12-numericStepButtonWidth, rect.Min.Y+86, rect.Max.X-12, rect.Min.Y+106) {
		t.Fatalf("increment rect = %#v", got)
	}
	if track != image.Rect(rect.Min.X+12+numericStepButtonWidth+numericStepButtonGap, rect.Min.Y+92, rect.Max.X-12-numericStepButtonWidth-numericStepButtonGap, rect.Min.Y+100) {
		t.Fatalf("slider track = %#v", track)
	}
	if got := game.valueDialogOKRect(); got != image.Rect(rect.Max.X-64, rect.Max.Y-34, rect.Max.X-12, rect.Max.Y-12) {
		t.Fatalf("ok rect = %#v", got)
	}

	game.overlays.value = valueDialog{Open: true, Text: "0", Min: 0, Max: 20}
	game.setValueDialogFromSlider(track.Min.X + track.Dx()/2)
	if game.overlays.value.Text != "10" {
		t.Fatalf("slider text = %q", game.overlays.value.Text)
	}
	game.setValueDialogFromSlider(track.Min.X - track.Dx())
	if game.overlays.value.Text != "0" {
		t.Fatalf("slider low text = %q", game.overlays.value.Text)
	}
	game.setValueDialogFromSlider(track.Max.X + track.Dx())
	if game.overlays.value.Text != "20" {
		t.Fatalf("slider high text = %q", game.overlays.value.Text)
	}
	game.overlays.value = valueDialog{Open: true, Text: "0", Min: 5, Max: 15}
	game.setValueDialogFromSlider(track.Min.X + track.Dx()/2)
	if game.overlays.value.Text != "10" {
		t.Fatalf("nonzero slider text = %q", game.overlays.value.Text)
	}
	game.clickValueDialog(game.valueDialogIncrementRect().Min.X+1, game.valueDialogIncrementRect().Min.Y+1)
	if game.overlays.value.Text != "10.1" || game.controls.activeValueStep != numericStepAmount {
		t.Fatalf("increment text=%q active=%f", game.overlays.value.Text, game.controls.activeValueStep)
	}
	game.controls.valueStepTicks = numericStepHoldDelayTicks - 1
	game.continueValueDialogStepHold()
	if game.overlays.value.Text != "10.2" {
		t.Fatalf("held increment text = %q", game.overlays.value.Text)
	}
	game.clickValueDialog(game.valueDialogDecrementRect().Min.X+1, game.valueDialogDecrementRect().Min.Y+1)
	if game.overlays.value.Text != "10.1" || game.controls.activeValueStep != -numericStepAmount {
		t.Fatalf("decrement text=%q active=%f", game.overlays.value.Text, game.controls.activeValueStep)
	}

	game.overlays.value = valueDialog{Open: true}
	game.clickValueDialog(rect.Max.X+1, rect.Max.Y+1)
	if game.overlays.value.Open {
		t.Fatal("outside click should close dialog")
	}
	applied := 0.0
	game.overlays.value = valueDialog{Open: true, Text: "4", Apply: func(value float64) { applied = value }}
	game.clickValueDialog(game.valueDialogOKRect().Min.X+1, game.valueDialogOKRect().Min.Y+1)
	if applied != 4 || game.overlays.value.Open {
		t.Fatalf("ok click applied=%f dialog=%#v", applied, game.overlays.value)
	}
	game.overlays.value = valueDialog{Open: true, Text: "bad", Apply: func(value float64) { applied = value }}
	game.applyValueDialog()
	if !game.overlays.value.Open || applied != 4 {
		t.Fatalf("invalid apply applied=%f dialog=%#v", applied, game.overlays.value)
	}
	game.overlays.value = valueDialog{Open: true, Text: "8"}
	game.applyValueDialog()
	if game.overlays.value.Open {
		t.Fatal("apply without callback should still close dialog")
	}

	game.overlays.value = valueDialog{Open: true}
	game.tickValueDialog()
	if game.overlays.value.Ticks != 1 || !game.valueDialogCursorVisible() {
		t.Fatalf("dialog ticks=%d cursor=%t", game.overlays.value.Ticks, game.valueDialogCursorVisible())
	}
	game.overlays.value.Ticks = valueCursorPeriod
	if game.valueDialogCursorVisible() {
		t.Fatal("cursor should hide after one period")
	}
	game.overlays.value.Open = false
	if game.valueDialogCursorVisible() {
		t.Fatal("closed dialog cursor should be hidden")
	}
}

func TestAppUnitSpringValueDialogAndDistance(t *testing.T) {
	game := appUnitGameWithMasses(
		sim.Mass{ID: 1, Position: sim.Vec2{X: 10, Y: 10}, Mass: 1},
		sim.Mass{ID: 2, Position: sim.Vec2{X: 30, Y: 10}, Mass: 1},
	)
	_ = game.World().AddSpring(sim.Spring{ID: 3, MassA: 1, MassB: 2, SpringConstant: 12})
	game.canvasYUp = false

	if game.openSpringConstantDialogAt(200, 200) {
		t.Fatal("far spring dialog should not open")
	}
	if !game.openSpringConstantDialogAt(20, 10) {
		t.Fatal("spring dialog did not open")
	}
	if !game.overlays.value.Open || game.overlays.value.Title != "Kspring Spring #3" || game.overlays.value.Text != "12" || game.overlays.value.Min != 0 || game.overlays.value.Max != 1000 {
		t.Fatalf("spring value dialog = %#v", game.overlays.value)
	}
	game.overlays.value.Text = "22"
	game.dirty = false
	game.applyValueDialog()
	spring, _ := game.World().SpringByID(3)
	if spring.SpringConstant != 22 || spring.Stiffness != 22 || !game.dirty || game.overlays.value.Open {
		t.Fatalf("spring after apply = %#v dirty=%t dialog=%#v", spring, game.dirty, game.overlays.value)
	}

	game.dirty = false
	game.setSpringConstant(99, 33)
	if game.dirty {
		t.Fatal("missing spring constant update should not mark dirty")
	}
	game.setSpringDamping(99, 33)
	if game.dirty {
		t.Fatal("missing spring damping update should not mark dirty")
	}
	game.setSpringRestLength(99, 33)
	if game.dirty {
		t.Fatal("missing spring rest length update should not mark dirty")
	}

	if id, ok := game.springAt(sim.Vec2{X: 20, Y: 16}); id != 3 || !ok {
		t.Fatalf("threshold spring hit = %d, %t", id, ok)
	}
	if id, ok := game.springAt(sim.Vec2{X: 100, Y: 100}); id != 0 || ok {
		t.Fatalf("far spring hit = %d, %t", id, ok)
	}

	assertNear(t, distanceToSegment(sim.Vec2{X: 13, Y: 14}, sim.Vec2{X: 10, Y: 10}, sim.Vec2{X: 10, Y: 10}), 5)
	assertNear(t, distanceToSegment(sim.Vec2{X: 0.5, Y: 2}, sim.Vec2{}, sim.Vec2{X: 1}), 2)
	assertNear(t, distanceToSegment(sim.Vec2{X: 20, Y: 10}, sim.Vec2{X: 10, Y: 10}, sim.Vec2{X: 30, Y: 30}), math.Sqrt(50))
	assertNear(t, distanceToSegment(sim.Vec2{X: 25, Y: 19}, sim.Vec2{X: 10, Y: 10}, sim.Vec2{X: 30, Y: 30}), math.Sqrt(18))
}

func assertNear(t *testing.T, got float64, want float64) {
	t.Helper()
	if math.IsNaN(got) || math.Abs(got-want) > 1e-9 {
		t.Fatalf("got %f, want %f", got, want)
	}
}

func TestAppUnitSpringTemperatureValueDialog(t *testing.T) {
	game := appUnitGameWithMasses(
		sim.Mass{ID: 1, Position: sim.Vec2{X: 10, Y: 10}, Mass: 1},
		sim.Mass{ID: 2, Position: sim.Vec2{X: 30, Y: 10}, Mass: 1},
	)
	_ = game.World().AddSpring(sim.Spring{ID: 3, MassA: 1, MassB: 2, Temperature: 2.5})

	if !game.SelectSpringContextMenuItem(3, "Temperature") {
		t.Fatal("Temperature spring menu item was not handled")
	}
	if !game.overlays.value.Open || game.overlays.value.Title != "Temperature Spring #3" || game.overlays.value.Text != "2.5" || game.overlays.value.Min != 0 || game.overlays.value.Max != 10 {
		t.Fatalf("temperature dialog = %#v", game.overlays.value)
	}
	game.overlays.value.Text = "7.5"
	game.applyValueDialog()
	spring, _ := game.World().SpringByID(3)
	if spring.Temperature != 7.5 {
		t.Fatalf("spring temperature = %#v", spring)
	}
}

func TestAppUnitDemoPickerSelection(t *testing.T) {
	dir := t.TempDir()
	first := filepath.Join(dir, "first.xsp")
	second := filepath.Join(dir, "second.xsp")
	if err := os.WriteFile(first, []byte("#1.0\nmass 1 10 20 1 0\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(second, []byte("#1.0\nmass 2 30 40 1 0\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	game := NewGame()
	game.ReplaceWorld(sim.NewWorld())
	game.controls.demoFiles = []string{first, second}
	game.controls.demoPickerOpen = true
	game.controls.demoPickerScroll = 1
	row := game.demoRowRect(0)
	game.clickDemoPicker(row.Min.X+2, row.Min.Y+2)
	if _, ok := game.World().MassByID(2); !ok || game.controls.demoPickerOpen {
		t.Fatalf("demo picker load failed masses=%#v open=%t", game.World().Masses, game.controls.demoPickerOpen)
	}

	third := filepath.Join(dir, "third.xsp")
	if err := os.WriteFile(third, []byte("#1.0\nmass 3 50 60 1 0\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	game.ReplaceWorld(sim.NewWorld())
	game.controls.demoFiles = []string{first, second, third}
	game.controls.demoPickerOpen = true
	game.controls.demoPickerScroll = 1
	row = game.demoRowRect(1)
	game.clickDemoPicker(row.Min.X+2, row.Min.Y+2)
	if _, ok := game.World().MassByID(3); !ok || game.controls.demoPickerOpen {
		t.Fatalf("scrolled row load failed masses=%#v open=%t", game.World().Masses, game.controls.demoPickerOpen)
	}

	game.controls.demoPickerOpen = true
	game.clickDemoPicker(demoPickerRect().Max.X+1, demoPickerRect().Max.Y+1)
	if game.controls.demoPickerOpen {
		t.Fatal("outside click should close demo picker")
	}
}

func TestAppUnitSaveFilenameDialogEditsAndWritesFile(t *testing.T) {
	root := t.TempDir()
	withAppUnitWorkingDirectory(t, root)
	game := appUnitGameWithMasses(sim.Mass{ID: 7, Position: sim.Vec2{X: 10, Y: 20}, Mass: 1})

	game.openSaveFilenameDialog()
	if !game.SaveFilenameDialogOpen() || game.SaveFilenameText() != ".xsp" || game.SaveFilenameCursor() != 0 {
		t.Fatalf("save dialog = open:%t text:%q cursor:%d", game.SaveFilenameDialogOpen(), game.SaveFilenameText(), game.SaveFilenameCursor())
	}

	game.EnterSaveFilenamePrefix("lab_scene")
	game.deleteSaveFilenameCharacter()
	game.EnterSaveFilenamePrefix("e")
	if err := game.SubmitSaveFilenameDialog(); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join("saves", "lab_scene.xsp")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "\nmass 7 ") {
		t.Fatalf("saved content missing mass 7:\n%s", string(content))
	}
	if game.CurrentFilePath() != path || game.SaveFilenameDialogOpen() {
		t.Fatalf("current path = %q open = %t", game.CurrentFilePath(), game.SaveFilenameDialogOpen())
	}

	game.openSaveFilenameDialog()
	game.clickSaveFilenameDialog(saveFilenameDialogRect().Max.X+1, saveFilenameDialogRect().Max.Y+1)
	if game.SaveFilenameDialogOpen() {
		t.Fatal("outside click should close save dialog")
	}

	game.openSaveFilenameDialog()
	game.EnterSaveFilenamePrefix("clicked")
	ok := game.saveFilenameDialogOKRect()
	game.clickSaveFilenameDialog(ok.Min.X+1, ok.Min.Y+1)
	if game.SaveFilenameDialogOpen() || game.CurrentFilePath() != filepath.Join("saves", "clicked.xsp") {
		t.Fatalf("ok click left dialog open=%t path=%q", game.SaveFilenameDialogOpen(), game.CurrentFilePath())
	}
}

func TestAppUnitSaveFilenameDeleteEdges(t *testing.T) {
	game := NewGame()
	game.overlays.save = saveFilenameDialog{Open: true, Text: ".xsp", Cursor: 0}
	game.deleteSaveFilenameCharacter()
	if game.SaveFilenameText() != ".xsp" || game.SaveFilenameCursor() != 0 {
		t.Fatalf("delete at start changed dialog text=%q cursor=%d", game.SaveFilenameText(), game.SaveFilenameCursor())
	}

	game.overlays.save = saveFilenameDialog{Open: true, Text: "", Cursor: 1}
	game.deleteSaveFilenameCharacter()
	if game.SaveFilenameText() != "" || game.SaveFilenameCursor() != 1 {
		t.Fatalf("delete empty changed dialog text=%q cursor=%d", game.SaveFilenameText(), game.SaveFilenameCursor())
	}

	game.overlays.save = saveFilenameDialog{Open: true, Text: "abc.xsp", Cursor: 3}
	game.deleteSaveFilenameCharacter()
	if game.SaveFilenameText() != "ab.xsp" || game.SaveFilenameCursor() != 2 {
		t.Fatalf("delete middle text=%q cursor=%d", game.SaveFilenameText(), game.SaveFilenameCursor())
	}
}

func TestAppUnitSaveFilenameDialogGeometry(t *testing.T) {
	game := NewGame()
	rect := saveFilenameDialogRect()
	wantRect := image.Rect(
		screenWidth/2-valueDialogWidth/2,
		screenHeight/2-valueDialogHeight/2,
		screenWidth/2+valueDialogWidth/2,
		screenHeight/2+valueDialogHeight/2,
	)
	if rect != wantRect {
		t.Fatalf("save dialog rect = %#v", rect)
	}
	if got := game.saveFilenameTextRect(); got != image.Rect(rect.Min.X+12, rect.Min.Y+42, rect.Max.X-12, rect.Min.Y+66) {
		t.Fatalf("save filename text rect = %#v", got)
	}
	if got := game.saveFilenameDialogOKRect(); got != image.Rect(rect.Max.X-64, rect.Max.Y-34, rect.Max.X-12, rect.Max.Y-12) {
		t.Fatalf("save filename OK rect = %#v", got)
	}
}

func TestAppUnitSaveFilenamePathNormalizesInput(t *testing.T) {
	for _, tc := range []struct {
		input string
		path  string
		ok    bool
	}{
		{input: " lab_scene.xsp ", path: filepath.Join("saves", "lab_scene.xsp"), ok: true},
		{input: filepath.Join("nested", "scene"), path: filepath.Join("saves", "scene.xsp"), ok: true},
		{input: ".xsp", ok: false},
	} {
		path, err := saveFilenamePath(tc.input)
		if tc.ok && (err != nil || path != tc.path) {
			t.Fatalf("saveFilenamePath(%q) = %q, %v; want %q", tc.input, path, err, tc.path)
		}
		if !tc.ok && err == nil {
			t.Fatalf("saveFilenamePath(%q) unexpectedly succeeded with %q", tc.input, path)
		}
	}
}

func TestAppUnitLoadPickerEntriesAndSelectionByName(t *testing.T) {
	root := t.TempDir()
	withAppUnitWorkingDirectory(t, root)
	savePath := filepath.Join("saves", "lab_scene.xsp")
	demoPath := filepath.Join("demos", "pendulum.xsp")
	originalPath := filepath.Join("demos", "original", "pend.xsp")
	for path, content := range map[string]string{
		savePath:     "#1.0\nmass 9 10 20 1 0\n",
		demoPath:     "#1.0\nmass 2 10 20 1 0\n",
		originalPath: "#1.0\nmass 3 10 20 1 0\n",
	} {
		appUnitWriteFile(t, path, content)
	}

	game := NewGame()
	want := []string{savePath, loadPickerSeparator, demoPath, originalPath}
	if got := game.LoadPickerEntries(); !sameStrings(got, want) {
		t.Fatalf("load picker entries = %#v, want %#v", got, want)
	}
	if game.ChooseLoadPickerEntry(loadPickerSeparator) {
		t.Fatal("separator should not load")
	}
	if !game.ChooseLoadPickerEntry("lab_scene.xsp") {
		t.Fatalf("saved file did not load: %s", game.LastFileError())
	}
	if _, ok := game.World().MassByID(9); !ok || game.CurrentFilePath() != savePath {
		t.Fatalf("loaded masses = %#v current path = %q", game.World().Masses, game.CurrentFilePath())
	}
	if game.ChooseLoadPickerEntry("missing.xsp") {
		t.Fatal("missing picker entry should not load")
	}
}

func TestAppUnitOpenDemoPickerRefreshesFileList(t *testing.T) {
	root := t.TempDir()
	withAppUnitWorkingDirectory(t, root)
	appUnitWriteFile(t, filepath.Join("saves", "old.xsp"), "#1.0\nmass 1 10 20 1 0\n")
	game := NewGame()

	game.openDemoPicker()
	if !sameStrings(game.LoadPickerEntries(), []string{filepath.Join("saves", "old.xsp"), loadPickerSeparator}) {
		t.Fatalf("initial picker entries = %#v", game.LoadPickerEntries())
	}

	appUnitWriteFile(t, filepath.Join("saves", "simple hex.xsp"), "#1.0\nmass 2 30 40 1 0\n")
	game.openDemoPicker()

	if !sameStrings(game.LoadPickerEntries(), []string{filepath.Join("saves", "old.xsp"), filepath.Join("saves", "simple hex.xsp"), loadPickerSeparator}) {
		t.Fatalf("refreshed picker entries = %#v", game.LoadPickerEntries())
	}
	if !game.ChooseLoadPickerEntry("simple hex.xsp") {
		t.Fatalf("new saved file did not load: %s", game.LastFileError())
	}
	if _, ok := game.World().MassByID(2); !ok || game.CurrentFilePath() != filepath.Join("saves", "simple hex.xsp") {
		t.Fatalf("loaded masses = %#v current path = %q", game.World().Masses, game.CurrentFilePath())
	}
}

func withAppUnitWorkingDirectory(t *testing.T, dir string) {
	t.Helper()
	previous, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(previous); err != nil {
			t.Fatalf("restore working directory: %v", err)
		}
	})
}

func appUnitWriteFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func sameStrings(left []string, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func TestAppUnitDemoPickerGeometryAndBounds(t *testing.T) {
	game := NewGame()
	rect := demoPickerRect()
	if rect != image.Rect(240, 96, screenWidth-240, screenHeight-96) {
		t.Fatalf("picker rect = %#v", rect)
	}
	row0 := game.demoRowRect(0)
	row1 := game.demoRowRect(1)
	if row0 != image.Rect(rect.Min.X+12, rect.Min.Y+40, rect.Max.X-12, rect.Min.Y+40+demoPickerRowHeight-2) {
		t.Fatalf("row0 = %#v", row0)
	}
	if row1.Min.Y-row0.Min.Y != demoPickerRowHeight || row1.Dx() != row0.Dx() || row1.Dy() != row0.Dy() {
		t.Fatalf("row spacing row0=%#v row1=%#v", row0, row1)
	}
	if rows := demoPickerVisibleRows(); rows != 31 {
		t.Fatalf("visible rows = %d, want 31", rows)
	}
	game.controls.demoFiles = []string{"first.xsp", "second.xsp", "third.xsp"}
	game.controls.demoPickerScroll = 0
	visible := game.visibleDemoPaths()
	if len(visible) != 3 || visible[0] != "first.xsp" {
		t.Fatalf("visible demo paths = %#v", visible)
	}
	game.controls.demoFiles = []string{filepath.Join("..", "..", "demos", "pendulum.xsp")}
	game.ReplaceWorld(sim.NewWorld())
	game.controls.demoPickerOpen = true
	game.loadDemoAt(0)
	if _, ok := game.World().MassByID(1); !ok || game.controls.demoPickerOpen {
		t.Fatalf("index 0 load failed masses=%#v open=%t", game.World().Masses, game.controls.demoPickerOpen)
	}
	for _, index := range []int{-1, 1} {
		game.controls.demoFiles = []string{filepath.Join("..", "..", "demos", "pendulum.xsp")}
		game.controls.demoPickerOpen = true
		if game.loadDemoAt(index) || !game.controls.demoPickerOpen {
			t.Fatalf("out-of-range index %d result open=%t", index, game.controls.demoPickerOpen)
		}
	}
	game.controls.demoFiles = []string{loadPickerSeparator}
	game.controls.demoPickerOpen = true
	if game.loadDemoAt(0) || !game.controls.demoPickerOpen {
		t.Fatalf("separator load result open=%t", game.controls.demoPickerOpen)
	}
	game.controls.demoFiles = []string{"missing.xsp"}
	game.controls.demoPickerOpen = true
	game.loadDemoAt(0)
	if game.controls.demoPickerOpen {
		t.Fatal("failed file read should still close demo picker")
	}
}

func TestAppUnitGroupedLoadPickerEntries(t *testing.T) {
	for _, tc := range []struct {
		name      string
		saves     []string
		starters  []string
		originals []string
		want      []string
	}{
		{
			name:     "saves before starters",
			saves:    []string{"saves/a.xsp"},
			starters: []string{"demos/b.xsp"},
			want:     []string{"saves/a.xsp", loadPickerSeparator, "demos/b.xsp"},
		},
		{
			name:      "saves before originals",
			saves:     []string{"saves/a.xsp"},
			originals: []string{"demos/original/c.xsp"},
			want:      []string{"saves/a.xsp", loadPickerSeparator, "demos/original/c.xsp"},
		},
		{
			name:  "saves only",
			saves: []string{"saves/a.xsp"},
			want:  []string{"saves/a.xsp", loadPickerSeparator},
		},
		{
			name:     "starters only",
			starters: []string{"demos/b.xsp"},
			want:     []string{"demos/b.xsp"},
		},
	} {
		if got := groupedLoadPickerEntries(tc.saves, tc.starters, tc.originals); !sameStrings(got, tc.want) {
			t.Fatalf("%s entries = %#v, want %#v", tc.name, got, tc.want)
		}
	}
}

func TestAppUnitMassContextMenuActionsAndGeometry(t *testing.T) {
	game := appUnitGameWithMasses(sim.Mass{ID: 1, Position: sim.Vec2{X: 20, Y: 20}, Mass: 3})
	game.canvasYUp = false

	if !game.openMassContextMenu(20, 20) {
		t.Fatal("mass context menu did not open")
	}
	if !game.overlays.massMenu.Open || game.overlays.massMenu.MassID != 1 || !game.selected {
		t.Fatalf("mass menu = %#v selected=%t", game.overlays.massMenu, game.selected)
	}
	rect := game.massContextMenuRect()
	if rect.Dx() != massMenuWidth || rect.Dy() != (massMenuTitleRows+len(game.massContextMenuItems()))*massMenuRowHeight {
		t.Fatalf("mass menu rect = %#v", rect)
	}
	row0 := game.massContextMenuRowRect(0)
	row1 := game.massContextMenuRowRect(1)
	if row0.Min.Y != rect.Min.Y+massMenuTitleRows*massMenuRowHeight || row1.Min.Y-row0.Min.Y != massMenuRowHeight {
		t.Fatalf("mass menu rows row0=%#v row1=%#v rect=%#v", row0, row1, rect)
	}

	game.clickMassContextMenu(row0.Min.X+1, row0.Min.Y+1)
	mass, _ := game.World().MassByID(1)
	if !mass.Fixed || game.overlays.massMenu.Open || !game.dirty {
		t.Fatalf("fixed toggle mass=%#v menu=%#v dirty=%t", mass, game.overlays.massMenu, game.dirty)
	}

	game.overlays.massMenu = massContextMenu{Open: true, MassID: 1, X: 20, Y: 20}
	row1 = game.massContextMenuRowRect(1)
	game.clickMassContextMenu(row1.Min.X+1, row1.Min.Y+1)
	if !game.overlays.value.Open || game.overlays.value.Target != "mass" || game.overlays.value.Text != "3" {
		t.Fatalf("value dialog = %#v", game.overlays.value)
	}
	game.overlays.value.Text = "7"
	game.dirty = false
	game.applyValueDialog()
	mass, _ = game.World().MassByID(1)
	if mass.Mass != 7 || !game.dirty {
		t.Fatalf("mass value = %f dirty=%t", mass.Mass, game.dirty)
	}

	game.overlays.massMenu = massContextMenu{Open: true, MassID: 1, X: 20, Y: 20}
	row2 := game.massContextMenuRowRect(2)
	game.dirty = false
	game.clickMassContextMenu(row2.Min.X+1, row2.Min.Y+1)
	if game.World().CenterMassID() != 1 || game.overlays.massMenu.Open || !game.dirty {
		t.Fatalf("center mass = %d menu=%#v dirty=%t", game.World().CenterMassID(), game.overlays.massMenu, game.dirty)
	}

	game.overlays.massMenu = massContextMenu{Open: true, MassID: 1, X: 0, Y: 0}
	anchored := game.massContextMenuRect()
	if anchored.Min != image.Pt(0, 0) {
		t.Fatalf("anchored menu rect = %#v", anchored)
	}

	game.overlays.massMenu = massContextMenu{Open: true, MassID: 1, X: screenWidth + 50, Y: screenHeight + 50}
	clamped := game.massContextMenuRect()
	if clamped.Max.X != screenWidth || clamped.Max.Y != screenHeight {
		t.Fatalf("clamped menu rect = %#v", clamped)
	}
	game.clickMassContextMenu(clamped.Max.X+1, clamped.Max.Y+1)
	if game.overlays.massMenu.Open {
		t.Fatal("outside context click should close menu")
	}
}

func TestAppUnitOpenContextAtChoosesSpringDialogAndIgnoresDemoPicker(t *testing.T) {
	game := appUnitGameWithMasses(
		sim.Mass{ID: 1, Position: sim.Vec2{X: 10, Y: 10}, Mass: 1},
		sim.Mass{ID: 2, Position: sim.Vec2{X: 110, Y: 10}, Mass: 1},
	)
	game.canvasYUp = false
	_ = game.World().AddSpring(sim.Spring{ID: 3, MassA: 1, MassB: 2, SpringConstant: 12})

	game.controls.demoPickerOpen = true
	game.openContextAt(10, 10)
	if game.overlays.massMenu.Open || game.overlays.value.Open {
		t.Fatal("demo picker should block context menu opening")
	}

	game.controls.demoPickerOpen = false
	game.overlays.springMenu.Open = true
	game.openContextAt(10, 10)
	if !game.overlays.massMenu.Open || game.overlays.massMenu.MassID != 1 || game.overlays.springMenu.Open || game.overlays.value.Open {
		t.Fatalf("mass context = %#v spring=%#v dialog=%#v", game.overlays.massMenu, game.overlays.springMenu, game.overlays.value)
	}

	game.overlays.massMenu.Open = true
	game.openContextAt(60, 10)
	if !game.overlays.springMenu.Open || game.overlays.springMenu.SpringID != 3 || game.overlays.massMenu.Open || game.overlays.value.Open {
		t.Fatalf("spring menu = %#v mass menu=%#v dialog=%#v", game.overlays.springMenu, game.overlays.massMenu, game.overlays.value)
	}

	if game.openSpringContextMenu(500, 500) {
		t.Fatal("empty spring context position should not open a spring menu")
	}

	game.overlays.springMenu.Open = false
	game.openContextAt(500, 500)
	if game.overlays.massMenu.Open || game.overlays.springMenu.Open || game.overlays.value.Open {
		t.Fatalf("empty context should close overlays mass=%#v spring=%#v dialog=%#v", game.overlays.massMenu, game.overlays.springMenu, game.overlays.value)
	}
}

func TestAppUnitSpringContextMenuActionsAndGeometry(t *testing.T) {
	game := appUnitGameWithMasses(
		sim.Mass{ID: 1, Position: sim.Vec2{X: 10, Y: 10}, Mass: 1},
		sim.Mass{ID: 2, Position: sim.Vec2{X: 110, Y: 10}, Mass: 1},
	)
	game.canvasYUp = false
	_ = game.World().AddSpring(sim.Spring{ID: 3, MassA: 1, MassB: 2, RestLength: 100, SpringConstant: 12, Damping: 0.4})

	if !game.openSpringContextMenu(60, 10) {
		t.Fatal("spring context menu did not open")
	}
	rect := game.springContextMenuRect()
	if rect.Dx() != springMenuWidth || rect.Dy() != (springMenuTitleRows+len(game.springContextMenuItems()))*springMenuRowHeight {
		t.Fatalf("spring menu rect = %#v", rect)
	}
	row0 := game.springContextMenuRowRect(0)
	row1 := game.springContextMenuRowRect(1)
	if row0.Min.Y != rect.Min.Y+springMenuTitleRows*springMenuRowHeight || row1.Min.Y-row0.Min.Y != springMenuRowHeight {
		t.Fatalf("spring menu rows row0=%#v row1=%#v rect=%#v", row0, row1, rect)
	}

	game.clickSpringContextMenu(row0.Min.X+1, row0.Min.Y+1)
	if !game.overlays.value.Open || game.overlays.value.Title != "Kspring Spring #3" || game.overlays.value.Text != "12" || game.overlays.value.Target != "spring" || game.overlays.springMenu.Open {
		t.Fatalf("kspring dialog = %#v spring menu=%#v", game.overlays.value, game.overlays.springMenu)
	}
	game.overlays.value.Text = "22"
	game.applyValueDialog()
	spring, _ := game.World().SpringByID(3)
	if spring.SpringConstant != 22 || spring.Stiffness != 22 {
		t.Fatalf("spring constant = %#v", spring)
	}

	game.overlays.springMenu = springContextMenu{Open: true, SpringID: 3, X: 60, Y: 10}
	row1 = game.springContextMenuRowRect(1)
	game.clickSpringContextMenu(row1.Min.X+1, row1.Min.Y+1)
	if !game.overlays.value.Open || game.overlays.value.Title != "Kdamp Spring #3" || game.overlays.value.Text != "0.4" {
		t.Fatalf("kdamp dialog = %#v", game.overlays.value)
	}
	game.overlays.value.Text = "0.9"
	game.applyValueDialog()
	spring, _ = game.World().SpringByID(3)
	if spring.Damping != 0.9 {
		t.Fatalf("spring damping = %#v", spring)
	}

	game.overlays.springMenu = springContextMenu{Open: true, SpringID: 3, X: 60, Y: 10}
	row2 := game.springContextMenuRowRect(2)
	game.clickSpringContextMenu(row2.Min.X+1, row2.Min.Y+1)
	if !game.overlays.value.Open || game.overlays.value.Title != "RestLen Spring #3" || game.overlays.value.Text != "100" {
		t.Fatalf("rest length dialog = %#v", game.overlays.value)
	}
	game.overlays.value.Text = "125"
	game.applyValueDialog()
	spring, _ = game.World().SpringByID(3)
	if spring.RestLength != 125 {
		t.Fatalf("spring rest length = %#v", spring)
	}

	game.overlays.springMenu = springContextMenu{Open: true, SpringID: 3, X: 60, Y: 10}
	game.dirty = false
	row3 := game.springContextMenuRowRect(3)
	game.clickSpringContextMenu(row3.Min.X+1, row3.Min.Y+1)
	spring, _ = game.World().SpringByID(3)
	if !spring.Wall || !game.dirty || game.overlays.springMenu.Open {
		t.Fatalf("wall toggle spring=%#v dirty=%t menu=%#v", spring, game.dirty, game.overlays.springMenu)
	}

	labels := game.SpringContextMenuLabelsForSpring(3)
	if len(labels) != 5 || labels[3] != "Wall" || labels[4] != "Temperature" {
		t.Fatalf("spring menu labels = %#v", labels)
	}
	if !game.SelectSpringContextMenuItem(3, "Wall") || game.overlays.springMenu.Open || game.SelectSpringContextMenuItem(3, "Missing") {
		t.Fatalf("spring menu selection failed, spring=%#v", game.overlays.springMenu)
	}
	spring, _ = game.World().SpringByID(3)
	if spring.Wall {
		t.Fatalf("programmatic wall toggle spring=%#v", spring)
	}

	game.overlays.springMenu = springContextMenu{Open: true, SpringID: 3, X: screenWidth + 50, Y: screenHeight + 50}
	clamped := game.springContextMenuRect()
	if clamped.Max.X != screenWidth || clamped.Max.Y != screenHeight {
		t.Fatalf("clamped spring menu rect = %#v", clamped)
	}
	game.clickSpringContextMenu(clamped.Max.X+1, clamped.Max.Y+1)
	if game.overlays.springMenu.Open {
		t.Fatal("outside spring menu click should close menu")
	}
}

func TestAppUnitContextMenuHelpers(t *testing.T) {
	called := false
	closed := false
	items := []contextMenuItem{{
		Label:  "Run",
		Action: func() { called = true },
	}}

	labels := contextMenuLabels(items)
	if len(labels) != 1 || labels[0] != "Run" {
		t.Fatalf("labels = %#v", labels)
	}
	if !selectContextMenuItem(items, "Run", func() { closed = true }) {
		t.Fatal("existing menu item should be selected")
	}
	if !called || !closed {
		t.Fatalf("selection side effects called=%t closed=%t", called, closed)
	}

	called = false
	closed = false
	if selectContextMenuItem(items, "Missing", func() { closed = true }) {
		t.Fatal("missing menu item should not be selected")
	}
	if called || closed {
		t.Fatalf("missing selection side effects called=%t closed=%t", called, closed)
	}
}
