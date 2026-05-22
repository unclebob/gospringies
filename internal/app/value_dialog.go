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

type springValueKind string

const (
	springValueKspring     springValueKind = "Kspring"
	springValueKdamp       springValueKind = "Kdamp"
	springValueRestLen     springValueKind = "RestLen"
	springValueTemperature springValueKind = "Temperature"
)

func (g *Game) openMassValueDialog(id int) {
	mass, ok := g.world.simulation.MassByID(id)
	if !ok {
		return
	}
	g.overlays.value = valueDialog{
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
	g.openSpringValueDialog(spring.ID, springValueKspring)
	return true
}

func (g *Game) openSpringValueDialog(id int, kind springValueKind) {
	spring, ok := g.world.simulation.SpringByID(id)
	if !ok {
		return
	}
	text, max, apply := g.springValueDialogSpec(id, spring, kind)
	g.overlays.value = valueDialog{
		Open:   true,
		Title:  fmt.Sprintf("%s Spring #%d", kind, id),
		Text:   text,
		Min:    0,
		Max:    max,
		Target: "spring",
		Apply:  apply,
	}
}

func (g *Game) springValueDialogSpec(id int, spring sim.Spring, kind springValueKind) (string, float64, func(float64)) {
	switch kind {
	case springValueKdamp:
		return formatControlFloat(spring.Damping), 1000, func(value float64) {
			g.setSpringDamping(id, value)
		}
	case springValueRestLen:
		return formatControlFloat(spring.RestLength), 1000, func(value float64) {
			g.setSpringRestLength(id, value)
		}
	case springValueTemperature:
		return formatControlFloat(spring.Temperature), 10, func(value float64) {
			g.setSpringTemperature(id, value)
		}
	default:
		return formatControlFloat(spring.SpringConstant), 1000, func(value float64) {
			g.setSpringConstant(id, value)
		}
	}
}

func (g *Game) tickValueDialog() {
	if !g.overlays.value.Open {
		return
	}
	g.overlays.value.Ticks++
}

func (g *Game) valueDialogCursorVisible() bool {
	if !g.overlays.value.Open {
		return false
	}
	return (g.overlays.value.Ticks/valueCursorPeriod)%2 == 0
}

func (g *Game) SpringTemperatureDialogRange() (float64, float64, bool) {
	return g.overlays.value.Min, g.overlays.value.Max, g.overlays.value.Open && strings.HasPrefix(g.overlays.value.Title, string(springValueTemperature)+" Spring #")
}

func (g *Game) ApplyValueDialogText(text string) bool {
	if !g.overlays.value.Open {
		return false
	}
	g.overlays.value.Text = text
	g.applyValueDialog()
	return true
}

func (g *Game) clickValueDialog(x int, y int) {
	point := image.Pt(x, y)
	if !point.In(valueDialogRect()) {
		g.overlays.value.Open = false
		return
	}
	if point.In(g.valueDialogDecrementRect()) {
		g.controls.activeValueStep = -numericStepAmount
		g.controls.valueStepTicks = 0
		g.stepValueDialog(g.controls.activeValueStep)
		return
	}
	if point.In(g.valueDialogIncrementRect()) {
		g.controls.activeValueStep = numericStepAmount
		g.controls.valueStepTicks = 0
		g.stepValueDialog(g.controls.activeValueStep)
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
			g.overlays.value.Text += string(char)
		}
	}
}

func (g *Game) deleteValueDialogCharacter() {
	if len(g.overlays.value.Text) > 0 {
		g.overlays.value.Text = g.overlays.value.Text[:len(g.overlays.value.Text)-1]
	}
}

func (g *Game) applyValueDialog() {
	value, err := strconv.ParseFloat(g.overlays.value.Text, 64)
	if err != nil {
		return
	}
	if g.overlays.value.Apply != nil {
		g.overlays.value.Apply(value)
	}
	g.overlays.value.Open = false
}

