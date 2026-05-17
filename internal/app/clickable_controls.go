package app

import (
	"image"

	"springs/internal/sim"
)

var modeControlModes = map[string]string{
	"select mode": "select",
	"mass mode":   "add mass",
	"spring mode": "add spring",
	"drag mode":   "drag",
}

var visibleControlCommands = map[string]string{
	"run command":    "run",
	"pause command":  "pause",
	"reset command":  "reset",
	"load command":   "load",
	"insert command": "insert",
	"save command":   "save",
	"quit command":   "quit",
}

// ClickAt handles pointer activation for the same rectangles Draw uses for controls.
func (g *Game) ClickAt(x int, y int) bool {
	control, ok := visibleControlAt(image.Pt(x, y))
	if !ok {
		return false
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
	return false
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

func (g *Game) Mode() string {
	return g.mode
}

func (g *Game) PathEntryCommand() string {
	return g.pathEntryCommand
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
