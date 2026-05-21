package acceptance

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"springs/internal/app"
)

func TestSaveLoadDialogStepsSaveNamedWorld(t *testing.T) {
	w := &world{}
	t.Cleanup(func() { cleanupSaveLoadWorkDir(t, w) })

	example := map[string]string{
		"world_state":     "simple masses",
		"field_text":      ".xsp",
		"cursor_position": "before xsp extension",
		"filename_prefix": "lab_scene",
		"saved_path":      filepath.Join("saves", "lab_scene.xsp"),
	}

	steps := []struct {
		name string
		run  func(*world, map[string]string) error
	}{
		{"create world", createCurrentWorldWithState},
		{"dialog open", assertSaveFilenameDialogOpen},
		{"field text", assertSaveFilenameFieldText},
		{"cursor position", assertSaveFilenameCursorPosition},
		{"enter filename", enterSaveFilenamePrefix},
		{"submit", submitSaveFilenameDialog},
		{"saved file exists", assertSavedXSPFileExists},
		{"saved file has content", assertSavedXSPFileContainsState},
	}
	for _, step := range steps {
		if err := step.run(w, example); err != nil {
			t.Fatalf("%s step returned error: %v", step.name, err)
		}
	}
}

func TestSaveLoadDialogStepsLoadSavedWorldAndOrderPicker(t *testing.T) {
	w := &world{}
	if err := ensureSaveLoadWorkDir(w); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { cleanupSaveLoadWorkDir(t, w) })

	example := map[string]string{
		"save_file":     "lab_scene.xsp",
		"demo_file":     "pendulum.xsp",
		"original_file": "pend.xsp",
		"separator":     "separator",
		"world_state":   "simple masses",
		"saved_path":    filepath.Join("saves", "lab_scene.xsp"),
	}

	setupSteps := []struct {
		name string
		run  func(*world, map[string]string) error
	}{
		{"plain saved file", createSavedXSPFile},
		{"saved file", createSavedXSPFileWithState},
		{"demo file", createDemoXSPFile},
		{"original file", createOriginalXSPFile},
	}
	for _, step := range setupSteps {
		if err := step.run(w, example); err != nil {
			t.Fatalf("%s step returned error: %v", step.name, err)
		}
	}

	w.appGame = app.NewGame()
	assertionSteps := []struct {
		name string
		run  func(*world, map[string]string) error
	}{
		{"save before separator", assertSaveFileBeforeSeparator},
		{"separator before demo", assertSeparatorBeforeDemoFile},
		{"demo before original", assertDemoFileBeforeOriginalFile},
		{"choose saved file", chooseLoadPickerEntry},
		{"loaded world has content", assertLoadedWorldIncludesState},
		{"current path tracks load", assertCurrentFilePath},
		{"missing order reports err", expectMissingLoadPickerOrder},
		{"unsupported cursor reports err", expectUnsupportedCursorPosition},
	}
	for _, step := range assertionSteps {
		if err := step.run(w, example); err != nil {
			t.Fatalf("%s step returned error: %v", step.name, err)
		}
	}
}

func expectMissingLoadPickerOrder(w *world, example map[string]string) error {
	missing := map[string]string{
		"save_file": "missing.xsp",
		"separator": example["separator"],
	}
	if err := assertSaveFileBeforeSeparator(w, missing); err == nil {
		return errors.New("missing load picker entry should fail")
	}
	return nil
}

func expectUnsupportedCursorPosition(w *world, example map[string]string) error {
	unsupported := map[string]string{
		"cursor_position": "after filename",
	}
	if err := assertSaveFilenameCursorPosition(w, unsupported); err == nil {
		return errors.New("unsupported cursor position should fail")
	}
	return nil
}

func cleanupSaveLoadWorkDir(t *testing.T, w *world) {
	t.Helper()
	if w.previousWorkDir != "" {
		if err := os.Chdir(w.previousWorkDir); err != nil {
			t.Fatalf("restore working directory: %v", err)
		}
	}
	if w.workDir != "" {
		if err := os.RemoveAll(w.workDir); err != nil {
			t.Fatalf("remove work dir: %v", err)
		}
	}
}
