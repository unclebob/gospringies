//go:build property

package app

import (
	"fmt"
	"image"
	"math"
	"math/rand"
	"sort"
	"strconv"
	"testing"
	"testing/quick"

	"springs/internal/sim"
)

func TestPropertyCanvasBoundsClampAndSnapStayConsistent(t *testing.T) {
	checkProperty(t, 11, 500, canvasBoundsClampAndSnapStayConsistent)
}

func TestPropertyCanvasCoordinatesRoundTrip(t *testing.T) {
	checkProperty(t, 12, 500, canvasCoordinatesRoundTrip)
}

func TestPropertyMassHitTestingMatchesDrawRadius(t *testing.T) {
	checkProperty(t, 13, 500, massHitTestingMatchesDrawRadius)
}

func TestPropertyDialogRectsStayInsideScreen(t *testing.T) {
	checkProperty(t, 14, 200, dialogRectsStayInsideScreen)
}

func TestPropertyVisibleControlLookupsAndNumericHelpersAreStable(t *testing.T) {
	checkProperty(t, 16, 500, visibleControlLookupsAndNumericHelpersAreStable)
}

func TestPropertyUpdateByIDOnlyUpdatesMatches(t *testing.T) {
	checkProperty(t, 17, 500, updateByIDOnlyUpdatesMatches)
}

func TestPropertyRenderSpringEndpointLookupMatchesValidEndpoints(t *testing.T) {
	checkProperty(t, 18, 500, renderSpringEndpointLookupMatchesValidEndpoints)
}

func TestPropertyValueDialogSliderAndRectsAreBounded(t *testing.T) {
	checkProperty(t, 19, 500, valueDialogSliderAndRectsAreBounded)
}

func TestPropertyDemoPickerGeometryAndClampAreBounded(t *testing.T) {
	checkProperty(t, 20, 500, demoPickerGeometryAndClampAreBounded)
}

func TestPropertyNumericSettingControlsRoundTrip(t *testing.T) {
	checkProperty(t, 21, 500, numericSettingControlsRoundTrip)
}

func TestPropertyClipboardPasteKeepsIDsAndReferencesValid(t *testing.T) {
	checkProperty(t, 22, 500, clipboardPasteKeepsIDsAndReferencesValid)
}

func TestPropertyRenderGeometryHelpersAreDeterministic(t *testing.T) {
	checkProperty(t, 23, 500, renderGeometryHelpersAreDeterministic)
}

func TestPropertyVisibleControlReportsAreInternallyConsistent(t *testing.T) {
	checkProperty(t, 24, 500, visibleControlReportsAreInternallyConsistent)
}

func checkProperty(t *testing.T, seed int64, maxCount int, property any) {
	t.Helper()
	config := &quick.Config{MaxCount: maxCount, Rand: rand.New(rand.NewSource(seed))}
	if err := quick.Check(property, config); err != nil {
		t.Fatal(err)
	}
}

func canvasBoundsClampAndSnapStayConsistent(xInput, yInput, heightInput, gridInput float64, yUp bool) bool {
	game := NewGame()
	game.run.canvasYUp = yUp
	height := propertyFloat(heightInput, 600, 2400)
	game.world.simulation.Bounds.Height = height

	minX, maxX, minY, maxY := game.canvasWorldBounds()
	heightMinX, heightMaxX, heightMinY, heightMaxY := game.canvasWorldBoundsForHeight(height)
	requireApprox("min x", minX, heightMinX)
	requireApprox("max x", maxX, heightMaxX)
	requireApprox("min y", minY, heightMinY)
	requireApprox("max y", maxY, heightMaxY)
	requireOrderedBounds(minX, maxX, minY, maxY)

	world := sim.NewWorld()
	world.Bounds.Height = height
	game.applyCanvasWallBounds(world)
	requireApprox("applied left", world.Bounds.Left, minX)
	requireApprox("applied right", world.Bounds.Right, maxX)
	requireApprox("applied bottom", world.Bounds.Bottom, minY)
	requireApprox("applied top", world.Bounds.Top, maxY)

	position := sim.Vec2{
		X: propertyFloat(xInput, minX-500, maxX+500),
		Y: propertyFloat(yInput, minY-500, maxY+500),
	}
	insideBefore := game.positionInCanvas(position)
	clamped := game.clampToCanvas(position)
	if !game.positionInCanvas(clamped) {
		panic(fmt.Sprintf("clamped position outside canvas: %#v bounds=%v", clamped, []float64{minX, maxX, minY, maxY}))
	}
	if insideBefore {
		requireVecApprox("inside position changed by clamp", clamped, position)
	}

	grid := propertyFloat(gridInput, 1, 100)
	game.world.simulation.Parameters.Set("grid snap", formatControlFloat(grid))
	snapped := game.snapToCanvas(position)
	if !game.positionInCanvas(snapped) {
		panic(fmt.Sprintf("snapped position outside canvas: %#v", snapped))
	}
	return true
}

func canvasCoordinatesRoundTrip(xInput, yInput, heightInput float64, yUp bool) bool {
	game := NewGame()
	game.run.canvasYUp = yUp
	height := propertyFloat(heightInput, 600, 2400)
	game.world.simulation.Bounds.Height = height
	position := sim.Vec2{
		X: propertyFloat(xInput, -500, 2500),
		Y: propertyFloat(yInput, -500, 2500),
	}
	world := game.screenToWorld(position)
	screen := game.worldToScreen(world)
	requireVecApprox("screen/world round trip", screen, position)
	requireVecApprox("canvasCoordinate agrees with screenToWorld", game.canvasCoordinate(position), world)
	flipped := game.flipCanvasY(position)
	requireVecApprox("flipCanvasY is self inverse", game.flipCanvasY(flipped), position)
	if yUp {
		requireApprox("flipped y", flipped.Y, height-position.Y)
	} else {
		requireVecApprox("y-down conversion unchanged", world, position)
	}
	return true
}

