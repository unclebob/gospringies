package acceptance

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"springs/internal/sim"
)

func createDomainWorld(w *world, _ map[string]string) error {
	w.domainWorld = sim.NewWorld()
	return nil
}

func assertDomainMassCount(w *world, example map[string]string) error {
	world, err := domainWorld(w)
	if err != nil {
		return err
	}
	expected, err := intValue(example, "mass_count")
	if err != nil {
		return err
	}
	return assertCount("masses", len(world.Masses), expected)
}

func assertDomainSpringCount(w *world, example map[string]string) error {
	world, err := domainWorld(w)
	if err != nil {
		return err
	}
	expected, err := intValue(example, "spring_count")
	if err != nil {
		return err
	}
	return assertCount("springs", len(world.Springs), expected)
}

func assertCount(name string, got, expected int) error {
	if got != expected {
		return fmt.Errorf("expected %d %s, got %d", expected, name, got)
	}
	return nil
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
	return updateMassFloat(w, example, "mass_value", func(mass *sim.Mass, value float64) {
		mass.Mass = value
	})
}

func setDomainMassElasticity(w *world, example map[string]string) error {
	return updateMassFloat(w, example, "elasticity", func(mass *sim.Mass, value float64) {
		mass.Elasticity = value
	})
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
	return updateMass(w, example, func(mass *sim.Mass) error {
		value, err := floatValue(example, key)
		if err != nil {
			return err
		}
		assign(mass, value)
		return nil
	})
}

func lookupDomainMass(w *world, example map[string]string) error {
	world, err := domainWorld(w)
	if err != nil {
		return err
	}
	id, err := intValue(example, "id")
	if err != nil {
		return err
	}
	mass, ok := world.MassByID(id)
	if !ok {
		return fmt.Errorf("mass %d not found", id)
	}
	w.lookedMass = mass
	return nil
}

func assertDomainMassPosition(w *world, example map[string]string) error {
	expected, err := vecFromExample(example, "x", "y")
	if err != nil {
		return err
	}
	return assertVec("position", w.lookedMass.Position, expected.X, expected.Y)
}

func assertDomainMassVelocity(w *world, example map[string]string) error {
	expected, err := vecFromExample(example, "vx", "vy")
	if err != nil {
		return err
	}
	return assertVec("velocity", w.lookedMass.Velocity, expected.X, expected.Y)
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
	return updateSpringFloat(w, example, "spring_constant", func(spring *sim.Spring, value float64) {
		spring.SpringConstant = value
		spring.Stiffness = value
	})
}

func setDomainSpringDamping(w *world, example map[string]string) error {
	return updateSpringFloat(w, example, "damping_constant", func(spring *sim.Spring, value float64) {
		spring.Damping = value
	})
}

func setDomainSpringRestLength(w *world, example map[string]string) error {
	return updateSpringFloat(w, example, "rest_length", func(spring *sim.Spring, value float64) {
		spring.RestLength = value
	})
}

func updateSpringFloat(w *world, example map[string]string, key string, assign func(*sim.Spring, float64)) error {
	return updateSpring(w, example, func(spring *sim.Spring) error {
		value, err := floatValue(example, key)
		if err != nil {
			return err
		}
		assign(spring, value)
		return nil
	})
}

func lookupDomainSpring(w *world, example map[string]string) error {
	world, err := domainWorld(w)
	if err != nil {
		return err
	}
	id, err := intValue(example, "spring_id")
	if err != nil {
		return err
	}
	spring, ok := world.SpringByID(id)
	if !ok {
		return fmt.Errorf("spring %d not found", id)
	}
	w.lookedSpring = spring
	return nil
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
		w.validationErr = world.AddMass(sim.Mass{ID: id, Mass: 1})
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
	world, err := domainWorld(w)
	if err != nil {
		return err
	}
	id, err := intValue(example, "id")
	if err != nil {
		return err
	}
	for i := range world.Masses {
		if world.Masses[i].ID == id {
			return update(&world.Masses[i])
		}
	}
	return fmt.Errorf("mass %d not found", id)
}

func updateSpring(w *world, example map[string]string, update func(*sim.Spring) error) error {
	world, err := domainWorld(w)
	if err != nil {
		return err
	}
	id, err := intValue(example, "spring_id")
	if err != nil {
		return err
	}
	for i := range world.Springs {
		if world.Springs[i].ID == id {
			return update(&world.Springs[i])
		}
	}
	return fmt.Errorf("spring %d not found", id)
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
