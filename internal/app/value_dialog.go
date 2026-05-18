package app

import (
	"fmt"
	"image"
	"math"
	"strconv"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"springs/internal/sim"
)

const (
	valueDialogWidth  = 320
	valueDialogHeight = 150
	valueCursorPeriod = 60
)

type valueDialog struct {
	Open   bool
	Title  string
	Text   string
	Min    float64
	Max    float64
	Apply  func(float64)
	Target string
	Ticks  int
}

func (g *Game) openMassValueDialog(id int) {
	mass, ok := g.simulation.MassByID(id)
	if !ok {
		return
	}
	g.valueDialog = valueDialog{
		Open:   true,
		Title:  fmt.Sprintf("Set Mass #%d", id),
		Text:   formatControlFloat(mass.Mass),
		Min:    0,
		Max:    20,
		Target: "mass",
		Apply: func(value float64) {
			g.setMassValue(id, value)
		},
	}
}

func (g *Game) openSpringConstantDialogAt(x int, y int) bool {
	id, ok := g.springAt(g.screenToWorld(simVec(x, y)))
	if !ok {
		return false
	}
	spring, ok := g.simulation.SpringByID(id)
	if !ok {
		return false
	}
	g.valueDialog = valueDialog{
		Open:   true,
		Title:  fmt.Sprintf("Set Spring #%d", id),
		Text:   formatControlFloat(spring.SpringConstant),
		Min:    0,
		Max:    50,
		Target: "spring",
		Apply: func(value float64) {
			g.setSpringConstant(id, value)
		},
	}
	return true
}

func (g *Game) drawValueDialog(screen *ebiten.Image) {
	rect := valueDialogRect()
	vector.DrawFilledRect(screen, float32(rect.Min.X), float32(rect.Min.Y), float32(rect.Dx()), float32(rect.Dy()), panelColor, false)
	ebitenutil.DebugPrintAt(screen, g.valueDialog.Title, rect.Min.X+12, rect.Min.Y+10)
	drawLabeledRect(screen, g.valueDialogTextRect(), controlColor, g.valueDialog.Text)
	g.drawValueDialogCursor(screen)
	track := g.valueDialogSliderTrack()
	vector.DrawFilledRect(screen, float32(track.Min.X), float32(track.Min.Y), float32(track.Dx()), float32(track.Dy()), sectionColor, false)
	fill := track
	fill.Max.X = fill.Min.X + int(g.valueDialogFraction()*float64(track.Dx()))
	vector.DrawFilledRect(screen, float32(fill.Min.X), float32(fill.Min.Y), float32(fill.Dx()), float32(fill.Dy()), activeControlColor, false)
	drawLabeledRect(screen, g.valueDialogOKRect(), activeControlColor, "OK")
}

func (g *Game) tickValueDialog() {
	if !g.valueDialog.Open {
		return
	}
	g.valueDialog.Ticks++
}

func (g *Game) drawValueDialogCursor(screen *ebiten.Image) {
	if !g.valueDialogCursorVisible() {
		return
	}
	rect := g.valueDialogTextRect()
	x := rect.Min.X + 4 + len(g.valueDialog.Text)*debugGlyphWidth
	if x > rect.Max.X-6 {
		x = rect.Max.X - 6
	}
	vector.DrawFilledRect(screen, float32(x), float32(rect.Min.Y+4), 2, float32(debugGlyphHeight-2), selectionColor, false)
}

func (g *Game) valueDialogCursorVisible() bool {
	if !g.valueDialog.Open {
		return false
	}
	return (g.valueDialog.Ticks/valueCursorPeriod)%2 == 0
}

func (g *Game) clickValueDialog(x int, y int) {
	point := image.Pt(x, y)
	if !point.In(valueDialogRect()) {
		g.valueDialog.Open = false
		return
	}
	if point.In(g.valueDialogSliderTrack()) {
		g.setValueDialogFromSlider(x)
		return
	}
	if point.In(g.valueDialogOKRect()) {
		g.applyValueDialog()
		return
	}
}