func massHitTestingMatchesDrawRadius(xInput, yInput, massInput, dxInput, dyInput float64) bool {
	game := NewGame()
	massValue := propertyFloat(massInput, 0.1, 200)
	position := sim.Vec2{X: propertyFloat(xInput, 100, 900), Y: propertyFloat(yInput, 100, 900)}
	game.world.simulation.Masses = []sim.Mass{{ID: 101, Position: position, Mass: massValue}}
	cx, cy, radius := massDrawCircle(game.world.simulation.Masses[0])
	requireApprox("circle x", float64(cx), position.X)
	requireApprox("circle y", float64(cy), position.Y)
	if radius != float32(sim.MassRadius(game.world.simulation.Masses[0])) {
		panic(fmt.Sprintf("radius = %v, want %v", radius, sim.MassRadius(game.world.simulation.Masses[0])))
	}

	angle := propertyFloat(dxInput, 0, 2*math.Pi)
	distance := propertyFloat(dyInput, 0, float64(radius)*0.99)
	id, ok := game.massAt(position.Add(sim.Vec2{X: math.Cos(angle) * distance, Y: math.Sin(angle) * distance}))
	if !ok || id != 101 {
		panic(fmt.Sprintf("massAt missed point inside radius: id=%d ok=%v radius=%v distance=%v", id, ok, radius, distance))
	}
	_, ok = game.massAt(position.Add(sim.Vec2{X: float64(radius) + 2, Y: 0}))
	if ok {
		panic("massAt hit point outside radius")
	}
	return true
}

func dialogRectsStayInsideScreen(widthInput, heightInput float64) bool {
	game := NewGame()
	width := int(propertyFloat(widthInput, 80, screenWidth))
	height := int(propertyFloat(heightInput, 80, screenHeight))
	rects := []image.Rectangle{
		centeredDialogRect(width, height),
		dialogTextRect(centeredDialogRect(width, height)),
		dialogOKRect(centeredDialogRect(width, height)),
		saveFilenameDialogRect(),
		game.saveFilenameTextRect(),
		game.saveFilenameDialogOKRect(),
		valueDialogRect(),
		game.valueDialogTextRect(),
	}
	for _, rect := range rects {
		requireRectInsideScreen(rect)
		if rect.Empty() {
			panic(fmt.Sprintf("empty dialog rect: %#v", rect))
		}
	}
	if !game.saveFilenameTextRect().In(saveFilenameDialogRect()) {
		panic("save filename text rect outside dialog")
	}
	if !game.saveFilenameDialogOKRect().In(saveFilenameDialogRect()) {
		panic("save filename ok rect outside dialog")
	}
	if !game.valueDialogTextRect().In(valueDialogRect()) {
		panic("value text rect outside dialog")
	}
	return true
}

func visibleControlLookupsAndNumericHelpersAreStable(valueInput, minInput, maxInput, xInput, gridInput float64) bool {
	game := NewGame()
	controls := game.visibleControls()
	if len(controls) == 0 {
		panic("no visible controls")
	}
	index := int(propertyFloat(valueInput, 0, float64(len(controls))))
	control := controls[index]
	uniqueLabel := control.Label != "" && controlLabelCount(controls, control.Label) == 1
	if got, ok := game.VisibleControlBounds(control.Label); uniqueLabel && (!ok || got != control.Rect) {
		panic(fmt.Sprintf("VisibleControlBounds(%q) = %#v, %v; want %#v", control.Label, got, ok, control.Rect))
	}
	if found, ok := game.visibleControlWithLabel(control.Label); uniqueLabel && (!ok || found != control) {
		panic("game visibleControlWithLabel failed")
	}
	if found, ok := game.visibleControlWithName(control.Name); !ok || found != control {
		panic("game visibleControlWithName failed")
	}
	if found, ok := game.visibleControlWithField(control.Name, func(control controlBox) string { return control.Name }); !ok || found != control {
		panic("game visibleControlWithField did not return first matching field")
	}
	center := image.Pt((control.Rect.Min.X+control.Rect.Max.X)/2, (control.Rect.Min.Y+control.Rect.Max.Y)/2)
	if found, ok := game.visibleControlAt(center); !ok || found != control {
		panic("game visibleControlAt failed")
	}
	if found, ok := controlAt(center, controls); !ok || found != control {
		panic("controlAt failed")
	}

	global := visibleControls()
	globalControl := global[0]
	globalCenter := image.Pt((globalControl.Rect.Min.X+globalControl.Rect.Max.X)/2, (globalControl.Rect.Min.Y+globalControl.Rect.Max.Y)/2)
	if found, ok := visibleControlAt(globalCenter); !ok || found != globalControl {
		panic("package visibleControlAt failed")
	}
	if found, ok := visibleControlWithLabel(globalControl.Label); globalControl.Label != "" && controlLabelCount(global, globalControl.Label) == 1 && (!ok || found != globalControl) {
		panic("package visibleControlWithLabel failed")
	}
	if found, ok := visibleControlWithName(globalControl.Name); !ok || found != globalControl {
		panic("package visibleControlWithName failed")
	}
	if found, ok := visibleControlWithField(globalControl.Name, func(control controlBox) string { return control.Name }); !ok || found != globalControl {
		panic("package visibleControlWithField failed")
	}

	minValue := propertyFloat(minInput, -100, 0)
	maxValue := propertyFloat(maxInput, 0.001, 100)
	value := propertyFloat(valueInput, minValue-100, maxValue+100)
	clamped := clampFloat(value, minValue, maxValue)
	if clamped < minValue || clamped > maxValue {
		panic(fmt.Sprintf("clampFloat out of range: %v not in [%v,%v]", clamped, minValue, maxValue))
	}
	rounded := roundControlFloat(value)
	formatted := formatControlFloat(value)
	parsed, err := strconv.ParseFloat(formatted, 64)
	if err != nil {
		panic(err)
	}
	requireApprox("formatControlFloat", parsed, rounded)

	track := image.Rect(10, 0, 110, 20)
	fraction := sliderFractionAt(track, int(propertyFloat(xInput, -50, 170)))
	if fraction < 0 || fraction > 1 {
		panic(fmt.Sprintf("slider fraction out of range: %v", fraction))
	}
	if sliderFractionAt(image.Rect(0, 0, 0, 10), 10) != 0 {
		panic("zero-width slider fraction should be zero")
	}

	game.world.simulation.Parameters.Set("grid snap", formatControlFloat(propertyFloat(gridInput, 1, 40)))
	snapped := game.snapToGrid(sim.Vec2{X: value, Y: value + 0.25})
	size := game.gridSnapSize()
	requireApprox("snapToGrid x", math.Round(snapped.X/size)*size, snapped.X)
	requireApprox("snapToGrid y", math.Round(snapped.Y/size)*size, snapped.Y)
	if game.parameterFloat("grid snap") != size {
		panic("parameterFloat/gridSnapSize mismatch")
	}
	if game.parameterForEditorControl("missing") != 0 {
		panic("missing editor control parameter should be zero")
	}

	force := sim.ForceConfig{Values: map[string]string{"magnitude": formatted}}
	requireApprox("forceValueFloat", forceValueFloat(force, "magnitude"), parsed)
	if nonNilStringMap(nil) == nil {
		panic("nonNilStringMap returned nil")
	}
	game.world.simulation.Parameters.Forces["gravity"] = sim.ForceConfig{Values: nil}
	if game.forceConfig("gravity").Values == nil {
		panic("forceConfig did not normalize nil Values")
	}
	game.editing().SelectedMasses = map[int]bool{3: true, 1: true, 2: false}
	ids := game.selectedMassIDs()
	sort.Ints(ids)
	if fmt.Sprint(ids) != "[1 3]" {
		panic(fmt.Sprintf("selectedMassIDs = %v", ids))
	}
	return true
}

