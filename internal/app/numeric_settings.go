package app

import (
	"fmt"
	"image"
	"strconv"
	"strings"
)

const (
	numericTextCursorPeriod   = 60
	numericStepButtonWidth    = 20
	numericStepButtonGap      = 4
	numericStepAmount         = 0.1
	numericStepHoldDelayTicks = 30
	numericStepRepeatTicks    = 6
	wallToggleButtonWidth     = 14
	wallToggleButtonGap       = 2
)

type numericSetting struct {
	Name      string
	Label     string
	Parameter string
	Control   string
	Force     string
	ForceKey  string
	Speed     bool
	Min       float64
	Max       float64
	Decimals  int
	Y         int
}

var numericSettings = []numericSetting{
	{Name: "Mass", Label: "Mass", Parameter: "current mass", Control: "mass", Min: 0, Max: 1000, Decimals: 1, Y: 68},
	{Name: "Elasticity", Label: "Elasticity", Parameter: "elasticity", Control: "elasticity", Min: 0, Max: 1, Decimals: 1, Y: 94},
	{Name: "Kspring", Label: "Kspring", Parameter: "spring constant", Control: "Kspring", Min: 0, Max: 1000, Decimals: 1, Y: 177},
	{Name: "Kdamp", Label: "Kdamp", Parameter: "damping", Control: "Kdamp", Min: 0, Max: 1000, Decimals: 1, Y: 203},
	{Name: "Gravity", Label: "Gravity", Force: "gravity", ForceKey: "magnitude", Min: 0, Max: 50, Decimals: 1, Y: 286},
	{Name: "Center Attraction", Label: "Center Attraction", Force: "center attraction", ForceKey: "magnitude", Min: 0, Max: 1000, Decimals: 1, Y: 312},
	{Name: "Center Of Mass Attraction", Label: "CM Attraction", Force: "center of mass attraction", ForceKey: "magnitude", Min: 0, Max: 1000, Decimals: 1, Y: 338},
	{Name: "Wall Repulsion", Label: "Wall Rep", Force: "wall repulsion", ForceKey: "magnitude", Min: 0, Max: 100000, Decimals: 1, Y: 364},
	{Name: "Viscosity", Label: "Viscosity", Parameter: "viscosity", Min: 0, Max: 2, Decimals: 1, Y: 421},
	{Name: "Stick", Label: "Stick", Parameter: "stickiness", Min: 0, Max: 10, Decimals: 1, Y: 447},
	{Name: "Speed", Label: "Speed", Speed: true, Min: 0, Max: maxSpeed, Decimals: 1, Y: 473},
	{Name: "Time Step", Label: "Time Step", Parameter: "timestep", Min: 0.0001, Max: 0.1, Decimals: 3, Y: 499},
	{Name: "Precision", Label: "Precision", Parameter: "precision", Min: 0.000001, Max: 0.01, Decimals: 3, Y: 525},
}

func numericSettingControls() []controlBox {
	var controls []controlBox
	for _, setting := range numericSettings {
		checkbox, label, decrement, slider, increment, text := numericSettingRects(setting)
		if checkboxName := numericSettingToggleControl(setting); checkboxName != "" {
			controls = append(controls, controlBox{Name: checkboxName, Label: "", Region: "right inspector", Rect: checkbox})
		}
		controls = append(controls, wallToggleControlsForSetting(setting)...)
		controls = append(controls,
			controlBox{Name: numericControlName(setting.Name, "label"), Label: setting.Label + ":", Region: "right inspector", Rect: label},
			controlBox{Name: numericControlName(setting.Name, "decrement"), Label: "<", Region: "right inspector", Rect: decrement},
			controlBox{Name: numericControlName(setting.Name, "slider"), Label: "", Region: "right inspector", Rect: slider},
			controlBox{Name: numericControlName(setting.Name, "increment"), Label: ">", Region: "right inspector", Rect: increment},
			controlBox{Name: numericControlName(setting.Name, "text field"), Label: "", Region: "right inspector", Rect: text},
		)
	}
	return controls
}

