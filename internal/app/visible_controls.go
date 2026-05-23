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

type namedLabel interface {
	nameLabel() (string, string)
}

func (control controlBox) nameLabel() (string, string) {
	return control.Name, control.Label
}

func (field statusField) nameLabel() (string, string) {
	return field.Name, field.Label
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
	return controlLabelMap(g.visibleControls())
}

func visibleInspectorSections() map[string]bool {
	return visibleInspectorSectionValues(func(controlBox) bool { return true })
}

func visibleInspectorSectionRects() map[string]image.Rectangle {
	return visibleInspectorSectionValues(func(section controlBox) image.Rectangle { return section.Rect })
}

func visibleInspectorSectionValues[V any](value func(controlBox) V) map[string]V {
	sections := map[string]V{}
	for _, section := range inspectorSections() {
		sections[section.Label] = value(section)
	}
	return sections
}

func (g *Game) visibleStatusFields() map[string]string {
	return statusFieldLabelMap(g.statusFields())
}

func controlLabelMap(controls []controlBox) map[string]string {
	return labelMap(controls)
}

func statusFieldLabelMap(fields []statusField) map[string]string {
	return labelMap(fields)
}

func labelMap[T namedLabel](items []T) map[string]string {
	labels := map[string]string{}
	for _, item := range items {
		name, label := item.nameLabel()
		labels[name] = label
	}
	return labels
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
// {"version":1,"tested_at":"2026-05-23T09:06:33-05:00","module_hash":"138c6833be36987d3579a9043782d5970cac3c3a4fe5fb9dcfbec2beb96d8297","functions":[{"id":"func/controlBox.nameLabel","name":"controlBox.nameLabel","line":37,"end_line":39,"hash":"820904392de93ef8332b721867c1af37e58e9ff1b8a56612597b39d54aa58c7e"},{"id":"func/statusField.nameLabel","name":"statusField.nameLabel","line":41,"end_line":43,"hash":"13d311afe79df84c637bf294872f3fe5ad87688d4c669ac5178d386e0b6a1ac7"},{"id":"func/isSliderControl","name":"isSliderControl","line":73,"end_line":76,"hash":"8ceba17f5b66651a2463fc0461d92e0d4aa8c678ab3978f667ea7d32cf8086f3"},{"id":"func/sliderTrack","name":"sliderTrack","line":78,"end_line":80,"hash":"46d823d72ae8a66fba02a8d2f7040c90cc2992092c0792df4a86e544d66832ce"},{"id":"func/Game.sliderFraction","name":"Game.sliderFraction","line":82,"end_line":88,"hash":"d1f33250a922ac07cd47ea5f1a1fb862dcc1fb2cbb69186673f38f7027239f0e"},{"id":"func/Game.sliderLabel","name":"Game.sliderLabel","line":90,"end_line":92,"hash":"9518a106db343932e617e1e39e5e16d832411f22f6c6b22eae415be6913c36ba"},{"id":"func/Game.activeControl","name":"Game.activeControl","line":94,"end_line":101,"hash":"cb05d4a22dae48b9d409a24f413426e956b5dc2dfcc80da005d53bb5d9b8969f"},{"id":"func/Game.activeRunControl","name":"Game.activeRunControl","line":103,"end_line":110,"hash":"4d533625b4090a8e7b13cb173d832be2f915de6d2fec0c4d89043804fef270bb"},{"id":"func/Game.activeForceControl","name":"Game.activeForceControl","line":112,"end_line":115,"hash":"c325d1b27113b9be28ff09d2864a00e4267a652e4dc43dcdc1ee04b7a8f1560c"},{"id":"func/Game.activeParameterControl","name":"Game.activeParameterControl","line":125,"end_line":138,"hash":"64fe357ad27032dcee16ee76a6e733f85c8d1980d883d16d53c0cda31477cf6f"},{"id":"func/Game.activeSelectedSpringControl","name":"Game.activeSelectedSpringControl","line":140,"end_line":155,"hash":"e5f6066e8764b31e854ad214b922572357de9eff89b1c97f183e9e279a94e01c"},{"id":"func/Game.activeWallControl","name":"Game.activeWallControl","line":157,"end_line":170,"hash":"70c0328208760cb2a63cb5074d116acfbe989380d2df61cbb2e3940ea1e2c3cd"},{"id":"func/Game.visibleControls","name":"Game.visibleControls","line":172,"end_line":176,"hash":"5852351c1a8067198eb2b1ea0d6a4c4ae1259058979021f4bb6be2e8eb2e20c8"},{"id":"func/visibleControls","name":"visibleControls","line":178,"end_line":180,"hash":"df5b4ca9f1bcdf2a0087db098f271b1f0532080d27bdea4756e546d0f65bd1f6"},{"id":"func/menuControls","name":"menuControls","line":182,"end_line":186,"hash":"7a632b167f77d172b422b42ab30311786f673b77fab82640220ac77b25e4fb31"},{"id":"func/Game.editMenuControls","name":"Game.editMenuControls","line":188,"end_line":197,"hash":"0ef9897e3339c54230ed4064fe379b47d72a66265ecc4a51d92bc4f8f376dbca"},{"id":"func/toolbarControls","name":"toolbarControls","line":199,"end_line":205,"hash":"ef6b992b59a86bfa58646d326682f9037168e35bbd2e8227951454a59975feee"},{"id":"func/Game.commandControls","name":"Game.commandControls","line":207,"end_line":218,"hash":"a83815e0695773f44924287a444a47f0aa9170212f109cf6af8e4de30644096e"},{"id":"func/Game.runPauseToggleLabel","name":"Game.runPauseToggleLabel","line":220,"end_line":225,"hash":"2ab723d8b45797cc772ff58c6b2975e83f227776b6c12465bf04c9a085da2ba7"},{"id":"func/inspectorControls","name":"inspectorControls","line":227,"end_line":245,"hash":"d0d8eeecbc2882d71ad35293a6e6a6cf77ce4b8ae72f7065b3ed7547455b2f82"},{"id":"func/inspectorSections","name":"inspectorSections","line":247,"end_line":257,"hash":"4b93cc6194ca93004f98e949d56e345fd0dfb761eea7a66042d3c1855a340ecf"},{"id":"func/sectionHeaderLabel","name":"sectionHeaderLabel","line":259,"end_line":261,"hash":"965cc3e940b8f377026c01743390675feaae023ac94b515c0f143f197cf11e36"},{"id":"func/inspectorLeft","name":"inspectorLeft","line":263,"end_line":265,"hash":"93e0ad43c5e22d7ec07af829c3f7383f3367334e90967f03635e5f28b8609df6"},{"id":"func/Game.forceEnabled","name":"Game.forceEnabled","line":267,"end_line":270,"hash":"e92eb6da3a8645ed7b2b94029d1e844afb59a9908991669db4215e1552b9ee2e"},{"id":"func/Game.parameterEnabled","name":"Game.parameterEnabled","line":272,"end_line":274,"hash":"8d6d66cd2a42184a3d094230de4fb8e8a517d6ffa02862e945166de4b7e83466"},{"id":"func/Game.wallEnabled","name":"Game.wallEnabled","line":276,"end_line":279,"hash":"7b351631c38f03ff3ab45efb859f86b1bf792b6fa0368e4eb5b00ecbd0bc7426"},{"id":"func/Game.gridSnapEnabled","name":"Game.gridSnapEnabled","line":281,"end_line":283,"hash":"fe4c82bb8c4f8c739d9a9752b22185bebeccc822495ba9cf97444b77cb24008b"},{"id":"func/Game.statusFields","name":"Game.statusFields","line":285,"end_line":295,"hash":"62ea5d5932abbe4f9e0db8a012ee2a64bb6ea33475f27f9bbc27a5dbc76c0449"},{"id":"func/Game.objectCountsStatusLabel","name":"Game.objectCountsStatusLabel","line":297,"end_line":299,"hash":"99b2c415cf24a06570d23751ec47ff2c3ee12005642ed77426c3b27c9822be78"},{"id":"func/Game.currentFileStatusLabel","name":"Game.currentFileStatusLabel","line":301,"end_line":309,"hash":"8a7468b87ec5f80df05a9faa2120b32bf2596bce12511d2650e328ca78bd6396"},{"id":"func/Game.selectedObjectCount","name":"Game.selectedObjectCount","line":311,"end_line":324,"hash":"37228a151a2c37a8334e0743e75e66c4035294b511f5b8142948fd3fb0b2671b"},{"id":"func/Game.DrawFrameReport","name":"Game.DrawFrameReport","line":326,"end_line":328,"hash":"21457be3cb4856c4bb131972199ad58004071195dbfef0cb1500d22d228c9ad7"},{"id":"func/analyzeDrawnFrame","name":"analyzeDrawnFrame","line":330,"end_line":347,"hash":"9def1a8d4761e6f8a3413923faf571c8dbfe7f4a094b546a1bd07a64af801b8b"},{"id":"func/Game.visibleActiveControls","name":"Game.visibleActiveControls","line":349,"end_line":357,"hash":"c41b4c8d1458255e1118f68ef0c6646de61324f62cd0d5820a53080b06c0a709"},{"id":"func/Game.visibleControlLabels","name":"Game.visibleControlLabels","line":359,"end_line":361,"hash":"b0948b14b96a8306586ca5a9684d8ab510b5c815021162e100be05042c40aa8a"},{"id":"func/visibleInspectorSections","name":"visibleInspectorSections","line":363,"end_line":365,"hash":"c0efd6cdb02fc79f4c4e46695decf101c1ca785e964ff4b8d8fc91baa597a6bc"},{"id":"func/visibleInspectorSectionRects","name":"visibleInspectorSectionRects","line":367,"end_line":369,"hash":"a76b6b32d0d6a5689aafef250bf75735a452d259ee85f9d677b39a98f8cea212"},{"id":"func/visibleInspectorSectionValues","name":"visibleInspectorSectionValues","line":371,"end_line":377,"hash":"fb8a9ce94305dd8fdd8d141ab0b444a57acb95f9f576b08d78f513312711e4e5"},{"id":"func/Game.visibleStatusFields","name":"Game.visibleStatusFields","line":379,"end_line":381,"hash":"50d07aa7312fc26f394f6a6876e5ff966683ac92ce3fd49026e183c01fa6c449"},{"id":"func/controlLabelMap","name":"controlLabelMap","line":383,"end_line":385,"hash":"bd8c5c640c96459f92bac26ffd67565436a5a235c3d75767bce54770811892b7"},{"id":"func/statusFieldLabelMap","name":"statusFieldLabelMap","line":387,"end_line":389,"hash":"2ac6d4ad90818cfed2a8f18f15d0c98dc0f5b556b82f78610352e3c2b759e90a"},{"id":"func/labelMap","name":"labelMap","line":391,"end_line":398,"hash":"865aff1bf622500851a2a5a7c62d2c59910c52e8e933e421e4d930cae577aa6b"},{"id":"func/Game.visibleRegionControlCounts","name":"Game.visibleRegionControlCounts","line":400,"end_line":410,"hash":"e59e899d1422fb42e7208802a98d7ac74a8062fcec0303cb32dab893395b13f3"},{"id":"func/visibleLabelsFit","name":"visibleLabelsFit","line":412,"end_line":416,"hash":"2c7d65d1bbe41e0a4137b4207c134d562e0b2a796f94ca93dae3c9193fd53f93"},{"id":"func/controlLabelsFit","name":"controlLabelsFit","line":418,"end_line":420,"hash":"6cb92c088807face5a2162c78d338eb1278ed8774354b8b1d9ea884286f97a42"},{"id":"func/statusLabelsFit","name":"statusLabelsFit","line":422,"end_line":424,"hash":"9786f63a658af2dce883ca37f2fac1b7365902b152510d13c9787f1d5bdd412b"},{"id":"func/labelsFitItems","name":"labelsFitItems","line":426,"end_line":434,"hash":"dde06934ee8b31db69fe38357ce4bd6c6b73d991ef5d6f3f389e012fc3965dae"},{"id":"func/labelFits","name":"labelFits","line":436,"end_line":438,"hash":"60447d380d9ed2d8710772539abc4bc8e5a7bf62719fc471d97c16bb3ae7b3b3"},{"id":"func/visibleRegionRects","name":"visibleRegionRects","line":440,"end_line":447,"hash":"074c67126025359bea7e31374985d3a06715c395ea70ac7acfda1af1a1ffb089"},{"id":"func/Game.visibleRegionPixels","name":"Game.visibleRegionPixels","line":449,"end_line":460,"hash":"c63bb0d200670f072220e6eeeaccd2de898122bccdb766cfbbea5c7480c0346e"},{"id":"func/regionControlPixels","name":"regionControlPixels","line":462,"end_line":470,"hash":"09baade91c8b970cf2ec8ce4dd168c323ca63cf59170b8f129019bedef0ea743"},{"id":"func/Game.regionStatusPixels","name":"Game.regionStatusPixels","line":472,"end_line":481,"hash":"d6af3c389249f2f33a402dd43ae8fd7f1294cd6ecfccc930756dab7b21f18adf"},{"id":"func/rectPixels","name":"rectPixels","line":483,"end_line":485,"hash":"be91e5c34925744d4284c9bb7111688f4f62eef60f40d9b92c77353a60433f83"},{"id":"func/visibleWorldPixels","name":"visibleWorldPixels","line":487,"end_line":503,"hash":"0e2c9df97d85ab73c9408b23f370e2619c7b3cd570f132f6890e35e4c9218a7a"}]}
// mutate4go-manifest-end