func updateByIDOnlyUpdatesMatches(idInput float64) bool {
	type item struct {
		id    int
		value int
	}
	items := []item{{id: 1, value: 10}, {id: 2, value: 20}, {id: 3, value: 30}}
	id := int(propertyFloat(idInput, 1, 5))
	updated := updateByID(items, id, func(item *item) int { return item.id }, func(item *item) { item.value += 100 })
	for _, item := range items {
		if item.id == id {
			if !updated || item.value < 100 {
				panic(fmt.Sprintf("matching item not updated: %#v updated=%v", items, updated))
			}
			continue
		}
		if item.value >= 100 {
			panic(fmt.Sprintf("nonmatching item updated: %#v", items))
		}
	}
	game := NewGame()
	game.clearDirty()
	updateByIDAndMarkDirty(game, items, id, func(item *item) int { return item.id }, func(item *item) { item.value++ })
	if game.editState.dirty != updated {
		panic(fmt.Sprintf("dirty = %v, want %v", game.editState.dirty, updated))
	}
	return true
}

func renderSpringEndpointLookupMatchesValidEndpoints(idMode bool, invalid bool) bool {
	game := NewGame()
	game.world.simulation.Masses = []sim.Mass{
		{ID: 10, Position: sim.Vec2{X: 1, Y: 2}, Mass: 1},
		{ID: 20, Position: sim.Vec2{X: 3, Y: 4}, Mass: 2},
	}
	var spring sim.Spring
	if idMode {
		spring = sim.Spring{MassA: 10, MassB: 20}
		if invalid {
			spring.MassB = 99
		}
	} else {
		spring = sim.Spring{A: 0, B: 1}
		if invalid {
			spring.B = 99
		}
	}
	a, b, ok := game.springEndpoints(spring)
	if invalid {
		if ok || game.validSpring(spring) {
			panic("invalid spring endpoints reported valid")
		}
		return true
	}
	if !ok || !game.validSpring(spring) {
		panic("valid spring endpoints rejected")
	}
	if a.ID != 10 || b.ID != 20 {
		panic(fmt.Sprintf("endpoints = %#v %#v", a, b))
	}
	if !validSpringIndex(0, game.world.simulation.Masses) || validSpringIndex(-1, game.world.simulation.Masses) || validSpringIndex(2, game.world.simulation.Masses) {
		panic("validSpringIndex boundary mismatch")
	}
	return true
}