func (g *Game) setValueDialogFromSlider(x int) {
	track := g.valueDialogSliderTrack()
	fraction := clampFloat(float64(x-track.Min.X)/float64(track.Dx()), 0, 1)
	value := g.overlays.value.Min + fraction*(g.overlays.value.Max-g.overlays.value.Min)
	g.overlays.value.Text = formatControlFloat(value)
}

func (g *Game) stepValueDialog(delta float64) {
	value, err := strconv.ParseFloat(g.overlays.value.Text, 64)
	if err != nil {
		value = g.overlays.value.Min
	}
	value = clampFloat(value+delta, g.overlays.value.Min, g.overlays.value.Max)
	g.overlays.value.Text = formatControlFloat(roundControlFloat(value))
}

func (g *Game) continueValueDialogStepHold() {
	if !g.overlays.value.Open || g.controls.activeValueStep == 0 {
		g.controls.activeValueStep = 0
		g.controls.valueStepTicks = 0
		return
	}
	g.controls.valueStepTicks++
	if g.controls.valueStepTicks < numericStepHoldDelayTicks {
		return
	}
	if (g.controls.valueStepTicks-numericStepHoldDelayTicks)%numericStepRepeatTicks == 0 {
		g.stepValueDialog(g.controls.activeValueStep)
	}
}

func (g *Game) valueDialogFraction() float64 {
	value, err := strconv.ParseFloat(g.overlays.value.Text, 64)
	if err != nil || g.overlays.value.Max <= g.overlays.value.Min {
		return 0
	}
	return clampFloat((value-g.overlays.value.Min)/(g.overlays.value.Max-g.overlays.value.Min), 0, 1)
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
	return image.Rect(rect.Min.X+12+numericStepButtonWidth+numericStepButtonGap, rect.Min.Y+92, rect.Max.X-12-numericStepButtonWidth-numericStepButtonGap, rect.Min.Y+100)
}

func (g *Game) valueDialogDecrementRect() image.Rectangle {
	rect := valueDialogRect()
	return image.Rect(rect.Min.X+12, rect.Min.Y+86, rect.Min.X+12+numericStepButtonWidth, rect.Min.Y+106)
}

func (g *Game) valueDialogIncrementRect() image.Rectangle {
	rect := valueDialogRect()
	return image.Rect(rect.Max.X-12-numericStepButtonWidth, rect.Min.Y+86, rect.Max.X-12, rect.Min.Y+106)
}

func (g *Game) valueDialogOKRect() image.Rectangle {
	rect := valueDialogRect()
	return image.Rect(rect.Max.X-64, rect.Max.Y-34, rect.Max.X-12, rect.Max.Y-12)
}

func (g *Game) setSpringConstant(id int, value float64) {
	g.setSpringFloat(id, springFloatConstant, value)
}

func (g *Game) setSpringDamping(id int, value float64) {
	g.setSpringFloat(id, springFloatDamping, value)
}

func (g *Game) setSpringRestLength(id int, value float64) {
	g.setSpringFloat(id, springFloatRestLength, value)
}

func (g *Game) setSpringTemperature(id int, value float64) {
	g.setSpringFloat(id, springFloatTemperature, value)
}

type springFloatField string

const (
	springFloatConstant    springFloatField = "constant"
	springFloatDamping     springFloatField = "damping"
	springFloatRestLength  springFloatField = "rest length"
	springFloatTemperature springFloatField = "temperature"
)

func (g *Game) setSpringFloat(id int, field springFloatField, value float64) {
	g.updateSpring(id, func(spring *sim.Spring) {
		applySpringFloat(spring, field, value)
	})
}

func applySpringFloat(spring *sim.Spring, field springFloatField, value float64) {
	switch field {
	case springFloatConstant:
		spring.SpringConstant = value
		spring.Stiffness = value
	case springFloatDamping:
		spring.Damping = value
	case springFloatRestLength:
		spring.RestLength = value
	case springFloatTemperature:
		spring.Temperature = value
	}
}

