package acceptance

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"springs/internal/sim"
)

func createDomainWorld(w *world, _ map[string]string) error {
	return setSimulation(&w.domainWorld, sim.NewWorld())
}

func assertDomainMassCount(w *world, example map[string]string) error {
	return assertDomainCount(w, example, "masses", "mass_count", massCount)
}

func assertDomainSpringCount(w *world, example map[string]string) error {
	return assertDomainCount(w, example, "springs", "spring_count", springCount)
}

func assertDomainCount(w *world, example map[string]string, name, key string, count func(*sim.Simulation) int) error {
	world, err := domainWorld(w)
	if err != nil {
		return err
	}
	expected, err := intValue(example, key)
	if err != nil {
		return err
	}
	return assertCount(name, count(world), expected)
}

func assertCount(name string, got, expected int) error {
	if got != expected {
		return fmt.Errorf("expected %d %s, got %d", expected, name, got)
	}
	return nil
}

func massCount(world *sim.Simulation) int {
	return len(world.Masses)
}

func springCount(world *sim.Simulation) int {
	return len(world.Springs)
}

func addDomainMass(w *world, example map[string]string) error {
	return addDomainMassFromKeys(w, example, "id", "x", "y")
}

func addDomainMassA(w *world, example map[string]string) error {
	return addDomainMassFromKeys(w, example, "mass_a", "x_a", "y_a")
}

func addDomainMassB(w *world, example map[string]string) error {
	return addDomainMassFromKeys(w, example, "mass_b", "x_b", "y_b")
}

func addExistingDomainMass(w *world, example map[string]string) error {
	return addDomainMassFromKeys(w, example, "existing_mass", "x", "y")
}

func addDomainMassFromKeys(w *world, example map[string]string, idKey, xKey, yKey string) error {
	world := ensureDomainWorld(w)
	id, x, y, err := massFields(example, idKey, xKey, yKey)
	if err != nil {
		return err
	}
	if _, ok := world.MassByID(id); ok {
		return nil
	}
	return world.AddMass(sim.Mass{ID: id, Position: sim.Vec2{X: x, Y: y}, Mass: 1})
}

func setDomainMassVelocity(w *world, example map[string]string) error {
	return updateMass(w, example, func(mass *sim.Mass) error {
		velocity, err := vecFromExample(example, "vx", "vy")
		if err != nil {
			return err
		}
		mass.Velocity = velocity
		return nil
	})
}

func setDomainMassValue(w *world, example map[string]string) error {
	return updateMassFloat(w, example, "mass_value", setMassValue)
}

func setDomainMassElasticity(w *world, example map[string]string) error {
	return updateMassFloat(w, example, "elasticity", setMassElasticity)
}

func setDomainMassFixed(w *world, example map[string]string) error {
	return updateMass(w, example, func(mass *sim.Mass) error {
		value, err := boolValue(example, "fixed")
		if err != nil {
			return err
		}
		mass.Fixed = value
		return nil
	})
}

func updateMassFloat(w *world, example map[string]string, key string, assign func(*sim.Mass, float64)) error {
	return updateMass(w, example, floatUpdate(example, key, assign))
}

func setMassValue(mass *sim.Mass, value float64) {
	mass.Mass = value
}

func setMassElasticity(mass *sim.Mass, value float64) {
	mass.Elasticity = value
}

func lookupDomainMass(w *world, example map[string]string) error {
	return lookupByExample(w, example, "id", "mass", (*sim.Simulation).MassByID, func(mass sim.Mass) { w.lookedMass = mass })
}

func assertDomainMassPosition(w *world, example map[string]string) error {
	return assertVecExample("position", w.lookedMass.Position, example, "x", "y")
}

func assertDomainMassVelocity(w *world, example map[string]string) error {
	return assertVecExample("velocity", w.lookedMass.Velocity, example, "vx", "vy")
}

func assertVecExample(name string, got sim.Vec2, example map[string]string, xKey, yKey string) error {
	expected, err := vecFromExample(example, xKey, yKey)
	if err != nil {
		return err
	}
	return assertVec(name, got, expected.X, expected.Y)
}

