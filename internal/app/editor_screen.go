package app

var editorRegions = []ScreenRegion{
	{"canvas", "edit and view the simulation world"},
	{"left toolbar", "run selection commands"},
	{"top bar", "run commands and file commands"},
	{"right inspector", "edit selected objects and world parameters"},
	{"status line", "show simulation state, counts, and file state"},
}

var editorCommandControls = []string{"run", "pause", "reset", "load", "insert", "save", "quit", "delete", "select all", "cut", "copy", "paste"}

var shortcutCommands = map[string]string{
	"Space":  "pause",
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

// mutate4go-manifest-begin
// {"version":1,"tested_at":"2026-05-19T12:04:31-05:00","module_hash":"69fc194e063da3ebc87a8b63d0bef989319ca79358be49e7d3b0ae3cfca51304","functions":[{"id":"func/Game.EditorScreen","name":"Game.EditorScreen","line":44,"end_line":58,"hash":"cb51d64ea30f16d1d7bb996487ba925fce6333e8b9bf6fe0c93fb43ecdb812d1"},{"id":"func/EditorScreen.RegionPurpose","name":"EditorScreen.RegionPurpose","line":60,"end_line":67,"hash":"816db9dc4c92e4e8586c2b6f13a8625624401fa62426fccf75b9137d01b5de95"},{"id":"func/EditorScreen.HasCommandControl","name":"EditorScreen.HasCommandControl","line":69,"end_line":71,"hash":"23b17abfa20aaf956e2941ac29a507e665f35d927a6d0351452faa2c5b987654"},{"id":"func/Game.SetSelected","name":"Game.SetSelected","line":73,"end_line":75,"hash":"625f8043db1dd482f56861abc98264348d94e820995f88ea2c7a7f41ac8e0f5c"},{"id":"func/Game.SetDirty","name":"Game.SetDirty","line":77,"end_line":79,"hash":"ebd484e63b9f7c36cc5c1d7a91261c2846ca62ca84f2f10f7a8f6a32d17c456b"},{"id":"func/Game.HandleShortcut","name":"Game.HandleShortcut","line":81,"end_line":88,"hash":"3d4844c6b08736d3dd7b7023ab012ac8283ae1bdc4e3e14b0ea2e63b8071bbf4"},{"id":"func/Game.LastCommand","name":"Game.LastCommand","line":90,"end_line":92,"hash":"3814381553baa93a33b597ce2804b2cda942d6d2d8f2281cbd71d15a06a87728"},{"id":"func/Game.simulationState","name":"Game.simulationState","line":94,"end_line":96,"hash":"cac39cf93aa6bd3b060df38b78300dc023b3017337f907bfbdfaca196f9fe664"},{"id":"func/Game.selectionState","name":"Game.selectionState","line":98,"end_line":100,"hash":"da852f34cf6e762cb63a5dfda2f3630741e6987eb91e067e44a5b68f20d0bf0a"},{"id":"func/Game.fileState","name":"Game.fileState","line":102,"end_line":104,"hash":"58912ba0ba7137c3fdd9f91d5e41459c6e0fe4b342fdceacadaa399647a60818"},{"id":"func/stateLabel","name":"stateLabel","line":106,"end_line":111,"hash":"8102475472161409422aafadaab16a68d061937c6be751de49729064a2f33497"},{"id":"func/contains","name":"contains","line":113,"end_line":120,"hash":"8d1951ea6ecaaeb43f50cc1cfaead778b9a46e20a7d635966c58f41777191329"}]}
// mutate4go-manifest-end
