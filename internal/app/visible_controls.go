package app

import (
	"fmt"
	"image"
	"image/color"
)

const (
	debugGlyphWidth  = 6
	debugGlyphHeight = 16
)

var (
	controlColor       = color.RGBA{R: 74, G: 88, B: 108, A: 255}
	activeControlColor = color.RGBA{R: 96, G: 132, B: 166, A: 255}
	sectionColor       = color.RGBA{R: 58, G: 70, B: 86, A: 255}
)

type controlBox struct {
	Name   string
	Label  string
	Region string
	Rect   image.Rectangle
}

type statusField struct {
	Name  string
	Label string
	Rect  image.Rectangle
}

type DrawFrameReport struct {
	RegionPixels          map[string]int
	Controls              map[string]string
	ActiveControls        map[string]bool
	InspectorSections     map[string]bool
	InspectorSectionRects map[string]image.Rectangle
	NumericSettings       map[string]NumericSettingFrame
	StatusFields          map[string]string
	CanvasWorldPixels     int
	ControlLabelsFit      bool
	RegionControlCounts   map[string]int
}

type NumericSettingFrame struct {
	CheckboxRect       image.Rectangle
	LabelRect          image.Rectangle
	DecrementRect      image.Rectangle
	SliderRect         image.Rectangle
	IncrementRect      image.Rectangle
	TextFieldRect      image.Rectangle
	InspectorRect      image.Rectangle
	Text               string
	SliderFraction     float64
	TextCursorVisible  bool
	TextHighlighted    bool
	LabelFitsInspector bool
}

func isSliderControl(name string) bool {
	_, ok := numericSettingForSlider(name)
	return ok
}

func sliderTrack(control controlBox) image.Rectangle {
	return image.Rect(control.Rect.Min.X+4, control.Rect.Min.Y+13, control.Rect.Max.X-4, control.Rect.Min.Y+17)
}

func (g *Game) sliderFraction(name string) float64 {
	setting, ok := numericSettingForSlider(name)
	if !ok {
		return 0
	}
	return g.numericSettingSliderFraction(setting)
}

func (g *Game) sliderLabel(control controlBox) string {
	return control.Label
}

func (g *Game) activeControl(name string) bool {
	return g.activeRunControl(name) ||
		g.activeForceControl(name) ||
		g.activeParameterControl(name) ||
		g.activeSelectedSpringControl(name) ||
		g.activeWallControl(name) ||
		g.controls.activeNumericStep == name
}

func (g *Game) activeRunControl(name string) bool {
	switch name {
	case "run pause toggle command":
		return true
	default:
		return false
	}
}

func (g *Game) activeForceControl(name string) bool {
	forceName, ok := activeForceControls[name]
	return ok && g.forceEnabled(forceName)
}

var activeForceControls = map[string]string{
	"gravity force":           "gravity",
	"center attraction force": "center attraction",
	"center mass force":       "center of mass attraction",
	"wall repulsion force":    "wall repulsion",
	"mass collision force":    "mass collision",
}

func (g *Game) activeParameterControl(name string) bool {
	switch name {
	case "fixed mass toggle":
		return g.parameterEnabled("fixed mass")
	case "show springs toggle":
		return g.parameterEnabled("show springs")
	case "adaptive timestep toggle":
		return g.parameterEnabled("adaptive timestep")
	case "grid snap toggle":
		return g.gridSnapEnabled()
	default:
		return false
	}
}

func (g *Game) activeSelectedSpringControl(name string) bool {
	if name != "spring wall toggle" {
		return false
	}
	selectedCount := 0
	for _, spring := range g.world.simulation.Springs {
		if !g.editing().SelectedSprings[spring.ID] {
			continue
		}
		selectedCount++
		if !spring.Wall {
			return false
		}
	}
	return selectedCount > 0
}

func (g *Game) activeWallControl(name string) bool {
	switch name {
	case "top wall toggle":
		return g.wallEnabled("top")
	case "left wall toggle":
		return g.wallEnabled("left")
	case "right wall toggle":
		return g.wallEnabled("right")
	case "bottom wall toggle":
		return g.wallEnabled("bottom")
	default:
		return false
	}
}

func (g *Game) visibleControls() []controlBox {
	controls := append(menuControls(), toolbarControls()...)
	controls = append(controls, g.commandControls()...)
	return append(controls, inspectorControls()...)
}

func visibleControls() []controlBox {
	return NewGame().visibleControls()
}