func vecFromExample(example map[string]string, xKey, yKey string) (sim.Vec2, error) {
	x, err := floatValue(example, xKey)
	if err != nil {
		return sim.Vec2{}, err
	}
	y, err := floatValue(example, yKey)
	if err != nil {
		return sim.Vec2{}, err
	}
	return sim.Vec2{X: x, Y: y}, nil
}

func assertDomainMassValue(w *world, example map[string]string) error {
	return assertFloatExample("mass value", w.lookedMass.Mass, example, "mass_value")
}

func assertDomainMassElasticity(w *world, example map[string]string) error {
	return assertFloatExample("elasticity", w.lookedMass.Elasticity, example, "elasticity")
}

func assertFloatExample(name string, got float64, example map[string]string, key string) error {
	expected, err := floatValue(example, key)
	if err != nil {
		return err
	}
	return assertFloat(name, got, expected)
}

func assertDomainMassFixed(w *world, example map[string]string) error {
	expected, err := boolValue(example, "fixed")
	if err != nil {
		return err
	}
	if w.lookedMass.Fixed != expected {
		return fmt.Errorf("expected fixed %t, got %t", expected, w.lookedMass.Fixed)
	}
	return nil
}

func addDomainSpring(w *world, example map[string]string) error {
	world := ensureDomainWorld(w)
	spring, err := springFromExample(example)
	if err != nil {
		return err
	}
	if _, ok := world.SpringByID(spring.ID); ok {
		return nil
	}
	return world.AddSpring(spring)
}

func setDomainSpringConstant(w *world, example map[string]string) error {
	return updateSpringFloat(w, example, "spring_constant", setSpringConstant)
}

func setDomainSpringDamping(w *world, example map[string]string) error {
	return updateSpringFloat(w, example, "damping_constant", setSpringDamping)
}

func setDomainSpringRestLength(w *world, example map[string]string) error {
	return updateSpringFloat(w, example, "rest_length", setSpringRestLength)
}

func updateSpringFloat(w *world, example map[string]string, key string, assign func(*sim.Spring, float64)) error {
	return updateSpring(w, example, floatUpdate(example, key, assign))
}

func setSpringDamping(spring *sim.Spring, value float64) {
	spring.Damping = value
}

func setSpringConstant(spring *sim.Spring, value float64) {
	spring.SpringConstant = value
	spring.Stiffness = value
}

func setSpringRestLength(spring *sim.Spring, value float64) {
	spring.RestLength = value
}

func lookupDomainSpring(w *world, example map[string]string) error {
	return lookupByExample(w, example, "spring_id", "spring", (*sim.Simulation).SpringByID, func(spring sim.Spring) { w.lookedSpring = spring })
}

func assertDomainSpringEndpoints(w *world, example map[string]string) error {
	massA, err := intValue(example, "mass_a")
	if err != nil {
		return err
	}
	massB, err := intValue(example, "mass_b")
	if err != nil {
		return err
	}
	if w.lookedSpring.MassA != massA || w.lookedSpring.MassB != massB {
		return fmt.Errorf("expected spring endpoints %d,%d got %d,%d", massA, massB, w.lookedSpring.MassA, w.lookedSpring.MassB)
	}
	return nil
}

func assertDomainSpringConstant(w *world, example map[string]string) error {
	return assertFloatExample("spring constant", w.lookedSpring.SpringConstant, example, "spring_constant")
}

func assertDomainSpringDamping(w *world, example map[string]string) error {
	return assertFloatExample("damping constant", w.lookedSpring.Damping, example, "damping_constant")
}

func assertDomainSpringRestLength(w *world, example map[string]string) error {
	return assertFloatExample("rest length", w.lookedSpring.RestLength, example, "rest_length")
}

func addExistingDomainObject(w *world, example map[string]string) error {
	world := ensureDomainWorld(w)
	objectType, id, err := objectTypeAndID(example)
	if err != nil {
		return err
	}
	if objectType == "mass" {
		return world.AddMass(sim.Mass{ID: id, Mass: 1})
	}
	return addExistingSpring(world, id)
}

