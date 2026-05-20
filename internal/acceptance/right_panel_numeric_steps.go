package acceptance

import (
	"fmt"
	"image"

	"springs/internal/app"
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
	return assertNumericSettingRectVisible(w, example, "slider", numericSettingSliderRect)
}

func assertNumericSettingVisibleTextField(w *world, example map[string]string) error {
	return assertNumericSettingRectVisible(w, example, "text field", numericSettingTextFieldRect)
}

func assertNumericSettingRectVisible(w *world, example map[string]string, name string, rect func(imageSettingFrame) image.Rectangle) error {
	report, err := numericSettingFrame(w, example)
	if err != nil {
		return err
	}
	if rect(report).Empty() {
		return fmt.Errorf("numeric setting %s was not visible", name)
	}
	return nil
}

func assertNumericSettingTextFieldShowsValue(w *world, example map[string]string) error {
	return assertNumericSettingTextFromExample(w, example, "value")
}

func assertNumericSettingTextFieldShowsNewValue(w *world, example map[string]string) error {
	return assertNumericSettingTextFromExample(w, example, "new_value")
}

func assertStickTextFieldShowsNewValue(w *world, example map[string]string) error {
	return assertFixedNumericSettingTextFromExample(w, example, "Stick", "new_value")
}

func assertEveryNumericSettingLabelFitsInspector(w *world, _ map[string]string) error {
	return assertEveryNumericSettingRectFitsInspector(w, "label", numericSettingLabelRect)
}

func assertEveryNumericSettingSliderFitsInspector(w *world, _ map[string]string) error {
	return assertEveryNumericSettingRectFitsInspector(w, "slider", numericSettingSliderRect)
}

func assertEveryNumericSettingTextFieldFitsInspector(w *world, _ map[string]string) error {
	return assertEveryNumericSettingRectFitsInspector(w, "text field", numericSettingTextFieldRect)
}

func assertEveryNumericSettingRectFitsInspector(w *world, controlName string, rect func(imageSettingFrame) image.Rectangle) error {
	for settingName, report := range w.visibleControlsFrame.NumericSettings {
		settingRect := rect(imageSettingFrame(report))
		if !settingRect.In(report.InspectorRect) {
			return fmt.Errorf("%s %s outside inspector: %#v", settingName, controlName, settingRect)
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
			namedRect{name + " checkbox", report.CheckboxRect},
			namedRect{name + " label", report.LabelRect},
			namedRect{name + " decrement", report.DecrementRect},
			namedRect{name + " slider", report.SliderRect},
			namedRect{name + " increment", report.IncrementRect},
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
			if numericSettingControlsOverlap(report, rect) {
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
	return changeNumericSettingWithSliderFromExample(w, example, "new_value")
}

func changeStickWithSlider(w *world, example map[string]string) error {
	return changeFixedNumericSettingWithSliderFromExample(w, example, "Stick", "new_value")
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

func changeNumericSettingWithSliderFromExample(w *world, example map[string]string, valueKey string) error {
	return changeNumericSettingWithSliderFromSetting(w, example, "", valueKey)
}

func changeFixedNumericSettingWithSliderFromExample(w *world, example map[string]string, setting string, valueKey string) error {
	return changeNumericSettingWithSliderFromSetting(w, example, setting, valueKey)
}

func changeNumericSettingWithSliderFromSetting(w *world, example map[string]string, fixedSetting string, valueKey string) error {
	return applyNumericSettingValue(w, example, fixedSetting, []string{valueKey}, changeNumericSettingWithSliderValue)
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
	value, err := numericTextEntryValue(example)
	if err != nil {
		return err
	}
	game, err := visibleControlsGame(w)
	if err != nil {
		return err
	}
	focusNumericSettingFromExample(game, example)
	if !game.EnterNumericSettingText(value) {
		return fmt.Errorf("numeric text entry %q was not handled", value)
	}
	w.visibleControlsFrame = game.DrawFrameReport()
	return nil
}

func assertNumericSettingHasValue(w *world, example map[string]string) error {
	return assertNumericSettingTextFromExample(w, example, "new_value")
}

func assertNumericSettingSliderShowsValue(w *world, example map[string]string) error {
	setting, value, err := numericSettingValueFromExample(example, "new_value", "final_value")
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

func assertNumericSettingTextFromExample(w *world, example map[string]string, valueKey string) error {
	return assertNumericSettingTextFromSetting(w, example, "", valueKey)
}

func assertFixedNumericSettingTextFromExample(w *world, example map[string]string, setting string, valueKey string) error {
	return assertNumericSettingTextFromSetting(w, example, setting, valueKey)
}

func numericSettingValueFromExample(example map[string]string, valueKeys ...string) (string, string, error) {
	return numericSettingValueWithFallback(example, "", valueKeys...)
}

func assertNumericSettingTextFromSetting(w *world, example map[string]string, fixedSetting string, valueKey string) error {
	return applyNumericSettingValue(w, example, fixedSetting, []string{valueKey}, assertNumericSettingText)
}

func applyNumericSettingValue(w *world, example map[string]string, fixedSetting string, valueKeys []string, action func(*world, string, string) error) error {
	setting, value, err := numericSettingValueWithFallback(example, fixedSetting, valueKeys...)
	if err != nil {
		return err
	}
	return action(w, setting, value)
}

func numericSettingValueWithFallback(example map[string]string, fixedSetting string, valueKeys ...string) (string, string, error) {
	var lastErr error
	for _, valueKey := range valueKeys {
		setting, value, err := numericSettingValueForKey(example, fixedSetting, valueKey)
		if err == nil {
			return setting, value, nil
		}
		lastErr = err
	}
	return "", "", lastErr
}

func numericSettingValueForKey(example map[string]string, fixedSetting string, valueKey string) (string, string, error) {
	if fixedSetting != "" {
		value, err := stringValue(example, valueKey)
		return fixedSetting, value, err
	}
	return stringPair(example, "setting", valueKey)
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
	CheckboxRect       image.Rectangle
	LabelRect          image.Rectangle
	DecrementRect      image.Rectangle
	SliderRect         image.Rectangle
	IncrementRect      image.Rectangle
	TextFieldRect      image.Rectangle
	InspectorRect      image.Rectangle
	Text               string
	SliderFraction     float64
	TextCursorVisible  bool
	TextHighlighted    bool
	LabelFitsInspector bool
}

func numericSettingLabelRect(report imageSettingFrame) image.Rectangle {
	return report.LabelRect
}

func numericSettingSliderRect(report imageSettingFrame) image.Rectangle {
	return report.SliderRect
}

func numericSettingTextFieldRect(report imageSettingFrame) image.Rectangle {
	return report.TextFieldRect
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

func numericSettingControlsOverlap(report app.NumericSettingFrame, rect image.Rectangle) bool {
	return rectanglesOverlap(report.LabelRect, rect) ||
		rectanglesOverlap(report.CheckboxRect, rect) ||
		rectanglesOverlap(report.DecrementRect, rect) ||
		rectanglesOverlap(report.SliderRect, rect) ||
		rectanglesOverlap(report.IncrementRect, rect) ||
		rectanglesOverlap(report.TextFieldRect, rect)
}

func numericTextEntryValue(example map[string]string) (string, error) {
	value, err := stringValue(example, "new_value")
	if err == nil {
		return value, nil
	}
	return stringValue(example, "final_value")
}

func focusNumericSettingFromExample(game *app.Game, example map[string]string) {
	setting, err := stringValue(example, "setting")
	if err != nil {
		return
	}
	_, _ = game.NumericSettingText(setting)
	game.FocusNumericSettingTextField(setting)
}