func numericSettingRects(setting numericSetting) (image.Rectangle, image.Rectangle, image.Rectangle, image.Rectangle, image.Rectangle, image.Rectangle) {
	left := inspectorLeft() + 16
	right := screenWidth - 16
	labelLeft := left
	labelRight := left + 168
	controlAnchor := left + 168
	checkbox := image.Rectangle{}
	if numericSettingToggleControl(setting) != "" {
		checkbox = image.Rect(left, setting.Y, left+numericStepButtonWidth, setting.Y+20)
		labelLeft = checkbox.Max.X + numericStepButtonGap
	}
	if setting.Force == "wall repulsion" {
		labelRight = labelLeft + 64
	}
	label := image.Rect(labelLeft, setting.Y, labelRight, setting.Y+20)
	decrement := image.Rect(controlAnchor+8, setting.Y, controlAnchor+8+numericStepButtonWidth, setting.Y+20)
	slider := image.Rect(decrement.Max.X+numericStepButtonGap, setting.Y, right-80-numericStepButtonWidth-numericStepButtonGap, setting.Y+20)
	increment := image.Rect(slider.Max.X+numericStepButtonGap, setting.Y, slider.Max.X+numericStepButtonGap+numericStepButtonWidth, setting.Y+20)
	text := image.Rect(right-72, setting.Y, right, setting.Y+20)
	return checkbox, label, decrement, slider, increment, text
}

func numericControlName(setting string, kind string) string {
	return strings.ToLower(setting) + " " + kind
}

func numericSettingForceToggleControl(setting numericSetting) string {
	return numericForceToggleControls[setting.Force]
}

func numericSettingParameterToggleControl(setting numericSetting) string {
	return numericParameterToggleControls[setting.Name]
}

func numericSettingToggleControl(setting numericSetting) string {
	if control := numericSettingForceToggleControl(setting); control != "" {
		return control
	}
	return numericSettingParameterToggleControl(setting)
}

var numericForceToggleControls = map[string]string{
	"gravity":                   "gravity force",
	"center attraction":         "center attraction force",
	"center of mass attraction": "center mass force",
	"wall repulsion":            "wall repulsion force",
}

var numericParameterToggleControls = map[string]string{
	"Precision": "adaptive timestep toggle",
}

func wallToggleControlsForSetting(setting numericSetting) []controlBox {
	if setting.Force != "wall repulsion" {
		return nil
	}
	_, label, _, _, _, _ := numericSettingRects(setting)
	left := label.Max.X + numericStepButtonGap
	specs := []struct {
		name  string
		label string
	}{
		{"top wall toggle", "T"},
		{"bottom wall toggle", "B"},
		{"left wall toggle", "L"},
		{"right wall toggle", "R"},
	}
	controls := make([]controlBox, 0, len(specs))
	for index, spec := range specs {
		x := left + index*(wallToggleButtonWidth+wallToggleButtonGap)
		controls = append(controls, controlBox{
			Name:   spec.name,
			Label:  spec.label,
			Region: "right inspector",
			Rect:   image.Rect(x, setting.Y, x+wallToggleButtonWidth, setting.Y+20),
		})
	}
	return controls
}

func numericSettingForSlider(name string) (numericSetting, bool) {
	return numericSettingForControl(name, "slider")
}

func numericSettingForStepButton(name string) (numericSetting, float64, bool) {
	if setting, ok := numericSettingForControl(name, "decrement"); ok {
		return setting, -numericStepAmount, true
	}
	if setting, ok := numericSettingForControl(name, "increment"); ok {
		return setting, numericStepAmount, true
	}
	return numericSetting{}, 0, false
}

func numericSettingForTextField(name string) (numericSetting, bool) {
	return numericSettingForControl(name, "text field")
}

func numericSettingForControl(name string, kind string) (numericSetting, bool) {
	for _, setting := range numericSettings {
		if name == numericControlName(setting.Name, kind) {
			return setting, true
		}
	}
	return numericSetting{}, false
}

func numericSettingByName(name string) (numericSetting, bool) {
	for _, setting := range numericSettings {
		if setting.Name == name {
			return setting, true
		}
	}
	return numericSetting{}, false
}

func (g *Game) numericSettingValue(setting numericSetting) float64 {
	value, _ := strconv.ParseFloat(g.committedNumericSettingValueText(setting), 64)
	return value
}

func (g *Game) numericSettingValueText(setting numericSetting) string {
	if g.controls.focusedNumeric == setting.Name {
		return g.controls.numericInputText
	}
	raw := g.rawNumericSettingValue(setting)
	return formatNumericSettingText(raw, setting.Decimals)
}

func (g *Game) committedNumericSettingValueText(setting numericSetting) string {
	return formatNumericSettingText(g.rawNumericSettingValue(setting), setting.Decimals)
}

