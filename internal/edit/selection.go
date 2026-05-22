package edit

import (
	"fmt"
	"math"

	"springs/internal/sim"
)

type DuplicatedObjects struct {
	MassIDs   []int
	SpringIDs []int
}

func (e *Editor) SelectMass(id int) error {
	return e.selectExisting(id, "mass", e.massExists, func() { e.SelectedMasses[id] = true })
}

func (e *Editor) AddMassSelection(id int) error {
	return e.selectExisting(id, "mass", e.massExists, func() { e.SelectedMasses[id] = true }, keepSelection)
}

func (e *Editor) SelectSpring(id int) error {
	return e.selectExisting(id, "spring", e.springExists, func() { e.SelectedSprings[id] = true })
}

func (e *Editor) SelectNearest(position sim.Vec2, toggle bool) error {
	id, ok := e.nearestMassID(position)
	if !ok {
		return fmt.Errorf("no object near pointer")
	}
	if toggle {
		e.toggleMassSelection(id)
		return nil
	}
	e.clearSelection()
	e.SelectedMasses[id] = true
	return nil
}

func (e *Editor) selectExisting(id int, objectType string, exists func(int) bool, selectObject func(), options ...func(*Editor)) error {
	if !exists(id) {
		return fmt.Errorf("%s %d not found", objectType, id)
	}
	if len(options) == 0 {
		e.clearSelection()
	}
	for _, option := range options {
		option(e)
	}
	selectObject()
	return nil
}

func keepSelection(*Editor) {}

func (e *Editor) SelectAll() {
	e.clearSelection()
	for _, mass := range e.World.Masses {
		e.SelectedMasses[mass.ID] = true
	}
	for _, spring := range e.World.Springs {
		e.SelectedSprings[spring.ID] = true
	}
}

func (e *Editor) BoxSelect(min sim.Vec2, max sim.Vec2, add bool) {
	if !add {
		e.clearSelection()
	}
	massesInBox := 0
	for _, mass := range e.World.Masses {
		if withinBox(mass.Position, min, max) {
			e.SelectedMasses[mass.ID] = true
			massesInBox++
		}
	}
	fullyEnclosedSprings := e.selectFullyEnclosedSprings(min, max)
	if massesInBox == 0 && fullyEnclosedSprings == 0 {
		e.selectSinglePartiallyEnclosedSpring(min, max)
	}
}

func (e *Editor) selectFullyEnclosedSprings(min sim.Vec2, max sim.Vec2) int {
	count := 0
	for _, spring := range e.World.Springs {
		a, okA := e.World.MassByID(spring.MassA)
		b, okB := e.World.MassByID(spring.MassB)
		if okA && okB && withinBox(a.Position, min, max) && withinBox(b.Position, min, max) {
			e.SelectedSprings[spring.ID] = true
			count++
		}
	}
	return count
}

func (e *Editor) selectSinglePartiallyEnclosedSpring(min sim.Vec2, max sim.Vec2) {
	selectedID := e.singlePartiallyEnclosedSpringID(min, max)
	if selectedID != 0 {
		e.SelectedSprings[selectedID] = true
	}
}

func (e *Editor) singlePartiallyEnclosedSpringID(min sim.Vec2, max sim.Vec2) int {
	selectedID := 0
	for _, spring := range e.World.Springs {
		a, okA := e.World.MassByID(spring.MassA)
		b, okB := e.World.MassByID(spring.MassB)
		if !okA || !okB || !segmentIntersectsBox(a.Position, b.Position, min, max) {
			continue
		}
		if selectedID != 0 {
			return 0
		}
		selectedID = spring.ID
	}
	return selectedID
}

func (e *Editor) MoveSelected(delta sim.Vec2) {
	for i := range e.World.Masses {
		if e.SelectedMasses[e.World.Masses[i].ID] && !e.World.Masses[i].Fixed {
			e.World.Masses[i].Position = e.World.Masses[i].Position.Add(delta)
		}
	}
}

func (e *Editor) ThrowSelected(velocity sim.Vec2) {
	for i := range e.World.Masses {
		if e.SelectedMasses[e.World.Masses[i].ID] && !e.World.Masses[i].Fixed {
			e.World.Masses[i].Velocity = velocity
		}
	}
}

func (e *Editor) MassSelected(id int) bool {
	return e.SelectedMasses[id]
}

func (e *Editor) SpringSelected(id int) bool {
	return e.SelectedSprings[id]
}