func addDuplicateDomainObject(w *world, example map[string]string) error {
	world := ensureDomainWorld(w)
	objectType, id, err := objectTypeAndID(example)
	if err != nil {
		return err
	}
	if objectType == "mass" {
		w.validationErr = world.AddMass(sim.Mass{ID: id})
	} else {
		w.validationErr = world.AddSpring(sim.Spring{ID: id, MassA: 1, MassB: 2})
	}
	return nil
}

func objectTypeAndID(example map[string]string) (string, int, error) {
	objectType, err := stringValue(example, "object_type")
	if err != nil {
		return "", 0, err
	}
	id, err := intValue(example, "id")
	if err != nil {
		return "", 0, err
	}
	return objectType, id, nil
}

func addExistingSpring(world *sim.Simulation, id int) error {
	if err := world.AddMass(sim.Mass{ID: 1, Mass: 1}); err != nil {
		return err
	}
	if err := world.AddMass(sim.Mass{ID: 2, Mass: 1}); err != nil {
		return err
	}
	return world.AddSpring(sim.Spring{ID: id, MassA: 1, MassB: 2})
}

func addInvalidDomainSpring(w *world, example map[string]string) error {
	world := ensureDomainWorld(w)
	spring, err := springFromExample(example)
	if err != nil {
		return err
	}
	w.validationErr = world.AddSpring(spring)
	return nil
}

func assertDomainValidationReason(w *world, example map[string]string) error {
	reason, err := stringValue(example, "reason")
	if err != nil {
		return err
	}
	return assertValidationReason(w.validationErr, reason)
}

func ensureDomainWorld(w *world) *sim.Simulation {
	if w.domainWorld == nil {
		w.domainWorld = sim.NewWorld()
	}
	return w.domainWorld
}

func domainWorld(w *world) (*sim.Simulation, error) {
	if w.domainWorld == nil {
		return nil, fmt.Errorf("domain world has not been created")
	}
	return w.domainWorld, nil
}

func updateMass(w *world, example map[string]string, update func(*sim.Mass) error) error {
	return updateDomainByExample(w, example, "id", "mass", masses, massID, update)
}

func updateSpring(w *world, example map[string]string, update func(*sim.Spring) error) error {
	return updateDomainByExample(w, example, "spring_id", "spring", springs, springID, update)
}

func updateDomainByExample[T any](w *world, example map[string]string, key, name string, items func(*sim.Simulation) []T, itemID func(T) int, update func(*T) error) error {
	world, err := domainWorld(w)
	if err != nil {
		return err
	}
	id, err := intValue(example, key)
	if err != nil {
		return err
	}
	return updateByID(items(world), id, name, itemID, update)
}

func masses(world *sim.Simulation) []sim.Mass {
	return world.Masses
}

func springs(world *sim.Simulation) []sim.Spring {
	return world.Springs
}

func massID(mass sim.Mass) int {
	return mass.ID
}

func springID(spring sim.Spring) int {
	return spring.ID
}

func floatUpdate[T any](example map[string]string, key string, assign func(*T, float64)) func(*T) error {
	return func(item *T) error {
		value, err := floatValue(example, key)
		if err != nil {
			return err
		}
		assign(item, value)
		return nil
	}
}

func lookupByExample[T any](w *world, example map[string]string, key, name string, lookup func(*sim.Simulation, int) (T, bool), assign func(T)) error {
	world, err := domainWorld(w)
	if err != nil {
		return err
	}
	id, err := intValue(example, key)
	if err != nil {
		return err
	}
	item, ok := lookup(world, id)
	if !ok {
		return fmt.Errorf("%s %d not found", name, id)
	}
	assign(item)
	return nil
}

func updateByID[T any](items []T, id int, name string, itemID func(T) int, update func(*T) error) error {
	for i := range items {
		if itemID(items[i]) == id {
			return update(&items[i])
		}
	}
	return fmt.Errorf("%s %d not found", name, id)
}

