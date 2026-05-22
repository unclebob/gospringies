package acceptance

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"springs/internal/edit"
	"springs/internal/sim"
)

var editInitialVelocity = sim.Vec2{X: 9, Y: 9}

func activateEditMode(w *world, _ map[string]string) error {
	editor := ensureMouseEditor(w)
	editor.Mode = edit.ModeEdit
	return nil
}

func addObjectNearPointer(w *world, example map[string]string) error {
	id, err := intValue(example, "object_id")
	if err != nil {
		return err
	}
	return ensureEditMass(w, id, sim.Vec2{X: float64(id * 10), Y: 0}, false)
}

func setInitialEditSelection(w *world, example map[string]string) error {
	selection, err := editIDList(example, "initial_selection")
	if err != nil {
		return err
	}
	editor := ensureMouseEditor(w)
	editor.SelectedMasses = map[int]bool{}
	for _, id := range selection {
		if err := ensureEditMass(w, id, sim.Vec2{X: float64(id * 10), Y: 0}, false); err != nil {
			return err
		}
		editor.SelectedMasses[id] = true
	}
	return nil
}

func clickEditObject(w *world, example map[string]string) error {
	id, err := intValue(example, "object_id")
	if err != nil {
		return err
	}
	action, err := stringValue(example, "click_action")
	if err != nil {
		return err
	}
	toggle, err := editClickToggle(action)
	if err != nil {
		return err
	}
	return ensureMouseEditor(w).SelectNearest(editPointerPosition(id), toggle)
}

func editClickToggle(action string) (bool, error) {
	switch action {
	case "left clicks":
		return false, nil
	case "shift left clicks":
		return true, nil
	default:
		return false, fmt.Errorf("unsupported click action %q", action)
	}
}

func addObjectsInsideSelectionBox(w *world, example map[string]string) error {
	return addEditObjects(w, example, "inside_objects", insideSelectionBoxPosition)
}

func addObjectsOutsideSelectionBox(w *world, example map[string]string) error {
	return addEditObjects(w, example, "outside_objects", outsideSelectionBoxPosition)
}

func dragSelectionBox(w *world, example map[string]string) error {
	modifier, err := stringValue(example, "modifier")
	if err != nil {
		return err
	}
	switch modifier {
	case "none":
		ensureMouseEditor(w).BoxSelect(sim.Vec2{}, sim.Vec2{X: 50, Y: 50}, false)
	case "shift":
		ensureMouseEditor(w).BoxSelect(sim.Vec2{}, sim.Vec2{X: 50, Y: 50}, true)
	default:
		return fmt.Errorf("unsupported selection-box modifier %q", modifier)
	}
	return nil
}

func addSelectedObjectAtStart(w *world, example map[string]string) error {
	id, err := intValue(example, "object_id")
	if err != nil {
		return err
	}
	position, err := positionValue(example, "start_position")
	if err != nil {
		return err
	}
	if err := ensureEditMass(w, id, position, false); err != nil {
		return err
	}
	ensureMouseEditor(w).SelectedMasses[id] = true
	return nil
}

func middleDragSelectedObjects(w *world, example map[string]string) error {
	return applyEditVector(w, example, "drag_delta", (*edit.Editor).MoveSelected)
}

func assertEditObjectPosition(w *world, example map[string]string) error {
	id, err := intValue(example, "object_id")
	if err != nil {
		return err
	}
	expected, err := positionValue(example, "expected_position")
	if err != nil {
		return err
	}
	mass, ok := ensureDomainWorld(w).MassByID(id)
	if !ok {
		return fmt.Errorf("mass %d not found", id)
	}
	if mass.Position != expected {
		return fmt.Errorf("mass %d position = %#v, want %#v", id, mass.Position, expected)
	}
	return nil
}

func addSelectedMassWithFixedState(w *world, example map[string]string) error {
	id, err := intValue(example, "mass_id")
	if err != nil {
		return err
	}
	fixed, err := boolValue(example, "fixed")
	if err != nil {
		return err
	}
	if err := ensureEditMass(w, id, sim.Vec2{X: float64(id * 10), Y: 0}, fixed); err != nil {
		return err
	}
	setEditMassVelocity(w, id, editInitialVelocity)
	ensureMouseEditor(w).SelectedMasses[id] = true
	return nil
}

