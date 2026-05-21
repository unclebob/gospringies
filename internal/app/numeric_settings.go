package app

import (
	"fmt"
	"image"
	"strconv"
	"strings"
)

const numericTextCursorPeriod = 60

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
	{Name: "Kspring", Label: "Kspring", Parameter: "spring constant", Control: "Kspring", Min: 0, Max: 1000, Decimals: 1, Y: 172},
	{Name: "Kdamp", Label: "Kdamp", Parameter: "damping", Control: "Kdamp", Min: 0, Max: 1000, Decimals: 1, Y: 198},
	{Name: "Gravity", Label: "Gravity", Force: "gravity", ForceKey: "magnitude", Min: 0, Max: 50, Decimals: 1, Y: 278},
	{Name: "Center Attraction", Label: "Center Attraction", Force: "center attraction", ForceKey: "magnitude", Min: 0, Max: 1000, Decimals: 1, Y: 304},
	{Name: "Center Of Mass Attraction", Label: "Center Of Mass Attraction", Force: "center of mass attraction", ForceKey: "magnitude", Min: 0, Max: 1000, Decimals: 1, Y: 330},
	{Name: "Wall Repulsion", Label: "Wall Repulsion", Force: "wall repulsion", ForceKey: "magnitude", Min: 0, Max: 100000, Decimals: 1, Y: 356},
	{Name: "Viscosity", Label: "Viscosity", Parameter: "viscosity", Min: 0, Max: 2, Decimals: 1, Y: 594},
	{Name: "Stick", Label: "Stick", Parameter: "stickiness", Min: 0, Max: 10, Decimals: 1, Y: 620},
	{Name: "Speed", Label: "Speed", Speed: true, Min: 0, Max: maxSpeed, Decimals: 1, Y: 646},
	{Name: "Time Step", Label: "Time Step", Parameter: "timestep", Min: 0.0001, Max: 0.1, Decimals: 3, Y: 672},
	{Name: "Precision", Label: "Precision", Parameter: "precision", Min: 0.000001, Max: 0.01, Decimals: 3, Y: 698},
}

func numericSettingControls() []controlBox {
	var controls []controlBox
	for _, setting := range numericSettings {
		label, slider, text := numericSettingRects(setting.Y)
		controls = append(controls,
			controlBox{Name: numericControlName(setting.Name, "label"), Label: setting.Label + ":", Region: "right inspector", Rect: label},
			controlBox{Name: numericControlName(setting.Name, "slider"), Label: "", Region: "right inspector", Rect: slider},
			controlBox{Name: numericControlName(setting.Name, "text field"), Label: "", Region: "right inspector", Rect: text},
		)
	}
	return controls
}

func numericSettingRects(y int) (image.Rectangle, image.Rectangle, image.Rectangle) {
	left := inspectorLeft() + 16
	right := screenWidth - 16
	label := image.Rect(left, y, left+168, y+20)
	slider := image.Rect(label.Max.X+8, y, right-80, y+20)
	text := image.Rect(right-72, y, right, y+20)
	return label, slider, text
}

func numericControlName(setting string, kind string) string {
	return strings.ToLower(setting) + " " + kind
}

func numericSettingForSlider(name string) (numericSetting, bool) {
	return numericSettingForControl(name, "slider")
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
	value, _ := strconv.ParseFloat(g.numericSettingValueText(setting), 64)
	return value
}

func (g *Game) numericSettingValueText(setting numericSetting) string {
	if g.focusedNumeric == setting.Name {
		return g.numericInputText
	}
	raw := g.rawNumericSettingValue(setting)
	return formatNumericSettingText(raw, setting.Decimals)
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
	_, slider, _ := numericSettingRects(setting.Y)
	fraction := sliderFractionAt(sliderTrack(controlBox{Rect: slider}), x)
	value := setting.Min + fraction*(setting.Max-setting.Min)
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
	g.focusedNumeric = setting.Name
	g.numericInputText = g.numericSettingValueText(setting)
	g.numericInputTicks = 0
	g.numericInputFresh = true
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
	if g.focusedNumeric == "" || len(g.numericInputText) == 0 {
		return
	}
	setting, ok := g.focusedNumericSetting()
	if !ok {
		return
	}
	g.numericInputText = g.numericInputText[:len(g.numericInputText)-1]
	g.numericInputFresh = false
	g.setNumericSettingValue(setting, g.numericInputText)
}

func (g *Game) focusedNumericSetting() (numericSetting, bool) {
	if g.focusedNumeric == "" {
		return numericSetting{}, false
	}
	return numericSettingByName(g.focusedNumeric)
}

func (g *Game) appendNumericSettingCharacter(setting numericSetting, char rune) {
	if !isNumericInputCharacter(char) {
		return
	}
	if g.numericInputFresh {
		g.numericInputText = ""
		g.numericInputFresh = false
	}
	g.numericInputText += string(char)
	g.setNumericSettingValue(setting, g.numericInputText)
}

func isNumericInputCharacter(char rune) bool {
	return strings.ContainsRune("0123456789.-", char)
}

func (g *Game) tickNumericTextField() {
	if g.focusedNumeric != "" {
		g.numericInputTicks++
	}
}

func (g *Game) numericTextCursorVisible(setting string) bool {
	return g.focusedNumeric == setting && (g.numericInputTicks/numericTextCursorPeriod)%2 == 0
}

func (g *Game) NumericSettingReport(settingName string) (NumericSettingFrame, bool) {
	setting, ok := numericSettingByName(settingName)
	if !ok {
		return NumericSettingFrame{}, false
	}
	label, slider, text := numericSettingRects(setting.Y)
	return NumericSettingFrame{
		LabelRect:          label,
		SliderRect:         slider,
		TextFieldRect:      text,
		InspectorRect:      inspectorRect(),
		Text:               g.numericSettingValueText(setting),
		SliderFraction:     g.numericSettingSliderFraction(setting),
		TextCursorVisible:  g.numericTextCursorVisible(setting.Name),
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
	_, slider, _ := numericSettingRects(setting.Y)
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
	if g.focusedNumeric == "" {
		return false
	}
	g.numericInputText = ""
	g.numericInputFresh = false
	g.appendNumericSettingInput([]rune(text))
	return true
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
	return g.numericSettingValueText(setting), true
}

func inspectorRect() image.Rectangle {
	return image.Rect(screenWidth-inspectorWidth, topBarHeight, screenWidth, screenHeight-statusHeight)
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
