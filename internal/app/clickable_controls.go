package app

import (
	"image"
	"math"
	"strconv"

	"springs/internal/sim"
)

var visibleControlCommands = map[string]string{
	"run pause toggle command": "pause toggle",
	"reset command":            "reset",
	"save state command":       "save state",
	"restore state command":    "restore state",
	"load command":             "load",
	"insert command":           "insert",
	"save command":             "save",
	"quit command":             "quit",
	"select all command":       "select all",
	"duplicate command":        "duplicate",
	"delete command":           "delete",
	"cut command":              "cut",
	"copy command":             "copy",
	"paste command":            "paste",
}

// ClickAt handles pointer activation for the same rectangles Draw uses for controls.
func (g *Game) ClickAt(x int, y int) bool {
	g.pointer.lastCursor = g.screenToWorld(simVec(x, y))
	if control, ok := g.editMenuControlAt(image.Pt(x, y)); ok {
		return g.activateVisibleControl(control)
	}
	if control, ok := g.visibleControlAt(image.Pt(x, y)); ok {
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
		g.controls.activeNumericStep = control.Name
		g.controls.numericStepTicks = 0
		g.stepNumericSetting(setting, delta)
		return true
	}
	if isSliderControl(control.Name) {
		g.controls.activeSlider = control.Name
		g.setSliderAt(control.Name, x)
		return true
	}
	return g.activateVisibleControl(control)
}

func (g *Game) clickAwayFromVisibleControls() bool {
	if g.controls.editMenuOpen {
		g.controls.editMenuOpen = false
		return true
	}
	g.cancelNumericSettingInput()
	return false
}

func (g *Game) ClickVisibleControl(label string) bool {
	control, ok := g.visibleControlWithLabel(label)
	if !ok {
		return false
	}
	return g.activateVisibleControl(control)
}

func (g *Game) VisibleControlBounds(label string) (image.Rectangle, bool) {
	control, ok := g.visibleControlWithLabel(label)
	if !ok {
		return image.Rectangle{}, false
	}
	return control.Rect, true
}

func (g *Game) activateVisibleControl(control controlBox) bool {
	if control.Name == "edit menu" {
		g.controls.editMenuOpen = !g.controls.editMenuOpen
		return true
	}
	if command, ok := visibleControlCommands[control.Name]; ok {
		g.controls.editMenuOpen = false
		g.RunCommand(command)
		return true
	}
	if g.activateInspectorControl(control.Name) {
		return true
	}
	return false
}

func (g *Game) editMenuControlAt(point image.Point) (controlBox, bool) {
	return controlAt(point, g.editMenuControls())
}

