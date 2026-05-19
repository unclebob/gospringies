package acceptance

import (
	"fmt"
	"image"
)

func renderRightInspector(w *world, _ map[string]string) error {
	game, err := visibleControlsGame(w)
	if err != nil {
		return err
	}
	w.visibleControlsFrame = game.DrawFrameReport()
	return nil
}

func assertNumericSettingVisibleSlider(w *world, example map[string]string) error {
	report, err := numericSettingFrame(w, example)
	if err != nil {
		return err
	}
	if report.SliderRect.Empty() {
		return fmt.Errorf("numeric setting slider was not visible")
	}
	return nil
}

func assertNumericSettingVisibleTextField(w *world, example map[string]string) error {
	report, err := numericSettingFrame(w, example)
	if err != nil {
		return err
	}
	if report.TextFieldRect.Empty() {
		return fmt.Errorf("numeric setting text field was not visible")
	}
	return nil
}

func assertNumericSettingTextFieldShowsValue(w *world, example map[string]string) error {
	setting, value, err := stringPair(example, "setting", "value")
	if err != nil {
		return err
	}
	return assertNumericSettingText(w, setting, value)
}

func assertNumericSettingTextFieldShowsNewValue(w *world, example map[string]string) error {
	setting, value, err := stringPair(example, "setting", "new_value")
	if err != nil {
		return err
	}
	return assertNumericSettingText(w, setting, value)
}

func assertStickTextFieldShowsNewValue(w *world, example map[string]string) error {
	value, err := stringValue(example, "new_value")
	if err != nil {
		return err
	}
	return assertNumericSettingText(w, "Stick", value)
}

func assertEveryNumericSettingLabelFitsInspector(w *world, _ map[string]string) error {
	for name, report := range w.visibleControlsFrame.NumericSettings {
		if !report.LabelRect.In(report.InspectorRect) {
			return fmt.Errorf("%s label outside inspector: %#v", name, report.LabelRect)
		}
	}
	return nil
}

func assertEveryNumericSettingSliderFitsInspector(w *world, _ map[string]string) error {
	for name, report := range w.visibleControlsFrame.NumericSettings {
		if !report.SliderRect.In(report.InspectorRect) {
			return fmt.Errorf("%s slider outside inspector: %#v", name, report.SliderRect)
		}
	}
	return nil
}

func assertEveryNumericSettingTextFieldFitsInspector(w *world, _ map[string]string) error {
	for name, report := range w.visibleControlsFrame.NumericSettings {
		if !report.TextFieldRect.In(report.InspectorRect) {
			return fmt.Errorf("%s text field outside inspector: %#v", name, report.TextFieldRect)
		}
	}
	return nil
}

func assertNumericSettingControlsDoNotOverlap(w *world, _ map[string]string) error {
	type namedRect struct {
		name string
		rect image.Rectangle
	}
	var rects []namedRect
	for name, report := range w.visibleControlsFrame.NumericSettings {
		rects = append(rects,
			namedRect{name + " label", report.LabelRect},
			namedRect{name + " slider", report.SliderRect},
			namedRect{name + " text field", report.TextFieldRect},
		)
	}
	for i := range rects {
		for j := i + 1; j < len(rects); j++ {
			if rectanglesOverlap(rects[i].rect, rects[j].rect) {
				return fmt.Errorf("%s overlaps %s", rects[i].name, rects[j].name)
			}
		}
	}
	return nil
}

func assertNumericSettingControlsDoNotOverlapHeadings(w *world, _ map[string]string) error {
	for name, report := range w.visibleControlsFrame.NumericSettings {
		for section, rect := range w.visibleControlsFrame.InspectorSectionRects {
			if rectanglesOverlap(report.LabelRect, rect) || rectanglesOverlap(report.SliderRect, rect) || rectanglesOverlap(report.TextFieldRect, rect) {
				return fmt.Errorf("%s controls overlap %s heading", name, section)
			}
		}
	}
	return nil
}

func renderNumericSetting(w *world, example map[string]string) error {
	return renderRightInspector(w, example)
}

func assertNumericSettingLabelLeftOfSlider(w *world, example map[string]string) error {
	report, err := numericSettingFrame(w, example)
	if err != nil {
		return err
	}
	if report.LabelRect.Max.X > report.SliderRect.Min.X {
		return fmt.Errorf("label %#v was not left of slider %#v", report.LabelRect, report.SliderRect)
	}
	return nil
}

func assertNumericSettingTextRightOfSlider(w *world, example map[string]string) error {
	report, err := numericSettingFrame(w, example)
	if err != nil {
		return err
	}
	if report.TextFieldRect.Min.X < report.SliderRect.Max.X {
		return fmt.Errorf("text field %#v was not right of slider %#v", report.TextFieldRect, report.SliderRect)
	}
	return nil
}

func assertNumericSettingSliderAndTextSameRow(w *world, example map[string]string) error {
	report, err := numericSettingFrame(w, example)
	if err != nil {
		return err
	}
	if report.SliderRect.Min.Y != report.TextFieldRect.Min.Y || report.SliderRect.Max.Y != report.TextFieldRect.Max.Y {
		return fmt.Errorf("slider %#v and text field %#v were not on same row", report.SliderRect, report.TextFieldRect)
	}
	return nil
}

func assertNumericSettingTextFieldFitsValue(w *world, example map[string]string) error {
	setting, value, err := stringPair(example, "setting", "longest_value")
	if err != nil {
		return err
	}
	report, ok := w.visibleControlsFrame.NumericSettings[setting]
	if !ok {
		return fmt.Errorf("missing numeric setting %q", setting)
	}
	if len(value)*6 > report.TextFieldRect.Dx()-8 {
		return fmt.Errorf("%s text field %#v does not fit %q", setting, report.TextFieldRect, value)
	}
	return nil
}