func valueDialogSliderAndRectsAreBounded(xInput, textInput, minInput, maxInput, deltaInput float64) bool {
	game := NewGame()
	minValue := propertyFloat(minInput, -50, 0)
	maxValue := propertyFloat(maxInput, 0.001, 100)
	game.overlays.value = valueDialog{
		Open: true,
		Min:  minValue,
		Max:  maxValue,
		Text: formatControlFloat(propertyFloat(textInput, minValue-100, maxValue+100)),
	}
	rect := valueDialogRect()
	for _, child := range []image.Rectangle{
		game.valueDialogSliderTrack(),
		game.valueDialogDecrementRect(),
		game.valueDialogIncrementRect(),
		game.valueDialogOKRect(),
	} {
		if !child.In(rect) || child.Empty() {
			panic(fmt.Sprintf("value dialog child rect invalid: %#v in %#v", child, rect))
		}
	}
	track := game.valueDialogSliderTrack()
	game.setValueDialogFromSlider(int(propertyFloat(xInput, float64(track.Min.X-100), float64(track.Max.X+100))))
	value, err := strconv.ParseFloat(game.overlays.value.Text, 64)
	if err != nil {
		panic(err)
	}
	if value < minValue-1e-6 || value > maxValue+1e-6 {
		panic(fmt.Sprintf("slider value out of range: %v not in [%v,%v]", value, minValue, maxValue))
	}
	fraction := game.valueDialogFraction()
	if fraction < 0 || fraction > 1 {
		panic(fmt.Sprintf("value dialog fraction out of range: %v", fraction))
	}
	game.stepValueDialog(propertyFloat(deltaInput, -200, 200))
	value, err = strconv.ParseFloat(game.overlays.value.Text, 64)
	if err != nil {
		panic(err)
	}
	if value < minValue-1e-6 || value > maxValue+1e-6 {
		panic(fmt.Sprintf("stepped value out of range: %v", value))
	}
	a := sim.Vec2{X: minValue, Y: maxValue}
	b := sim.Vec2{X: maxValue + 10, Y: minValue - 10}
	distanceAB := distanceToSegment(a, a, b)
	requireApprox("distance to segment endpoint", distanceAB, 0)
	p := sim.Vec2{X: propertyFloat(xInput, -100, 100), Y: propertyFloat(textInput, -100, 100)}
	requireApprox("distance endpoint reversal", distanceToSegment(p, a, b), distanceToSegment(p, b, a))
	return true
}

func demoPickerGeometryAndClampAreBounded(valueInput, lowerInput, upperInput float64) bool {
	rect := demoPickerRect()
	requireRectInsideScreen(rect)
	if demoPickerVisibleRows() <= 0 {
		panic("demo picker visible rows must be positive")
	}
	game := NewGame()
	game.controls.demoFiles = groupedLoadPickerEntries([]string{"save.xsp"}, []string{"demo.xsp"}, []string{"original.xsp"})
	game.controls.demoPickerScroll = int(propertyFloat(valueInput, -10, 10))
	visible := game.visibleDemoPaths()
	if len(visible) > demoPickerVisibleRows() || len(visible) > len(game.controls.demoFiles) {
		panic(fmt.Sprintf("visibleDemoPaths length = %d", len(visible)))
	}
	entries := game.LoadPickerEntries()
	if len(entries) != len(game.controls.demoFiles) {
		panic("LoadPickerEntries length mismatch")
	}
	entries[0] = "mutated"
	if game.controls.demoFiles[0] == "mutated" {
		panic("LoadPickerEntries aliases demo list")
	}
	if path, ok := game.demoPathAt(0); !ok || path != "save.xsp" {
		panic(fmt.Sprintf("demoPathAt valid = %q %v", path, ok))
	}
	if path, ok := game.demoPathAt(1); ok || path != "" {
		panic("demoPathAt accepted separator")
	}
	if globXSP("definitely-missing-demo-dir") != nil {
		panic("globXSP missing dir should return nil")
	}
	_ = NewGame().buildDemoList()
	rowIndex := int(propertyFloat(valueInput, 0, float64(demoPickerVisibleRows())))
	row := game.demoRowRect(rowIndex)
	if !row.In(rect) || row.Empty() {
		panic(fmt.Sprintf("demo row outside picker: %#v in %#v", row, rect))
	}
	lower := int(propertyFloat(lowerInput, -100, 0))
	upper := int(propertyFloat(upperInput, 1, 100))
	if lower > upper {
		lower, upper = upper, lower
	}
	value := int(propertyFloat(valueInput, float64(lower-100), float64(upper+100)))
	clamped := clampInt(value, lower, upper)
	if clamped < lower || clamped > upper {
		panic(fmt.Sprintf("clampInt out of range: %d not in [%d,%d]", clamped, lower, upper))
	}
	if !loadPickerEntryMatches("saves/example.xsp", "example.xsp") || !loadPickerEntryMatches("saves/example.xsp", "saves/example.xsp") {
		panic("loadPickerEntryMatches rejected path or base name")
	}
	if loadPickerEntryMatches(loadPickerSeparator, loadPickerSeparator) {
		panic("loadPickerEntryMatches accepted separator")
	}
	grouped := groupedLoadPickerEntries([]string{"b"}, []string{"a"}, []string{"c"})
	if fmt.Sprint(grouped) != "[b separator a c]" {
		panic(fmt.Sprintf("groupedLoadPickerEntries = %v", grouped))
	}
	return true
}

