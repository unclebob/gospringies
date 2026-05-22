package acceptance

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"springs/internal/sim"
)

func ensureSaveLoadWorkDir(w *world) error {
	if w.workDir != "" {
		return nil
	}
	previous, err := os.Getwd()
	if err != nil {
		return err
	}
	dir, err := os.MkdirTemp("", "springs-save-load-*")
	if err != nil {
		return err
	}
	if err := os.Chdir(dir); err != nil {
		_ = os.RemoveAll(dir)
		return err
	}
	w.previousWorkDir = previous
	w.workDir = dir
	return nil
}

func assertSaveFilenameDialogOpen(w *world, _ map[string]string) error {
	game, err := visibleControlsGame(w)
	if err != nil {
		return err
	}
	if !game.SaveFilenameDialogOpen() {
		return fmt.Errorf("save filename dialog was not open")
	}
	return nil
}

func assertSaveFilenameFieldText(w *world, example map[string]string) error {
	expected, err := stringValue(example, "field_text")
	if err != nil {
		return err
	}
	game, err := visibleControlsGame(w)
	if err != nil {
		return err
	}
	if got := game.SaveFilenameText(); got != expected {
		return fmt.Errorf("save filename field = %q, want %q", got, expected)
	}
	return nil
}

func assertSaveFilenameCursorPosition(w *world, example map[string]string) error {
	position, err := stringValue(example, "cursor_position")
	if err != nil {
		return err
	}
	if position != "before xsp extension" {
		return fmt.Errorf("unsupported cursor position %q", position)
	}
	game, err := visibleControlsGame(w)
	if err != nil {
		return err
	}
	if game.SaveFilenameCursor() != 0 {
		return fmt.Errorf("save filename cursor = %d, want before extension", game.SaveFilenameCursor())
	}
	return nil
}

func createCurrentWorldWithState(w *world, example map[string]string) error {
	if err := ensureSaveLoadWorkDir(w); err != nil {
		return err
	}
	state, err := stringValue(example, "world_state")
	if err != nil {
		return err
	}
	world, err := worldForSaveLoadState(state)
	if err != nil {
		return err
	}
	game := newApplicationDriverGame()
	game.ReplaceWorld(world)
	game.RunCommand("save")
	w.appGame = game
	return nil
}

func enterSaveFilenamePrefix(w *world, example map[string]string) error {
	prefix, err := stringValue(example, "filename_prefix")
	if err != nil {
		return err
	}
	game, err := visibleControlsGame(w)
	if err != nil {
		return err
	}
	game.EnterSaveFilenamePrefix(prefix)
	return nil
}

func submitSaveFilenameDialog(w *world, _ map[string]string) error {
	game, err := visibleControlsGame(w)
	if err != nil {
		return err
	}
	return game.SubmitSaveFilenameDialog()
}

func assertSavedXSPFileExists(w *world, example map[string]string) error {
	path, err := stringValue(example, "saved_path")
	if err != nil {
		return err
	}
	if _, err := os.Stat(path); err != nil {
		return err
	}
	return nil
}

