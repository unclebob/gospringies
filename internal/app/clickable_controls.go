package app

import (
	"image"
	"math"
	"strconv"

	"springs/internal/sim"
)

var visibleControlCommands = map[string]string{
	"run command":           "run",
	"pause command":         "pause",
	"reset command":         "reset",
	"save state command":    "save state",
	"restore state command": "restore state",
	"load command":          "load",
	"insert command":        "insert",
	"save command":          "save",
	"quit command":          "quit",
	"select all command":    "select all",
	"duplicate command":     "duplicate",
	"delete command":        "delete",
	"cut command":           "cut",
	"copy command":          "copy",
	"paste command":         "paste",
}

// ClickAt handles pointer activation for the same rectangles Draw uses for controls.
func (g *Game) ClickAt(x int, y int) bool {
	g.pointer.lastCursor = g.screenToWorld(simVec(x, y))
	if control, ok := g.editMenuControlAt(image.Pt(x, y)); ok {
		return g.activateVisibleControl(control)
	}
	if control, ok := visibleControlAt(image.Pt(x, y)); ok {
		return g.clickVisibleControl(control, x)
	}
	return g.clickAwayFromVisibleControls()
}

func (g *Game) clickVisibleControl(control controlBox, x int) bool {
	if setting, ok := numericSettingForTextField(control.Name); ok {
		g.focusNumericSettingTextField(setting)
		return true
	}
	if setting, delta, ok := numericSettingForStepButton(control.Name); ok {
		g.activeNumericStep = control.Name
		g.numericStepTicks = 0
		g.stepNumericSetting(setting, delta)
		return true
	}
	if isSliderControl(control.Name) {
		g.activeSlider = control.Name
		g.setSliderAt(control.Name, x)
		return true
	}
	return g.activateVisibleControl(control)
}

func (g *Game) clickAwayFromVisibleControls() bool {
	if g.editMenuOpen {
		g.editMenuOpen = false
		return true
	}
	g.cancelNumericSettingInput()
	return false
}

func (g *Game) ClickVisibleControl(label string) bool {
	control, ok := visibleControlWithLabel(label)
	if !ok {
		return false
	}
	return g.activateVisibleControl(control)
}

func (g *Game) VisibleControlBounds(label string) (image.Rectangle, bool) {
	control, ok := visibleControlWithLabel(label)
	if !ok {
		return image.Rectangle{}, false
	}
	return control.Rect, true
}

func (g *Game) activateVisibleControl(control controlBox) bool {
	if control.Name == "edit menu" {
		g.editMenuOpen = !g.editMenuOpen
		return true
	}
	if command, ok := visibleControlCommands[control.Name]; ok {
		g.editMenuOpen = false
		g.RunCommand(command)
		return true
	}
	if g.activateInspectorControl(control.Name) {
		return true
	}
	return false
}

func (g *Game) editMenuControlAt(point image.Point) (controlBox, bool) {
	for _, control := range g.editMenuControls() {
		if point.In(control.Rect) {
			return control, true
		}
	}
	return controlBox{}, false
}

func (g *Game) activateInspectorControl(name string) bool {
	action, ok := inspectorControlActions[name]
	if !ok {
		return false
	}
	action(g)
	g.dirty = true
	return true
}