func changeNumericSettingWithSlider(w *world, example map[string]string) error {
	setting, value, err := stringPair(example, "setting", "new_value")
	if err != nil {
		return err
	}
	return changeNumericSettingWithSliderValue(w, setting, value)
}

func changeStickWithSlider(w *world, example map[string]string) error {
	value, err := stringValue(example, "new_value")
	if err != nil {
		return err
	}
	return changeNumericSettingWithSliderValue(w, "Stick", value)
}

func changeNumericSettingWithSliderValue(w *world, setting string, value string) error {
	game, err := visibleControlsGame(w)
	if err != nil {
		return err
	}
	if !game.ChangeNumericSettingWithSlider(setting, value) {
		return fmt.Errorf("slider change for %s to %s was not handled", setting, value)
	}
	w.visibleControlsFrame = game.DrawFrameReport()
	return nil
}

func assertParameterValueFromExample(w *world, example map[string]string) error {
	value, err := stringValue(example, "new_value")
	if err != nil {
		return err
	}
	game, err := visibleControlsGame(w)
	if err != nil {
		return err
	}
	if got := game.World().Parameters.Value("stickiness"); got != value {
		return fmt.Errorf("parameter stickiness = %q, want %q", got, value)
	}
	return nil
}

func setNumericSettingValue(w *world, example map[string]string) error {
	setting, value, err := stringPair(example, "setting", "old_value")
	if err != nil {
		return err
	}
	game, err := visibleControlsGame(w)
	if err != nil {
		return err
	}
	if !game.SetNumericSettingValue(setting, value) {
		return fmt.Errorf("setting %s to %s was not handled", setting, value)
	}
	w.visibleControlsFrame = game.DrawFrameReport()
	return nil
}

func focusNumericSettingTextField(w *world, example map[string]string) error {
	setting, err := stringValue(example, "setting")
	if err != nil {
		return err
	}
	game, err := visibleControlsGame(w)
	if err != nil {
		return err
	}
	if !game.FocusNumericSettingTextField(setting) {
		return fmt.Errorf("focus for numeric setting %q was not handled", setting)
	}
	w.visibleControlsFrame = game.DrawFrameReport()
	return nil
}

func assertNumericSettingCursorBlinks(w *world, example map[string]string) error {
	report, err := numericSettingFrame(w, example)
	if err != nil {
		return err
	}
	if !report.TextCursorVisible {
		return fmt.Errorf("numeric setting cursor was not visible")
	}
	return nil
}

func enterNumericTextValue(w *world, example map[string]string) error {
	value, err := stringValue(example, "new_value")
	if err != nil {
		value, err = stringValue(example, "final_value")
	}
	if err != nil {
		return err
	}
	game, err := visibleControlsGame(w)
	if err != nil {
		return err
	}
	if setting, settingErr := stringValue(example, "setting"); settingErr == nil {
		_, _ = game.NumericSettingText(setting)
		game.FocusNumericSettingTextField(setting)
	}
	if !game.EnterNumericSettingText(value) {
		return fmt.Errorf("numeric text entry %q was not handled", value)
	}
	w.visibleControlsFrame = game.DrawFrameReport()
	return nil
}

func assertNumericSettingHasValue(w *world, example map[string]string) error {
	setting, value, err := stringPair(example, "setting", "new_value")
	if err != nil {
		return err
	}
	return assertNumericSettingText(w, setting, value)
}

func assertNumericSettingSliderShowsValue(w *world, example map[string]string) error {
	setting, value, err := stringPair(example, "setting", "new_value")
	if err != nil {
		setting, value, err = stringPair(example, "setting", "final_value")
	}
	if err != nil {
		return err
	}
	game, err := visibleControlsGame(w)
	if err != nil {
		return err
	}
	got, ok := game.NumericSettingSliderValue(setting)
	if !ok {
		return fmt.Errorf("missing numeric setting %q", setting)
	}
	if got != value {
		return fmt.Errorf("%s slider value = %q, want %q", setting, got, value)
	}
	return nil
}

func numericSettingFrame(w *world, example map[string]string) (imageSettingFrame, error) {
	setting, err := stringValue(example, "setting")
	if err != nil {
		return imageSettingFrame{}, err
	}
	report, ok := w.visibleControlsFrame.NumericSettings[setting]
	if !ok {
		return imageSettingFrame{}, fmt.Errorf("missing numeric setting %q", setting)
	}
	return imageSettingFrame(report), nil
}

type imageSettingFrame struct {
	LabelRect          image.Rectangle
	SliderRect         image.Rectangle
	TextFieldRect      image.Rectangle
	InspectorRect      image.Rectangle
	Text               string
	SliderFraction     float64
	TextCursorVisible  bool
	LabelFitsInspector bool
}

func assertNumericSettingText(w *world, setting string, value string) error {
	report, ok := w.visibleControlsFrame.NumericSettings[setting]
	if !ok {
		return fmt.Errorf("missing numeric setting %q", setting)
	}
	if report.Text != value {
		return fmt.Errorf("%s text field = %q, want %q", setting, report.Text, value)
	}
	return nil
}

func rectanglesOverlap(a image.Rectangle, b image.Rectangle) bool {
	return a.Min.X < b.Max.X && a.Max.X > b.Min.X && a.Min.Y < b.Max.Y && a.Max.Y > b.Min.Y
}
