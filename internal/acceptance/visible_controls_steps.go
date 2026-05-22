package acceptance

import (
	"fmt"
	"strings"
)

var visibleControlStateSetters = map[string]func(*driverGame){
	"running":       func(game *driverGame) { game.SetPaused(false) },
	"object counts": func(*driverGame) {},
	"saved":         func(game *driverGame) { game.SetDirty(false) },
}

func drawApplicationFrame(w *world, _ map[string]string) error {
	game, err := visibleControlsGame(w)
	if err != nil {
		return err
	}
	w.visibleControlsFrame = game.DrawFrameReport()
	return nil
}

func assertFrameRegionHasPixels(w *world, example map[string]string) error {
	return assertVisibleFrameRegion(example, w.visibleControlsFrame.RegionPixels, "had no non-background pixels")
}

func assertFrameRegionNotOnlyDebugText(w *world, example map[string]string) error {
	return assertVisibleFrameRegion(example, w.visibleControlsFrame.RegionControlCounts, "had no rendered controls")
}

func assertVisibleFrameRegion(example map[string]string, values map[string]int, message string) error {
	region, err := stringValue(example, "region")
	if err != nil {
		return err
	}
	if values[region] == 0 {
		return fmt.Errorf("screen region %q %s", region, message)
	}
	return nil
}

func assertVisibleControlReadableLabel(w *world, example map[string]string) error {
	control, label, err := stringPair(example, "control", "label")
	if err != nil {
		return err
	}
	if w.visibleControlsFrame.Controls[control] != label {
		return fmt.Errorf("control %q label = %q, want %q", control, w.visibleControlsFrame.Controls[control], label)
	}
	return nil
}

func assertInspectorSectionVisible(w *world, example map[string]string) error {
	section, err := stringValue(example, "section")
	if err != nil {
		return err
	}
	if !w.visibleControlsFrame.InspectorSections[section] {
		return fmt.Errorf("inspector section %q was not visible", section)
	}
	return nil
}

func setVisibleControlsApplicationState(w *world, example map[string]string) error {
	state, err := stringValue(example, "state")
	if err != nil {
		return err
	}
	game, err := visibleControlsGame(w)
	if err != nil {
		return err
	}
	setState, ok := visibleControlStateSetters[state]
	if !ok {
		return fmt.Errorf("unsupported visible controls state %q", state)
	}
	setState(game)
	return nil
}

func assertStatusFieldVisible(w *world, example map[string]string) error {
	field, err := stringValue(example, "field")
	if err != nil {
		return err
	}
	if w.visibleControlsFrame.StatusFields[field] == "" {
		return fmt.Errorf("status field %q was not visible", field)
	}
	return nil
}

func assertStatusFieldShows(w *world, example map[string]string) error {
	field, state, err := stringPair(example, "field", "state")
	if err != nil {
		return err
	}
	if !strings.Contains(w.visibleControlsFrame.StatusFields[field], state) {
		return fmt.Errorf("status field %q = %q, want it to show %q", field, w.visibleControlsFrame.StatusFields[field], state)
	}
	return nil
}

func assertCanvasWorldContentVisible(w *world, _ map[string]string) error {
	if w.visibleControlsFrame.CanvasWorldPixels == 0 {
		return fmt.Errorf("canvas had no visible world content")
	}
	return nil
}

func assertChromeDoesNotCoverAllWorldContent(w *world, _ map[string]string) error {
	return assertCanvasWorldContentVisible(w, nil)
}

func drawApplicationFrameDefaultSize(w *world, example map[string]string) error {
	return drawApplicationFrame(w, example)
}

func assertVisibleControlLabelsFit(w *world, _ map[string]string) error {
	if !w.visibleControlsFrame.ControlLabelsFit {
		return fmt.Errorf("visible control labels did not fit")
	}
	return nil
}

func visibleControlsGame(w *world) (*driverGame, error) {
	if w.appGame == nil {
		startApplicationDriver(w)
	}
	return concreteApplicationDriverWithMessage(w, "application was not started")
}