func (e *Editor) DeleteSelected() {
	e.deleteSelectedMasses()
	e.deleteSelectedSprings()
	e.reindexSprings()
	e.clearSelection()
}

func (e *Editor) DuplicateSelected() (DuplicatedObjects, error) {
	duplicated := DuplicatedObjects{}
	massIDs := e.duplicateMasses(&duplicated)
	if err := e.duplicateSprings(massIDs, &duplicated); err != nil {
		return DuplicatedObjects{}, err
	}
	e.clearSelection()
	for _, id := range duplicated.MassIDs {
		e.SelectedMasses[id] = true
	}
	for _, id := range duplicated.SpringIDs {
		e.SelectedSprings[id] = true
	}
	return duplicated, nil
}

func (e *Editor) clearSelection() {
	e.SelectedMasses = map[int]bool{}
	e.SelectedSprings = map[int]bool{}
}

func (e *Editor) ClearSelection() {
	e.clearSelection()
}

func (e *Editor) toggleMassSelection(id int) {
	if e.SelectedMasses[id] {
		delete(e.SelectedMasses, id)
		return
	}
	e.SelectedMasses[id] = true
}

func (e *Editor) deleteSelectedMasses() {
	masses := e.World.Masses[:0]
	for _, mass := range e.World.Masses {
		if !e.SelectedMasses[mass.ID] {
			masses = append(masses, mass)
		}
	}
	e.World.Masses = masses
}

func (e *Editor) deleteSelectedSprings() {
	springs := e.World.Springs[:0]
	for _, spring := range e.World.Springs {
		if e.keepSpring(spring) {
			springs = append(springs, spring)
		}
	}
	e.World.Springs = springs
}

func (e *Editor) keepSpring(spring sim.Spring) bool {
	return !e.SelectedSprings[spring.ID] && !e.SelectedMasses[spring.MassA] && !e.SelectedMasses[spring.MassB]
}

func (e *Editor) duplicateMasses(duplicated *DuplicatedObjects) map[int]int {
	next := nextMassID(e.World)
	massIDs := map[int]int{}
	for _, mass := range e.World.Masses {
		if !e.SelectedMasses[mass.ID] {
			continue
		}
		originalID := mass.ID
		mass.ID = next
		next++
		e.World.Masses = append(e.World.Masses, mass)
		massIDs[originalID] = mass.ID
		duplicated.MassIDs = append(duplicated.MassIDs, mass.ID)
	}
	return massIDs
}

func (e *Editor) duplicateSprings(massIDs map[int]int, duplicated *DuplicatedObjects) error {
	next := nextSpringID(e.World)
	for _, spring := range e.World.Springs {
		if !e.SelectedSprings[spring.ID] {
			continue
		}
		spring.ID = next
		next++
		spring.MassA = replacementID(massIDs, spring.MassA)
		spring.MassB = replacementID(massIDs, spring.MassB)
		if err := e.World.AddSpring(spring); err != nil {
			return err
		}
		duplicated.SpringIDs = append(duplicated.SpringIDs, spring.ID)
	}
	return nil
}

func (e *Editor) reindexSprings() {
	for i := range e.World.Springs {
		a, okA := e.worldIndexByMassID(e.World.Springs[i].MassA)
		b, okB := e.worldIndexByMassID(e.World.Springs[i].MassB)
		if okA && okB {
			e.World.Springs[i].A = a
			e.World.Springs[i].B = b
		}
	}
}

func (e *Editor) massExists(id int) bool {
	return objectExists(func() (sim.Mass, bool) { return e.World.MassByID(id) })
}

func (e *Editor) springExists(id int) bool {
	return objectExists(func() (sim.Spring, bool) { return e.World.SpringByID(id) })
}

func objectExists[T any](lookup func() (T, bool)) bool {
	_, ok := lookup()
	return ok
}

func (e *Editor) worldIndexByMassID(id int) (int, bool) {
	for i, mass := range e.World.Masses {
		if mass.ID == id {
			return i, true
		}
	}
	return 0, false
}

func (e *Editor) nearestMassID(position sim.Vec2) (int, bool) {
	if len(e.World.Masses) == 0 {
		return 0, false
	}
	nearestID := e.World.Masses[0].ID
	nearestDistance := math.MaxFloat64
	for _, mass := range e.World.Masses {
		if d := distance(mass.Position, position); d < nearestDistance {
			nearestID = mass.ID
			nearestDistance = d
		}
	}
	return nearestID, true
}