func numericSettingControlsRoundTrip(valueInput, xInput float64) bool {
	game := NewGame()
	controls := numericSettingControls()
	if len(controls) == 0 {
		panic("numericSettingControls returned no controls")
	}
	seenNames := map[string]bool{}
	for _, control := range controls {
		if control.Name == "" || seenNames[control.Name] {
			panic(fmt.Sprintf("invalid or duplicate numeric control name: %q", control.Name))
		}
		seenNames[control.Name] = true
		if control.Region != "right inspector" || control.Rect.Empty() || !control.Rect.In(inspectorRect()) {
			panic(fmt.Sprintf("numeric control has invalid region/rect: %#v", control))
		}
	}
	for _, setting := range numericSettings {
		checkbox, label, decrement, slider, increment, text := numericSettingRects(setting)
		if numericSettingToggleControl(setting) == "" && !checkbox.Empty() {
			panic("setting without toggle has checkbox rect")
		}
		for _, rect := range []image.Rectangle{label, decrement, slider, increment, text} {
			if rect.Empty() || !rect.In(inspectorRect()) {
				panic(fmt.Sprintf("numeric setting rect invalid: %#v for %#v", rect, setting))
			}
		}
		if numericControlName(setting.Name, "slider") == "" {
			panic("numericControlName returned empty name")
		}
		if got, ok := numericSettingForSlider(numericControlName(setting.Name, "slider")); !ok || got.Name != setting.Name {
			panic("numericSettingForSlider failed")
		}
		if got, delta, ok := numericSettingForStepButton(numericControlName(setting.Name, "decrement")); !ok || got.Name != setting.Name || delta >= 0 {
			panic("numericSettingForStepButton decrement failed")
		}
		if got, delta, ok := numericSettingForStepButton(numericControlName(setting.Name, "increment")); !ok || got.Name != setting.Name || delta <= 0 {
			panic("numericSettingForStepButton increment failed")
		}
		if got, ok := numericSettingForTextField(numericControlName(setting.Name, "text field")); !ok || got.Name != setting.Name {
			panic("numericSettingForTextField failed")
		}
		if got, ok := numericSettingForControl(numericControlName(setting.Name, "label"), "label"); !ok || got.Name != setting.Name {
			panic("numericSettingForControl failed")
		}
		if got, ok := numericSettingByName(setting.Name); !ok || got.Name != setting.Name {
			panic("numericSettingByName failed")
		}

		rawValue := formatControlFloat(propertyFloat(valueInput, setting.Min, setting.Max))
		if setting.Speed {
			game.run.simulationSpeed, _ = strconv.ParseFloat(rawValue, 64)
		} else if setting.Force != "" {
			game.setForceValue(setting.Force, setting.ForceKey, propertyFloat(valueInput, setting.Min, setting.Max))
		} else {
			game.world.simulation.Parameters.Set(setting.Parameter, rawValue)
		}
		textValue := game.numericSettingValueText(setting)
		committed := game.committedNumericSettingValueText(setting)
		if textValue != committed {
			panic("unfocused numeric text should equal committed text")
		}
		if game.rawNumericSettingValue(setting) == "" && !setting.Speed {
			panic("raw numeric setting value should not be empty after setup")
		}
		fraction := game.numericSettingSliderFraction(setting)
		if fraction < 0 || fraction > 1 {
			panic(fmt.Sprintf("numeric slider fraction out of range: %v", fraction))
		}
		game.setNumericSettingFromSlider(setting, int(propertyFloat(xInput, float64(slider.Min.X-100), float64(slider.Max.X+100))))
		game.stepNumericSetting(setting, propertyFloat(valueInput, -1000, 1000))
		sliderValue := formatNumericSettingSliderValue(propertyFloat(valueInput, setting.Min, setting.Max), setting.Decimals)
		if _, err := strconv.ParseFloat(sliderValue, 64); err != nil {
			panic(fmt.Sprintf("slider value is not numeric: %q", sliderValue))
		}
		if _, ok := game.NumericSettingReport(setting.Name); !ok {
			panic("NumericSettingReport rejected valid setting")
		}
		if _, ok := game.NumericSettingText(setting.Name); !ok {
			panic("NumericSettingText rejected valid setting")
		}
		if _, ok := game.NumericSettingSliderValue(setting.Name); !ok {
			panic("NumericSettingSliderValue rejected valid setting")
		}
		if validateNumericSetting(setting.Name) != nil {
			panic("validateNumericSetting rejected valid setting")
		}
	}
	if formatNumericSettingText("1", 2) != "1.00" {
		panic("formatNumericSettingText did not apply decimals")
	}
	if !isNumericInputCharacter('5') || isNumericInputCharacter('x') {
		panic("isNumericInputCharacter mismatch")
	}
	if _, ok := game.NumericSettingReport("missing"); ok {
		panic("NumericSettingReport accepted missing setting")
	}
	return true
}

func clipboardPasteKeepsIDsAndReferencesValid(xInput, yInput float64) bool {
	game := NewGame()
	game.world.simulation.Masses = []sim.Mass{
		{ID: 3, Position: sim.Vec2{X: 10, Y: 20}, Mass: 1},
		{ID: 7, Position: sim.Vec2{X: 30, Y: 40}, Mass: 2},
	}
	game.world.simulation.Springs = []sim.Spring{{ID: 5, A: 0, B: 1, MassA: 3, MassB: 7, SpringConstant: 1}}
	game.editing().SelectedMasses = map[int]bool{3: true, 7: true}
	game.editing().SelectedSprings = map[int]bool{}
	game.copySelection()
	if len(game.editState.clipboard.Masses) != 2 || len(game.editState.clipboard.Springs) != 1 {
		panic(fmt.Sprintf("copySelection clipboard = %#v", game.editState.clipboard))
	}
	origin := game.editState.clipboard.origin()
	requireVecApprox("clipboard origin", origin, sim.Vec2{X: 10, Y: 20})
	if game.nextMassID() != 8 || game.nextSpringID() != 6 {
		panic(fmt.Sprintf("next ids = %d %d", game.nextMassID(), game.nextSpringID()))
	}
	if nextID([]int{10, 2, 11}, func(value int) int { return value }) != 12 {
		panic("nextID failed")
	}
	target := sim.Vec2{X: propertyFloat(xInput, 100, 300), Y: propertyFloat(yInput, 100, 300)}
	if !game.pasteSelectionAt(target) {
		panic("pasteSelectionAt failed")
	}
	if len(game.world.simulation.Masses) != 4 || len(game.world.simulation.Springs) != 2 {
		panic(fmt.Sprintf("paste counts masses=%d springs=%d", len(game.world.simulation.Masses), len(game.world.simulation.Springs)))
	}
	pastedMassIDs := map[int]bool{}
	for _, mass := range game.world.simulation.Masses[2:] {
		if mass.ID <= 7 || pastedMassIDs[mass.ID] {
			panic(fmt.Sprintf("invalid pasted mass id: %#v", mass))
		}
		pastedMassIDs[mass.ID] = true
		if !game.positionInCanvas(mass.Position) {
			panic(fmt.Sprintf("pasted mass outside canvas: %#v", mass))
		}
	}
	pastedSpring := game.world.simulation.Springs[1]
	if !pastedMassIDs[pastedSpring.MassA] || !pastedMassIDs[pastedSpring.MassB] {
		panic(fmt.Sprintf("pasted spring endpoints invalid: %#v ids=%v", pastedSpring, pastedMassIDs))
	}
	empty := NewGame()
	if empty.pasteSelectionAt(target) {
		panic("empty clipboard pasted")
	}
	return true
}

