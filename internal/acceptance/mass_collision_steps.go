package acceptance

import (
	"fmt"

	"springs/internal/sim"
)

type collisionMassProperties struct {
	id         int
	mass       float64
	elasticity float64
	fixed      bool
}

func createCollisionMassA(w *world, example map[string]string) error {
	return createCollisionMass(w, example, "mass_a", "x_a", "y_a", "vx_a", "vy_a")
}

func createCollisionMassB(w *world, example map[string]string) error {
	return createCollisionMass(w, example, "mass_b", "x_b", "y_b", "vx_b", "vy_b")
}

func createCollisionMass(w *world, example map[string]string, idKey, xKey, yKey, vxKey, vyKey string) error {
	world := ensureDomainWorld(w)
	id, err := intValue(example, idKey)
	if err != nil {
		return err
	}
	position, err := vecFromExample(example, xKey, yKey)
	if err != nil {
		return err
	}
	velocity, err := vecFromExample(example, vxKey, vyKey)
	if err != nil {
		return err
	}
	return world.AddMass(sim.Mass{ID: id, Position: position, Velocity: velocity, Mass: 1, Elasticity: 1})
}

func setCollisionMassAProperties(w *world, example map[string]string) error {
	return setCollisionMassProperties(w, example, "mass_a", "mass_value_a", "elasticity_a", "fixed_a")
}

func setCollisionMassBProperties(w *world, example map[string]string) error {
	return setCollisionMassProperties(w, example, "mass_b", "mass_value_b", "elasticity_b", "fixed_b")
}

func setCollisionMassProperties(w *world, example map[string]string, idKey, massKey, elasticityKey, fixedKey string) error {
	properties, err := collisionMassPropertiesFromExample(example, idKey, massKey, elasticityKey, fixedKey)
	if err != nil {
		return err
	}
	return applyCollisionMassProperties(ensureDomainWorld(w), properties)
}

func collisionMassPropertiesFromExample(example map[string]string, idKey, massKey, elasticityKey, fixedKey string) (collisionMassProperties, error) {
	id, err := intValue(example, idKey)
	if err != nil {
		return collisionMassProperties{}, err
	}
	massValue, err := floatValue(example, massKey)
	if err != nil {
		return collisionMassProperties{}, err
	}
	elasticity, err := floatValue(example, elasticityKey)
	if err != nil {
		return collisionMassProperties{}, err
	}
	fixed, err := boolValue(example, fixedKey)
	if err != nil {
		return collisionMassProperties{}, err
	}
	return collisionMassProperties{id: id, mass: massValue, elasticity: elasticity, fixed: fixed}, nil
}

func applyCollisionMassProperties(world *sim.Simulation, properties collisionMassProperties) error {
	for i := range world.Masses {
		if world.Masses[i].ID == properties.id {
			world.Masses[i].Mass = properties.mass
			world.Masses[i].Elasticity = properties.elasticity
			world.Masses[i].Fixed = properties.fixed
			return nil
		}
	}
	return fmt.Errorf("mass %d not found", properties.id)
}

func enableMassCollision(w *world, _ map[string]string) error {
	ensureDomainWorld(w).Parameters.EnableForce("mass collision", map[string]string{})
	return nil
}

func advanceThroughMassCollision(w *world, _ map[string]string) error {
	ensureDomainWorld(w).Step(sim.DefaultParameters().StepDuration())
	return nil
}

func assertCollisionMassAVelocity(w *world, example map[string]string) error {
	return assertCollisionMassVelocity(w, example, "mass_a", "expected_vx_a", "expected_vy_a")
}

func assertCollisionMassBVelocity(w *world, example map[string]string) error {
	return assertCollisionMassVelocity(w, example, "mass_b", "expected_vx_b", "expected_vy_b")
}

func assertCollisionMassVelocity(w *world, example map[string]string, idKey, vxKey, vyKey string) error {
	id, err := intValue(example, idKey)
	if err != nil {
		return err
	}
	expected, err := vecFromExample(example, vxKey, vyKey)
	if err != nil {
		return err
	}
	mass, ok := ensureDomainWorld(w).MassByID(id)
	if !ok {
		return fmt.Errorf("mass %d not found", id)
	}
	return assertVec("collision velocity", mass.Velocity, expected.X, expected.Y)
}