func (g *Game) rawNumericSettingValue(setting numericSetting) string {
	switch {
	case setting.Speed:
		return formatControlFloat(g.simulationSpeed)
	case setting.Force != "":
		force, _ := g.simulation.Parameters.Force(setting.Force)
		return force.Values[setting.ForceKey]
	default:
		return g.simulation.Parameters.Value(setting.Parameter)
	}
}

func formatNumericSettingText(raw string, decimals int) string {
	if raw == "" {
		raw = "0"
	}
	if strings.ContainsAny(raw, ".eE") || decimals <= 0 {
		return raw
	}
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return raw
	}
	return strconv.FormatFloat(value, 'f', decimals, 64)
}

func (g *Game) numericSettingSliderFraction(setting numericSetting) float64 {
	if setting.Max <= setting.Min {
		return 0
	}
	value := g.numericSettingValue(setting)
	return clampFloat((value-setting.Min)/(setting.Max-setting.Min), 0, 1)
}

func (g *Game) setNumericSettingFromSlider(setting numericSetting, x int) {
	_, _, _, slider, _, _ := numericSettingRects(setting)
	fraction := sliderFractionAt(sliderTrack(controlBox{Rect: slider}), x)
	value := setting.Min + fraction*(setting.Max-setting.Min)
	g.setNumericSettingValue(setting, formatNumericSettingSliderValue(value, setting.Decimals))
}

func (g *Game) stepNumericSetting(setting numericSetting, delta float64) {
	value := clampFloat(g.numericSettingValue(setting)+delta, setting.Min, setting.Max)
	g.setNumericSettingValue(setting, formatNumericSettingSliderValue(value, setting.Decimals))
}

func formatNumericSettingSliderValue(value float64, decimals int) string {
	if decimals > 0 {
		return strconv.FormatFloat(roundControlFloat(value), 'f', decimals, 64)
	}
	return formatControlFloat(value)
}

func (g *Game) setNumericSettingValue(setting numericSetting, text string) bool {
	value, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return false
	}
	switch {
	case setting.Speed:
		g.simulationSpeed = clampFloat(value, setting.Min, setting.Max)
	case setting.Force != "":
		g.setForceValue(setting.Force, setting.ForceKey, value)
	default:
		g.setParameterNumericSetting(setting, text)
	}
	g.dirty = true
	return true
}

func (g *Game) setParameterNumericSetting(setting numericSetting, text string) {
	g.simulation.Parameters.Set(setting.Parameter, text)
	if setting.Control != "" {
		_ = g.editing().ChangeControl(setting.Control, text)
	}
}

func (g *Game) focusNumericSettingTextField(setting numericSetting) {
	g.controls.focusedNumeric = setting.Name
	g.controls.numericInputText = g.committedNumericSettingValueText(setting)
	g.controls.numericInputTicks = 0
	g.controls.numericInputFresh = true
}

func (g *Game) appendNumericSettingInput(chars []rune) {
	setting, ok := g.focusedNumericSetting()
	if !ok {
		return
	}
	for _, char := range chars {
		g.appendNumericSettingCharacter(setting, char)
	}
}

func (g *Game) deleteNumericSettingCharacter() {
	if g.controls.focusedNumeric == "" || len(g.controls.numericInputText) == 0 {
		return
	}
	if _, ok := g.focusedNumericSetting(); !ok {
		g.cancelNumericSettingInput()
		return
	}
	g.controls.numericInputText = g.controls.numericInputText[:len(g.controls.numericInputText)-1]
	g.controls.numericInputFresh = false
}

func (g *Game) commitNumericSettingInput() bool {
	if g.controls.focusedNumeric == "" {
		return false
	}
	setting, ok := g.focusedNumericSetting()
	if !ok {
		g.cancelNumericSettingInput()
		return false
	}
	if !g.setNumericSettingValue(setting, g.controls.numericInputText) {
		return false
	}
	g.controls.focusedNumeric = ""
	g.controls.numericInputFresh = false
	return true
}

func (g *Game) cancelNumericSettingInput() {
	g.controls.focusedNumeric = ""
	g.controls.numericInputText = ""
	g.controls.numericInputFresh = false
}

func (g *Game) focusedNumericSetting() (numericSetting, bool) {
	if g.controls.focusedNumeric == "" {
		return numericSetting{}, false
	}
	return numericSettingByName(g.controls.focusedNumeric)
}

