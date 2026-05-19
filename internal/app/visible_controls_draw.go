//go:build !appunit

package app

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func (g *Game) drawVisibleControls(screen *ebiten.Image) {
	for _, control := range visibleControls() {
		g.drawControl(screen, control)
	}
	for _, control := range g.editMenuControls() {
		g.drawControl(screen, control)
	}
	for _, section := range inspectorSections() {
		drawLabeledRect(screen, section.Rect, sectionColor, section.Label)
	}
	for _, field := range g.statusFields() {
		drawLabeledRect(screen, field.Rect, controlColor, field.Label)
	}
}

func (g *Game) drawControl(screen *ebiten.Image, control controlBox) {
	if isSliderControl(control.Name) {
		g.drawSlider(screen, control)
		return
	}
	if setting, ok := numericSettingForTextField(control.Name); ok {
		g.drawNumericTextField(screen, control, setting)
		return
	}
	fill := controlColor
	if g.activeControl(control.Name) {
		fill = activeControlColor
	}
	drawLabeledRect(screen, control.Rect, fill, control.Label)
}

func (g *Game) drawSlider(screen *ebiten.Image, control controlBox) {
	drawLabeledRect(screen, control.Rect, controlColor, g.sliderLabel(control))
	track := sliderTrack(control)
	vector.DrawFilledRect(screen, float32(track.Min.X), float32(track.Min.Y), float32(track.Dx()), float32(track.Dy()), sectionColor, false)
	fill := track
	fill.Max.X = track.Min.X + int(g.sliderFraction(control.Name)*float64(track.Dx()))
	vector.DrawFilledRect(screen, float32(fill.Min.X), float32(fill.Min.Y), float32(fill.Dx()), float32(fill.Dy()), activeControlColor, false)
}

func (g *Game) drawNumericTextField(screen *ebiten.Image, control controlBox, setting numericSetting) {
	drawLabeledRect(screen, control.Rect, controlColor, g.numericSettingValueText(setting))
	if !g.numericTextCursorVisible(setting.Name) {
		return
	}
	x := control.Rect.Min.X + 4 + len(g.numericSettingValueText(setting))*debugGlyphWidth
	vector.DrawFilledRect(screen, float32(x), float32(control.Rect.Min.Y+4), 1, debugGlyphHeight, activeControlColor, false)
}

func drawLabeledRect(screen *ebiten.Image, rect image.Rectangle, fill color.RGBA, label string) {
	vector.DrawFilledRect(screen, float32(rect.Min.X), float32(rect.Min.Y), float32(rect.Dx()), float32(rect.Dy()), fill, false)
	ebitenutil.DebugPrintAt(screen, label, rect.Min.X+4, rect.Min.Y+4)
}