var inspectorControlActions = map[string]func(*Game){
	"mass parameter":          func(g *Game) { g.stepEditorControl("mass", 1) },
	"elasticity parameter":    func(g *Game) { g.stepEditorControl("elasticity", 0.1) },
	"fixed mass toggle":       func(g *Game) { g.toggleFixedMass() },
	"kspring parameter":       func(g *Game) { g.stepEditorControl("Kspring", 1) },
	"kdamp parameter":         func(g *Game) { g.stepEditorControl("Kdamp", 0.1) },
	"set rest length command": func(g *Game) { _ = g.editing().SetRestLength() },
	"spring wall toggle":      func(g *Game) { g.toggleSelectedSpringWall() },
	"gravity force":           func(g *Game) { g.toggleForce("gravity", map[string]string{"magnitude": "10", "direction": "0"}) },
	"center attraction force": func(g *Game) {
		g.toggleForce("center attraction", map[string]string{"magnitude": "10", "exponent": "0"})
	},
	"center mass force": func(g *Game) {
		g.toggleForce("center of mass attraction", map[string]string{"magnitude": "5", "damping": "2"})
	},
	"wall repulsion force": func(g *Game) {
		g.toggleForce("wall repulsion", map[string]string{"magnitude": "10000", "exponent": "1"})
	},
	"mass collision force":     func(g *Game) { g.toggleForce("mass collision", map[string]string{}) },
	"set center command":       func(g *Game) { g.simulation.SetForceCenter(g.selectedMassIDs()) },
	"top wall toggle":          func(g *Game) { g.toggleWall("top") },
	"bottom wall toggle":       func(g *Game) { g.toggleWall("bottom") },
	"left wall toggle":         func(g *Game) { g.toggleWall("left") },
	"right wall toggle":        func(g *Game) { g.toggleWall("right") },
	"grid snap toggle":         func(g *Game) { g.toggleGridSnap() },
	"show springs toggle":      func(g *Game) { g.toggleParameter("show springs") },
	"viscosity parameter":      func(g *Game) { g.stepParameter("viscosity", 0.1) },
	"stickiness parameter":     func(g *Game) { g.stepParameter("stickiness", 0.1) },
	"timestep parameter":       func(g *Game) { g.stepParameter("timestep", 0.001) },
	"precision parameter":      func(g *Game) { g.stepParameter("precision", 0.001) },
	"adaptive timestep toggle": func(g *Game) { g.toggleParameter("adaptive timestep") },
}

func (g *Game) setSliderAt(name string, x int) {
	setting, ok := numericSettingForSlider(name)
	if !ok {
		return
	}
	g.setNumericSettingFromSlider(setting, x)
}

func (g *Game) continueNumericStepHold() {
	setting, delta, ok := numericSettingForStepButton(g.activeNumericStep)
	if !ok {
		g.activeNumericStep = ""
		g.numericStepTicks = 0
		return
	}
	g.numericStepTicks++
	if g.numericStepTicks < numericStepHoldDelayTicks {
		return
	}
	if (g.numericStepTicks-numericStepHoldDelayTicks)%numericStepRepeatTicks == 0 {
		g.stepNumericSetting(setting, delta)
	}
}

func sliderFractionAt(track image.Rectangle, x int) float64 {
	width := track.Dx()
	if width <= 0 {
		return 0
	}
	return clampFloat(float64(x-track.Min.X)/float64(width), 0, 1)
}

func (g *Game) stepEditorControl(control string, delta float64) {
	value := g.parameterForEditorControl(control) + delta
	_ = g.editing().ChangeControl(control, formatControlFloat(value))
}

func (g *Game) parameterForEditorControl(control string) float64 {
	parameter, ok := editorControlParameters[control]
	if !ok {
		return 0
	}
	return g.parameterFloat(parameter)
}

var editorControlParameters = map[string]string{
	"mass":       "current mass",
	"elasticity": "elasticity",
	"Kspring":    "spring constant",
	"Kdamp":      "damping",
}

func (g *Game) toggleFixedMass() {
	value := strconv.FormatBool(!g.parameterEnabled("fixed mass"))
	g.simulation.Parameters.Set("fixed mass", value)
	_ = g.editing().ChangeControl("fixed", value)
}

func (g *Game) toggleSelectedSpringWall() {
	value := strconv.FormatBool(!g.activeSelectedSpringControl("spring wall toggle"))
	_ = g.editing().ChangeControl("Wall", value)
}

func (g *Game) toggleForce(name string, defaults map[string]string) {
	force, _ := g.simulation.Parameters.Force(name)
	if force.Enabled == "true" {
		force.Enabled = "false"
		if force.Values == nil {
			force.Values = map[string]string{}
		}
		g.simulation.Parameters.Forces[name] = force
		g.simulation.Parameters.SelectForce(name)
		return
	}
	g.simulation.Parameters.EnableForce(name, defaults)
}