func renderGeometryHelpersAreDeterministic(xInput, yInput, endXInput, endYInput, massInput, gridInput float64, selected bool) bool {
	game := NewGame()
	grid := propertyFloat(gridInput, 2, 80)
	game.world.simulation.Parameters.Set("grid snap", formatControlFloat(grid))
	if !validGridSnapSize(grid) || validGridSnapSize(0) {
		panic("validGridSnapSize boundary mismatch")
	}
	canvas := visibleRegionRects()["canvas"]
	first := firstGridCoordinateAtOrAfter(float64(canvas.Min.X), grid)
	if first < float64(canvas.Min.X) || first-grid >= float64(canvas.Min.X) {
		panic(fmt.Sprintf("first grid coordinate = %v, min = %v, grid = %v", first, canvas.Min.X, grid))
	}
	points := game.gridPoints()
	rects := game.gridPointRects()
	if len(points) == 0 || len(points) != len(rects) {
		panic(fmt.Sprintf("grid points/rects mismatch: %d %d", len(points), len(rects)))
	}
	for i, point := range points {
		if point.X < float64(canvas.Min.X) || point.X >= float64(canvas.Max.X) {
			panic(fmt.Sprintf("grid point outside x bounds: %#v", point))
		}
		if i > 0 && point.Y < points[i-1].Y {
			panic("grid points are not monotonic by row")
		}
		rect := rects[i]
		if rect.width != gridPointPixelSize() || rect.height != gridPointPixelSize() || rect.color != gridPointColor || rect.antiAlias != gridPointAntiAlias() {
			panic(fmt.Sprintf("grid rect draw metadata mismatch: %#v", rect))
		}
	}

	start := sim.Vec2{X: propertyFloat(xInput, -200, 800), Y: propertyFloat(yInput, -200, 800)}
	end := sim.Vec2{X: propertyFloat(endXInput, -200, 800), Y: propertyFloat(endYInput, -200, 800)}
	requireClosedRectangle("selection rectangle", selectionRectangleLines(start, end), start, end)
	requireClosedRectangle("selection rectangle reversed", selectionRectangleLines(end, start), start, end)

	mass := sim.Mass{ID: 1, Position: start, Mass: propertyFloat(massInput, 0.1, 100)}
	outline := selectionOutline(mass)
	requireClosedRectangle("selection outline", outline, sim.Vec2{X: outline[0].x1, Y: outline[0].y1}, sim.Vec2{X: outline[1].x2, Y: outline[1].y2})
	if len(selectedMassOutline([]sim.Mass{mass, {ID: 2, Position: end, Mass: 2}})) != 8 {
		panic("selectedMassOutline did not create four lines per mass")
	}

	game.world.simulation.Masses = []sim.Mass{mass, {ID: 2, Position: end, Mass: 2}}
	game.world.simulation.Springs = []sim.Spring{{ID: 3, MassA: 1, MassB: 2}}
	game.editing().SelectedMasses[mass.ID] = selected
	game.editing().SelectedSprings[3] = true
	if len(game.selectedSpringLines()) != 1 {
		panic("selectedSpringLines missing selected spring")
	}
	explicit := game.explicitSelectedMasses()
	if selected && len(explicit) != 1 {
		panic(fmt.Sprintf("explicit selected masses = %v", explicit))
	}
	if !selected && len(explicit) != 0 {
		panic(fmt.Sprintf("unexpected explicit selected masses = %v", explicit))
	}
	game.editing().SelectedSprings = map[int]bool{}
	game.editing().SelectedMasses = map[int]bool{}
	game.editState.selected = true
	if !game.allMassesImplicitlySelected() || len(game.selectedMasses()) != len(game.world.simulation.Masses) {
		panic("implicit all-mass selection failed")
	}

	game.pointer.pendingSpringID = 1
	game.pointer.pendingSpringEnd = end
	line, ok := game.pendingSpringLine()
	if !ok || line.x1 != start.X || line.y1 != start.Y || line.x2 != end.X || line.y2 != end.Y {
		panic(fmt.Sprintf("pendingSpringLine = %#v %v", line, ok))
	}
	if springDrawColor(sim.Spring{Wall: true}) != wallSpringColor || springDrawColor(sim.Spring{}) != springColor {
		panic("springDrawColor mismatch")
	}
	if massDrawColor(sim.Mass{Fixed: true}) != fixedMassColor || massDrawColor(sim.Mass{}) != massColor {
		panic("massDrawColor mismatch")
	}
	if drawColorFor(true, wallColor, massColor) != wallColor || drawColorFor(false, wallColor, massColor) != massColor {
		panic("drawColorFor mismatch")
	}
	walls := wallDrawLines(game.world.simulation.Bounds)
	if len(walls) != 4 {
		panic(fmt.Sprintf("wallDrawLines count = %d", len(walls)))
	}
	picker := demoPickerRect()
	if !demoPickerTitlePoint(picker).In(picker) {
		panic("demoPickerTitlePoint outside picker")
	}
	row := game.demoRowRect(0)
	if !demoPickerRowTextPoint(row).In(row) {
		panic("demoPickerRowTextPoint outside row")
	}
	if demoPickerRowFill(0) != controlColor || demoPickerRowFill(1) != sectionColor {
		panic("demoPickerRowFill parity mismatch")
	}
	return true
}

