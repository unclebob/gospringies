package app

import (
	"image"
	"math"
	"strconv"

	"springs/internal/sim"
)

var modeControlModes = map[string]string{
	"select mode": "select",
	"mass mode":   "add mass",
	"spring mode": "add spring",
	"drag mode":   "drag",
}

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
}

// ClickAt handles pointer activation for the same rectangles Draw uses for controls.
func (g *Game) ClickAt(x int, y int) bool {
	control, ok := visibleControlAt(image.Pt(x, y))
	if !ok {
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
	if mode, ok := modeControlModes[control.Name]; ok {
		g.SetMode(mode)
		return true
	}
	if command, ok := visibleControlCommands[control.Name]; ok {
		g.RunCommand(command)
		return true
	}
	if g.activateInspectorControl(control.Name) {
		return true
	}
	return false
}

func (g *Game) activateInspectorControl(name string) bool {
	switch name {
	case "mass parameter":
		g.stepEditorControl("mass", 1)
	case "elasticity parameter":
		g.stepEditorControl("elasticity", 0.1)
	case "fixed mass toggle":
		g.toggleFixedMass()
	case "kspring parameter":
		g.stepEditorControl("Kspring", 1)
	case "kdamp parameter":
		g.stepEditorControl("Kdamp", 0.1)
	case "set rest length command":
		_ = g.editing().SetRestLength()
	case "gravity force":
		g.toggleForce("gravity", map[string]string{"magnitude": "10", "direction": "0"})
	case "center attraction force":
		g.toggleForce("center attraction", map[string]string{"magnitude": "10", "exponent": "0"})
	case "center mass force":
		g.toggleForce("center of mass attraction", map[string]string{"magnitude": "5", "damping": "2"})
	case "wall repulsion force":
		g.toggleForce("wall repulsion", map[string]string{"magnitude": "10000", "exponent": "1"})
	case "set center command":
		g.simulation.SetForceCenter(g.selectedMassIDs())
	case "top wall toggle":
		g.toggleWall("top")
	case "bottom wall toggle":
		g.toggleWall("bottom")
	case "left wall toggle":
		g.toggleWall("left")
	case "right wall toggle":
		g.toggleWall("right")
	case "grid snap toggle":
		g.toggleGridSnap()
	case "show springs toggle":
		g.toggleParameter("show springs")
	case "viscosity parameter":
		g.stepParameter("viscosity", 0.1)
	case "stickiness parameter":
		g.stepParameter("stickiness", 0.1)
	case "timestep parameter":
		g.stepParameter("timestep", 0.001)
	case "precision parameter":
		g.stepParameter("precision", 0.001)
	case "adaptive timestep toggle":
		g.toggleParameter("adaptive timestep")
	default:
		return false
	}
	g.dirty = true
	return true
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
	switch control {
	case "mass":
		return g.parameterFloat("current mass")
	case "elasticity":
		return g.parameterFloat("elasticity")
	case "Kspring":
		return g.parameterFloat("spring constant")
	case "Kdamp":
		return g.parameterFloat("damping")
	default:
		return 0
	}
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

func (g *Game) Mode() string {
	return g.mode
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
	if g.mode != "drag" {
		return false
	}
	for i := range g.simulation.Masses {
		if g.simulation.Masses[i].ID == id {
			if !g.simulation.Masses[i].Fixed {
				g.simulation.Masses[i].Position = position
				g.dirty = true
			}
			return true
		}
	}
	return false
}