func (g *Game) stepForceValue(forceName string, valueName string, delta float64) {
	g.setForceValue(forceName, valueName, forceValueFloat(g.forceConfig(forceName), valueName)+delta)
}

func (g *Game) setForceValue(forceName string, valueName string, value float64) {
	force := g.forceConfig(forceName)
	if force.Values == nil {
		force.Values = map[string]string{}
	}
	if force.Enabled != "true" {
		force.Enabled = "true"
	}
	force.Values[valueName] = formatControlFloat(value)
	g.simulation.Parameters.Forces[forceName] = force
	g.simulation.Parameters.SelectForce(forceName)
}

func (g *Game) forceConfig(forceName string) sim.ForceConfig {
	force, _ := g.simulation.Parameters.Force(forceName)
	if force.Values == nil {
		force.Values = map[string]string{}
	}
	return force
}

func forceValueFloat(force sim.ForceConfig, name string) float64 {
	value, _ := strconv.ParseFloat(force.Values[name], 64)
	return value
}

func (g *Game) toggleWall(name string) {
	g.simulation.Parameters.Walls[name] = !g.simulation.Parameters.Walls[name]
}

func (g *Game) toggleGridSnap() {
	if g.gridSnapEnabled() {
		g.simulation.Parameters.Set("grid snap", "0")
		return
	}
	g.simulation.Parameters.Set("grid snap", "10")
}

func (g *Game) toggleParameter(name string) {
	g.simulation.Parameters.Set(name, strconv.FormatBool(!g.parameterEnabled(name)))
}

func (g *Game) stepParameter(name string, delta float64) {
	g.simulation.Parameters.Set(name, formatControlFloat(g.parameterFloat(name)+delta))
}

func (g *Game) selectedMassIDs() []int {
	var ids []int
	for id, selected := range g.editing().SelectedMasses {
		if selected {
			ids = append(ids, id)
		}
	}
	return ids
}

func (g *Game) parameterFloat(name string) float64 {
	value, _ := strconv.ParseFloat(g.simulation.Parameters.Value(name), 64)
	return value
}

func (g *Game) gridSnapSize() float64 {
	return g.parameterFloat("grid snap")
}

func formatControlFloat(value float64) string {
	return strconv.FormatFloat(roundControlFloat(value), 'f', -1, 64)
}

func roundControlFloat(value float64) float64 {
	return math.Round(value*1000000) / 1000000
}

func clampFloat(value float64, min float64, max float64) float64 {
	return math.Min(math.Max(value, min), max)
}

func visibleControlAt(point image.Point) (controlBox, bool) {
	for _, control := range visibleControls() {
		if point.In(control.Rect) {
			return control, true
		}
	}
	return controlBox{}, false
}

func visibleControlWithLabel(label string) (controlBox, bool) {
	for _, control := range visibleControls() {
		if control.Label == label {
			return control, true
		}
	}
	return controlBox{}, false
}

func visibleControlWithName(name string) (controlBox, bool) {
	for _, control := range visibleControls() {
		if control.Name == name {
			return control, true
		}
	}
	return controlBox{}, false
}

func (g *Game) PathEntryCommand() string {
	return g.pathEntryCommand
}

func (g *Game) DemoPickerOpen() bool {
	return g.demoPickerOpen
}

func (g *Game) VisibleControlActive(label string) bool {
	control, ok := visibleControlWithLabel(label)
	return ok && g.activeControl(control.Name)
}

func (g *Game) DragMass(id int, position sim.Vec2) bool {
	if g.editing().MassSelected(id) {
		return g.dragSelectedMasses(position)
	}
	return g.dragSingleMass(id, position)
}

func (g *Game) dragSelectedMasses(position sim.Vec2) bool {
	if len(g.pointer.draggingOffsets) > 0 {
		g.applyDraggingOffsets(position)
	} else {
		g.moveSelectedMasses(position.Sub(g.pointer.draggingLast))
	}
	g.finishMassDragStep(position)
	return true
}

