package app

var editorRegions = []ScreenRegion{
	{"canvas", "edit and view the simulation world"},
	{"left toolbar", "choose editing modes"},
	{"top bar", "run commands and file commands"},
	{"right inspector", "edit selected objects and world parameters"},
	{"status line", "show mode, simulation state, counts, and file state"},
}

var editorModeControls = []string{"select", "add mass", "add spring", "drag"}

var editorCommandControls = []string{"run", "pause", "reset", "load", "insert", "save", "quit", "delete", "select all"}

var shortcutCommands = map[string]string{
	"Space":  "pause",
	"Delete": "delete",
	"Ctrl+A": "select all",
	"R":      "reset",
	"Ctrl+S": "save",
	"Ctrl+O": "load",
	"Ctrl+I": "insert",
	"Q":      "quit",
}

type EditorScreen struct {
	Editor          bool
	LandingPage     bool
	Regions         []ScreenRegion
	ModeControls    []string
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
		ModeControls:    editorModeControls,
		CommandControls: editorCommandControls,
		Indicators: map[string]string{
			"active mode":      g.mode + " mode",
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

func (s EditorScreen) HasModeControl(mode string) bool {
	return contains(s.ModeControls, mode)
}

func (s EditorScreen) HasCommandControl(command string) bool {
	return contains(s.CommandControls, command)
}

func (g *Game) SetMode(mode string) {
	g.mode = mode
}

func (g *Game) SetSelected(selected bool) {
	g.selected = selected
}

func (g *Game) SetDirty(dirty bool) {
	g.dirty = dirty
}

func (g *Game) HandleShortcut(shortcut string) bool {
	command, ok := shortcutCommands[shortcut]
	if !ok {
		return false
	}
	g.RunCommand(command)
	return true
}

func (g *Game) LastCommand() string {
	return g.lastCommand
}

func (g *Game) simulationState() string {
	return stateLabel(g.paused, "paused", "running")
}

func (g *Game) selectionState() string {
	return stateLabel(g.selected, "object selected", "nothing selected")
}

func (g *Game) fileState() string {
	return stateLabel(g.dirty, "unsaved changes", "saved")
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