func replacementID(ids map[int]int, id int) int {
	if replacement, ok := ids[id]; ok {
		return replacement
	}
	return id
}

func withinBox(position sim.Vec2, min sim.Vec2, max sim.Vec2) bool {
	lowX, highX := ordered(min.X, max.X)
	lowY, highY := ordered(min.Y, max.Y)
	return position.X >= lowX && position.X <= highX && position.Y >= lowY && position.Y <= highY
}

func segmentIntersectsBox(a sim.Vec2, b sim.Vec2, min sim.Vec2, max sim.Vec2) bool {
	if segmentFullyWithinBox(a, b, min, max) {
		return true
	}
	lowX, highX := ordered(min.X, max.X)
	lowY, highY := ordered(min.Y, max.Y)
	corners := []sim.Vec2{
		{X: lowX, Y: lowY},
		{X: highX, Y: lowY},
		{X: highX, Y: highY},
		{X: lowX, Y: highY},
	}
	for i := range corners {
		if segmentsIntersect(a, b, corners[i], corners[(i+1)%len(corners)]) {
			return true
		}
	}
	return false
}

func segmentFullyWithinBox(a sim.Vec2, b sim.Vec2, min sim.Vec2, max sim.Vec2) bool {
	if !withinBox(a, min, max) {
		return false
	}
	if !withinBox(b, min, max) {
		return false
	}
	return true
}

func segmentsIntersect(a sim.Vec2, b sim.Vec2, c sim.Vec2, d sim.Vec2) bool {
	o1 := orientation(a, b, c)
	o2 := orientation(a, b, d)
	o3 := orientation(c, d, a)
	o4 := orientation(c, d, b)
	if hasCollinearEndpoint(a, b, c, d, o1, o2, o3, o4) {
		return true
	}
	return oppositeSides(o1, o2) && oppositeSides(o3, o4)
}

func oppositeSides(first, second float64) bool {
	return (first > 0 && second < 0) || (first < 0 && second > 0)
}

func hasCollinearEndpoint(a sim.Vec2, b sim.Vec2, c sim.Vec2, d sim.Vec2, o1, o2, o3, o4 float64) bool {
	return collinearEndpointOnSegment(a, b, c, o1) ||
		collinearEndpointOnSegment(a, b, d, o2) ||
		collinearEndpointOnSegment(c, d, a, o3) ||
		collinearEndpointOnSegment(c, d, b, o4)
}

func collinearEndpointOnSegment(start sim.Vec2, end sim.Vec2, point sim.Vec2, orientation float64) bool {
	return orientation == 0 && onSegment(start, point, end)
}

func orientation(a sim.Vec2, b sim.Vec2, c sim.Vec2) float64 {
	value := (b.X-a.X)*(c.Y-a.Y) - (b.Y-a.Y)*(c.X-a.X)
	if math.Abs(value) < 1e-9 {
		return 0
	}
	return value
}

func onSegment(a sim.Vec2, b sim.Vec2, c sim.Vec2) bool {
	lowX, highX := ordered(a.X, c.X)
	lowY, highY := ordered(a.Y, c.Y)
	return between(b.X, lowX, highX) && between(b.Y, lowY, highY)
}

func between(value float64, low float64, high float64) bool {
	return value >= low && value <= high
}

func ordered(a float64, b float64) (float64, float64) {
	return math.Min(a, b), math.Max(a, b)
}

