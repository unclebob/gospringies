package acceptance

import (
	"errors"
	"fmt"
	"math"

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
	if len(world.Masses) != expected {
		return fmt.Errorf("expected %d masses, got %d", expected, len(world.Masses))
	}
	return nil
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
	if len(world.Springs) != expected {
		return fmt.Errorf("expected %d springs, got %d", expected, len(world.Springs))
	}
	return nil
}

func createDomainMassFromID(w *world, example map[string]string) error {
	return createDomainMass(w, example, "id", "x", "y")
}

func createDomainMassA(w *world, example map[string]string) error {
	return createDomainMass(w, example, "mass_a", "x_a", "y_a")
}

func createDomainMassB(w *world, example map[string]string) error {
	return createDomainMass(w, example, "mass_b", "x_b", "y_b")
}

func createExistingDomainMass(w *world, example map[string]string) error {
	return createDomainMass(w, example, "existing_mass", "x", "y")
}

func createDomainMass(w *world, example map[string]string, idKey, xKey, yKey string) error {
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

func setMassVelocity(w *world, example map[string]string) error {
	return updateMass(w, example, func(mass *sim.Mass) error {
		vx, err := floatValue(example, "vx")
		if err != nil {
			return err
		}
		vy, err := floatValue(example, "vy")
		if err != nil {
			return err
		}
		mass.Velocity = sim.Vec2{X: vx, Y: vy}
		return nil
	})
}

func setMassValue(w *world, example map[string]string) error {
	return updateMass(w, example, func(mass *sim.Mass) error {
		value, err := floatValue(example, "mass_value")
		if err != nil {
			return err
		}
		mass.Mass = value
		return nil
	})
}

func setMassElasticity(w *world, example map[string]string) error {
	return updateMass(w, example, func(mass *sim.Mass) error {
		value, err := floatValue(example, "elasticity")
		if err != nil {
			return err
		}
		mass.Elasticity = value
		return nil
	})
}

func setMassFixed(w *world, example map[string]string) error {
	return updateMass(w, example, func(mass *sim.Mass) error {
		value, err := boolValue(example, "fixed")
		if err != nil {
			return err
		}
		mass.Fixed = value
		return nil
	})
}

func lookupMass(w *world, example map[string]string) error {
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

func assertMassPosition(w *world, example map[string]string) error {
	x, err := floatValue(example, "x")
	if err != nil {
		return err
	}
	y, err := floatValue(example, "y")
	if err != nil {
		return err
	}
	return assertVec("position", w.lookedMass.Position, x, y)
}

func assertMassVelocity(w *world, example map[string]string) error {
	vx, err := floatValue(example, "vx")
	if err != nil {
		return err
	}
	vy, err := floatValue(example, "vy")
	if err != nil {
		return err
	}
	return assertVec("velocity", w.lookedMass.Velocity, vx, vy)
}

func assertMassValue(w *world, example map[string]string) error {
	expected, err := floatValue(example, "mass_value")
	if err != nil {
		return err
	}
	return assertFloat("mass value", w.lookedMass.Mass, expected)
}

func assertMassElasticity(w *world, example map[string]string) error {
	expected, err := floatValue(example, "elasticity")
	if err != nil {
		return err
	}
	return assertFloat("elasticity", w.lookedMass.Elasticity, expected)
}

func assertMassFixed(w *world, example map[string]string) error {
	expected, err := boolValue(example, "fixed")
	if err != nil {
		return err
	}
	if w.lookedMass.Fixed != expected {
		return fmt.Errorf("expected fixed %t, got %t", expected, w.lookedMass.Fixed)
	}
	return nil
}

func createDomainSpring(w *world, example map[string]string) error {
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

func setSpringConstant(w *world, example map[string]string) error {
	return updateSpring(w, example, func(spring *sim.Spring) error {
		value, err := floatValue(example, "spring_constant")
		if err != nil {
			return err
		}
		spring.SpringConstant = value
		spring.Stiffness = value
		return nil
	})
}

func setSpringDamping(w *world, example map[string]string) error {
	return updateSpring(w, example, func(spring *sim.Spring) error {
		value, err := floatValue(example, "damping_constant")
		if err != nil {
			return err
		}
		spring.Damping = value
		return nil
	})
}

func setSpringRestLength(w *world, example map[string]string) error {
	return updateSpring(w, example, func(spring *sim.Spring) error {
		value, err := floatValue(example, "rest_length")
		if err != nil {
			return err
		}
		spring.RestLength = value
		return nil
	})
}

func lookupSpring(w *world, example map[string]string) error {
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

func assertSpringEndpoints(w *world, example map[string]string) error {
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

func assertSpringConstant(w *world, example map[string]string) error {
	expected, err := floatValue(example, "spring_constant")
	if err != nil {
		return err
	}
	return assertFloat("spring constant", w.lookedSpring.SpringConstant, expected)
}

func assertSpringDamping(w *world, example map[string]string) error {
	expected, err := floatValue(example, "damping_constant")
	if err != nil {
		return err
	}
	return assertFloat("damping constant", w.lookedSpring.Damping, expected)
}

func assertSpringRestLength(w *world, example map[string]string) error {
	expected, err := floatValue(example, "rest_length")
	if err != nil {
		return err
	}
	return assertFloat("rest length", w.lookedSpring.RestLength, expected)
}

func createDuplicateSubject(w *world, example map[string]string) error {
	world := ensureDomainWorld(w)
	objectType, err := stringValue(example, "object_type")
	if err != nil {
		return err
	}
	id, err := intValue(example, "id")
	if err != nil {
		return err
	}
	if objectType == "mass" {
		return world.AddMass(sim.Mass{ID: id, Mass: 1})
	}
	if err := world.AddMass(sim.Mass{ID: 1, Mass: 1}); err != nil {
		return err
	}
	if err := world.AddMass(sim.Mass{ID: 2, Mass: 1}); err != nil {
		return err
	}
	return world.AddSpring(sim.Spring{ID: id, MassA: 1, MassB: 2})
}

func addDuplicateSubject(w *world, example map[string]string) error {
	world := ensureDomainWorld(w)
	objectType, err := stringValue(example, "object_type")
	if err != nil {
		return err
	}
	id, err := intValue(example, "id")
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

func addDomainSpringForValidation(w *world, example map[string]string) error {
	world := ensureDomainWorld(w)
	spring, err := springFromExample(example)
	if err != nil {
		return err
	}
	w.validationErr = world.AddSpring(spring)
	return nil
}

func assertValidationFailure(w *world, example map[string]string) error {
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
	switch reason {
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