func rightDragSelectedMasses(w *world, example map[string]string) error {
	return applyEditVector(w, example, "release_velocity", (*edit.Editor).ThrowSelected)
}

func assertEditMassVelocity(w *world, example map[string]string) error {
	id, err := intValue(example, "mass_id")
	if err != nil {
		return err
	}
	mass, err := editMassByID(w, id)
	if err != nil {
		return err
	}
	return assertEditMassExpectedVelocity(id, mass, example)
}

func assertEditMassExpectedVelocity(id int, mass sim.Mass, example map[string]string) error {
	expectedText, err := stringValue(example, "expected_velocity")
	if err != nil {
		return err
	}
	if expectedText == "unchanged" {
		return assertEditVelocityUnchanged(id, mass)
	}
	expected, err := positionValue(example, "expected_velocity")
	if err != nil {
		return err
	}
	return assertEditVelocityEquals(id, mass, expected)
}

func assertEditSelection(w *world, example map[string]string) error {
	expected, err := editIDList(example, "expected_selection")
	if err != nil {
		return err
	}
	actual := selectedEditMassIDs(ensureMouseEditor(w))
	if strings.Join(intStrings(actual), ",") != strings.Join(intStrings(expected), ",") {
		return fmt.Errorf("selection = %v, want %v", actual, expected)
	}
	return nil
}

func addEditObjects(w *world, example map[string]string, key string, position func(int) sim.Vec2) error {
	ids, err := editIDList(example, key)
	if err != nil {
		return err
	}
	for index, id := range ids {
		if err := ensureEditMass(w, id, position(index)); err != nil {
			return err
		}
	}
	return nil
}

func editPointerPosition(id int) sim.Vec2 {
	return sim.Vec2{X: float64(id * 10), Y: 0}
}

func insideSelectionBoxPosition(index int) sim.Vec2 {
	return sim.Vec2{X: float64(10 + index*10), Y: 10}
}

func outsideSelectionBoxPosition(index int) sim.Vec2 {
	return sim.Vec2{X: float64(100 + index*10), Y: 100}
}

func applyEditVector(w *world, example map[string]string, key string, action func(*edit.Editor, sim.Vec2)) error {
	vector, err := positionValue(example, key)
	if err != nil {
		return err
	}
	action(ensureMouseEditor(w), vector)
	return nil
}

func ensureEditMass(w *world, id int, position sim.Vec2, fixed ...bool) error {
	if _, ok := ensureDomainWorld(w).MassByID(id); ok {
		return nil
	}
	isFixed := len(fixed) > 0 && fixed[0]
	return ensureDomainWorld(w).AddMass(sim.Mass{ID: id, Position: position, Mass: 1, Fixed: isFixed})
}

func setEditMassVelocity(w *world, id int, velocity sim.Vec2) {
	for index := range ensureDomainWorld(w).Masses {
		if ensureDomainWorld(w).Masses[index].ID == id {
			ensureDomainWorld(w).Masses[index].Velocity = velocity
			return
		}
	}
}

func editMassByID(w *world, id int) (sim.Mass, error) {
	mass, ok := ensureDomainWorld(w).MassByID(id)
	if !ok {
		return sim.Mass{}, fmt.Errorf("mass %d not found", id)
	}
	return mass, nil
}

func editIDList(example map[string]string, key string) ([]int, error) {
	value, err := stringValue(example, key)
	if err != nil {
		return nil, err
	}
	return parseEditIDList(key, value)
}