func (g *Game) dragSingleMass(id int, position sim.Vec2) bool {
	for i := range g.simulation.Masses {
		if g.simulation.Masses[i].ID == id {
			g.simulation.Masses[i].Position = g.snapToCanvas(position)
			g.simulation.Masses[i].Velocity = sim.Vec2{}
			g.finishMassDragStep(position)
			return true
		}
	}
	return false
}

func (g *Game) finishMassDragStep(position sim.Vec2) {
	g.pointer.draggingLast = position
	if !selectionClick(g.pointer.draggingStart, position) {
		g.pointer.dragMoved = true
	}
	g.dirty = true
}

func (g *Game) moveSelectedMasses(delta sim.Vec2) {
	for i := range g.simulation.Masses {
		if g.editing().MassSelected(g.simulation.Masses[i].ID) {
			g.simulation.Masses[i].Position = g.snapToCanvas(g.simulation.Masses[i].Position.Add(delta))
			g.simulation.Masses[i].Velocity = sim.Vec2{}
		}
	}
}

func (g *Game) applyDraggingOffsets(cursor sim.Vec2) {
	for i := range g.simulation.Masses {
		offset, ok := g.pointer.draggingOffsets[g.simulation.Masses[i].ID]
		if ok {
			g.simulation.Masses[i].Position = g.snapToCanvas(cursor.Add(offset))
			g.simulation.Masses[i].Velocity = sim.Vec2{}
		}
	}
}

func (g *Game) snapToGrid(position sim.Vec2) sim.Vec2 {
	size := g.gridSnapSize()
	if size <= 0 {
		return position
	}
	return sim.Vec2{
		X: math.Round(position.X/size) * size,
		Y: math.Round(position.Y/size) * size,
	}
}

