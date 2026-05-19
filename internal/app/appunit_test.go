//go:build appunit

package app

import (
	"image"
	"math"
	"os"
	"path/filepath"
	"testing"

	"springs/internal/sim"
)

func appUnitGameWithMasses(masses ...sim.Mass) *Game {
	game := NewGame()
	world := sim.NewWorld()
	for _, mass := range masses {
		if mass.Mass == 0 {
			mass.Mass = 1
		}
		_ = world.AddMass(mass)
	}
	game.ReplaceWorld(world)
	return game
}

func TestAppUnitVisibleControlsAndSliders(t *testing.T) {
	game := NewGame()
	game.World().Parameters.Forces["gravity"] = sim.ForceConfig{Enabled: "false", Values: map[string]string{"magnitude": "0", "direction": "90"}}
	game.dirty = false

	if !game.ClickVisibleControl("Gravity") {
		t.Fatal("Gravity control click was not handled")
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

func TestAppUnitVisibleControlLayoutAndReport(t *testing.T) {
	game := appUnitGameWithMasses(
		sim.Mass{ID: 1, Position: sim.Vec2{X: 120, Y: 120}, Mass: 1},
		sim.Mass{ID: 2, Position: sim.Vec2{X: 180, Y: 120}, Mass: 1},
	)
	_ = game.World().AddSpring(sim.Spring{ID: 1, MassA: 1, MassB: 2})

	if inspectorLeft() != screenWidth-inspectorWidth {
		t.Fatalf("inspector left = %d", inspectorLeft())
	}
	for _, name := range []string{"gravity slider", "speed slider", "viscosity slider"} {
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
	wantControls := map[string]image.Rectangle{
		"mass parameter":           image.Rect(x, 68, x+half, 88),
		"elasticity parameter":     image.Rect(x+half+8, 68, right, 88),
		"kspring parameter":        image.Rect(x, 148, x+half, 168),
		"kdamp parameter":          image.Rect(x+half+8, 148, right, 168),
		"gravity force":            image.Rect(x, 230, right, 250),
		"gravity slider":           image.Rect(x, 256, right, 286),
		"center attraction force":  image.Rect(x, 292, x+half, 312),
		"center mass force":        image.Rect(x+half+8, 292, right, 312),
		"wall repulsion force":     image.Rect(x, 318, x+half, 338),
		"mass collision force":     image.Rect(x+half+8, 318, right, 338),
		"top wall toggle":          image.Rect(x, 414, x+half, 434),
		"bottom wall toggle":       image.Rect(x+half+8, 414, right, 434),
		"left wall toggle":         image.Rect(x, 440, x+half, 460),
		"right wall toggle":        image.Rect(x+half+8, 440, right, 460),
		"grid snap toggle":         image.Rect(x, 514, x+half, 534),
		"show springs toggle":      image.Rect(x+half+8, 514, right, 534),
		"timestep parameter":       image.Rect(x, 638, x+half, 658),
		"precision parameter":      image.Rect(x+half+8, 638, right, 658),
		"adaptive timestep toggle": image.Rect(x, 664, right, 684),
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
	if len(sections) != 5 || sections[0].Rect != image.Rect(x, 44, right, 64) || sections[4].Rect != image.Rect(x, 486, right, 508) {
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
	if report.RegionControlCounts["right inspector"] != len(inspectorControls())+len(inspectorSections()) {
		t.Fatalf("right inspector count = %d", report.RegionControlCounts["right inspector"])
	}
	if report.RegionControlCounts["status line"] != len(game.statusFields()) {
		t.Fatalf("status count = %d", report.RegionControlCounts["status line"])
	}

	game.World().Parameters.EnableForce("gravity", map[string]string{"magnitude": "25"})
	game.World().Parameters.Set("fixed mass", "true")
	game.World().Parameters.EnableWall("top")
	game.simulationSpeed = maxSpeed / 2
	game.World().Parameters.Set("viscosity", "1")
	if !game.activeControl("gravity force") || !game.activeControl("fixed mass toggle") || !game.activeControl("top wall toggle") {
		t.Fatal("enabled controls should be active")
	}
	for _, name := range []string{"missing", "pause command", "center attraction force", "adaptive timestep toggle", "left wall toggle"} {
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
	if !active["Gravity"] || !active["gravity force"] || active["gravity slider"] {
		t.Fatalf("visible active controls = %#v", active)
	}
	sectionsMap := visibleInspectorSections()
	if !sectionsMap["Mass"] || !sectionsMap["Spring"] || !sectionsMap["Forces"] || !sectionsMap["Walls"] || !sectionsMap["Simulation"] {
		t.Fatalf("visible inspector sections = %#v", sectionsMap)
	}

	status := game.statusFields()
	wantStatus := []image.Rectangle{
		image.Rect(8, screenHeight-22, 104, screenHeight-2),
		image.Rect(112, screenHeight-22, 212, screenHeight-2),
		image.Rect(220, screenHeight-22, 290, screenHeight-2),
		image.Rect(298, screenHeight-22, 412, screenHeight-2),
		image.Rect(420, screenHeight-22, 512, screenHeight-2),
		image.Rect(520, screenHeight-22, 612, screenHeight-2),
		image.Rect(620, screenHeight-22, 872, screenHeight-2),
	}
	for i, want := range wantStatus {
		if status[i].Rect != want {
			t.Fatalf("status rect %d = %#v, want %#v", i, status[i].Rect, want)
		}
	}

	regions := visibleRegionRects()
	wantRegions := map[string]image.Rectangle{
		"canvas":          image.Rect(toolbarWidth, topBarHeight, screenWidth-inspectorWidth, screenHeight-statusHeight),
		"left toolbar":    image.Rect(0, topBarHeight, toolbarWidth, screenHeight-statusHeight),
		"top command bar": image.Rect(toolbarWidth, 0, screenWidth-inspectorWidth, topBarHeight),
		"right inspector": image.Rect(screenWidth-inspectorWidth, topBarHeight, screenWidth, screenHeight-statusHeight),
		"status line":     image.Rect(0, screenHeight-statusHeight, screenWidth, screenHeight),
	}
	for name, want := range wantRegions {
		if got := regions[name]; got != want {
			t.Fatalf("%s region = %#v, want %#v", name, got, want)
		}
	}
	if got := game.visibleRegionPixels("top command bar"); got != regionControlPixels(visibleControls(), "top command bar")+regionControlPixels(inspectorSections(), "top command bar") {
		t.Fatalf("top command pixels = %d", got)
	}
	if got := game.visibleRegionPixels("right inspector"); got != regionControlPixels(visibleControls(), "right inspector")+regionControlPixels(inspectorSections(), "right inspector") {
		t.Fatalf("right inspector pixels = %d", got)
	}
	if got := game.visibleRegionPixels("status line"); got != game.regionStatusPixels("status line") {
		t.Fatalf("status line pixels = %d", got)
	}
	if got := game.regionStatusPixels("status line"); got != rectPixels(status[0].Rect)+rectPixels(status[1].Rect)+rectPixels(status[2].Rect)+rectPixels(status[3].Rect)+rectPixels(status[4].Rect)+rectPixels(status[5].Rect)+rectPixels(status[6].Rect) {
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
	game.pathEntryCommand = "this/status/path/is/far/too/long/to/fit"
	if visibleLabelsFit(game) {
		t.Fatal("long status label should not fit")
	}
}

func TestAppUnitDragMassSnapsToGrid(t *testing.T) {
	game := appUnitGameWithMasses(
		sim.Mass{ID: 1, Position: sim.Vec2{X: 10, Y: 10}, Velocity: sim.Vec2{X: 1, Y: 1}, Mass: 1},
		sim.Mass{ID: 2, Position: sim.Vec2{X: 20, Y: 20}, Velocity: sim.Vec2{X: 2, Y: 2}, Mass: 1},
	)
	game.World().Parameters.Set("grid snap", "10")

	if !game.DragMass(1, sim.Vec2{X: 123, Y: 87}) {
		t.Fatal("single mass drag should succeed")
	}
	mass, _ := game.World().MassByID(1)
	if mass.Position != (sim.Vec2{X: 120, Y: 90}) || mass.Velocity != (sim.Vec2{}) || !game.dirty || !game.dragMoved {
		t.Fatalf("single dragged mass = %#v dirty=%t moved=%t", mass, game.dirty, game.dragMoved)
	}

	game.dirty = false
	game.dragMoved = false
	game.draggingOffsets = map[int]sim.Vec2{1: {X: 2, Y: 3}, 2: {X: -3, Y: -2}}
	_ = game.editing().SelectMass(1)
	_ = game.editing().AddMassSelection(2)
	if !game.DragMass(1, sim.Vec2{X: 13, Y: 14}) {
		t.Fatal("selected mass drag should succeed")
	}
	first, _ := game.World().MassByID(1)
	second, _ := game.World().MassByID(2)
	if first.Position != (sim.Vec2{X: 15, Y: 17}) || second.Position != (sim.Vec2{X: 10, Y: 12}) {
		t.Fatalf("selected dragged masses = %#v %#v", first, second)
	}
	if first.Velocity != (sim.Vec2{}) || second.Velocity != (sim.Vec2{}) || !game.dirty || !game.dragMoved {
		t.Fatalf("selected drag state first=%#v second=%#v dirty=%t moved=%t", first, second, game.dirty, game.dragMoved)
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
	if !screen.HasCommandControl("run") || screen.HasCommandControl("missing") {
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
	game.lastCursor = sim.Vec2{X: 100, Y: 120}
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
	if game.valueDialog.Open {
		t.Fatal("missing mass opened value dialog")
	}
	game.openMassValueDialog(1)
	if !game.valueDialog.Open || game.valueDialog.Title != "Set Mass #1" || game.valueDialog.Text != "5" || game.valueDialog.Min != 0 || game.valueDialog.Max != 20 {
		t.Fatalf("mass value dialog = %#v", game.valueDialog)
	}

	game.valueDialog.Text = "1"
	game.appendValueDialogInput([]rune{'2', 'x', '.', '-', '3'})
	if game.valueDialog.Text != "12.-3" {
		t.Fatalf("dialog text = %q", game.valueDialog.Text)
	}
	game.deleteValueDialogCharacter()
	if game.valueDialog.Text != "12.-" {
		t.Fatalf("dialog text after delete = %q", game.valueDialog.Text)
	}
	game.valueDialog.Text = ""
	game.deleteValueDialogCharacter()
	if game.valueDialog.Text != "" {
		t.Fatalf("empty dialog text after delete = %q", game.valueDialog.Text)
	}
	game.valueDialog.Text = "a"
	game.deleteValueDialogCharacter()
	if game.valueDialog.Text != "" {
		t.Fatalf("single-character dialog text after delete = %q", game.valueDialog.Text)
	}

	game.valueDialog = valueDialog{Open: true, Text: "15", Min: 0, Max: 10}
	if got := game.valueDialogFraction(); got != 1 {
		t.Fatalf("high fraction = %f", got)
	}
	game.valueDialog.Text = "-5"
	if got := game.valueDialogFraction(); got != 0 {
		t.Fatalf("low fraction = %f", got)
	}
	game.valueDialog.Text = "5"
	if got := game.valueDialogFraction(); got != 0.5 {
		t.Fatalf("middle fraction = %f", got)
	}
	game.valueDialog = valueDialog{Text: "10", Min: 5, Max: 15}
	if got := game.valueDialogFraction(); got != 0.5 {
		t.Fatalf("nonzero range fraction = %f", got)
	}
	game.valueDialog.Text = "invalid"
	if got := game.valueDialogFraction(); got != 0 {
		t.Fatalf("invalid fraction = %f", got)
	}
	game.valueDialog = valueDialog{Text: "invalid", Min: -10, Max: 10}
	if got := game.valueDialogFraction(); got != 0 {
		t.Fatalf("invalid nonzero-range fraction = %f", got)
	}
	game.valueDialog = valueDialog{Text: "10", Min: 10, Max: 10}
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
	if track != image.Rect(rect.Min.X+12, rect.Min.Y+92, rect.Max.X-12, rect.Min.Y+100) {
		t.Fatalf("slider track = %#v", track)
	}
	if got := game.valueDialogOKRect(); got != image.Rect(rect.Max.X-64, rect.Max.Y-34, rect.Max.X-12, rect.Max.Y-12) {
		t.Fatalf("ok rect = %#v", got)
	}

	game.valueDialog = valueDialog{Open: true, Text: "0", Min: 0, Max: 20}
	game.setValueDialogFromSlider(track.Min.X + track.Dx()/2)
	if game.valueDialog.Text != "10" {
		t.Fatalf("slider text = %q", game.valueDialog.Text)
	}
	game.setValueDialogFromSlider(track.Min.X - track.Dx())
	if game.valueDialog.Text != "0" {
		t.Fatalf("slider low text = %q", game.valueDialog.Text)
	}
	game.setValueDialogFromSlider(track.Max.X + track.Dx())
	if game.valueDialog.Text != "20" {
		t.Fatalf("slider high text = %q", game.valueDialog.Text)
	}
	game.valueDialog = valueDialog{Open: true, Text: "0", Min: 5, Max: 15}
	game.setValueDialogFromSlider(track.Min.X + track.Dx()/2)
	if game.valueDialog.Text != "10" {
		t.Fatalf("nonzero slider text = %q", game.valueDialog.Text)
	}

	game.valueDialog = valueDialog{Open: true}
	game.clickValueDialog(rect.Max.X+1, rect.Max.Y+1)
	if game.valueDialog.Open {
		t.Fatal("outside click should close dialog")
	}
	applied := 0.0
	game.valueDialog = valueDialog{Open: true, Text: "4", Apply: func(value float64) { applied = value }}
	game.clickValueDialog(game.valueDialogOKRect().Min.X+1, game.valueDialogOKRect().Min.Y+1)
	if applied != 4 || game.valueDialog.Open {
		t.Fatalf("ok click applied=%f dialog=%#v", applied, game.valueDialog)
	}
	game.valueDialog = valueDialog{Open: true, Text: "bad", Apply: func(value float64) { applied = value }}
	game.applyValueDialog()
	if !game.valueDialog.Open || applied != 4 {
		t.Fatalf("invalid apply applied=%f dialog=%#v", applied, game.valueDialog)
	}
	game.valueDialog = valueDialog{Open: true, Text: "8"}
	game.applyValueDialog()
	if game.valueDialog.Open {
		t.Fatal("apply without callback should still close dialog")
	}

	game.valueDialog = valueDialog{Open: true}
	game.tickValueDialog()
	if game.valueDialog.Ticks != 1 || !game.valueDialogCursorVisible() {
		t.Fatalf("dialog ticks=%d cursor=%t", game.valueDialog.Ticks, game.valueDialogCursorVisible())
	}
	game.valueDialog.Ticks = valueCursorPeriod
	if game.valueDialogCursorVisible() {
		t.Fatal("cursor should hide after one period")
	}
	game.valueDialog.Open = false
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
	if !game.valueDialog.Open || game.valueDialog.Title != "Set Spring #3" || game.valueDialog.Text != "12" || game.valueDialog.Min != 0 || game.valueDialog.Max != 50 {
		t.Fatalf("spring value dialog = %#v", game.valueDialog)
	}
	game.valueDialog.Text = "22"
	game.dirty = false
	game.applyValueDialog()
	spring, _ := game.World().SpringByID(3)
	if spring.SpringConstant != 22 || spring.Stiffness != 22 || !game.dirty || game.valueDialog.Open {
		t.Fatalf("spring after apply = %#v dirty=%t dialog=%#v", spring, game.dirty, game.valueDialog)
	}

	game.dirty = false
	game.setSpringConstant(99, 33)
	if game.dirty {
		t.Fatal("missing spring constant update should not mark dirty")
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
	game.demoFiles = []string{first, second}
	game.demoPickerOpen = true
	game.demoPickerScroll = 1
	row := game.demoRowRect(0)
	game.clickDemoPicker(row.Min.X+2, row.Min.Y+2)
	if _, ok := game.World().MassByID(2); !ok || game.demoPickerOpen {
		t.Fatalf("demo picker load failed masses=%#v open=%t", game.World().Masses, game.demoPickerOpen)
	}

	third := filepath.Join(dir, "third.xsp")
	if err := os.WriteFile(third, []byte("#1.0\nmass 3 50 60 1 0\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	game.ReplaceWorld(sim.NewWorld())
	game.demoFiles = []string{first, second, third}
	game.demoPickerOpen = true
	game.demoPickerScroll = 1
	row = game.demoRowRect(1)
	game.clickDemoPicker(row.Min.X+2, row.Min.Y+2)
	if _, ok := game.World().MassByID(3); !ok || game.demoPickerOpen {
		t.Fatalf("scrolled row load failed masses=%#v open=%t", game.World().Masses, game.demoPickerOpen)
	}

	game.demoPickerOpen = true
	game.clickDemoPicker(demoPickerRect().Max.X+1, demoPickerRect().Max.Y+1)
	if game.demoPickerOpen {
		t.Fatal("outside click should close demo picker")
	}
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
	game.demoFiles = []string{"first.xsp", "second.xsp", "third.xsp"}
	game.demoPickerScroll = 0
	visible := game.visibleDemoPaths()
	if len(visible) != 3 || visible[0] != "first.xsp" {
		t.Fatalf("visible demo paths = %#v", visible)
	}
	game.demoFiles = []string{filepath.Join("..", "..", "demos", "pendulum.xsp")}
	game.ReplaceWorld(sim.NewWorld())
	game.demoPickerOpen = true
	game.loadDemoAt(0)
	if _, ok := game.World().MassByID(1); !ok || game.demoPickerOpen {
		t.Fatalf("index 0 load failed masses=%#v open=%t", game.World().Masses, game.demoPickerOpen)
	}
	for _, index := range []int{-1, 1} {
		game.demoFiles = []string{filepath.Join("..", "..", "demos", "pendulum.xsp")}
		game.demoPickerOpen = true
		game.loadDemoAt(index)
		if !game.demoPickerOpen {
			t.Fatalf("out-of-range index %d closed demo picker", index)
		}
	}
	game.demoFiles = []string{"missing.xsp"}
	game.demoPickerOpen = true
	game.loadDemoAt(0)
	if game.demoPickerOpen {
		t.Fatal("failed file read should still close demo picker")
	}
}

func TestAppUnitMassContextMenuActionsAndGeometry(t *testing.T) {
	game := appUnitGameWithMasses(sim.Mass{ID: 1, Position: sim.Vec2{X: 20, Y: 20}, Mass: 3})
	game.canvasYUp = false

	if !game.openMassContextMenu(20, 20) {
		t.Fatal("mass context menu did not open")
	}
	if !game.massMenu.Open || game.massMenu.MassID != 1 || !game.selected {
		t.Fatalf("mass menu = %#v selected=%t", game.massMenu, game.selected)
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
	if !mass.Fixed || game.massMenu.Open || !game.dirty {
		t.Fatalf("fixed toggle mass=%#v menu=%#v dirty=%t", mass, game.massMenu, game.dirty)
	}

	game.massMenu = massContextMenu{Open: true, MassID: 1, X: 20, Y: 20}
	row1 = game.massContextMenuRowRect(1)
	game.clickMassContextMenu(row1.Min.X+1, row1.Min.Y+1)
	if !game.valueDialog.Open || game.valueDialog.Target != "mass" || game.valueDialog.Text != "3" {
		t.Fatalf("value dialog = %#v", game.valueDialog)
	}
	game.valueDialog.Text = "7"
	game.dirty = false
	game.applyValueDialog()
	mass, _ = game.World().MassByID(1)
	if mass.Mass != 7 || !game.dirty {
		t.Fatalf("mass value = %f dirty=%t", mass.Mass, game.dirty)
	}

	game.massMenu = massContextMenu{Open: true, MassID: 1, X: 0, Y: 0}
	anchored := game.massContextMenuRect()
	if anchored.Min != image.Pt(0, 0) {
		t.Fatalf("anchored menu rect = %#v", anchored)
	}

	game.massMenu = massContextMenu{Open: true, MassID: 1, X: screenWidth + 50, Y: screenHeight + 50}
	clamped := game.massContextMenuRect()
	if clamped.Max.X != screenWidth || clamped.Max.Y != screenHeight {
		t.Fatalf("clamped menu rect = %#v", clamped)
	}
	game.clickMassContextMenu(clamped.Max.X+1, clamped.Max.Y+1)
	if game.massMenu.Open {
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

	game.demoPickerOpen = true
	game.openContextAt(10, 10)
	if game.massMenu.Open || game.valueDialog.Open {
		t.Fatal("demo picker should block context menu opening")
	}

	game.demoPickerOpen = false
	game.openContextAt(10, 10)
	if !game.massMenu.Open || game.massMenu.MassID != 1 || game.valueDialog.Open {
		t.Fatalf("mass context = %#v dialog=%#v", game.massMenu, game.valueDialog)
	}

	game.massMenu.Open = true
	game.openContextAt(60, 10)
	if !game.valueDialog.Open || game.valueDialog.Target != "spring" || game.massMenu.Open {
		t.Fatalf("spring dialog = %#v mass menu=%#v", game.valueDialog, game.massMenu)
	}

	game.valueDialog.Open = false
	game.openContextAt(500, 500)
	if game.massMenu.Open || game.valueDialog.Open {
		t.Fatalf("empty context should close overlays menu=%#v dialog=%#v", game.massMenu, game.valueDialog)
	}
}
