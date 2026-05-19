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
	LabelRect          image.Rectangle
	SliderRect         image.Rectangle
	TextFieldRect      image.Rectangle
	InspectorRect      image.Rectangle
	Text               string
	SliderFraction     float64
	TextCursorVisible  bool
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
		g.activeWallControl(name)
}

func (g *Game) activeRunControl(name string) bool {
	switch name {
	case "pause command":
		return g.paused
	case "run command":
		return !g.paused
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

func visibleControls() []controlBox {
	controls := append(menuControls(), toolbarControls()...)
	controls = append(controls, commandControls()...)
	return append(controls, inspectorControls()...)
}

func menuControls() []controlBox {
	return []controlBox{
		{Name: "edit menu", Label: "Edit", Region: "top command bar", Rect: image.Rect(8, 8, 52, 28)},
	}
}

func (g *Game) editMenuControls() []controlBox {
	if !g.editMenuOpen {
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

func commandControls() []controlBox {
	return []controlBox{
		{Name: "run command", Label: "Run", Region: "top command bar", Rect: image.Rect(76, 8, 110, 28)},
		{Name: "pause command", Label: "Pause", Region: "top command bar", Rect: image.Rect(112, 8, 158, 28)},
		{Name: "reset command", Label: "Reset", Region: "top command bar", Rect: image.Rect(160, 8, 206, 28)},
		{Name: "save state command", Label: "State+", Region: "top command bar", Rect: image.Rect(208, 8, 260, 28)},
		{Name: "restore state command", Label: "State", Region: "top command bar", Rect: image.Rect(262, 8, 310, 28)},
		{Name: "load command", Label: "Load", Region: "top command bar", Rect: image.Rect(312, 8, 354, 28)},
		{Name: "insert command", Label: "Insert", Region: "top command bar", Rect: image.Rect(356, 8, 408, 28)},
		{Name: "save command", Label: "Save", Region: "top command bar", Rect: image.Rect(410, 8, 452, 28)},
		{Name: "quit command", Label: "Quit", Region: "top command bar", Rect: image.Rect(454, 8, 496, 28)},
	}
}

func inspectorControls() []controlBox {
	controls := numericSettingControls()
	x := inspectorLeft() + 16
	right := screenWidth - 16
	half := (right - x - 8) / 2
	controls = append(controls, []controlBox{
		{Name: "fixed mass toggle", Label: "Fixed", Region: "right inspector", Rect: image.Rect(x, 120, right, 140)},
		{Name: "set rest length command", Label: "RestLen", Region: "right inspector", Rect: image.Rect(x, 224, right, 244)},
		{Name: "gravity force", Label: "Gravity", Region: "right inspector", Rect: image.Rect(x, 382, x+half, 402)},
		{Name: "center attraction force", Label: "Center", Region: "right inspector", Rect: image.Rect(x+half+8, 382, right, 402)},
		{Name: "center mass force", Label: "CMass", Region: "right inspector", Rect: image.Rect(x, 408, x+half, 428)},
		{Name: "wall repulsion force", Label: "WallRep", Region: "right inspector", Rect: image.Rect(x+half+8, 408, right, 428)},
		{Name: "mass collision force", Label: "Collide", Region: "right inspector", Rect: image.Rect(x, 434, right, 454)},
		{Name: "set center command", Label: "SetCtr", Region: "right inspector", Rect: image.Rect(x, 460, right, 480)},
		{Name: "top wall toggle", Label: "Top", Region: "right inspector", Rect: image.Rect(x, 510, x+half, 530)},
		{Name: "bottom wall toggle", Label: "Bot", Region: "right inspector", Rect: image.Rect(x+half+8, 510, right, 530)},
		{Name: "left wall toggle", Label: "Left", Region: "right inspector", Rect: image.Rect(x, 536, x+half, 556)},
		{Name: "right wall toggle", Label: "Right", Region: "right inspector", Rect: image.Rect(x+half+8, 536, right, 556)},
		{Name: "grid snap toggle", Label: "Grid", Region: "right inspector", Rect: image.Rect(x, 724, x+half, 744)},
		{Name: "show springs toggle", Label: "Springs", Region: "right inspector", Rect: image.Rect(x+half+8, 724, right, 744)},
		{Name: "adaptive timestep toggle", Label: "Adapt", Region: "right inspector", Rect: image.Rect(x, 750, right, 770)},
	}...)
	return controls
}

func inspectorSections() []controlBox {
	x := inspectorLeft() + 16
	right := screenWidth - 16
	return []controlBox{
		{Label: "Mass", Region: "right inspector", Rect: image.Rect(x, 44, right, 64)},
		{Label: "Spring", Region: "right inspector", Rect: image.Rect(x, 148, right, 168)},
		{Label: "Forces", Region: "right inspector", Rect: image.Rect(x, 254, right, 274)},
		{Label: "Walls", Region: "right inspector", Rect: image.Rect(x, 486, right, 506)},
		{Label: "Simulation", Region: "right inspector", Rect: image.Rect(x, 570, right, 590)},
	}
}

func inspectorLeft() int {
	return screenWidth - inspectorWidth
}

func (g *Game) forceEnabled(name string) bool {
	force, ok := g.simulation.Parameters.Force(name)
	return ok && force.Enabled == "true"
}

func (g *Game) parameterEnabled(name string) bool {
	return g.simulation.Parameters.Value(name) == "true"
}

func (g *Game) wallEnabled(name string) bool {
	enabled, _ := g.simulation.Parameters.WallEnabled(name)
	return enabled
}

func (g *Game) gridSnapEnabled() bool {
	return g.gridSnapSize() > 0
}

func (g *Game) statusFields() []statusField {
	return []statusField{
		{Name: "run state", Label: g.simulationState(), Rect: image.Rect(8, screenHeight-22, 104, screenHeight-2)},
		{Name: "object counts", Label: "object counts", Rect: image.Rect(112, screenHeight-22, 212, screenHeight-2)},
		{Name: "selected object count", Label: fmt.Sprintf("%d sel", g.selectedObjectCount()), Rect: image.Rect(220, screenHeight-22, 290, screenHeight-2)},
		{Name: "current file", Label: g.currentFileStatusLabel(), Rect: image.Rect(298, screenHeight-22, 412, screenHeight-2)},
		{Name: "dirty state", Label: g.fileState(), Rect: image.Rect(420, screenHeight-22, 512, screenHeight-2)},
		{Name: "file state", Label: g.fileState(), Rect: image.Rect(520, screenHeight-22, 612, screenHeight-2)},
		{Name: "last error", Label: "", Rect: image.Rect(620, screenHeight-22, 872, screenHeight-2)},
	}
}

func (g *Game) currentFileStatusLabel() string {
	if g.pathEntryCommand != "" {
		return g.pathEntryCommand
	}
	return g.currentFilePath
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
		Controls:              visibleControlLabels(),
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
	for _, control := range visibleControls() {
		isActive := g.activeControl(control.Name)
		active[control.Name] = isActive
		active[control.Label] = active[control.Label] || isActive
	}
	return active
}

func visibleControlLabels() map[string]string {
	labels := map[string]string{}
	for _, control := range visibleControls() {
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
	for _, control := range visibleControls() {
		counts[control.Region]++
	}
	for _, section := range inspectorSections() {
		counts[section.Region]++
	}
	counts["status line"] = len(g.statusFields())
	return counts
}

func visibleLabelsFit(game *Game) bool {
	return controlLabelsFit(visibleControls()) &&
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
		"canvas":          image.Rect(toolbarWidth, topBarHeight, screenWidth-inspectorWidth, screenHeight-statusHeight),
		"left toolbar":    image.Rect(0, topBarHeight, toolbarWidth, screenHeight-statusHeight),
		"top command bar": image.Rect(toolbarWidth, 0, screenWidth-inspectorWidth, topBarHeight),
		"right inspector": image.Rect(screenWidth-inspectorWidth, topBarHeight, screenWidth, screenHeight-statusHeight),
		"status line":     image.Rect(0, screenHeight-statusHeight, screenWidth, screenHeight),
	}
}

func (g *Game) visibleRegionPixels(region string) int {
	count := regionControlPixels(visibleControls(), region) +
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
	if region != "status line" {
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
	for _, mass := range game.simulation.Masses {
		screenPosition := game.worldToScreen(mass.Position)
		point := image.Pt(int(screenPosition.X), int(screenPosition.Y))
		if point.In(canvas) {
			count += 25
		}
	}
	for _, spring := range game.simulation.Springs {
		if game.validSpring(spring) {
			count += 50
		}
	}
	return count
}

// mutate4go-manifest-begin
// {"version":1,"tested_at":"2026-05-19T12:27:13-05:00","module_hash":"c0cf4e54683868cce9965971629c7ff4a0f8d5548484e216bb031cc9a6cea8c0","functions":[{"id":"func/isSliderControl","name":"isSliderControl","line":44,"end_line":46,"hash":"3732aebb9cc30f2b4fe05fb4cf1ac9327b1dfbc31521f3e1a3c6c9adaa2a8e2d"},{"id":"func/sliderTrack","name":"sliderTrack","line":48,"end_line":50,"hash":"cc5079771830a307eab4bb5eb1b33e63fa5819b994dde7c843e350f028c28471"},{"id":"func/Game.sliderFraction","name":"Game.sliderFraction","line":52,"end_line":64,"hash":"487c69697c7f1e9982809551dfdc38c25a119e65dbf06f08b1ba9e5792c67877"},{"id":"func/Game.sliderLabel","name":"Game.sliderLabel","line":66,"end_line":78,"hash":"a7b99076f669ab2dbbd6172d698a913a8ae06cb31509b411aec9a4d1499fbed9"},{"id":"func/Game.activeControl","name":"Game.activeControl","line":80,"end_line":85,"hash":"3eed94f83cd14347182b84c107b64c4f0c19de813cd00cb638c34d347ef6c16e"},{"id":"func/Game.activeRunControl","name":"Game.activeRunControl","line":87,"end_line":96,"hash":"a74f7cf31e2e9e64f2f6721614dee596a37e20131db80b533046b8ca93933446"},{"id":"func/Game.activeForceControl","name":"Game.activeForceControl","line":98,"end_line":101,"hash":"c325d1b27113b9be28ff09d2864a00e4267a652e4dc43dcdc1ee04b7a8f1560c"},{"id":"func/Game.activeParameterControl","name":"Game.activeParameterControl","line":111,"end_line":124,"hash":"64fe357ad27032dcee16ee76a6e733f85c8d1980d883d16d53c0cda31477cf6f"},{"id":"func/Game.activeWallControl","name":"Game.activeWallControl","line":126,"end_line":139,"hash":"70c0328208760cb2a63cb5074d116acfbe989380d2df61cbb2e3940ea1e2c3cd"},{"id":"func/visibleControls","name":"visibleControls","line":141,"end_line":145,"hash":"14e0313a1743b30c30268cf867272f99ac5a1880ef27fab3ac7641592af5459a"},{"id":"func/menuControls","name":"menuControls","line":147,"end_line":151,"hash":"7a632b167f77d172b422b42ab30311786f673b77fab82640220ac77b25e4fb31"},{"id":"func/Game.editMenuControls","name":"Game.editMenuControls","line":153,"end_line":162,"hash":"def493ef01ca37219a140b40fdc37d17ca821523d597ad4dad819ec2aaed486f"},{"id":"func/toolbarControls","name":"toolbarControls","line":164,"end_line":170,"hash":"ef6b992b59a86bfa58646d326682f9037168e35bbd2e8227951454a59975feee"},{"id":"func/commandControls","name":"commandControls","line":172,"end_line":184,"hash":"c8d380a126ca7b6b48433d0da82dcfce7908127eeda183a70ddce75a823d43f9"},{"id":"func/inspectorControls","name":"inspectorControls","line":186,"end_line":217,"hash":"3ca2ed7ef59ce7ce1d5834f30265b1180ac1f67f53e2be99c0ee76bfe6d68b6e"},{"id":"func/inspectorSections","name":"inspectorSections","line":219,"end_line":229,"hash":"5fdf15504029b40f9cd06bb7d711ab503fd2eef07e034797e07ab8e06552c120"},{"id":"func/inspectorLeft","name":"inspectorLeft","line":231,"end_line":233,"hash":"93e0ad43c5e22d7ec07af829c3f7383f3367334e90967f03635e5f28b8609df6"},{"id":"func/Game.forceEnabled","name":"Game.forceEnabled","line":235,"end_line":238,"hash":"56fb34648f486f07d22bfd46863fc1f92c7bd6b6e5c72769a1eb80e546ae49f7"},{"id":"func/Game.parameterEnabled","name":"Game.parameterEnabled","line":240,"end_line":242,"hash":"f418f5472049db1d0662156c0b97168f97de2048ae03ff6a5460834c55e2cc2c"},{"id":"func/Game.wallEnabled","name":"Game.wallEnabled","line":244,"end_line":247,"hash":"95636da6c99db771737f08d1638b66cd7b667ee7a3a1b5fbf23c23bca91629bf"},{"id":"func/Game.gridSnapEnabled","name":"Game.gridSnapEnabled","line":249,"end_line":251,"hash":"fe4c82bb8c4f8c739d9a9752b22185bebeccc822495ba9cf97444b77cb24008b"},{"id":"func/Game.statusFields","name":"Game.statusFields","line":253,"end_line":263,"hash":"bc12718557a80b85225052a0ec5890470eed24bc531abab8eec4e23fe811c9e1"},{"id":"func/Game.selectedObjectCount","name":"Game.selectedObjectCount","line":265,"end_line":278,"hash":"37228a151a2c37a8334e0743e75e66c4035294b511f5b8142948fd3fb0b2671b"},{"id":"func/Game.DrawFrameReport","name":"Game.DrawFrameReport","line":280,"end_line":282,"hash":"21457be3cb4856c4bb131972199ad58004071195dbfef0cb1500d22d228c9ad7"},{"id":"func/analyzeDrawnFrame","name":"analyzeDrawnFrame","line":284,"end_line":299,"hash":"d0700ff4f457c7c1e070bbfcb65b482f3218d571f25faf66f60670612ad08822"},{"id":"func/Game.visibleActiveControls","name":"Game.visibleActiveControls","line":301,"end_line":309,"hash":"a5dc9fa130f937bef9431d0f108f159dcded3984e1decb5c9ef5bda78d1a5cbc"},{"id":"func/visibleControlLabels","name":"visibleControlLabels","line":311,"end_line":317,"hash":"afac7bcb5fc0d6d4c034c1af2c507c9bbd31f7830feed44a0c4813e9016d77c2"},{"id":"func/visibleInspectorSections","name":"visibleInspectorSections","line":319,"end_line":325,"hash":"ae27149a3bf3e13945c3a1fe67e87d42d1929718201b50499740d6a243b046b9"},{"id":"func/Game.visibleStatusFields","name":"Game.visibleStatusFields","line":327,"end_line":333,"hash":"38815d71e80400a37fa2e84f2abcbd950d1814697f03bc89239fd64a0b12fcd8"},{"id":"func/Game.visibleRegionControlCounts","name":"Game.visibleRegionControlCounts","line":335,"end_line":345,"hash":"0e874837f3da80aa8aba265ae4a5e4b6ceae779a13512ddcf5804b91eb092ac4"},{"id":"func/visibleLabelsFit","name":"visibleLabelsFit","line":347,"end_line":351,"hash":"0fd4cb254764943e007f36732d9aa1d7ad707a59746a9cca2e31aab766226f88"},{"id":"func/controlLabelsFit","name":"controlLabelsFit","line":353,"end_line":355,"hash":"6cb92c088807face5a2162c78d338eb1278ed8774354b8b1d9ea884286f97a42"},{"id":"func/statusLabelsFit","name":"statusLabelsFit","line":357,"end_line":359,"hash":"9786f63a658af2dce883ca37f2fac1b7365902b152510d13c9787f1d5bdd412b"},{"id":"func/labelsFitItems","name":"labelsFitItems","line":361,"end_line":369,"hash":"dde06934ee8b31db69fe38357ce4bd6c6b73d991ef5d6f3f389e012fc3965dae"},{"id":"func/labelFits","name":"labelFits","line":371,"end_line":373,"hash":"60447d380d9ed2d8710772539abc4bc8e5a7bf62719fc471d97c16bb3ae7b3b3"},{"id":"func/visibleRegionRects","name":"visibleRegionRects","line":375,"end_line":383,"hash":"3b404313058f4f3c7f18cdbe0dad285f1fd85c8210a4de1b3f4f8eb272eda4b8"},{"id":"func/Game.visibleRegionPixels","name":"Game.visibleRegionPixels","line":385,"end_line":396,"hash":"d6393cc85bc819f0606076cc4256b5d8d14dc8bb4f0dbdd03b4d5b701bc636dc"},{"id":"func/regionControlPixels","name":"regionControlPixels","line":398,"end_line":406,"hash":"09baade91c8b970cf2ec8ce4dd168c323ca63cf59170b8f129019bedef0ea743"},{"id":"func/Game.regionStatusPixels","name":"Game.regionStatusPixels","line":408,"end_line":417,"hash":"06db0a8e72e4a78fe3bc2316755c0e94a6a9249933e36bfd46789da7ab09742e"},{"id":"func/rectPixels","name":"rectPixels","line":419,"end_line":421,"hash":"be91e5c34925744d4284c9bb7111688f4f62eef60f40d9b92c77353a60433f83"},{"id":"func/visibleWorldPixels","name":"visibleWorldPixels","line":423,"end_line":439,"hash":"ba9902c2e580234ff3188d18afd2e517404988a22ad8ef20dd544fae9f51866c"}]}
// mutate4go-manifest-end