func (g *Game) updateSpring(id int, update func(*sim.Spring)) {
	for i := range g.world.simulation.Springs {
		if g.world.simulation.Springs[i].ID == id {
			update(&g.world.simulation.Springs[i])
			g.editState.dirty = true
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
	for _, spring := range g.world.simulation.Springs {
		a, okA := g.world.simulation.MassByID(spring.MassA)
		b, okB := g.world.simulation.MassByID(spring.MassB)
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
// {"version":1,"tested_at":"2026-05-22T08:41:56-05:00","module_hash":"18f6b08e0551cc3d568a1c665f44db9fdcec8eb7be9b8b07bac462486b5e93d5","functions":[{"id":"func/Game.openMassValueDialog","name":"Game.openMassValueDialog","line":39,"end_line":55,"hash":"b31882475c234d5b8c385bea972fcf0094176fa697d83196ea1b2e16eda410d6"},{"id":"func/Game.openSpringConstantDialogAt","name":"Game.openSpringConstantDialogAt","line":57,"end_line":64,"hash":"a2529f506ecb0526db639013d768e776db8ba2c07d094c49155c27c933376036"},{"id":"func/Game.openSpringValueDialog","name":"Game.openSpringValueDialog","line":66,"end_line":81,"hash":"e0a24c22c761141b0bb2a2962998be15e0f7138f2408835476cda9ff6111597f"},{"id":"func/Game.springValueDialogSpec","name":"Game.springValueDialogSpec","line":83,"end_line":102,"hash":"1566a87d2734c74f6eac4ed3f9d0e7e98f609aae72c1bb9b165e5634434eaf17"},{"id":"func/Game.tickValueDialog","name":"Game.tickValueDialog","line":104,"end_line":109,"hash":"5ecbea692c676e4a443763d653034b72eb5d6f1ffb119fa644eebd7489861210"},{"id":"func/Game.valueDialogCursorVisible","name":"Game.valueDialogCursorVisible","line":111,"end_line":116,"hash":"0f2ea255701e85dc6a797adea23d098bc98d95d0fa7ad9d07f0469316750b250"},{"id":"func/Game.SpringTemperatureDialogRange","name":"Game.SpringTemperatureDialogRange","line":118,"end_line":120,"hash":"1f95acde50530f897c7ea50cf9b62020d959bc1dfebecdf8133c15f3297f03c3"},{"id":"func/Game.ApplyValueDialogText","name":"Game.ApplyValueDialogText","line":122,"end_line":129,"hash":"1da2989ed54bc60ed3e05627c4d8f2a206ae643d23198cff8fddfb3d295d36ff"},{"id":"func/Game.clickValueDialog","name":"Game.clickValueDialog","line":131,"end_line":157,"hash":"8c7628b9c7c3d77622505051cb27b5682fa332af08f21bf384b844006cb61fd7"},{"id":"func/Game.appendValueDialogInput","name":"Game.appendValueDialogInput","line":159,"end_line":165,"hash":"d3a6a3bb79c378b6a3bda2a6f2c2282c3b8c47404c0465dc81d443430b57a645"},{"id":"func/Game.deleteValueDialogCharacter","name":"Game.deleteValueDialogCharacter","line":167,"end_line":171,"hash":"ded1f6d046f1a2d64a3e1c661062f34ac7dce12f6b358c8c0b450fdcce36916c"},{"id":"func/Game.applyValueDialog","name":"Game.applyValueDialog","line":173,"end_line":182,"hash":"beb33c06a29f37fb0224a4ae148c48135a647e29fe1ddd204ea05fd9b14e59ea"},{"id":"func/Game.setValueDialogFromSlider","name":"Game.setValueDialogFromSlider","line":184,"end_line":189,"hash":"723ff44ec8410fc15752cf10cc775d2875c36a001da934b9ebd0d6c7dd724094"},{"id":"func/Game.stepValueDialog","name":"Game.stepValueDialog","line":191,"end_line":198,"hash":"2c15259417884e41b86b0b001a9a11cb8bd70eb8da713d6b371f840dbd4b60cb"},{"id":"func/Game.continueValueDialogStepHold","name":"Game.continueValueDialogStepHold","line":200,"end_line":213,"hash":"4fbe423b2e19061fade444c593c1ca44382c4c59fc72c08a3efc049127e48dd7"},{"id":"func/Game.valueDialogFraction","name":"Game.valueDialogFraction","line":215,"end_line":221,"hash":"12defd493b8a76e35f13a7bab84ae100a5c585080a226578fed958817827a021"},{"id":"func/valueDialogRect","name":"valueDialogRect","line":223,"end_line":227,"hash":"302a55a957839dc502a50df300d3113e769d517c064813924e4c6149bd06d101"},{"id":"func/Game.valueDialogTextRect","name":"Game.valueDialogTextRect","line":229,"end_line":232,"hash":"3570f909e9c7d78dbf0c83c67822c84c818235f6f9d6e131a93cde97f85bfde3"},{"id":"func/Game.valueDialogSliderTrack","name":"Game.valueDialogSliderTrack","line":234,"end_line":237,"hash":"85ff1a4ab0451139dcc102cd3eed6b93f4fe5c99f26cc6b9fd09b2b83ad3c6df"},{"id":"func/Game.valueDialogDecrementRect","name":"Game.valueDialogDecrementRect","line":239,"end_line":242,"hash":"6684c6b564605bebfe396f610eebcd9babe25bc51a5a783a23e103e31ac843ab"},{"id":"func/Game.valueDialogIncrementRect","name":"Game.valueDialogIncrementRect","line":244,"end_line":247,"hash":"7aa7ef75d21c9d75c624d0f32dc4a03daa630771708d6d51b1721b108e0d813e"},{"id":"func/Game.valueDialogOKRect","name":"Game.valueDialogOKRect","line":249,"end_line":252,"hash":"4b5942653e3e664fcae0e3e57be322af27e842d3a886b16c7bb7581d45c005f6"},{"id":"func/Game.setSpringConstant","name":"Game.setSpringConstant","line":254,"end_line":256,"hash":"57518721b0699d29dc79d46dbf8d4af3c02a23777998c4222df6b3f12f981f14"},{"id":"func/Game.setSpringDamping","name":"Game.setSpringDamping","line":258,"end_line":260,"hash":"17a15671c2f21ed06d4300167c286b764d4a447403e071d7594de99023cdf820"},{"id":"func/Game.setSpringRestLength","name":"Game.setSpringRestLength","line":262,"end_line":264,"hash":"1a9e647e66676753d4b2a14024047e79cb96d01f904b7b49a21d17a01e350cfe"},{"id":"func/Game.setSpringTemperature","name":"Game.setSpringTemperature","line":266,"end_line":268,"hash":"7f3d0e1edae2e20ca223549d99831465fa90242564d2407bb3f55e9f9bb3f81d"},{"id":"func/Game.setSpringFloat","name":"Game.setSpringFloat","line":279,"end_line":283,"hash":"eb3b0fea93313eab3bd3cb66caffa0dd68fe54d1078b273b593ae1eb88276373"},{"id":"func/applySpringFloat","name":"applySpringFloat","line":285,"end_line":297,"hash":"8657afeeec76fa27b9fca45c7712781545b34f88639bebae7869185fde86bba4"},{"id":"func/Game.updateSpring","name":"Game.updateSpring","line":299,"end_line":307,"hash":"171660b01f5fbaadac33659c2b190db9b7ccbaa870a5b47c75efcd91d60f70bf"},{"id":"func/Game.springAt","name":"Game.springAt","line":309,"end_line":315,"hash":"004f335189f88e1e03b4d55dd33b7c3f196ff07ab7e577c918c3d4758991b10f"},{"id":"func/Game.springAtPosition","name":"Game.springAtPosition","line":317,"end_line":326,"hash":"f84bd1387b0fc3e0bba5f1ec438e776c4077320f6d82e2dc2214e634b7c508d7"},{"id":"func/distanceToSegment","name":"distanceToSegment","line":328,"end_line":338,"hash":"780d9bda4d7679a39d2aca892c903e2b8e4232701a027c90768e710967e61f14"}]}
// mutate4go-manifest-end
