package app

import "fmt"

var editorRegions = []ScreenRegion{
	{"canvas", "edit and view the simulation world"},
	{"left toolbar", "run selection commands"},
	{"top bar", "run commands and file commands"},
	{"right inspector", "edit selected objects and world parameters and show simulation state"},
}

var editorCommandControls = []string{"pause toggle", "reset", "load", "insert", "save", "quit", "delete", "select all", "cut", "copy", "paste"}

var shortcutCommands = map[string]string{
	"Space":  "pause toggle",
	"Delete": "delete",
	"Ctrl+A": "select all",
	"Ctrl+X": "cut",
	"Ctrl+C": "copy",
	"Ctrl+V": "paste",
	"Ctrl+D": "duplicate",
	"R":      "reset",
	"Ctrl+S": "save",
	"Ctrl+O": "load",
	"Ctrl+I": "insert",
	"Q":      "quit",
	"Esc":    "clear selection",
}

type EditorScreen struct {
	Editor          bool
	LandingPage     bool
	Regions         []ScreenRegion
	CommandControls []string
	Indicators      map[string]string
	CanvasVisible   bool
	ControlsUsable  bool
}

type ScreenRegion struct {
	Name    string
	Purpose string
}

func (g *Game) EditorScreen() EditorScreen {
	return EditorScreen{
		Editor:          true,
		LandingPage:     false,
		Regions:         editorRegions,
		CommandControls: editorCommandControls,
		Indicators: map[string]string{
			"simulation state": g.simulationState(),
			"selection":        g.selectionState(),
			"file state":       g.fileState(),
		},
		CanvasVisible:  true,
		ControlsUsable: true,
	}
}

func (s EditorScreen) RegionPurpose(name string) (string, bool) {
	for _, region := range s.Regions {
		if region.Name == name {
			return region.Purpose, true
		}
	}
	return "", false
}

func (s EditorScreen) HasCommandControl(command string) bool {
	return contains(s.CommandControls, command)
}

func (g *Game) SetSelected(selected bool) {
	g.setSelected(selected)
}

func (g *Game) SelectSpring(id int) error {
	return g.editing().SelectSpring(id)
}

func (g *Game) SelectSprings(ids ...int) error {
	editor := g.editing()
	editor.ClearSelection()
	for _, id := range ids {
		if _, ok := g.world.simulation.SpringByID(id); !ok {
			return fmt.Errorf("spring %d not found", id)
		}
		editor.SelectedSprings[id] = true
	}
	return nil
}

func (g *Game) SetDirty(dirty bool) {
	if dirty {
		g.markDirty()
		return
	}
	g.clearDirty()
}

func (g *Game) HandleShortcut(shortcut string) bool {
	if shortcut == "Esc" && g.pointer.pendingSpringID != 0 {
		g.clearPendingSpring()
		return true
	}
	command, ok := shortcutCommands[shortcut]
	if !ok {
		return false
	}
	g.RunCommand(command)
	return true
}

func (g *Game) LastCommand() string {
	return g.editState.lastCommand
}

func (g *Game) simulationState() string {
	return stateLabel(g.run.paused, "paused", "running")
}

func (g *Game) selectionState() string {
	return stateLabel(g.editState.selected, "object selected", "nothing selected")
}

func (g *Game) fileState() string {
	return stateLabel(g.editState.dirty, "unsaved changes", "saved")
}

func stateLabel(condition bool, trueLabel string, falseLabel string) string {
	if condition {
		return trueLabel
	}
	return falseLabel
}

func contains(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}

// mutate4go-manifest-begin
// {"version":1,"tested_at":"2026-05-22T08:52:06-05:00","module_hash":"d914d0b654711476d4bf89502583d23f4471a9117ae666062c8cd7576a1bd5c5","functions":[{"id":"func/Game.EditorScreen","name":"Game.EditorScreen","line":45,"end_line":59,"hash":"cb51d64ea30f16d1d7bb996487ba925fce6333e8b9bf6fe0c93fb43ecdb812d1"},{"id":"func/EditorScreen.RegionPurpose","name":"EditorScreen.RegionPurpose","line":61,"end_line":68,"hash":"816db9dc4c92e4e8586c2b6f13a8625624401fa62426fccf75b9137d01b5de95"},{"id":"func/EditorScreen.HasCommandControl","name":"EditorScreen.HasCommandControl","line":70,"end_line":72,"hash":"23b17abfa20aaf956e2941ac29a507e665f35d927a6d0351452faa2c5b987654"},{"id":"func/Game.SetSelected","name":"Game.SetSelected","line":74,"end_line":76,"hash":"39313dc8acee90b71536c47efc5991c0f30f1f7937a39e35401c8d039da35ea7"},{"id":"func/Game.SelectSpring","name":"Game.SelectSpring","line":78,"end_line":80,"hash":"0210e030e94bd341a8b80dfaff397945e15d014ed5d172ee536233ccf2130a8a"},{"id":"func/Game.SelectSprings","name":"Game.SelectSprings","line":82,"end_line":92,"hash":"bd69aeb47d43797d0b145d40ac73eb061677b628191957a8b9108285bebb88be"},{"id":"func/Game.SetDirty","name":"Game.SetDirty","line":94,"end_line":100,"hash":"cffc93c07a9ae5575aa4bf483996a296b89a91e2a6bd04ca8ec900f13299ca3c"},{"id":"func/Game.HandleShortcut","name":"Game.HandleShortcut","line":102,"end_line":113,"hash":"b74bedc0f2e7545fd797ce9758f8a103de3af79a37de2043eab8e16aeca9a90d"},{"id":"func/Game.LastCommand","name":"Game.LastCommand","line":115,"end_line":117,"hash":"d17f34806c8d3f2a266197710c1fa9e6083e061f95294589ff9bf7750946d7ca"},{"id":"func/Game.simulationState","name":"Game.simulationState","line":119,"end_line":121,"hash":"d8731f5b24ba6cb15ad05db2441fee8b90ce2d659882cfdccb01340998cdccea"},{"id":"func/Game.selectionState","name":"Game.selectionState","line":123,"end_line":125,"hash":"5f03f6ac45016b672402ed490e44b6d2edd5dc9f76fc169f2b2183ec2cb197c6"},{"id":"func/Game.fileState","name":"Game.fileState","line":127,"end_line":129,"hash":"b5587de68ad529f8cc6f164ac820ab0dd564f6455e6739a1c832b350e9367999"},{"id":"func/stateLabel","name":"stateLabel","line":131,"end_line":136,"hash":"8102475472161409422aafadaab16a68d061937c6be751de49729064a2f33497"},{"id":"func/contains","name":"contains","line":138,"end_line":145,"hash":"8d1951ea6ecaaeb43f50cc1cfaead778b9a46e20a7d635966c58f41777191329"}]}
// mutate4go-manifest-end