func menuControls() []controlBox {
	return []controlBox{
		{Name: "edit menu", Label: "Edit", Region: "top command bar", Rect: image.Rect(8, 8, 52, 28)},
	}
}

func (g *Game) editMenuControls() []controlBox {
	if !g.controls.editMenuOpen {
		return nil
	}
	return []controlBox{
		{Name: "cut command", Label: "Cut     Ctrl+X", Region: "top command bar", Rect: image.Rect(8, 30, 132, 54)},
		{Name: "copy command", Label: "Copy    Ctrl+C", Region: "top command bar", Rect: image.Rect(8, 54, 132, 78)},
		{Name: "paste command", Label: "Paste   Ctrl+V", Region: "top command bar", Rect: image.Rect(8, 78, 132, 102)},
	}
}

func toolbarControls() []controlBox {
	return []controlBox{
		{Name: "select all command", Label: "All", Region: "left toolbar", Rect: image.Rect(8, 48, 64, 68)},
		{Name: "duplicate command", Label: "Dup", Region: "left toolbar", Rect: image.Rect(8, 78, 64, 98)},
		{Name: "delete command", Label: "Del", Region: "left toolbar", Rect: image.Rect(8, 108, 64, 128)},
	}
}

func (g *Game) commandControls() []controlBox {
	return []controlBox{
		{Name: "run pause toggle command", Label: g.runPauseToggleLabel(), Region: "top command bar", Rect: image.Rect(76, 8, 128, 28)},
		{Name: "reset command", Label: "Reset", Region: "top command bar", Rect: image.Rect(130, 8, 176, 28)},
		{Name: "save state command", Label: "State+", Region: "top command bar", Rect: image.Rect(178, 8, 230, 28)},
		{Name: "restore state command", Label: "State", Region: "top command bar", Rect: image.Rect(232, 8, 280, 28)},
		{Name: "load command", Label: "Load", Region: "top command bar", Rect: image.Rect(282, 8, 324, 28)},
		{Name: "insert command", Label: "Insert", Region: "top command bar", Rect: image.Rect(326, 8, 378, 28)},
		{Name: "save command", Label: "Save", Region: "top command bar", Rect: image.Rect(380, 8, 422, 28)},
		{Name: "quit command", Label: "Quit", Region: "top command bar", Rect: image.Rect(424, 8, 466, 28)},
	}
}

func (g *Game) runPauseToggleLabel() string {
	if g.run.paused {
		return "Run"
	}
	return "Pause"
}

func inspectorControls() []controlBox {
	controls := numericSettingControls()
	x := inspectorLeft() + 16
	right := screenWidth - 16
	half := (right - x - 8) / 2
	third := (right - x - 16) / 3
	second := x + third + 8
	thirdStart := second + third + 8
	controls = append(controls, []controlBox{
		{Name: "fixed mass toggle", Label: "Fixed", Region: "right inspector", Rect: image.Rect(x, 120, x+third, 140)},
		{Name: "set center command", Label: "Set Center", Region: "right inspector", Rect: image.Rect(second, 120, second+third, 140)},
		{Name: "mass collision force", Label: "Collide", Region: "right inspector", Rect: image.Rect(thirdStart, 120, right, 140)},
		{Name: "set rest length command", Label: "RestLen", Region: "right inspector", Rect: image.Rect(x, 229, x+half, 249)},
		{Name: "spring wall toggle", Label: "Wall", Region: "right inspector", Rect: image.Rect(x+half+8, 229, right, 249)},
		{Name: "grid snap toggle", Label: "Grid", Region: "right inspector", Rect: image.Rect(x, 608, x+half, 628)},
		{Name: "show springs toggle", Label: "Springs", Region: "right inspector", Rect: image.Rect(x+half+8, 608, right, 628)},
	}...)
	return controls
}

func inspectorSections() []controlBox {
	x := inspectorLeft() + 16
	right := screenWidth - 16
	return []controlBox{
		{Label: sectionHeaderLabel("Selected Mass(es)"), Region: "right inspector", Rect: image.Rect(x, 44, right, 64)},
		{Label: sectionHeaderLabel("Selected Spring(s)"), Region: "right inspector", Rect: image.Rect(x, 153, right, 173)},
		{Label: sectionHeaderLabel("Forces"), Region: "right inspector", Rect: image.Rect(x, 262, right, 282)},
		{Label: sectionHeaderLabel("Simulation"), Region: "right inspector", Rect: image.Rect(x, 397, right, 417)},
		{Label: sectionHeaderLabel("Display"), Region: "right inspector", Rect: image.Rect(x, 584, right, 604)},
	}
}