func (g *Game) activateInspectorControl(name string) bool {
	action, ok := inspectorControlActions[name]
	if !ok {
		return false
	}
	action(g)
	g.markDirty()
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
	"set center command":       func(g *Game) { g.world.simulation.SetForceCenter(g.selectedMassIDs()) },
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
	setting, delta, ok := numericSettingForStepButton(g.controls.activeNumericStep)
	if !ok {
		g.controls.activeNumericStep = ""
		g.controls.numericStepTicks = 0
		return
	}
	g.controls.numericStepTicks++
	if g.controls.numericStepTicks < numericStepHoldDelayTicks {
		return
	}
	if g.controls.numericStepTicks%numericStepRepeatTicks == 0 {
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
	g.world.simulation.Parameters.Set("fixed mass", value)
	_ = g.editing().ChangeControl("fixed", value)
}

func (g *Game) toggleSelectedSpringWall() {
	value := strconv.FormatBool(!g.activeSelectedSpringControl("spring wall toggle"))
	_ = g.editing().ChangeControl("Wall", value)
}

func (g *Game) toggleForce(name string, defaults map[string]string) {
	force, _ := g.world.simulation.Parameters.Force(name)
	if force.Enabled == "true" {
		force.Enabled = "false"
		if force.Values == nil {
			force.Values = map[string]string{}
		}
		g.world.simulation.Parameters.Forces[name] = force
		g.world.simulation.Parameters.SelectForce(name)
		return
	}
	g.world.simulation.Parameters.EnableForce(name, defaults)
}

func (g *Game) stepForceValue(forceName string, valueName string, delta float64) {
	g.setForceValue(forceName, valueName, forceValueFloat(g.forceConfig(forceName), valueName)+delta)
}

func (g *Game) setForceValue(forceName string, valueName string, value float64) {
	force := g.forceConfig(forceName)
	force.Values = nonNilStringMap(force.Values)
	if force.Enabled != "true" {
		force.Enabled = "true"
	}
	force.Values[valueName] = formatControlFloat(value)
	g.world.simulation.Parameters.Forces[forceName] = force
	g.world.simulation.Parameters.SelectForce(forceName)
}

func (g *Game) forceConfig(forceName string) sim.ForceConfig {
	force, _ := g.world.simulation.Parameters.Force(forceName)
	force.Values = nonNilStringMap(force.Values)
	return force
}

func nonNilStringMap(values map[string]string) map[string]string {
	if values != nil {
		return values
	}
	return map[string]string{}
}

func forceValueFloat(force sim.ForceConfig, name string) float64 {
	value, _ := strconv.ParseFloat(force.Values[name], 64)
	return value
}

func (g *Game) toggleWall(name string) {
	g.world.simulation.Parameters.Walls[name] = !g.world.simulation.Parameters.Walls[name]
}

func (g *Game) toggleGridSnap() {
	if g.gridSnapEnabled() {
		g.world.simulation.Parameters.Set("grid snap", "0")
		return
	}
	g.world.simulation.Parameters.Set("grid snap", "10")
}

func (g *Game) toggleParameter(name string) {
	g.world.simulation.Parameters.Set(name, strconv.FormatBool(!g.parameterEnabled(name)))
}

func (g *Game) stepParameter(name string, delta float64) {
	g.world.simulation.Parameters.Set(name, formatControlFloat(g.parameterFloat(name)+delta))
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
	value, _ := strconv.ParseFloat(g.world.simulation.Parameters.Value(name), 64)
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

func (g *Game) visibleControlAt(point image.Point) (controlBox, bool) {
	return controlAt(point, g.visibleControls())
}

func controlAt(point image.Point, controls []controlBox) (controlBox, bool) {
	for _, control := range controls {
		if point.In(control.Rect) {
			return control, true
		}
	}
	return controlBox{}, false
}

func (g *Game) visibleControlWithLabel(label string) (controlBox, bool) {
	return g.visibleControlWithField(label, func(control controlBox) string { return control.Label })
}

func (g *Game) visibleControlWithName(name string) (controlBox, bool) {
	return g.visibleControlWithField(name, func(control controlBox) string { return control.Name })
}

func (g *Game) visibleControlWithField(value string, field func(controlBox) string) (controlBox, bool) {
	for _, control := range g.visibleControls() {
		if field(control) == value {
			return control, true
		}
	}
	return controlBox{}, false
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
	return visibleControlWithField(label, func(control controlBox) string { return control.Label })
}

func visibleControlWithName(name string) (controlBox, bool) {
	return visibleControlWithField(name, func(control controlBox) string { return control.Name })
}

func visibleControlWithField(value string, field func(controlBox) string) (controlBox, bool) {
	for _, control := range visibleControls() {
		if field(control) == value {
			return control, true
		}
	}
	return controlBox{}, false
}

func (g *Game) PathEntryCommand() string {
	return g.document.pathEntryCommand
}

func (g *Game) DemoPickerOpen() bool {
	return g.controls.demoPickerOpen
}

func (g *Game) VisibleControlActive(label string) bool {
	control, ok := g.visibleControlWithLabel(label)
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
	for i := range g.world.simulation.Masses {
		if g.world.simulation.Masses[i].ID == id {
			g.world.simulation.Masses[i].Position = g.snapToCanvas(position)
			g.world.simulation.Masses[i].Velocity = sim.Vec2{}
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
	g.markDirty()
}

func (g *Game) moveSelectedMasses(delta sim.Vec2) {
	for i := range g.world.simulation.Masses {
		if g.editing().MassSelected(g.world.simulation.Masses[i].ID) {
			g.world.simulation.Masses[i].Position = g.snapToCanvas(g.world.simulation.Masses[i].Position.Add(delta))
			g.world.simulation.Masses[i].Velocity = sim.Vec2{}
		}
	}
}

func (g *Game) applyDraggingOffsets(cursor sim.Vec2) {
	for i := range g.world.simulation.Masses {
		offset, ok := g.pointer.draggingOffsets[g.world.simulation.Masses[i].ID]
		if ok {
			g.world.simulation.Masses[i].Position = g.snapToCanvas(cursor.Add(offset))
			g.world.simulation.Masses[i].Velocity = sim.Vec2{}
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
// {"version":1,"tested_at":"2026-05-22T18:07:46-05:00","module_hash":"bf74de085dc4c38c2ba2bf9bff0856d63e46fc0048d253e3c27c8a59dd010e5c","functions":[{"id":"func/Game.ClickAt","name":"Game.ClickAt","line":29,"end_line":38,"hash":"874db5531ea44665abbf641f47124862abeaa930acf20834552e948b865a1786"},{"id":"func/Game.clickVisibleControl","name":"Game.clickVisibleControl","line":40,"end_line":57,"hash":"23fccde47def1951406b74e466dd4b772ac7bfe667c203c23d5f59ad6dfad7d7"},{"id":"func/Game.clickAwayFromVisibleControls","name":"Game.clickAwayFromVisibleControls","line":59,"end_line":66,"hash":"c1a348181832e9f0ac292e8331c17347a39e134d9786559006fac9084a7646f5"},{"id":"func/Game.ClickVisibleControl","name":"Game.ClickVisibleControl","line":68,"end_line":74,"hash":"de5ef559290d773af5aa912fa27fbc12dbda540aab1b9d0fe5994fc643ab9b37"},{"id":"func/Game.VisibleControlBounds","name":"Game.VisibleControlBounds","line":76,"end_line":82,"hash":"455f62f053098234bb6b0c7f1dfdb959363310089a709e60b10b82aae59ed683"},{"id":"func/Game.activateVisibleControl","name":"Game.activateVisibleControl","line":84,"end_line":98,"hash":"5134e0f80211e06a5f1f23e39f67c35f06bd6e25e54495357b4ba0237e5685fd"},{"id":"func/Game.editMenuControlAt","name":"Game.editMenuControlAt","line":100,"end_line":102,"hash":"c56508a83ca332354319ee2b40e108ff4cbfaac8dead41ee5b2e0dabeecab1d6"},{"id":"func/Game.activateInspectorControl","name":"Game.activateInspectorControl","line":104,"end_line":112,"hash":"a12881093386af2b234bafcb3b0e071d3c0ef7a628798e7df621d6e540622fc9"},{"id":"func/Game.setSliderAt","name":"Game.setSliderAt","line":147,"end_line":153,"hash":"47889db5165c085b486cab5c237e6f1f4c52fab5a7059db51badb7a196813d93"},{"id":"func/Game.continueNumericStepHold","name":"Game.continueNumericStepHold","line":155,"end_line":169,"hash":"0d1e6fcd617acdd319eb8c732517ceeecb6bfb24b013d19ae4305e1781f302bd"},{"id":"func/sliderFractionAt","name":"sliderFractionAt","line":171,"end_line":177,"hash":"8df037c485beccdda8a0cbe074c1cfcf010b61cecc12bd50c1dbfebb22762b8e"},{"id":"func/Game.stepEditorControl","name":"Game.stepEditorControl","line":179,"end_line":182,"hash":"908abc50cefb5bef5cba1cd1aa89359d53aba81e8de8c7a6ba1e04ff966d6f9e"},{"id":"func/Game.parameterForEditorControl","name":"Game.parameterForEditorControl","line":184,"end_line":190,"hash":"f0cb615eff406cfcaad996ceed2b2043a543338a8d9cd7603c4c35ccd54bb061"},{"id":"func/Game.toggleFixedMass","name":"Game.toggleFixedMass","line":199,"end_line":203,"hash":"47cf18b8b950b843a02574b7b1cd37ff7ac98afc50a65b49f9dbeff3db0cdbfc"},{"id":"func/Game.toggleSelectedSpringWall","name":"Game.toggleSelectedSpringWall","line":205,"end_line":208,"hash":"2d0dab9591647348c65edfdb7e93475c41d737726ea497ce623062177785615e"},{"id":"func/Game.toggleForce","name":"Game.toggleForce","line":210,"end_line":222,"hash":"eda87ab2868dc0ee663e6c6029734a6e00fab05e4fd6e3f0589206157ccf43d1"},{"id":"func/Game.stepForceValue","name":"Game.stepForceValue","line":224,"end_line":226,"hash":"0e88dce860c544e1d7ae4a45d509db1b277425eb54c51ea88858648c9152c154"},{"id":"func/Game.setForceValue","name":"Game.setForceValue","line":228,"end_line":237,"hash":"bd1a96a405506df7162278452e8a1ad1e6655a6a0d6093e27bd5c2934d88371f"},{"id":"func/Game.forceConfig","name":"Game.forceConfig","line":239,"end_line":243,"hash":"524bb1bb411d52159c25ca4c416b8825d03abc0cf54864adc81ca6c6e5cd873f"},{"id":"func/nonNilStringMap","name":"nonNilStringMap","line":245,"end_line":250,"hash":"6eeb8683720d95e89c0e8fe633a4778bec13895f307c8aece1bf1b2e82cd7a0b"},{"id":"func/forceValueFloat","name":"forceValueFloat","line":252,"end_line":255,"hash":"49ace1f2c99bcee1c74beb91d9c41aa20c5d3fcdfe64218b52a24987d28e2ddb"},{"id":"func/Game.toggleWall","name":"Game.toggleWall","line":257,"end_line":259,"hash":"d9aafe71be296e264e0fa2cd7bbaf0c789e1e113910bd3d1fc995a7169185ee9"},{"id":"func/Game.toggleGridSnap","name":"Game.toggleGridSnap","line":261,"end_line":267,"hash":"87a649884f58f1ba2e40a24c475b82357c91cbcd8c782bddc4d02fd3ed9ddab5"},{"id":"func/Game.toggleParameter","name":"Game.toggleParameter","line":269,"end_line":271,"hash":"2dc62c76aa4c67e5fd8772727bb04fa956d3affdf0e00c7fc7098ae7345abd49"},{"id":"func/Game.stepParameter","name":"Game.stepParameter","line":273,"end_line":275,"hash":"307945d810c511e42519b1be4fa4388065546f9b76d9ac0a3fe63d513b522856"},{"id":"func/Game.selectedMassIDs","name":"Game.selectedMassIDs","line":277,"end_line":285,"hash":"c22bc0b7c51e606264c25a2b1f13d5770a815f27bbf9edb542a1a7a95215ccd2"},{"id":"func/Game.parameterFloat","name":"Game.parameterFloat","line":287,"end_line":290,"hash":"b1b75434b4e1913a8f95165a651d741172c5d48351ae4add5ccd7a12d7bfc664"},{"id":"func/Game.gridSnapSize","name":"Game.gridSnapSize","line":292,"end_line":294,"hash":"2f848d29a5ed78e7601534d822f5588903cc01e5006b92b46e55fcfbf479b594"},{"id":"func/formatControlFloat","name":"formatControlFloat","line":296,"end_line":298,"hash":"b19b7aacb786da1791f10cb92f30c3b9902c9804613d34efcd6f7da207db4a9d"},{"id":"func/roundControlFloat","name":"roundControlFloat","line":300,"end_line":302,"hash":"5c3c02047df5405fa39aa22ef6c906165e410f44d101726a323894fcf3c84a75"},{"id":"func/clampFloat","name":"clampFloat","line":304,"end_line":306,"hash":"ab5544e501ecda24c07fea5d3a7db0d688bd2961d0e4deaae4b8a8e6eb5f6c52"},{"id":"func/Game.visibleControlAt","name":"Game.visibleControlAt","line":308,"end_line":310,"hash":"2aa00df037c964ad3a9740e2182052329d9215fb39364f48d4357a3e745be575"},{"id":"func/controlAt","name":"controlAt","line":312,"end_line":319,"hash":"96124e2df5fa8bb2750eec7dd57ed14570dad4deff46d2974dba923b9cab3c35"},{"id":"func/Game.visibleControlWithLabel","name":"Game.visibleControlWithLabel","line":321,"end_line":323,"hash":"7a8c9bd653f31b2fa6a9594623f8ba8ce2c9998214016bb08ba2f9718dbb295c"},{"id":"func/Game.visibleControlWithName","name":"Game.visibleControlWithName","line":325,"end_line":327,"hash":"2fab6527d49165fa5c4f391872fe5ff7d16ad3543580b41ddd17c60203e6c8b2"},{"id":"func/Game.visibleControlWithField","name":"Game.visibleControlWithField","line":329,"end_line":336,"hash":"9c365a137d56e41084c8dbf56f0c3d25061ce8ccbd6d0abd8e0feddc3bc889c3"},{"id":"func/visibleControlAt","name":"visibleControlAt","line":338,"end_line":345,"hash":"f049ae667df6f18d325f52daa1c49dea3a0f91b5bdd6e131fc4453478fe3f119"},{"id":"func/visibleControlWithLabel","name":"visibleControlWithLabel","line":347,"end_line":349,"hash":"12466197c75f303e2046e9040707712b50105ab0775ae48effa98c255c8b6473"},{"id":"func/visibleControlWithName","name":"visibleControlWithName","line":351,"end_line":353,"hash":"bf1da831cc99be8af7c973d3ae709681d724a3f1190917a8039604bebd1e6bba"},{"id":"func/visibleControlWithField","name":"visibleControlWithField","line":355,"end_line":362,"hash":"a86699b1155622f62d2bedac4f62b1ab2b47837fd617eaa135881750273cef08"},{"id":"func/Game.PathEntryCommand","name":"Game.PathEntryCommand","line":364,"end_line":366,"hash":"90b70f23deb7f8da1cece97e4fa6d3bdd382724bcd575e352633d8ce5122ab91"},{"id":"func/Game.DemoPickerOpen","name":"Game.DemoPickerOpen","line":368,"end_line":370,"hash":"5c9273633ac604733fe616cfa7f57e3e41dfaf021cdb9f7207f6fae944b7de68"},{"id":"func/Game.VisibleControlActive","name":"Game.VisibleControlActive","line":372,"end_line":375,"hash":"ea18610628e27129a6a6f1fbf75e0d23fd7d97f9f7675fb1c971e292c59d5445"},{"id":"func/Game.DragMass","name":"Game.DragMass","line":377,"end_line":382,"hash":"54e365a24dad0e7d7d748b16f5c3583af78add366f83ad03f81db661f1c9355f"},{"id":"func/Game.dragSelectedMasses","name":"Game.dragSelectedMasses","line":384,"end_line":392,"hash":"24b4d06566a81bd5006ae43ea6808d1170f96df1c322fb777a5ef18e7da6df32"},{"id":"func/Game.dragSingleMass","name":"Game.dragSingleMass","line":394,"end_line":404,"hash":"b2c68d8d37dfaac3417d2500bc988b5aed989998d397c1e167efc2074fefff35"},{"id":"func/Game.finishMassDragStep","name":"Game.finishMassDragStep","line":406,"end_line":412,"hash":"d26aca679650bf34bb7b46804a0531c43a27318d1eeeb8b2e07ac0ac2c98f79a"},{"id":"func/Game.moveSelectedMasses","name":"Game.moveSelectedMasses","line":414,"end_line":421,"hash":"f6ba940b1ed1114b687b0b84b3482bffa82119ef0a1dcc0418ac14bc9ba60d77"},{"id":"func/Game.applyDraggingOffsets","name":"Game.applyDraggingOffsets","line":423,"end_line":431,"hash":"0e6bb5212ea85b8de36949f8c249d170bb269d917f0bedeb497713ab275edece"},{"id":"func/Game.snapToGrid","name":"Game.snapToGrid","line":433,"end_line":442,"hash":"e7b93e39ef986bc33f6bedbab4eb997d9433a2e7267e430186f91dffd802870a"}]}
// mutate4go-manifest-end
