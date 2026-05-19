package app

import (
	"fmt"
	"image"
	"math"
	"strconv"
	"strings"

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
	spring, ok := g.springAtPosition(g.screenToWorld(simVec(x, y)))
	if !ok {
		return false
	}
	id := spring.ID
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

func (g *Game) tickValueDialog() {
	if !g.valueDialog.Open {
		return
	}
	g.valueDialog.Ticks++
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

func (g *Game) appendValueDialogInput(chars []rune) {
	for _, char := range chars {
		if strings.ContainsRune("0123456789.-", char) {
			g.valueDialog.Text += string(char)
		}
	}
}

func (g *Game) deleteValueDialogCharacter() {
	if len(g.valueDialog.Text) > 0 {
		g.valueDialog.Text = g.valueDialog.Text[:len(g.valueDialog.Text)-1]
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
	spring, ok := g.springAtPosition(position)
	if !ok {
		return 0, false
	}
	return spring.ID, true
}

func (g *Game) springAtPosition(position sim.Vec2) (sim.Spring, bool) {
	for _, spring := range g.simulation.Springs {
		a, okA := g.simulation.MassByID(spring.MassA)
		b, okB := g.simulation.MassByID(spring.MassB)
		if okA && okB && distanceToSegment(position, a.Position, b.Position) <= 6 {
			return spring, true
		}
	}
	return sim.Spring{}, false
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

// mutate4go-manifest-begin
// {"version":1,"tested_at":"2026-05-19T12:19:36-05:00","module_hash":"b27b87cc8011e5deb1c12b2666f2405d71a3c9db6ec540d3545f8bfe4da8b354","functions":[{"id":"func/Game.openMassValueDialog","name":"Game.openMassValueDialog","line":30,"end_line":46,"hash":"1cd83d04cc55bc04268a3bb1815db6792e431b081c2900c5a10e15a91eda0274"},{"id":"func/Game.openSpringConstantDialogAt","name":"Game.openSpringConstantDialogAt","line":48,"end_line":66,"hash":"5cddcb8c08f260cefe065bf3c7ddea6bbc71368c455a9c845a78c80646b3d7fc"},{"id":"func/Game.tickValueDialog","name":"Game.tickValueDialog","line":68,"end_line":73,"hash":"17cc428cae8a89641a9f7536359e37295c0bb73863f13a15e6472e099e3b2f4a"},{"id":"func/Game.valueDialogCursorVisible","name":"Game.valueDialogCursorVisible","line":75,"end_line":80,"hash":"9f3e225681082c9b1aaa4187cad32987e01dcda3e715c69901f71fcf788db199"},{"id":"func/Game.clickValueDialog","name":"Game.clickValueDialog","line":82,"end_line":96,"hash":"6a6cd0a10816a4ba5065fc57ef741befd9b36354be61b4f32ae8405ac6aca21e"},{"id":"func/Game.appendValueDialogInput","name":"Game.appendValueDialogInput","line":98,"end_line":104,"hash":"b512d053f88630b3f6adfd233fe96aacb82434b63db28506066a835c7870a168"},{"id":"func/Game.deleteValueDialogCharacter","name":"Game.deleteValueDialogCharacter","line":106,"end_line":110,"hash":"04fb2cf2958a9514bb103ef0aab66806183db7efd1812814a39b3fa5b43e214e"},{"id":"func/Game.applyValueDialog","name":"Game.applyValueDialog","line":112,"end_line":121,"hash":"1470ec33108d4ac9d21fb764b9e3a13d6519278089a9e4ccbba9a10ec13578b7"},{"id":"func/Game.setValueDialogFromSlider","name":"Game.setValueDialogFromSlider","line":123,"end_line":128,"hash":"b9591c2222d131b9368978ef5c9111eaea81c982ed3c873ed91589ee867319ef"},{"id":"func/Game.valueDialogFraction","name":"Game.valueDialogFraction","line":130,"end_line":136,"hash":"51f872d0716b555a22a16e3c10711e74902cdd676748a6613a2e5168ceaf0532"},{"id":"func/valueDialogRect","name":"valueDialogRect","line":138,"end_line":142,"hash":"302a55a957839dc502a50df300d3113e769d517c064813924e4c6149bd06d101"},{"id":"func/Game.valueDialogTextRect","name":"Game.valueDialogTextRect","line":144,"end_line":147,"hash":"3570f909e9c7d78dbf0c83c67822c84c818235f6f9d6e131a93cde97f85bfde3"},{"id":"func/Game.valueDialogSliderTrack","name":"Game.valueDialogSliderTrack","line":149,"end_line":152,"hash":"c00d5aa5016af8c21fafce55e78dcdf3fe313322ea8db7438907225381cb81ed"},{"id":"func/Game.valueDialogOKRect","name":"Game.valueDialogOKRect","line":154,"end_line":157,"hash":"4b5942653e3e664fcae0e3e57be322af27e842d3a886b16c7bb7581d45c005f6"},{"id":"func/Game.setSpringConstant","name":"Game.setSpringConstant","line":159,"end_line":168,"hash":"c07640a6e6c0889d505c9ee749a6311fcb4c352ec19b920dbaecc6a20a3c5fc5"},{"id":"func/Game.springAt","name":"Game.springAt","line":170,"end_line":176,"hash":"004f335189f88e1e03b4d55dd33b7c3f196ff07ab7e577c918c3d4758991b10f"},{"id":"func/Game.springAtPosition","name":"Game.springAtPosition","line":178,"end_line":187,"hash":"786bdda02f6d6aecfecfdd9370cb3be2f2c0f02ef22af39cda0ba6f344c2cbef"},{"id":"func/distanceToSegment","name":"distanceToSegment","line":189,"end_line":199,"hash":"780d9bda4d7679a39d2aca892c903e2b8e4232701a027c90768e710967e61f14"}]}
// mutate4go-manifest-end