func sectionHeaderLabel(label string) string {
	return "----- " + label + " -----"
}

func inspectorLeft() int {
	return screenWidth - inspectorWidth
}

func (g *Game) forceEnabled(name string) bool {
	force, ok := g.world.simulation.Parameters.Force(name)
	return ok && force.Enabled == "true"
}

func (g *Game) parameterEnabled(name string) bool {
	return g.world.simulation.Parameters.Value(name) == "true"
}

func (g *Game) wallEnabled(name string) bool {
	enabled, _ := g.world.simulation.Parameters.WallEnabled(name)
	return enabled
}

func (g *Game) gridSnapEnabled() bool {
	return g.gridSnapSize() > 0
}

func (g *Game) statusFields() []statusField {
	x := inspectorLeft() + 16
	right := screenWidth - 16
	row1 := screenHeight - 56
	row2 := screenHeight - 30
	return []statusField{
		{Name: "object counts", Label: g.objectCountsStatusLabel(), Rect: image.Rect(x, row1, x+120, row1+20)},
		{Name: "file state", Label: g.fileState(), Rect: image.Rect(x+128, row1, right, row1+20)},
		{Name: "current file", Label: g.currentFileStatusLabel(), Rect: image.Rect(x, row2, right, row2+20)},
	}
}

func (g *Game) objectCountsStatusLabel() string {
	return fmt.Sprintf("Masses: %d", len(g.world.simulation.Masses))
}

func (g *Game) currentFileStatusLabel() string {
	if g.document.pathEntryCommand != "" {
		return g.document.pathEntryCommand
	}
	if g.document.currentFilePath == "" {
		return "File: (untitled)"
	}
	return g.document.currentFilePath
}

func (g *Game) selectedObjectCount() int {
	count := 0
	for _, selected := range g.editing().SelectedMasses {
		if selected {
			count++
		}
	}
	for _, selected := range g.editing().SelectedSprings {
		if selected {
			count++
		}
	}
	return count
}

func (g *Game) DrawFrameReport() DrawFrameReport {
	return analyzeDrawnFrame(g)
}

func analyzeDrawnFrame(game *Game) DrawFrameReport {
	report := DrawFrameReport{
		RegionPixels:          map[string]int{},
		Controls:              game.visibleControlLabels(),
		ActiveControls:        game.visibleActiveControls(),
		InspectorSections:     visibleInspectorSections(),
		InspectorSectionRects: visibleInspectorSectionRects(),
		NumericSettings:       numericSettingReports(game),
		StatusFields:          game.visibleStatusFields(),
		RegionControlCounts:   game.visibleRegionControlCounts(),
		ControlLabelsFit:      visibleLabelsFit(game),
	}
	for name := range visibleRegionRects() {
		report.RegionPixels[name] = game.visibleRegionPixels(name)
	}
	report.CanvasWorldPixels = visibleWorldPixels(game)
	return report
}

func (g *Game) visibleActiveControls() map[string]bool {
	active := map[string]bool{}
	for _, control := range g.visibleControls() {
		isActive := g.activeControl(control.Name)
		active[control.Name] = isActive
		active[control.Label] = active[control.Label] || isActive
	}
	return active
}

func (g *Game) visibleControlLabels() map[string]string {
	labels := map[string]string{}
	for _, control := range g.visibleControls() {
		labels[control.Name] = control.Label
	}
	return labels
}

func visibleInspectorSections() map[string]bool {
	sections := map[string]bool{}
	for _, section := range inspectorSections() {
		sections[section.Label] = true
	}
	return sections
}

func visibleInspectorSectionRects() map[string]image.Rectangle {
	sections := map[string]image.Rectangle{}
	for _, section := range inspectorSections() {
		sections[section.Label] = section.Rect
	}
	return sections
}

func (g *Game) visibleStatusFields() map[string]string {
	fields := map[string]string{}
	for _, field := range g.statusFields() {
		fields[field.Name] = field.Label
	}
	return fields
}

func (g *Game) visibleRegionControlCounts() map[string]int {
	counts := map[string]int{}
	for _, control := range g.visibleControls() {
		counts[control.Region]++
	}
	for _, section := range inspectorSections() {
		counts[section.Region]++
	}
	counts["right inspector"] += len(g.statusFields())
	return counts
}