// mutate4go-manifest-begin
// {"version":1,"tested_at":"2026-05-22T10:50:24-05:00","module_hash":"f5e2587b3990a8dd88ffa09178f51822412dd22bdabd1239cdb423047191c458","functions":[{"id":"func/Editor.SelectMass","name":"Editor.SelectMass","line":15,"end_line":17,"hash":"f51bca3953e32b2084532bfd5e067feb1d7fa5dffb4536c5d540cb09df2cba22"},{"id":"func/Editor.AddMassSelection","name":"Editor.AddMassSelection","line":19,"end_line":21,"hash":"057fcb0f89b625f0157e663e35b83a5a9c21d7274bb8757087a5a4f74f22b44a"},{"id":"func/Editor.SelectSpring","name":"Editor.SelectSpring","line":23,"end_line":25,"hash":"ba721a8c37ccf28f8978e8232d2269d2a4a82163dd6d7bbbb0218f60553749e6"},{"id":"func/Editor.SelectNearest","name":"Editor.SelectNearest","line":27,"end_line":39,"hash":"e9817e0b7719b1e495ba797560cdf5c583c097d58bb43c9fd81f0fa5585c4b6b"},{"id":"func/Editor.selectExisting","name":"Editor.selectExisting","line":41,"end_line":53,"hash":"1ce7b3a6a3ba7c3f696e7337c0cac23e0b98c925f029150905ccfacb0f21b072"},{"id":"func/keepSelection","name":"keepSelection","line":55,"end_line":55,"hash":"4f9d4d498888366efb12efb0c2d22826cf213ef183aede17276594e794010f97"},{"id":"func/Editor.SelectAll","name":"Editor.SelectAll","line":57,"end_line":65,"hash":"e469da45ee05eda24609282d70ddb9ac40abc74a168f95ae21f678ee53b87e50"},{"id":"func/Editor.BoxSelect","name":"Editor.BoxSelect","line":67,"end_line":82,"hash":"ff4c0687f9c76d9148d519da24f363e6031a6a707f00b64bb69bc38b80158423"},{"id":"func/Editor.selectFullyEnclosedSprings","name":"Editor.selectFullyEnclosedSprings","line":84,"end_line":95,"hash":"ee84d12488e0920b7e6c7cca6c3353be208dd703017a76df1a87bdb8939c4b7c"},{"id":"func/Editor.selectSinglePartiallyEnclosedSpring","name":"Editor.selectSinglePartiallyEnclosedSpring","line":97,"end_line":102,"hash":"55973f84d46dc98a09c3808545730b297e2be6aef4cc179b285ef8d67757534e"},{"id":"func/Editor.singlePartiallyEnclosedSpringID","name":"Editor.singlePartiallyEnclosedSpringID","line":104,"end_line":118,"hash":"298cee217d2623ae4060d67f40fa2609d633efa19c41d396ba470c3a34c432ee"},{"id":"func/Editor.MoveSelected","name":"Editor.MoveSelected","line":120,"end_line":126,"hash":"9f842a91e9c859092f6ee56f81335f52a28773d4d84deac7f5e847095c813eb4"},{"id":"func/Editor.ThrowSelected","name":"Editor.ThrowSelected","line":128,"end_line":134,"hash":"172ec31f86008d0a246ea571ac35a5c177b9335021566d686d3270a741cd7e1a"},{"id":"func/Editor.MassSelected","name":"Editor.MassSelected","line":136,"end_line":138,"hash":"48dc4ab21ad0752b5dc8beb607d83c7bb7f44c4dcc78063d106a809d981cd691"},{"id":"func/Editor.SpringSelected","name":"Editor.SpringSelected","line":140,"end_line":142,"hash":"0e24b40e816c5d9545f6e756ee157e159e1d05022d39ca0a97dedf954b9415bc"},{"id":"func/Editor.DeleteSelected","name":"Editor.DeleteSelected","line":144,"end_line":149,"hash":"4afdbe3329aa4c864af3b78f5173102b0deb7425af459f896f90f8ee19c26788"},{"id":"func/Editor.DuplicateSelected","name":"Editor.DuplicateSelected","line":151,"end_line":165,"hash":"24217db691ea2d410994e65604747708996238cda5788c2d52fee660c0295340"},{"id":"func/Editor.clearSelection","name":"Editor.clearSelection","line":167,"end_line":170,"hash":"1416bd63cb2404297dff43fd155f17c1a03721765475f8f6742e71cc173418e6"},{"id":"func/Editor.ClearSelection","name":"Editor.ClearSelection","line":172,"end_line":174,"hash":"72f96325e7288a4121e7de5423dbb9bf5d6b411545bf66f5dd2579b23413ef8e"},{"id":"func/Editor.toggleMassSelection","name":"Editor.toggleMassSelection","line":176,"end_line":182,"hash":"f9e7e7cc913986ab5679eba912e6c48913be371b58abf6184427d80a62484b7a"},{"id":"func/Editor.deleteSelectedMasses","name":"Editor.deleteSelectedMasses","line":184,"end_line":192,"hash":"86532c69c730ba43fffaf0cde7f0e500e96e164112e4f1f6fbcb26339764710a"},{"id":"func/Editor.deleteSelectedSprings","name":"Editor.deleteSelectedSprings","line":194,"end_line":202,"hash":"9eaa765d15f8adb1bf7868c917d9ce53941e0a5381b618a20c35d0bc38a9e5aa"},{"id":"func/Editor.keepSpring","name":"Editor.keepSpring","line":204,"end_line":206,"hash":"75d0c6616ce695613b3b9d39607b5379a8f55099580c42c03311c036362c5277"},{"id":"func/Editor.duplicateMasses","name":"Editor.duplicateMasses","line":208,"end_line":223,"hash":"e2d351253c9ce13eee7b4f9d68b55c54a968fa49e5ba0979af27e5f8b99a260d"},{"id":"func/Editor.duplicateSprings","name":"Editor.duplicateSprings","line":225,"end_line":241,"hash":"ad38fef6213758640d3e4fb4bcc7211f30736d93a3c39a7cc338fc7ea0d39860"},{"id":"func/Editor.reindexSprings","name":"Editor.reindexSprings","line":243,"end_line":252,"hash":"869de77e6e5517526a4dffb9cf8d2b05473cba9a1c58775bea9071edabd31ea4"},{"id":"func/Editor.massExists","name":"Editor.massExists","line":254,"end_line":256,"hash":"25a87dfed1eb28dd6558672aab1869dfc15076d600ae2c8d5cd7fadf8a118d85"},{"id":"func/Editor.springExists","name":"Editor.springExists","line":258,"end_line":260,"hash":"e0fb8d38a26fea0cdae904163c5f6197d30784f43c3a375bc00700f538e3ca17"},{"id":"func/objectExists","name":"objectExists","line":262,"end_line":265,"hash":"884a8598176501b5afc8f641e409be429391787af1643ab35c08111964ac12c2"},{"id":"func/Editor.worldIndexByMassID","name":"Editor.worldIndexByMassID","line":267,"end_line":274,"hash":"9e8dcfa94073bcb0475516de0c3dd317e5bb48767f7de1c8b65f0ddc2b58762a"},{"id":"func/Editor.nearestMassID","name":"Editor.nearestMassID","line":276,"end_line":289,"hash":"1340cf53457185a8d3ebadb76ecc048cf8f288f8ec502455a5ee1b87bb9c01b0"},{"id":"func/replacementID","name":"replacementID","line":291,"end_line":296,"hash":"3e3416fc3d2df3659d9a34287c34f4ececdec6f198d4d4ef40679cb5f54a2141"},{"id":"func/withinBox","name":"withinBox","line":298,"end_line":302,"hash":"4a0bffa7827d629d5d91b46e722a4ad9d94feb7188381f53a423974b839644ff"},{"id":"func/segmentIntersectsBox","name":"segmentIntersectsBox","line":304,"end_line":322,"hash":"3dbf1d4d72cbf90ceaf8032e13250fa58905d9b078da7bbbfe46ba971dce3842"},{"id":"func/segmentFullyWithinBox","name":"segmentFullyWithinBox","line":324,"end_line":332,"hash":"099f6b63f116e791ffc025977f6b441e4f3798088c9380022f800c6a75457ae1"},{"id":"func/segmentsIntersect","name":"segmentsIntersect","line":334,"end_line":343,"hash":"fc4fba67912108b78ebd374aa0a7463d9bf2c592f1ceb765ff7a775de8afac3e"},{"id":"func/oppositeSides","name":"oppositeSides","line":345,"end_line":347,"hash":"f384f9b82e42ba8444c25c75ec5efeb076b15bb6200b5199ddd1bf55757341f7"},{"id":"func/hasCollinearEndpoint","name":"hasCollinearEndpoint","line":349,"end_line":354,"hash":"23fd495aec39e873ee2af29d82970a6414676ff8f365571c4ebcf69269a729dc"},{"id":"func/collinearEndpointOnSegment","name":"collinearEndpointOnSegment","line":356,"end_line":358,"hash":"e9dbd037ef34003af71ca091bfefb61304fd67a1e0763ac7370019ec3e2be84c"},{"id":"func/orientation","name":"orientation","line":360,"end_line":366,"hash":"7cad692c1f546cef6d0b51dccb8334fe39c99b7dbe5264574da86cc939cb7922"},{"id":"func/onSegment","name":"onSegment","line":368,"end_line":372,"hash":"fa1817d5f22c6ce63e32a35ebc5b050a95ddf6dae9c2a5e85a5e09e471842d93"},{"id":"func/between","name":"between","line":374,"end_line":376,"hash":"bc0b3444f09516cb3a31d1a1cc0697a4e3a2ab25079deac65427dd0242bda13e"},{"id":"func/ordered","name":"ordered","line":378,"end_line":380,"hash":"9d8fed3963191743a5744a3eeeaccff19e09988dc0611dc16a077f26647c1321"}]}
// mutate4go-manifest-end