func (g *Game) pollValueDialogKeyboard() {
	if !g.valueDialog.Open {
		return
	}
	for _, char := range ebiten.AppendInputChars(nil) {
		if strings.ContainsRune("0123456789.-", char) {
			g.valueDialog.Text += string(char)
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) && len(g.valueDialog.Text) > 0 {
		g.valueDialog.Text = g.valueDialog.Text[:len(g.valueDialog.Text)-1]
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyKPEnter) {
		g.applyValueDialog()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.valueDialog.Open = false
	}
}

func (g *Game) applyValueDialog() {
	value, err := strconv.ParseFloat(g.valueDialog.Text, 64)
	if err != nil {
		return
	}
	if g.valueDialog.Apply != nil {
		g.valueDialog.Apply(value)
	}
	g.valueDialog.Open = false
}

func (g *Game) setValueDialogFromSlider(x int) {
	track := g.valueDialogSliderTrack()
	if track.Dx() <= 0 {
		return
	}
	fraction := clampFloat(float64(x-track.Min.X)/float64(track.Dx()), 0, 1)
	value := g.valueDialog.Min + fraction*(g.valueDialog.Max-g.valueDialog.Min)
	g.valueDialog.Text = formatControlFloat(value)
}

func (g *Game) valueDialogFraction() float64 {
	value, err := strconv.ParseFloat(g.valueDialog.Text, 64)
	if err != nil || g.valueDialog.Max <= g.valueDialog.Min {
		return 0
	}
	return clampFloat((value-g.valueDialog.Min)/(g.valueDialog.Max-g.valueDialog.Min), 0, 1)
}

func valueDialogRect() image.Rectangle {
	x := screenWidth/2 - valueDialogWidth/2
	y := screenHeight/2 - valueDialogHeight/2
	return image.Rect(x, y, x+valueDialogWidth, y+valueDialogHeight)
}

func (g *Game) valueDialogTextRect() image.Rectangle {
	rect := valueDialogRect()
	return image.Rect(rect.Min.X+12, rect.Min.Y+42, rect.Max.X-12, rect.Min.Y+66)
}

func (g *Game) valueDialogSliderTrack() image.Rectangle {
	rect := valueDialogRect()
	return image.Rect(rect.Min.X+12, rect.Min.Y+92, rect.Max.X-12, rect.Min.Y+100)
}

func (g *Game) valueDialogOKRect() image.Rectangle {
	rect := valueDialogRect()
	return image.Rect(rect.Max.X-64, rect.Max.Y-34, rect.Max.X-12, rect.Max.Y-12)
}

func (g *Game) setSpringConstant(id int, value float64) {
	for i := range g.simulation.Springs {
		if g.simulation.Springs[i].ID == id {
			g.simulation.Springs[i].SpringConstant = value
			g.simulation.Springs[i].Stiffness = value
			g.dirty = true
			return
		}
	}
}

func (g *Game) springAt(position sim.Vec2) (int, bool) {
	for _, spring := range g.simulation.Springs {
		a, okA := g.simulation.MassByID(spring.MassA)
		b, okB := g.simulation.MassByID(spring.MassB)
		if okA && okB && distanceToSegment(position, a.Position, b.Position) <= 6 {
			return spring.ID, true
		}
	}
	return 0, false
}

func distanceToSegment(p sim.Vec2, a sim.Vec2, b sim.Vec2) float64 {
	ab := b.Sub(a)
	ap := p.Sub(a)
	lengthSquared := ab.X*ab.X + ab.Y*ab.Y
	if lengthSquared == 0 {
		return math.Hypot(ap.X, ap.Y)
	}
	t := clampFloat((ap.X*ab.X+ap.Y*ab.Y)/lengthSquared, 0, 1)
	closest := a.Add(ab.Scale(t))
	return math.Hypot(p.X-closest.X, p.Y-closest.Y)
}