func visibleLabelsFit(game *Game) bool {
	return controlLabelsFit(game.visibleControls()) &&
		controlLabelsFit(inspectorSections()) &&
		statusLabelsFit(game.statusFields())
}

func controlLabelsFit(boxes []controlBox) bool {
	return labelsFitItems(boxes, func(box controlBox) (string, image.Rectangle) { return box.Label, box.Rect })
}

func statusLabelsFit(fields []statusField) bool {
	return labelsFitItems(fields, func(field statusField) (string, image.Rectangle) { return field.Label, field.Rect })
}

func labelsFitItems[T any](items []T, labelAndRect func(T) (string, image.Rectangle)) bool {
	for _, item := range items {
		label, rect := labelAndRect(item)
		if !labelFits(label, rect) {
			return false
		}
	}
	return true
}

func labelFits(label string, rect image.Rectangle) bool {
	return len(label)*debugGlyphWidth <= rect.Dx()-8 && debugGlyphHeight <= rect.Dy()
}

func visibleRegionRects() map[string]image.Rectangle {
	return map[string]image.Rectangle{
		"canvas":          image.Rect(toolbarWidth, topBarHeight, screenWidth-inspectorWidth, screenHeight),
		"left toolbar":    image.Rect(0, topBarHeight, toolbarWidth, screenHeight),
		"top command bar": image.Rect(toolbarWidth, 0, screenWidth-inspectorWidth, topBarHeight),
		"right inspector": image.Rect(screenWidth-inspectorWidth, topBarHeight, screenWidth, screenHeight),
	}
}

func (g *Game) visibleRegionPixels(region string) int {
	count := regionControlPixels(g.visibleControls(), region) +
		regionControlPixels(inspectorSections(), region) +
		g.regionStatusPixels(region)
	if region != "canvas" {
		return count
	}
	if count == 0 {
		return rectPixels(visibleRegionRects()["canvas"])
	}
	return count
}

func regionControlPixels(controls []controlBox, region string) int {
	count := 0
	for _, control := range controls {
		if control.Region == region {
			count += rectPixels(control.Rect)
		}
	}
	return count
}

func (g *Game) regionStatusPixels(region string) int {
	if region != "right inspector" {
		return 0
	}
	count := 0
	for _, field := range g.statusFields() {
		count += rectPixels(field.Rect)
	}
	return count
}

func rectPixels(rect image.Rectangle) int {
	return rect.Dx() * rect.Dy()
}

func visibleWorldPixels(game *Game) int {
	canvas := visibleRegionRects()["canvas"]
	count := 0
	for _, mass := range game.world.simulation.Masses {
		screenPosition := game.worldToScreen(mass.Position)
		point := image.Pt(int(screenPosition.X), int(screenPosition.Y))
		if point.In(canvas) {
			count += 25
		}
	}
	for _, spring := range game.world.simulation.Springs {
		if game.validSpring(spring) {
			count += 50
		}
	}
	return count
}