func assertSavedXSPFileContainsState(w *world, example map[string]string) error {
	path, state, err := stringPair(example, "saved_path", "world_state")
	if err != nil {
		return err
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return assertSaveLoadStateContent(string(content), state)
}

func createSavedXSPFile(w *world, example map[string]string) error {
	return createSavedXSPFileFromKey(w, example, "save_file")
}

func createOldSavedXSPFile(w *world, example map[string]string) error {
	return createSavedXSPFileFromKey(w, example, "old_save_file")
}

func createSavedXSPFileFromKey(w *world, example map[string]string, key string) error {
	if err := ensureSaveLoadWorkDir(w); err != nil {
		return err
	}
	name, err := stringValue(example, key)
	if err != nil {
		return err
	}
	return writeSaveLoadXSP(filepath.Join("saves", name), simpleSceneXSP())
}

func createDemoXSPFile(w *world, example map[string]string) error {
	if err := ensureSaveLoadWorkDir(w); err != nil {
		return err
	}
	name, err := stringValue(example, "demo_file")
	if err != nil {
		return err
	}
	return writeSaveLoadXSP(filepath.Join("demos", name), "#1.0\nmass 3 10 20 1 0\n")
}

func createOriginalXSPFile(w *world, example map[string]string) error {
	if err := ensureSaveLoadWorkDir(w); err != nil {
		return err
	}
	name, err := stringValue(example, "original_file")
	if err != nil {
		return err
	}
	return writeSaveLoadXSP(filepath.Join("demos", "original", name), "#1.0\nmass 4 10 20 1 0\n")
}

func createSavedXSPFileWithState(w *world, example map[string]string) error {
	return createSavedXSPFileWithStateFromKey(w, example, "save_file")
}

func createNewSavedXSPFileWithState(w *world, example map[string]string) error {
	return createSavedXSPFileWithStateFromKey(w, example, "new_save_file")
}

func createSavedXSPFileWithStateFromKey(w *world, example map[string]string, key string) error {
	if err := ensureSaveLoadWorkDir(w); err != nil {
		return err
	}
	name, state, err := stringPair(example, key, "world_state")
	if err != nil {
		return err
	}
	world, err := worldForSaveLoadState(state)
	if err != nil {
		return err
	}
	return writeSaveLoadXSP(filepath.Join("saves", name), xspForSaveLoadWorld(world))
}

func openLoadPicker(w *world, _ map[string]string) error {
	if err := ensureSaveLoadWorkDir(w); err != nil {
		return err
	}
	return clickInsideRenderedBoundsOfControl(w, "Load")
}

func chooseLoadPickerEntry(w *world, example map[string]string) error {
	return chooseLoadPickerEntryFromKey(w, example, "save_file")
}

func chooseNewLoadPickerEntry(w *world, example map[string]string) error {
	return chooseLoadPickerEntryFromKey(w, example, "new_save_file")
}

func chooseLoadPickerEntryFromKey(w *world, example map[string]string, key string) error {
	name, err := stringValue(example, key)
	if err != nil {
		return err
	}
	game, err := visibleControlsGame(w)
	if err != nil {
		return err
	}
	if !game.ChooseLoadPickerEntry(name) {
		return fmt.Errorf("load picker entry %q was not loaded from %#v: %s", name, game.LoadPickerEntries(), game.LastFileError())
	}
	return nil
}

func assertLoadedWorldIncludesState(w *world, example map[string]string) error {
	state, err := stringValue(example, "world_state")
	if err != nil {
		return err
	}
	game, err := visibleControlsGame(w)
	if err != nil {
		return err
	}
	return assertSaveLoadWorldState(game.World(), state)
}

func assertCurrentFilePath(w *world, example map[string]string) error {
	expected, err := stringValue(example, "saved_path")
	if err != nil {
		return err
	}
	game, err := visibleControlsGame(w)
	if err != nil {
		return err
	}
	if got := game.CurrentFilePath(); got != filepath.Clean(expected) {
		return fmt.Errorf("current file path = %q, want %q", got, expected)
	}
	return nil
}

func assertLoadPickerEntryBefore(w *world, example map[string]string, firstKey string, secondKey string) error {
	first, second, err := stringPair(example, firstKey, secondKey)
	if err != nil {
		return err
	}
	game, err := visibleControlsGame(w)
	if err != nil {
		return err
	}
	entries := game.LoadPickerEntries()
	return requireLoadPickerEntryOrder(entries, first, second)
}

func requireLoadPickerEntryOrder(entries []string, first string, second string) error {
	firstIndex := loadPickerEntryIndex(entries, first)
	secondIndex := loadPickerEntryIndex(entries, second)
	if firstIndex < 0 {
		return fmt.Errorf("load picker entry %q not found in %#v", first, entries)
	}
	if secondIndex < 0 {
		return fmt.Errorf("load picker entry %q not found in %#v", second, entries)
	}
	if firstIndex >= secondIndex {
		return fmt.Errorf("load picker order %q before %q not found in %#v", first, second, entries)
	}
	return nil
}

func assertSaveFileBeforeSeparator(w *world, example map[string]string) error {
	return assertLoadPickerEntryBefore(w, example, "save_file", "separator")
}

func assertNewSaveFileBeforeSeparator(w *world, example map[string]string) error {
	return assertLoadPickerEntryBefore(w, example, "new_save_file", "separator")
}

func assertSeparatorBeforeDemoFile(w *world, example map[string]string) error {
	return assertLoadPickerEntryBefore(w, example, "separator", "demo_file")
}

func assertDemoFileBeforeOriginalFile(w *world, example map[string]string) error {
	return assertLoadPickerEntryBefore(w, example, "demo_file", "original_file")
}

func loadPickerEntryIndex(entries []string, name string) int {
	for i, entry := range entries {
		if entry == name || filepath.Base(entry) == name {
			return i
		}
	}
	return -1
}

func writeSaveLoadXSP(path string, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o600)
}

func worldForSaveLoadState(state string) (*sim.Simulation, error) {
	if state != "simple masses" {
		return nil, fmt.Errorf("unsupported world state %q", state)
	}
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 0, Y: 0}, Mass: 1})
	_ = world.AddMass(sim.Mass{ID: 2, Position: sim.Vec2{X: 10, Y: 0}, Mass: 1})
	return world, nil
}

func xspForSaveLoadWorld(world *sim.Simulation) string {
	var builder strings.Builder
	builder.WriteString("#1.0\n")
	for _, mass := range world.Masses {
		builder.WriteString(fmt.Sprintf("mass %d %.0f %.0f %.0f 0\n", mass.ID, mass.Position.X, mass.Position.Y, mass.Mass))
	}
	return builder.String()
}

func assertSaveLoadStateContent(content string, state string) error {
	if state != "simple masses" {
		return fmt.Errorf("unsupported world state %q", state)
	}
	if !strings.Contains(content, "\nmass 1 ") || !strings.Contains(content, "\nmass 2 ") {
		return fmt.Errorf("saved content did not include simple masses:\n%s", content)
	}
	return nil
}

func assertSaveLoadWorldState(world *sim.Simulation, state string) error {
	if state != "simple masses" {
		return fmt.Errorf("unsupported world state %q", state)
	}
	if _, ok := world.MassByID(1); !ok {
		return fmt.Errorf("loaded world missing mass 1: %#v", world.Masses)
	}
	if _, ok := world.MassByID(2); !ok {
		return fmt.Errorf("loaded world missing mass 2: %#v", world.Masses)
	}
	return nil
}