func parseEditIDList(key string, value string) ([]int, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "none" || trimmed == "" {
		return nil, nil
	}
	parts := strings.Split(trimmed, ",")
	ids := make([]int, 0, len(parts))
	for _, part := range parts {
		id, err := parseEditIDPart(part, key, value)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	sort.Ints(ids)
	return ids, nil
}

func parseEditIDPart(part string, key string, value string) (int, error) {
	id, err := strconv.Atoi(strings.TrimSpace(part))
	if err != nil {
		return 0, fmt.Errorf("invalid id list %s=%q", key, value)
	}
	return id, nil
}

func assertEditVelocityUnchanged(id int, mass sim.Mass) error {
	if mass.Velocity != editInitialVelocity {
		return fmt.Errorf("mass %d velocity changed to %#v", id, mass.Velocity)
	}
	return nil
}

func assertEditVelocityEquals(id int, mass sim.Mass, expected sim.Vec2) error {
	if mass.Velocity != expected {
		return fmt.Errorf("mass %d velocity = %#v, want %#v", id, mass.Velocity, expected)
	}
	return nil
}

func selectedEditMassIDs(editor *edit.Editor) []int {
	ids := make([]int, 0, len(editor.SelectedMasses))
	for id, selected := range editor.SelectedMasses {
		if selected {
			ids = append(ids, id)
		}
	}
	sort.Ints(ids)
	return ids
}

func intStrings(ids []int) []string {
	values := make([]string, len(ids))
	for i, id := range ids {
		values[i] = strconv.Itoa(id)
	}
	return values
}

// mutate4go-manifest-begin
// {"version":1,"tested_at":"2026-05-22T10:58:15-05:00","module_hash":"d9009528f9d0273ecb2ecc51b3e6729a609be097ea464c14d876f6a9b80487f5","functions":[{"id":"func/activateEditMode","name":"activateEditMode","line":15,"end_line":19,"hash":"db115eeaaa0ef208a217fc380d77553c9d3c02ce600dd47bdf9897ceb4ba4c73"},{"id":"func/addObjectNearPointer","name":"addObjectNearPointer","line":21,"end_line":27,"hash":"82e406c4998fa1e88a3d8906d7b52f703ed4359a4fd29a8d330f4df6208faf01"},{"id":"func/setInitialEditSelection","name":"setInitialEditSelection","line":29,"end_line":43,"hash":"c95d68c5414e0d7287735c8436d71d063b06e657df5d53e00eb3ba31e8a2a32f"},{"id":"func/clickEditObject","name":"clickEditObject","line":45,"end_line":59,"hash":"bcf02c6ec99c3b5bb9cb23fca1cf9265ac3597125884fd74d250e94865075279"},{"id":"func/editClickToggle","name":"editClickToggle","line":61,"end_line":70,"hash":"50922bc1c6f8ca13a54c4b09f9062980b3195cd4f6655ee15c79401049489598"},{"id":"func/addObjectsInsideSelectionBox","name":"addObjectsInsideSelectionBox","line":72,"end_line":74,"hash":"08f54ea4cb1a83c3b21863f36b6390e4ff5a761a3dfd89c5c3856b4b004e0a66"},{"id":"func/addObjectsOutsideSelectionBox","name":"addObjectsOutsideSelectionBox","line":76,"end_line":78,"hash":"8c5da9bb68f32e770c63dbb57047ad27d56a3c97c51e7f7d6c764f0aaf1e2dcf"},{"id":"func/dragSelectionBox","name":"dragSelectionBox","line":80,"end_line":94,"hash":"6b11d595522416fd31a2b423b3c606642efdb34b05e8da382a7f4a17da04dfcc"},{"id":"func/addSelectedObjectAtStart","name":"addSelectedObjectAtStart","line":96,"end_line":110,"hash":"240e381f2a565299848695f4cf9f306274eaf612f709b898ef0f2c5c61baf8b3"},{"id":"func/middleDragSelectedObjects","name":"middleDragSelectedObjects","line":112,"end_line":114,"hash":"31a8401fd62eef226edafe552f67adcaa4c589d151240f0a10b25ba131619c22"},{"id":"func/assertEditObjectPosition","name":"assertEditObjectPosition","line":116,"end_line":133,"hash":"e435f6570a4e226c3bc7b72241295f1d1b17837eec5cdc0ef70a72d3393cf600"},{"id":"func/addSelectedMassWithFixedState","name":"addSelectedMassWithFixedState","line":135,"end_line":150,"hash":"862ab0d0bbd839d5b3bcde77d6736f0a3f16dbea23380e83965463a7c3596b02"},{"id":"func/rightDragSelectedMasses","name":"rightDragSelectedMasses","line":152,"end_line":154,"hash":"74499c7166009a3bc896ff36b1e00841a968069679ee02778aeb5db77d9ff085"},{"id":"func/assertEditMassVelocity","name":"assertEditMassVelocity","line":156,"end_line":166,"hash":"2a33b37fab03395d1b332f60c662d8207358c8b4ffedc1b96dc597c0956266e9"},{"id":"func/assertEditMassExpectedVelocity","name":"assertEditMassExpectedVelocity","line":168,"end_line":181,"hash":"17f2ee6a540747a102c94307c39a4b2bf088a506419cf0d11c0aa46ad6f77916"},{"id":"func/assertEditSelection","name":"assertEditSelection","line":183,"end_line":193,"hash":"f9237e715fbaf6b1e6f5440d93b9685c91750bf62bb000358a46837517daefe2"},{"id":"func/addEditObjects","name":"addEditObjects","line":195,"end_line":206,"hash":"237ad4f7ef439b2b15b0d58e0da6f49e7a0137be272473425c27296a269086a9"},{"id":"func/editPointerPosition","name":"editPointerPosition","line":208,"end_line":210,"hash":"5a465e97342738f714c39fccfc871b6a1b12a2044959425f85425d3433492818"},{"id":"func/insideSelectionBoxPosition","name":"insideSelectionBoxPosition","line":212,"end_line":214,"hash":"1f338f4a08ea7b29a7e34b5ba11dbac412a5b956ece7b692f11425b044c12da7"},{"id":"func/outsideSelectionBoxPosition","name":"outsideSelectionBoxPosition","line":216,"end_line":218,"hash":"b583e625e92bdc72a0480a8348b68bc99281192299603d4dd49f5f4137125461"},{"id":"func/applyEditVector","name":"applyEditVector","line":220,"end_line":227,"hash":"cbd8d1aa3c8ef14af1fb569a38fed3395acb974c092e7052da9ba9322e666ab3"},{"id":"func/ensureEditMass","name":"ensureEditMass","line":229,"end_line":235,"hash":"442e7f2e9bcb1466b004c72571a38c6e85e734e471093eccd7d51d76e79ce408"},{"id":"func/setEditMassVelocity","name":"setEditMassVelocity","line":237,"end_line":244,"hash":"ca0d56164894385f37863e120cc4e5e69e4f471f82ee72ccf8e408fa792d4dc9"},{"id":"func/editMassByID","name":"editMassByID","line":246,"end_line":252,"hash":"ac5b5f235e73b5e2470c2678c2198fd63115772a82bbc8d627cfb87a2a6a82ad"},{"id":"func/editIDList","name":"editIDList","line":254,"end_line":260,"hash":"f603504c548ba7083e43c87c46cd97fccf0b72ba8093ccf93717ca81e106eee2"},{"id":"func/parseEditIDList","name":"parseEditIDList","line":262,"end_line":278,"hash":"c01685a576bd592177d758f5adeda2bd1a3f44d8660444398e36c42654249853"},{"id":"func/parseEditIDPart","name":"parseEditIDPart","line":280,"end_line":286,"hash":"cf4fbcf4a9c44fa9bf654e1680fc9411e73ee7698ce426d1edbdce5a9ada967e"},{"id":"func/assertEditVelocityUnchanged","name":"assertEditVelocityUnchanged","line":288,"end_line":293,"hash":"09976ba30c900d0323c072f33c2d3ddf1eb41e14f70c010d0e281cf99f28fcfb"},{"id":"func/assertEditVelocityEquals","name":"assertEditVelocityEquals","line":295,"end_line":300,"hash":"ac35195d3c6c78012fac542a1946429a8a216f6c9fefae862d43c7ffe8ae4f86"},{"id":"func/selectedEditMassIDs","name":"selectedEditMassIDs","line":302,"end_line":311,"hash":"94b09cbb5eaf5c46b6860b46d8df50d2fdb1d387b36ec781042e34ab29d29524"},{"id":"func/intStrings","name":"intStrings","line":313,"end_line":319,"hash":"2cb5e22e242b2fa305d956b2dd5f925f9e653e672461998051d002c1669d3ab1"}]}
// mutate4go-manifest-end