func springFromExample(example map[string]string) (sim.Spring, error) {
	id, err := intValue(example, "spring_id")
	if err != nil {
		return sim.Spring{}, err
	}
	massA, err := intValue(example, "mass_a")
	if err != nil {
		return sim.Spring{}, err
	}
	massB, err := intValue(example, "mass_b")
	if err != nil {
		return sim.Spring{}, err
	}
	return sim.Spring{ID: id, MassA: massA, MassB: massB}, nil
}

func assertValidationReason(err error, reason string) error {
	if err == nil {
		return fmt.Errorf("validation succeeded, expected %s", reason)
	}
	switch strings.TrimSpace(reason) {
	case "duplicate id":
		if errors.Is(err, sim.ErrDuplicateID) {
			return nil
		}
	case "missing spring endpoint":
		if errors.Is(err, sim.ErrMissingSpringEndpoint) {
			return nil
		}
	}
	return fmt.Errorf("expected validation reason %q, got %v", reason, err)
}

func assertVec(name string, got sim.Vec2, expectedX, expectedY float64) error {
	if math.Abs(got.X-expectedX) > 0.000001 || math.Abs(got.Y-expectedY) > 0.000001 {
		return fmt.Errorf("expected %s %f,%f got %f,%f", name, expectedX, expectedY, got.X, got.Y)
	}
	return nil
}

func assertFloat(name string, got, expected float64) error {
	if math.Abs(got-expected) > 0.000001 {
		return fmt.Errorf("expected %s %f got %f", name, expected, got)
	}
	return nil
}

func massFields(example map[string]string, idKey, xKey, yKey string) (int, float64, float64, error) {
	id, err := intValue(example, idKey)
	if err != nil {
		return 0, 0, 0, err
	}
	x, err := floatValue(example, xKey)
	if err != nil {
		return 0, 0, 0, err
	}
	y, err := floatValue(example, yKey)
	if err != nil {
		return 0, 0, 0, err
	}
	return id, x, y, nil
}