func (g *Game) appendNumericSettingCharacter(setting numericSetting, char rune) {
	if !isNumericInputCharacter(char) {
		return
	}
	if g.controls.numericInputFresh {
		g.controls.numericInputText = ""
		g.controls.numericInputFresh = false
	}
	g.controls.numericInputText += string(char)
}

func isNumericInputCharacter(char rune) bool {
	return strings.ContainsRune("0123456789.-", char)
}

func (g *Game) tickNumericTextField() {
	if g.controls.focusedNumeric != "" {
		g.controls.numericInputTicks++
	}
}

func (g *Game) numericTextCursorVisible(setting string) bool {
	return g.controls.focusedNumeric == setting && (g.controls.numericInputTicks/numericTextCursorPeriod)%2 == 0
}

func (g *Game) numericTextHighlighted(setting string) bool {
	return g.controls.focusedNumeric == setting && g.controls.numericInputFresh
}

func (g *Game) NumericSettingReport(settingName string) (NumericSettingFrame, bool) {
	setting, ok := numericSettingByName(settingName)
	if !ok {
		return NumericSettingFrame{}, false
	}
	checkbox, label, decrement, slider, increment, text := numericSettingRects(setting)
	return NumericSettingFrame{
		CheckboxRect:       checkbox,
		LabelRect:          label,
		DecrementRect:      decrement,
		SliderRect:         slider,
		IncrementRect:      increment,
		TextFieldRect:      text,
		InspectorRect:      inspectorRect(),
		Text:               g.numericSettingValueText(setting),
		SliderFraction:     g.numericSettingSliderFraction(setting),
		TextCursorVisible:  g.numericTextCursorVisible(setting.Name),
		TextHighlighted:    g.numericTextHighlighted(setting.Name),
		LabelFitsInspector: label.In(inspectorRect()),
	}, true
}

func (g *Game) SetNumericSettingValue(settingName string, text string) bool {
	setting, ok := numericSettingByName(settingName)
	return ok && g.setNumericSettingValue(setting, text)
}

func (g *Game) ChangeNumericSettingWithSlider(settingName string, text string) bool {
	setting, ok := numericSettingByName(settingName)
	if !ok {
		return false
	}
	value, err := strconv.ParseFloat(text, 64)
	if err != nil || setting.Max <= setting.Min {
		return false
	}
	_, _, _, slider, _, _ := numericSettingRects(setting)
	track := sliderTrack(controlBox{Rect: slider})
	fraction := clampFloat((value-setting.Min)/(setting.Max-setting.Min), 0, 1)
	x := track.Min.X + int(fraction*float64(track.Dx()))
	return g.ClickAt(x, slider.Min.Y+slider.Dy()/2)
}

func (g *Game) FocusNumericSettingTextField(settingName string) bool {
	setting, ok := numericSettingByName(settingName)
	if !ok {
		return false
	}
	g.focusNumericSettingTextField(setting)
	return true
}

func (g *Game) EnterNumericSettingText(text string) bool {
	if !g.TypeNumericSettingText(text) {
		return false
	}
	return g.CommitNumericSettingText()
}

func (g *Game) TypeNumericSettingText(text string) bool {
	if g.controls.focusedNumeric == "" {
		return false
	}
	g.controls.numericInputText = ""
	g.controls.numericInputFresh = false
	g.appendNumericSettingInput([]rune(text))
	return true
}

func (g *Game) CommitNumericSettingText() bool {
	return g.commitNumericSettingInput()
}

func (g *Game) NumericSettingText(settingName string) (string, bool) {
	report, ok := g.NumericSettingReport(settingName)
	return report.Text, ok
}

func (g *Game) NumericSettingSliderValue(settingName string) (string, bool) {
	setting, ok := numericSettingByName(settingName)
	if !ok {
		return "", false
	}
	return g.committedNumericSettingValueText(setting), true
}

func inspectorRect() image.Rectangle {
	return image.Rect(screenWidth-inspectorWidth, topBarHeight, screenWidth, screenHeight)
}

func numericSettingReports(g *Game) map[string]NumericSettingFrame {
	reports := map[string]NumericSettingFrame{}
	for _, setting := range numericSettings {
		report, _ := g.NumericSettingReport(setting.Name)
		reports[setting.Name] = report
	}
	return reports
}

func validateNumericSetting(setting string) error {
	if _, ok := numericSettingByName(setting); !ok {
		return fmt.Errorf("unsupported numeric setting %q", setting)
	}
	return nil
}