// mutate4go-manifest-begin
// {"version":1,"tested_at":"2026-05-22T08:41:56-05:00","module_hash":"ca6574842396aa13e7b2833d27c46d3e99bebe2e95bbb9157c15b905550bc008","functions":[{"id":"func/isSliderControl","name":"isSliderControl","line":61,"end_line":64,"hash":"8ceba17f5b66651a2463fc0461d92e0d4aa8c678ab3978f667ea7d32cf8086f3"},{"id":"func/sliderTrack","name":"sliderTrack","line":66,"end_line":68,"hash":"46d823d72ae8a66fba02a8d2f7040c90cc2992092c0792df4a86e544d66832ce"},{"id":"func/Game.sliderFraction","name":"Game.sliderFraction","line":70,"end_line":76,"hash":"d1f33250a922ac07cd47ea5f1a1fb862dcc1fb2cbb69186673f38f7027239f0e"},{"id":"func/Game.sliderLabel","name":"Game.sliderLabel","line":78,"end_line":80,"hash":"9518a106db343932e617e1e39e5e16d832411f22f6c6b22eae415be6913c36ba"},{"id":"func/Game.activeControl","name":"Game.activeControl","line":82,"end_line":89,"hash":"cb05d4a22dae48b9d409a24f413426e956b5dc2dfcc80da005d53bb5d9b8969f"},{"id":"func/Game.activeRunControl","name":"Game.activeRunControl","line":91,"end_line":100,"hash":"94d055de600771425a34536848b927f427d7de0721232c1e4d0a0d2c56ee2755"},{"id":"func/Game.activeForceControl","name":"Game.activeForceControl","line":102,"end_line":105,"hash":"c325d1b27113b9be28ff09d2864a00e4267a652e4dc43dcdc1ee04b7a8f1560c"},{"id":"func/Game.activeParameterControl","name":"Game.activeParameterControl","line":115,"end_line":128,"hash":"64fe357ad27032dcee16ee76a6e733f85c8d1980d883d16d53c0cda31477cf6f"},{"id":"func/Game.activeSelectedSpringControl","name":"Game.activeSelectedSpringControl","line":130,"end_line":145,"hash":"e5f6066e8764b31e854ad214b922572357de9eff89b1c97f183e9e279a94e01c"},{"id":"func/Game.activeWallControl","name":"Game.activeWallControl","line":147,"end_line":160,"hash":"70c0328208760cb2a63cb5074d116acfbe989380d2df61cbb2e3940ea1e2c3cd"},{"id":"func/visibleControls","name":"visibleControls","line":162,"end_line":166,"hash":"14e0313a1743b30c30268cf867272f99ac5a1880ef27fab3ac7641592af5459a"},{"id":"func/menuControls","name":"menuControls","line":168,"end_line":172,"hash":"7a632b167f77d172b422b42ab30311786f673b77fab82640220ac77b25e4fb31"},{"id":"func/Game.editMenuControls","name":"Game.editMenuControls","line":174,"end_line":183,"hash":"0ef9897e3339c54230ed4064fe379b47d72a66265ecc4a51d92bc4f8f376dbca"},{"id":"func/toolbarControls","name":"toolbarControls","line":185,"end_line":191,"hash":"ef6b992b59a86bfa58646d326682f9037168e35bbd2e8227951454a59975feee"},{"id":"func/commandControls","name":"commandControls","line":193,"end_line":205,"hash":"c8d380a126ca7b6b48433d0da82dcfce7908127eeda183a70ddce75a823d43f9"},{"id":"func/inspectorControls","name":"inspectorControls","line":207,"end_line":225,"hash":"d0d8eeecbc2882d71ad35293a6e6a6cf77ce4b8ae72f7065b3ed7547455b2f82"},{"id":"func/inspectorSections","name":"inspectorSections","line":227,"end_line":237,"hash":"4b93cc6194ca93004f98e949d56e345fd0dfb761eea7a66042d3c1855a340ecf"},{"id":"func/sectionHeaderLabel","name":"sectionHeaderLabel","line":239,"end_line":241,"hash":"965cc3e940b8f377026c01743390675feaae023ac94b515c0f143f197cf11e36"},{"id":"func/inspectorLeft","name":"inspectorLeft","line":243,"end_line":245,"hash":"93e0ad43c5e22d7ec07af829c3f7383f3367334e90967f03635e5f28b8609df6"},{"id":"func/Game.forceEnabled","name":"Game.forceEnabled","line":247,"end_line":250,"hash":"e92eb6da3a8645ed7b2b94029d1e844afb59a9908991669db4215e1552b9ee2e"},{"id":"func/Game.parameterEnabled","name":"Game.parameterEnabled","line":252,"end_line":254,"hash":"8d6d66cd2a42184a3d094230de4fb8e8a517d6ffa02862e945166de4b7e83466"},{"id":"func/Game.wallEnabled","name":"Game.wallEnabled","line":256,"end_line":259,"hash":"7b351631c38f03ff3ab45efb859f86b1bf792b6fa0368e4eb5b00ecbd0bc7426"},{"id":"func/Game.gridSnapEnabled","name":"Game.gridSnapEnabled","line":261,"end_line":263,"hash":"fe4c82bb8c4f8c739d9a9752b22185bebeccc822495ba9cf97444b77cb24008b"},{"id":"func/Game.statusFields","name":"Game.statusFields","line":265,"end_line":280,"hash":"032de12369f670d8a011778c870d19b0af96a9b8b1841187ded89cafd728e662"},{"id":"func/Game.currentFileStatusLabel","name":"Game.currentFileStatusLabel","line":282,"end_line":287,"hash":"497dff6f613db6818b3991cc8aa42c109f6907a274f1b5ce0656edcebb335d70"},{"id":"func/Game.selectedObjectCount","name":"Game.selectedObjectCount","line":289,"end_line":302,"hash":"37228a151a2c37a8334e0743e75e66c4035294b511f5b8142948fd3fb0b2671b"},{"id":"func/Game.DrawFrameReport","name":"Game.DrawFrameReport","line":304,"end_line":306,"hash":"21457be3cb4856c4bb131972199ad58004071195dbfef0cb1500d22d228c9ad7"},{"id":"func/analyzeDrawnFrame","name":"analyzeDrawnFrame","line":308,"end_line":325,"hash":"02544d7ddda41da6437759d9548c9fc72101f26a3c69ecce6a6686a782d1ee8d"},{"id":"func/Game.visibleActiveControls","name":"Game.visibleActiveControls","line":327,"end_line":335,"hash":"a5dc9fa130f937bef9431d0f108f159dcded3984e1decb5c9ef5bda78d1a5cbc"},{"id":"func/visibleControlLabels","name":"visibleControlLabels","line":337,"end_line":343,"hash":"afac7bcb5fc0d6d4c034c1af2c507c9bbd31f7830feed44a0c4813e9016d77c2"},{"id":"func/visibleInspectorSections","name":"visibleInspectorSections","line":345,"end_line":351,"hash":"ae27149a3bf3e13945c3a1fe67e87d42d1929718201b50499740d6a243b046b9"},{"id":"func/visibleInspectorSectionRects","name":"visibleInspectorSectionRects","line":353,"end_line":359,"hash":"193f6258b350ffa39cd94b7f120d0144b14f99bc40771430b17820ab69fffdac"},{"id":"func/Game.visibleStatusFields","name":"Game.visibleStatusFields","line":361,"end_line":367,"hash":"38815d71e80400a37fa2e84f2abcbd950d1814697f03bc89239fd64a0b12fcd8"},{"id":"func/Game.visibleRegionControlCounts","name":"Game.visibleRegionControlCounts","line":369,"end_line":379,"hash":"5ab00d00237ba48d513eaa578c5eb21b12db49f2414202231996a3ba84333790"},{"id":"func/visibleLabelsFit","name":"visibleLabelsFit","line":381,"end_line":385,"hash":"0fd4cb254764943e007f36732d9aa1d7ad707a59746a9cca2e31aab766226f88"},{"id":"func/controlLabelsFit","name":"controlLabelsFit","line":387,"end_line":389,"hash":"6cb92c088807face5a2162c78d338eb1278ed8774354b8b1d9ea884286f97a42"},{"id":"func/statusLabelsFit","name":"statusLabelsFit","line":391,"end_line":393,"hash":"9786f63a658af2dce883ca37f2fac1b7365902b152510d13c9787f1d5bdd412b"},{"id":"func/labelsFitItems","name":"labelsFitItems","line":395,"end_line":403,"hash":"dde06934ee8b31db69fe38357ce4bd6c6b73d991ef5d6f3f389e012fc3965dae"},{"id":"func/labelFits","name":"labelFits","line":405,"end_line":407,"hash":"60447d380d9ed2d8710772539abc4bc8e5a7bf62719fc471d97c16bb3ae7b3b3"},{"id":"func/visibleRegionRects","name":"visibleRegionRects","line":409,"end_line":416,"hash":"074c67126025359bea7e31374985d3a06715c395ea70ac7acfda1af1a1ffb089"},{"id":"func/Game.visibleRegionPixels","name":"Game.visibleRegionPixels","line":418,"end_line":429,"hash":"d6393cc85bc819f0606076cc4256b5d8d14dc8bb4f0dbdd03b4d5b701bc636dc"},{"id":"func/regionControlPixels","name":"regionControlPixels","line":431,"end_line":439,"hash":"09baade91c8b970cf2ec8ce4dd168c323ca63cf59170b8f129019bedef0ea743"},{"id":"func/Game.regionStatusPixels","name":"Game.regionStatusPixels","line":441,"end_line":450,"hash":"d6af3c389249f2f33a402dd43ae8fd7f1294cd6ecfccc930756dab7b21f18adf"},{"id":"func/rectPixels","name":"rectPixels","line":452,"end_line":454,"hash":"be91e5c34925744d4284c9bb7111688f4f62eef60f40d9b92c77353a60433f83"},{"id":"func/visibleWorldPixels","name":"visibleWorldPixels","line":456,"end_line":472,"hash":"0e2c9df97d85ab73c9408b23f370e2619c7b3cd570f132f6890e35e4c9218a7a"}]}
// mutate4go-manifest-end