// mutate4go-manifest-begin
// {"version":1,"tested_at":"2026-05-19T08:17:07-05:00","module_hash":"2c711f77c94b0dcc3b82e862bfbcfbc13483ab7c5471f59af32841a4a40330e1","functions":[{"id":"func/createDomainWorld","name":"createDomainWorld","line":12,"end_line":14,"hash":"8f33e51ba5f682bdc5b8b5d6a8e0dbfc8cc1ef1785f087ded7dea922cce981fb"},{"id":"func/assertDomainMassCount","name":"assertDomainMassCount","line":16,"end_line":18,"hash":"2fb2b2dc0359789482abe53924195118fe1aaad753dbc47eebac1c17dc0eb40e"},{"id":"func/assertDomainSpringCount","name":"assertDomainSpringCount","line":20,"end_line":22,"hash":"91f0cff79adbba2a67ac18f6f5bbfd004cac61240f0ca1899802add92bffc9a4"},{"id":"func/assertDomainCount","name":"assertDomainCount","line":24,"end_line":34,"hash":"855260f108997ae570d47620aae1b4882ca15b9569b9560c78c30a2fe92d224d"},{"id":"func/assertCount","name":"assertCount","line":36,"end_line":41,"hash":"52706b633ba01a53820631e91cdf6baa276c0fa231ef24349ba3851b0b2d4c75"},{"id":"func/massCount","name":"massCount","line":43,"end_line":45,"hash":"cba107267a3db2fb08a215465fb758d63edd221ccbbc651183fc71d3deefa13d"},{"id":"func/springCount","name":"springCount","line":47,"end_line":49,"hash":"9ba8fad7ca26a5792142490b7c575277a17515ff23afac2006e63cd5d3cbeb46"},{"id":"func/addDomainMass","name":"addDomainMass","line":51,"end_line":53,"hash":"7f02dea7e421fdac66d31eda8660cf7134e2c1f0033addfdaa070445419ccecb"},{"id":"func/addDomainMassA","name":"addDomainMassA","line":55,"end_line":57,"hash":"a07205025ed9ccf60ae0a886fe685f95bc8421666885312b01b12b02e03cf160"},{"id":"func/addDomainMassB","name":"addDomainMassB","line":59,"end_line":61,"hash":"459fe7026d228a83671536a71d8606723d88dca69f9873895ee54ba95f0d2602"},{"id":"func/addExistingDomainMass","name":"addExistingDomainMass","line":63,"end_line":65,"hash":"65d285175e188eaf8138f47964184e276d8307051b7f7244893d656831927bd8"},{"id":"func/addDomainMassFromKeys","name":"addDomainMassFromKeys","line":67,"end_line":77,"hash":"6c9adcaa24f45bb412c9d8d8e4552f8a8d7657a78addd77481027ba3415af597"},{"id":"func/setDomainMassVelocity","name":"setDomainMassVelocity","line":79,"end_line":88,"hash":"99b6ba84484865a579ecdeadfda004cbcac35e3017660943957b4629ea6555db"},{"id":"func/setDomainMassValue","name":"setDomainMassValue","line":90,"end_line":92,"hash":"444d0439034e571e9c679b88f4a091efb5efaecabba25ec84bfc9484106ac5d3"},{"id":"func/setDomainMassElasticity","name":"setDomainMassElasticity","line":94,"end_line":96,"hash":"82e45c0c2fe95755c2d285623150c3c39cb83f76bcb2197075036a38c2b53f52"},{"id":"func/setDomainMassFixed","name":"setDomainMassFixed","line":98,"end_line":107,"hash":"0f443ff4cc337704bc16cc46fe5ed573e87cdf9a895dffce26ea3f5948d6d3ee"},{"id":"func/updateMassFloat","name":"updateMassFloat","line":109,"end_line":111,"hash":"eae17e432cc939b54c2035a03f359184137cd028dc397d0a12a45159c9502d12"},{"id":"func/setMassValue","name":"setMassValue","line":113,"end_line":115,"hash":"665efc1870363f0cca42b9003f7e604fbfa46b6abedb7e089db7b53d9b017edb"},{"id":"func/setMassElasticity","name":"setMassElasticity","line":117,"end_line":119,"hash":"8d6ac8ab50e0aee16f4769ff9313264aaf6cfe3ada6d868824aa0872f3ca980d"},{"id":"func/lookupDomainMass","name":"lookupDomainMass","line":121,"end_line":123,"hash":"34ae16da411dce60cc7bfa507fa648dd30833e3461b34086d486a81dc2ded6db"},{"id":"func/assertDomainMassPosition","name":"assertDomainMassPosition","line":125,"end_line":127,"hash":"25edf1f43564108c5377228e2d861efa3d17f3b47e722a3e960a1151f2e0a89a"},{"id":"func/assertDomainMassVelocity","name":"assertDomainMassVelocity","line":129,"end_line":131,"hash":"d6b808b60cc32c97d1ee186b6a89cc87985e205b1f7978e3d0664696bc392e47"},{"id":"func/assertVecExample","name":"assertVecExample","line":133,"end_line":139,"hash":"62ffecd3ad8df47cf98fc8844617aa41d09c5fe0eed49e5d4d91b30e74fce6ea"},{"id":"func/vecFromExample","name":"vecFromExample","line":141,"end_line":151,"hash":"3c22e3993b99b42c5158fabb2a8cbfd8832acb077376176303bc1b4ae2907b69"},{"id":"func/assertDomainMassValue","name":"assertDomainMassValue","line":153,"end_line":155,"hash":"5c1abd41af069c45ee275a0acf1ad1850109dea28fae4c880385c4b288376aa3"},{"id":"func/assertDomainMassElasticity","name":"assertDomainMassElasticity","line":157,"end_line":159,"hash":"2011c5addcd02418eebdf24f42b358dbe4e20f79ff11f229a44a15efc9e56e41"},{"id":"func/assertFloatExample","name":"assertFloatExample","line":161,"end_line":167,"hash":"f13c209d9c109d8b16eaf831f63142e8b2b2170ae31c8f86826ac5f7a27e0a27"},{"id":"func/assertDomainMassFixed","name":"assertDomainMassFixed","line":169,"end_line":178,"hash":"57b5cc8685cea6f7697672156a0a04019e8d86230c16977e8d1bb94590631d05"},{"id":"func/addDomainSpring","name":"addDomainSpring","line":180,"end_line":190,"hash":"2b81705feab8fe986d452d2c923509dc37497ec299aac54af45361af1ecccf09"},{"id":"func/setDomainSpringConstant","name":"setDomainSpringConstant","line":192,"end_line":194,"hash":"89a89a0409ae32617eeaf72fb4eb0a24b3c74204c04edbd03d96a46ba3f44341"},{"id":"func/setDomainSpringDamping","name":"setDomainSpringDamping","line":196,"end_line":198,"hash":"6a583bdc6a6d9974986bbca92354e8710bf7b6fccb0f1098c1aedb8ce1d41352"},{"id":"func/setDomainSpringRestLength","name":"setDomainSpringRestLength","line":200,"end_line":202,"hash":"f03039d7c2ca7b41a2d72d73c7d0edc6cc2261c11814e529ae0c954fcdbb437d"},{"id":"func/updateSpringFloat","name":"updateSpringFloat","line":204,"end_line":206,"hash":"c68f71cebf83791cfb548776c4134126f41290bcf8c8cd5dfe590ce281c10c25"},{"id":"func/setSpringDamping","name":"setSpringDamping","line":208,"end_line":210,"hash":"7b273ab4ac7ffe6702bb846c2e318d66f3528ada4809726b1faf116af6e6e991"},{"id":"func/setSpringConstant","name":"setSpringConstant","line":212,"end_line":215,"hash":"bfd9c2f7ed6176a9abbb4ecac512f7a13081000ce360d54947d6f336ea10957a"},{"id":"func/setSpringRestLength","name":"setSpringRestLength","line":217,"end_line":219,"hash":"c8055c79d4647c2d775a2aac601c21f7aba5c45b51db0ec7ae65c11205a13648"},{"id":"func/lookupDomainSpring","name":"lookupDomainSpring","line":221,"end_line":223,"hash":"0d42ef78bb488954482bb9598fbefa783f1b49e2c6bb43ebd6b9e8ed5f4a663f"},{"id":"func/assertDomainSpringEndpoints","name":"assertDomainSpringEndpoints","line":225,"end_line":238,"hash":"4e40d61b081d26e4acd41dec7298723203f26618117feccbf26472c5cb30848b"},{"id":"func/assertDomainSpringConstant","name":"assertDomainSpringConstant","line":240,"end_line":242,"hash":"16432e549ee34e9d81340c097275054b49260d539408034c4f4654081b1e86ad"},{"id":"func/assertDomainSpringDamping","name":"assertDomainSpringDamping","line":244,"end_line":246,"hash":"1f6f863460c146521fb6cbfe86590a5d251476f3800de16be46ae445091b1755"},{"id":"func/assertDomainSpringRestLength","name":"assertDomainSpringRestLength","line":248,"end_line":250,"hash":"9b9f8115cc799c3066d2b15fc8a5fabf9230ff9eef48681ea3d10f4d6823c992"},{"id":"func/addExistingDomainObject","name":"addExistingDomainObject","line":252,"end_line":262,"hash":"a44253509b4a430a0457e67725b772a0083bc818482294e3e9b937ec13e55f64"},{"id":"func/addDuplicateDomainObject","name":"addDuplicateDomainObject","line":264,"end_line":276,"hash":"ba127689bd82d4007d42dc62be414704be5af3edae39365afd2263697d04b5fc"},{"id":"func/objectTypeAndID","name":"objectTypeAndID","line":278,"end_line":288,"hash":"514d3be50887ba74cddab8a8b732ce8686ebc1caaccd71453d15f915b26b639d"},{"id":"func/addExistingSpring","name":"addExistingSpring","line":290,"end_line":298,"hash":"099268e72c6054c0eb1371f0945f2406c453100d65fe55560826237b4f8636bc"},{"id":"func/addInvalidDomainSpring","name":"addInvalidDomainSpring","line":300,"end_line":308,"hash":"d16ec5725879b4340f989a76dee633a711842a4c04a93a74de6c1922bedb92d0"},{"id":"func/assertDomainValidationReason","name":"assertDomainValidationReason","line":310,"end_line":316,"hash":"18d890d7e127e6d8bda59deadf1ba7943a656e9c5b3a6cb3456d12c6b5ec22af"},{"id":"func/ensureDomainWorld","name":"ensureDomainWorld","line":318,"end_line":323,"hash":"d1bf884d014401328fc65d549f54c582723feadf7b3096bbd91bbb6439107246"},{"id":"func/domainWorld","name":"domainWorld","line":325,"end_line":330,"hash":"51b53aa33c5bf836b0aec59e7f11c0c54322923b51971ae6266761fd1f09cde9"},{"id":"func/updateMass","name":"updateMass","line":332,"end_line":334,"hash":"19cad73ae01315dce40aca90421e83d2e5babc2a82348cc72539acabb8faa698"},{"id":"func/updateSpring","name":"updateSpring","line":336,"end_line":338,"hash":"d7ab0c33796da293ef75dde86fa961a464c3a3a41b50b589b07faec25f8f8ec6"},{"id":"func/updateDomainByExample","name":"updateDomainByExample","line":340,"end_line":350,"hash":"86c14f7cdc164a0a1bf4ed1dc968469cdeda1a6008115cbe2d37bc018df98d6c"},{"id":"func/masses","name":"masses","line":352,"end_line":354,"hash":"8feb66e891c4d819075d7962bf90b18a73ffad678ba0962be73edf71cd225956"},{"id":"func/springs","name":"springs","line":356,"end_line":358,"hash":"ef8cd217224a25f6ed112db2f3f35418af8f9b3ff3f144d33056b8c937bbc7f3"},{"id":"func/massID","name":"massID","line":360,"end_line":362,"hash":"58f8283c19220dee19dc6fbdc983a3e7990318140f8c0d10fa99559d5e51aa3d"},{"id":"func/springID","name":"springID","line":364,"end_line":366,"hash":"afcb32a6ae139ce417bc7bad9bf47317ca2eb993a42060facbfbaa77eee226d9"},{"id":"func/floatUpdate","name":"floatUpdate","line":368,"end_line":377,"hash":"cc3d74081b2f1e54eb8ea26e7cb8a09a090be1a6fe3d97e5bea83b3cb66db083"},{"id":"func/lookupByExample","name":"lookupByExample","line":379,"end_line":394,"hash":"d5b6ef8c1db7318dde543cc0581027631e7b3835fb3bcd18a1d530316dc23fb2"},{"id":"func/updateByID","name":"updateByID","line":396,"end_line":403,"hash":"332c2529f1ee9a7803507ce76754018bfa692df32409ebbbafab4408e2d4885b"},{"id":"func/springFromExample","name":"springFromExample","line":405,"end_line":419,"hash":"e52b219578522625b477fe1429633839f7b1700027d77bb9e0c5a7980f4ad992"},{"id":"func/assertValidationReason","name":"assertValidationReason","line":421,"end_line":436,"hash":"7a89cbe81daa1e9f74db5e540ab5ff5178c4a77a9ad4d68f356da949b83364ad"},{"id":"func/assertVec","name":"assertVec","line":438,"end_line":443,"hash":"cec8ef1a2f584bac65f12feb4c516f0ec66b548b1db35ca40ff0ee7192221467"},{"id":"func/assertFloat","name":"assertFloat","line":445,"end_line":450,"hash":"4d4f35b2204102df1838d33200f742d6b2ffc6706b827b78eeed8a1c1e683e69"},{"id":"func/massFields","name":"massFields","line":452,"end_line":466,"hash":"dc6d9c33354b4032102712ef42d3b7c45ff921b3116895d2da056cf3cbfac6ca"}]}
// mutate4go-manifest-end
