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
	g.lastCursor = g.screenToWorld(simVec(x, y))
	if control, ok := g.editMenuControlAt(image.Pt(x, y)); ok {
		return g.activateVisibleControl(control)
	}
	control, ok := visibleControlAt(image.Pt(x, y))
	if !ok {
		if g.editMenuOpen {
			g.editMenuOpen = false
			return true
		}
		return false
	}
	if isSliderControl(control.Name) {
		g.activeSlider = control.Name
		g.setSliderAt(control.Name, x)
		return true
	}
	return g.activateVisibleControl(control)
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
	control, ok := visibleControlWithName(name)
	if !ok {
		return
	}
	track := sliderTrack(control)
	fraction := 0.0
	if track.Dx() > 0 {
		fraction = clampFloat(float64(x-track.Min.X)/float64(track.Dx()), 0, 1)
	}
	switch name {
	case "gravity slider":
		g.setForceValue("gravity", "magnitude", fraction*50)
	case "speed slider":
		g.simulationSpeed = fraction * maxSpeed
		return
	case "viscosity slider":
		g.simulation.Parameters.Set("viscosity", formatControlFloat(fraction*2))
	}
	g.dirty = true
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
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
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
	if len(g.draggingOffsets) > 0 {
		g.applyDraggingOffsets(position)
	} else {
		g.moveSelectedMasses(position.Sub(g.draggingLast))
	}
	g.finishMassDragStep(position)
	return true
}

func (g *Game) dragSingleMass(id int, position sim.Vec2) bool {
	for i := range g.simulation.Masses {
		if g.simulation.Masses[i].ID == id {
			g.simulation.Masses[i].Position = g.snapToGrid(position)
			g.simulation.Masses[i].Velocity = sim.Vec2{}
			g.finishMassDragStep(position)
			return true
		}
	}
	return false
}

func (g *Game) finishMassDragStep(position sim.Vec2) {
	g.draggingLast = position
	if !selectionClick(g.draggingStart, position) {
		g.dragMoved = true
	}
	g.dirty = true
}

func (g *Game) moveSelectedMasses(delta sim.Vec2) {
	for i := range g.simulation.Masses {
		if g.editing().MassSelected(g.simulation.Masses[i].ID) {
			g.simulation.Masses[i].Position = g.simulation.Masses[i].Position.Add(delta)
			g.simulation.Masses[i].Velocity = sim.Vec2{}
		}
	}
}

func (g *Game) applyDraggingOffsets(cursor sim.Vec2) {
	for i := range g.simulation.Masses {
		offset, ok := g.draggingOffsets[g.simulation.Masses[i].ID]
		if ok {
			g.simulation.Masses[i].Position = g.snapToGrid(cursor.Add(offset))
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