func visibleControlReportsAreInternallyConsistent(massCountInput, springCountInput, scrollInput float64, paused bool, editMenuOpen bool, selectedSpringWall bool) bool {
	game := NewGame()
	game.run.paused = paused
	game.controls.editMenuOpen = editMenuOpen
	game.controls.demoPickerScroll = int(propertyFloat(scrollInput, -5, 5))
	game.controls.activeNumericStep = "mass decrement"
	game.document.currentFilePath = "saves/current.xsp"

	massCount := int(propertyFloat(massCountInput, 1, 5))
	springCount := int(propertyFloat(springCountInput, 0, float64(massCount)))
	game.world.simulation.Masses = nil
	for i := 0; i < massCount; i++ {
		game.world.simulation.Masses = append(game.world.simulation.Masses, sim.Mass{
			ID:       i + 1,
			Position: sim.Vec2{X: 100 + float64(i*20), Y: 120 + float64(i*10)},
			Mass:     float64(i + 1),
			Fixed:    i%2 == 0,
		})
	}
	game.world.simulation.Springs = nil
	for i := 0; i < springCount; i++ {
		a := i % massCount
		b := (i + 1) % massCount
		game.world.simulation.Springs = append(game.world.simulation.Springs, sim.Spring{
			ID:    i + 10,
			A:     a,
			B:     b,
			MassA: game.world.simulation.Masses[a].ID,
			MassB: game.world.simulation.Masses[b].ID,
			Wall:  selectedSpringWall,
		})
		game.editing().SelectedSprings[i+10] = selectedSpringWall
	}
	game.editing().SelectedMasses[1] = true
	game.world.simulation.Parameters.EnableForce("gravity", map[string]string{"magnitude": "10"})
	game.world.simulation.Parameters.Set("fixed mass", "true")
	game.world.simulation.Parameters.Set("show springs", "true")
	game.world.simulation.Parameters.Set("adaptive timestep", "true")
	game.world.simulation.Parameters.Set("grid snap", "10")
	game.world.simulation.Parameters.EnableWall("top")

	allControls := game.visibleControls()
	if len(visibleControls()) == 0 || len(menuControls()) == 0 || len(toolbarControls()) == 0 || len(game.commandControls()) == 0 || len(inspectorControls()) == 0 || len(allControls) == 0 {
		panic("expected visible/menu/toolbar/command/inspector controls")
	}
	if editMenuOpen != (len(game.editMenuControls()) > 0) {
		panic("edit menu visibility mismatch")
	}
	for _, control := range allControls {
		if control.Name == "" && control.Label == "" {
			panic(fmt.Sprintf("control missing name and label: %#v", control))
		}
		if control.Rect.Empty() {
			panic(fmt.Sprintf("control has empty rect: %#v", control))
		}
		if _, ok := visibleRegionRects()[control.Region]; !ok {
			panic(fmt.Sprintf("control has unknown region: %#v", control))
		}
		if isSliderControl(control.Name) {
			track := sliderTrack(control)
			if track.Empty() || !track.In(control.Rect) {
				panic(fmt.Sprintf("slider track invalid: %#v in %#v", track, control.Rect))
			}
			fraction := game.sliderFraction(control.Name)
			if fraction < 0 || fraction > 1 {
				panic(fmt.Sprintf("slider fraction out of range: %v", fraction))
			}
		}
		if game.sliderLabel(control) != control.Label {
			panic("sliderLabel did not return control label")
		}
	}
	if !game.activeRunControl("run pause toggle command") || game.activeRunControl("missing") {
		panic("activeRunControl mismatch")
	}
	if !game.activeForceControl("gravity force") || game.activeForceControl("missing") {
		panic("activeForceControl mismatch")
	}
	if !game.activeParameterControl("fixed mass toggle") || !game.activeParameterControl("show springs toggle") || !game.activeParameterControl("adaptive timestep toggle") || !game.activeParameterControl("grid snap toggle") || game.activeParameterControl("missing") {
		panic("activeParameterControl mismatch")
	}
	if game.activeSelectedSpringControl("spring wall toggle") != (selectedSpringWall && springCount > 0) || game.activeSelectedSpringControl("missing") {
		panic("activeSelectedSpringControl mismatch")
	}
	if !game.activeWallControl("top wall toggle") || game.activeWallControl("missing") {
		panic("activeWallControl mismatch")
	}
	if !game.activeControl("run pause toggle command") || !game.activeControl("gravity force") || !game.activeControl("mass decrement") {
		panic("activeControl did not compose active states")
	}
	if game.runPauseToggleLabel() != map[bool]string{true: "Run", false: "Pause"}[paused] {
		panic("runPauseToggleLabel mismatch")
	}
	if sectionHeaderLabel("Forces") != "----- Forces -----" || inspectorLeft() != screenWidth-inspectorWidth {
		panic("section header or inspector left mismatch")
	}
	if !game.forceEnabled("gravity") || !game.parameterEnabled("fixed mass") || !game.wallEnabled("top") || !game.gridSnapEnabled() {
		panic("enabled state helpers mismatch")
	}
	if game.objectCountsStatusLabel() != fmt.Sprintf("Masses: %d", massCount) {
		panic("objectCountsStatusLabel mismatch")
	}
	if game.currentFileStatusLabel() != game.document.currentFilePath {
		panic("currentFileStatusLabel should use current file path")
	}
	game.document.pathEntryCommand = "Save As"
	if game.currentFileStatusLabel() != "Save As" {
		panic("currentFileStatusLabel should prefer path entry command")
	}
	if game.selectedObjectCount() < 1 {
		panic("selectedObjectCount should include selected mass")
	}

	status := game.statusFields()
	if len(status) != 3 {
		panic(fmt.Sprintf("statusFields count = %d", len(status)))
	}
	for _, field := range status {
		if field.Name == "" || field.Label == "" || field.Rect.Empty() {
			panic(fmt.Sprintf("invalid status field: %#v", field))
		}
	}
	active := game.visibleActiveControls()
	labels := game.visibleControlLabels()
	sections := visibleInspectorSections()
	sectionRects := visibleInspectorSectionRects()
	statusMap := game.visibleStatusFields()
	counts := game.visibleRegionControlCounts()
	if len(active) == 0 || len(labels) == 0 || len(sections) == 0 || len(sectionRects) != len(sections) || len(statusMap) != len(status) || len(counts) == 0 {
		panic("visible report component maps should be populated")
	}
	for label := range sections {
		if _, ok := sectionRects[label]; !ok {
			panic("missing section rect")
		}
	}
	if !visibleLabelsFit(game) || !controlLabelsFit(game.visibleControls()) || !controlLabelsFit(inspectorSections()) || !statusLabelsFit(status) {
		panic("expected labels to fit visible controls")
	}
	if !labelsFitItems([]controlBox{{Label: "OK", Rect: image.Rect(0, 0, 40, 20)}}, func(box controlBox) (string, image.Rectangle) { return box.Label, box.Rect }) {
		panic("labelsFitItems rejected fitting label")
	}
	if !labelFits("OK", image.Rect(0, 0, 40, 20)) || labelFits("too long", image.Rect(0, 0, 12, 10)) {
		panic("labelFits boundary mismatch")
	}
	regions := visibleRegionRects()
	for name, rect := range regions {
		if rect.Empty() || rectPixels(rect) <= 0 {
			panic(fmt.Sprintf("invalid visible region rect: %s %#v", name, rect))
		}
		pixels := game.visibleRegionPixels(name)
		if pixels <= 0 {
			panic(fmt.Sprintf("visibleRegionPixels invalid: %s %d", name, pixels))
		}
	}
	if regionControlPixels([]controlBox{{Region: "x", Rect: image.Rect(0, 0, 2, 3)}}, "x") != 6 {
		panic("regionControlPixels mismatch")
	}
	if game.regionStatusPixels("canvas") != 0 || game.regionStatusPixels("right inspector") <= 0 {
		panic("regionStatusPixels mismatch")
	}
	worldPixels := visibleWorldPixels(game)
	if worldPixels < len(game.world.simulation.Springs)*50 {
		panic(fmt.Sprintf("visibleWorldPixels too small: %d", worldPixels))
	}
	report := game.DrawFrameReport()
	report2 := analyzeDrawnFrame(game)
	if report.CanvasWorldPixels != worldPixels || report2.CanvasWorldPixels != worldPixels || len(report.RegionPixels) != len(regions) || !report.ControlLabelsFit {
		panic("draw frame report mismatch")
	}
	return true
}