// mutate4go-manifest-begin
// {"version":1,"tested_at":"2026-05-22T07:52:46-05:00","module_hash":"0ad46f4bab251df8dd005d000c434b071a53091a8d4684ddc2fb9350653275c6","functions":[{"id":"func/Game.ClickAt","name":"Game.ClickAt","line":30,"end_line":39,"hash":"abd8c850c09ded9a7bbbef67650b9dd812456e2232d07c865c4cfb530daefaf2"},{"id":"func/Game.clickVisibleControl","name":"Game.clickVisibleControl","line":41,"end_line":58,"hash":"dac4b1a08a4c94c7551d9a7f81040ce0c3423734f98a834631f7c9f857d66017"},{"id":"func/Game.clickAwayFromVisibleControls","name":"Game.clickAwayFromVisibleControls","line":60,"end_line":67,"hash":"7f6cee53d5433f287591fefd83c8d03e139a83be8138dd4cb2b890b1119b6ba2"},{"id":"func/Game.ClickVisibleControl","name":"Game.ClickVisibleControl","line":69,"end_line":75,"hash":"62c99a99facde1559b3bb3d3d4e7689177529881a2c1525986f8f4858af3f1bf"},{"id":"func/Game.VisibleControlBounds","name":"Game.VisibleControlBounds","line":77,"end_line":83,"hash":"c31a71819009fcde81e5dc37f01757124583e6d4d8143a37da74a8cb668768cf"},{"id":"func/Game.activateVisibleControl","name":"Game.activateVisibleControl","line":85,"end_line":99,"hash":"6793dc6c2b5bc5363f90fe8bb47d76f63d87ffb0da71284367eb7761c56ce174"},{"id":"func/Game.editMenuControlAt","name":"Game.editMenuControlAt","line":101,"end_line":108,"hash":"5d644c398ced9c63cca3dbc0689576ce08c17f1c301c10f0f5746faa73e8b60d"},{"id":"func/Game.activateInspectorControl","name":"Game.activateInspectorControl","line":110,"end_line":118,"hash":"af198a5b9677d55494e8f3c3958befbdde6f9f25cc9fcfcadcf956e6c0fe8bc5"},{"id":"func/Game.setSliderAt","name":"Game.setSliderAt","line":153,"end_line":159,"hash":"47889db5165c085b486cab5c237e6f1f4c52fab5a7059db51badb7a196813d93"},{"id":"func/Game.continueNumericStepHold","name":"Game.continueNumericStepHold","line":161,"end_line":175,"hash":"383137d6375bc5b1a6a52213bbe260083030f7da9097829e72748afa5d8b32c6"},{"id":"func/sliderFractionAt","name":"sliderFractionAt","line":177,"end_line":183,"hash":"8df037c485beccdda8a0cbe074c1cfcf010b61cecc12bd50c1dbfebb22762b8e"},{"id":"func/Game.stepEditorControl","name":"Game.stepEditorControl","line":185,"end_line":188,"hash":"908abc50cefb5bef5cba1cd1aa89359d53aba81e8de8c7a6ba1e04ff966d6f9e"},{"id":"func/Game.parameterForEditorControl","name":"Game.parameterForEditorControl","line":190,"end_line":196,"hash":"f0cb615eff406cfcaad996ceed2b2043a543338a8d9cd7603c4c35ccd54bb061"},{"id":"func/Game.toggleFixedMass","name":"Game.toggleFixedMass","line":205,"end_line":209,"hash":"7375fb7830a8240deb32bc922fdae697c28b685ea7a17ec3df61f0ab8e4bf580"},{"id":"func/Game.toggleSelectedSpringWall","name":"Game.toggleSelectedSpringWall","line":211,"end_line":214,"hash":"2d0dab9591647348c65edfdb7e93475c41d737726ea497ce623062177785615e"},{"id":"func/Game.toggleForce","name":"Game.toggleForce","line":216,"end_line":228,"hash":"e4885d2a6af0304cdd3a404b54d7a9e84acb84d7ed62326b2d8c7e9040080acd"},{"id":"func/Game.stepForceValue","name":"Game.stepForceValue","line":230,"end_line":232,"hash":"0e88dce860c544e1d7ae4a45d509db1b277425eb54c51ea88858648c9152c154"},{"id":"func/Game.setForceValue","name":"Game.setForceValue","line":234,"end_line":245,"hash":"9dc52a916bd907b92a2466a83057b1992e5c4d3c430365a4715b170bb4ea299b"},{"id":"func/Game.forceConfig","name":"Game.forceConfig","line":247,"end_line":253,"hash":"268a1f3034abaf9e5b29178f8b515eb9f92dbe1c754f478b58cf8cd498fdb590"},{"id":"func/forceValueFloat","name":"forceValueFloat","line":255,"end_line":258,"hash":"49ace1f2c99bcee1c74beb91d9c41aa20c5d3fcdfe64218b52a24987d28e2ddb"},{"id":"func/Game.toggleWall","name":"Game.toggleWall","line":260,"end_line":262,"hash":"dc9babca921d07624cfb1c5a31cc22bc56f19c433370b48247fc2f81798eb032"},{"id":"func/Game.toggleGridSnap","name":"Game.toggleGridSnap","line":264,"end_line":270,"hash":"2817114e27d8051a7538662f0883e3b7ac917693ae26b29bfb8b6172301800aa"},{"id":"func/Game.toggleParameter","name":"Game.toggleParameter","line":272,"end_line":274,"hash":"e244dff943c38b54b303f95fd4443bd14cbfeb7a19d2528a60c8e894e0a34598"},{"id":"func/Game.stepParameter","name":"Game.stepParameter","line":276,"end_line":278,"hash":"49ccea30a073fdc685a71d778da794146bcd1ed22a6e310db5e8fe61d828601a"},{"id":"func/Game.selectedMassIDs","name":"Game.selectedMassIDs","line":280,"end_line":288,"hash":"c22bc0b7c51e606264c25a2b1f13d5770a815f27bbf9edb542a1a7a95215ccd2"},{"id":"func/Game.parameterFloat","name":"Game.parameterFloat","line":290,"end_line":293,"hash":"763e76f48f23a676153bc74047d3a5b5931a900bdf4930b5d5756604ab102da0"},{"id":"func/Game.gridSnapSize","name":"Game.gridSnapSize","line":295,"end_line":297,"hash":"2f848d29a5ed78e7601534d822f5588903cc01e5006b92b46e55fcfbf479b594"},{"id":"func/formatControlFloat","name":"formatControlFloat","line":299,"end_line":301,"hash":"b19b7aacb786da1791f10cb92f30c3b9902c9804613d34efcd6f7da207db4a9d"},{"id":"func/roundControlFloat","name":"roundControlFloat","line":303,"end_line":305,"hash":"5c3c02047df5405fa39aa22ef6c906165e410f44d101726a323894fcf3c84a75"},{"id":"func/clampFloat","name":"clampFloat","line":307,"end_line":309,"hash":"ab5544e501ecda24c07fea5d3a7db0d688bd2961d0e4deaae4b8a8e6eb5f6c52"},{"id":"func/visibleControlAt","name":"visibleControlAt","line":311,"end_line":318,"hash":"f049ae667df6f18d325f52daa1c49dea3a0f91b5bdd6e131fc4453478fe3f119"},{"id":"func/visibleControlWithLabel","name":"visibleControlWithLabel","line":320,"end_line":327,"hash":"62b1458cf500759991a56926e8ba0dba602ab67f870e8f1dd4ceb47232ab6341"},{"id":"func/visibleControlWithName","name":"visibleControlWithName","line":329,"end_line":336,"hash":"c4ef38e860c00f329a88beae2957a5e872ed77f9dd8e6ba9143d7acb4a024302"},{"id":"func/Game.PathEntryCommand","name":"Game.PathEntryCommand","line":338,"end_line":340,"hash":"1fb1f605408b9e1533433a584f43ee538d97ae21e42c6db0e6bd17f65fcc6d9b"},{"id":"func/Game.DemoPickerOpen","name":"Game.DemoPickerOpen","line":342,"end_line":344,"hash":"e560d2fb41f9965ce66f7f0bfba1e96eb38d25c1e5d26c2b4301438e3a752ed5"},{"id":"func/Game.VisibleControlActive","name":"Game.VisibleControlActive","line":346,"end_line":349,"hash":"62378e0437078c58522cd0fb73fdf959154000d20d4a06ce524d11d50ac3bbeb"},{"id":"func/Game.DragMass","name":"Game.DragMass","line":351,"end_line":356,"hash":"54e365a24dad0e7d7d748b16f5c3583af78add366f83ad03f81db661f1c9355f"},{"id":"func/Game.dragSelectedMasses","name":"Game.dragSelectedMasses","line":358,"end_line":366,"hash":"24b4d06566a81bd5006ae43ea6808d1170f96df1c322fb777a5ef18e7da6df32"},{"id":"func/Game.dragSingleMass","name":"Game.dragSingleMass","line":368,"end_line":378,"hash":"0535634cd4a66f20ee188edad84e632b598c3a13807db600954650169bb64f18"},{"id":"func/Game.finishMassDragStep","name":"Game.finishMassDragStep","line":380,"end_line":386,"hash":"a7954983890cf003679bc5fe76584b50be113169f481bb2774775dab7c1d145f"},{"id":"func/Game.moveSelectedMasses","name":"Game.moveSelectedMasses","line":388,"end_line":395,"hash":"36e44ff1e6837878f5fe7e3cbc286a00250c07e42550de26e549974da402f291"},{"id":"func/Game.applyDraggingOffsets","name":"Game.applyDraggingOffsets","line":397,"end_line":405,"hash":"78cc5ffc8788be91ab55d2200afbb7998aff6470633931df764879e573b6dc30"},{"id":"func/Game.snapToGrid","name":"Game.snapToGrid","line":407,"end_line":416,"hash":"e7b93e39ef986bc33f6bedbab4eb997d9433a2e7267e430186f91dffd802870a"}]}
// mutate4go-manifest-end