func propertyFloat(input float64, minimum float64, maximum float64) float64 {
	if math.IsNaN(input) || math.IsInf(input, 0) {
		return minimum
	}
	return minimum + math.Mod(math.Abs(input), maximum-minimum)
}

func requireOrderedBounds(minX, maxX, minY, maxY float64) {
	if !(minX < maxX && minY < maxY) {
		panic(fmt.Sprintf("invalid bounds: %v %v %v %v", minX, maxX, minY, maxY))
	}
}

func requireVecApprox(label string, got sim.Vec2, want sim.Vec2) {
	requireApprox(label+" x", got.X, want.X)
	requireApprox(label+" y", got.Y, want.Y)
}

func requireApprox(label string, got float64, want float64) {
	const tolerance = 1e-9
	if math.Abs(got-want) > tolerance {
		panic(fmt.Sprintf("%s = %.12f, want %.12f", label, got, want))
	}
}

func requireRectInsideScreen(rect image.Rectangle) {
	screen := image.Rect(0, 0, screenWidth, screenHeight)
	if !rect.In(screen) {
		panic(fmt.Sprintf("rect outside screen: %#v", rect))
	}
}

func controlLabelCount(controls []controlBox, label string) int {
	count := 0
	for _, control := range controls {
		if control.Label == label {
			count++
		}
	}
	return count
}

func requireClosedRectangle(label string, lines []selectionLine, start sim.Vec2, end sim.Vec2) {
	if len(lines) != 4 {
		panic(fmt.Sprintf("%s line count = %d", label, len(lines)))
	}
	left := math.Min(start.X, end.X)
	right := math.Max(start.X, end.X)
	top := math.Min(start.Y, end.Y)
	bottom := math.Max(start.Y, end.Y)
	expected := []selectionLine{
		{x1: left, y1: top, x2: right, y2: top},
		{x1: right, y1: top, x2: right, y2: bottom},
		{x1: right, y1: bottom, x2: left, y2: bottom},
		{x1: left, y1: bottom, x2: left, y2: top},
	}
	for i := range expected {
		if lines[i] != expected[i] {
			panic(fmt.Sprintf("%s lines = %#v, want %#v", label, lines, expected))
		}
	}
}
