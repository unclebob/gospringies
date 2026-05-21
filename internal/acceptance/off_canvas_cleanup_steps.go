package acceptance

import (
	"fmt"

	"springs/internal/sim"
)

func init() {
	for step, handler := range map[string]stepHandler{
		"the off-canvas cleanup task is accepted":                                          acceptStep,
		"a cleanup canvas with width <canvas_width> and height <canvas_height>":            createCleanupCanvas,
		"cleanup mass <mass_a> starts at <x_a>, <y_a>":                                     createCleanupMassA,
		"cleanup mass <mass_b> starts at <x_b>, <y_b>":                                     createCleanupMassB,
		"cleanup spring <spring_id> connects mass <spring_mass_a> to mass <spring_mass_b>": createCleanupSpring,
		"the coder advances off-canvas cleanup":                                            advanceOffCanvasCleanup,
		"the cleanup world should contain <expected_mass_count> masses":                    assertCleanupMassCount,
		"the cleanup world should contain <expected_spring_count> springs":                 assertCleanupSpringCount,
		"cleanup mass <remaining_mass> should remain present":                              assertCleanupMassPresent,
	} {
		stepHandlers[step] = handler
	}
}

func createCleanupCanvas(w *world, example map[string]string) error {
	values, err := floatValues(example, "canvas_width", "canvas_height")
	if err != nil {
		return err
	}
	ensureDomainWorld(w).Bounds = sim.Bounds{Width: values[0], Height: values[1]}
	return nil
}

func createCleanupMassA(w *world, example map[string]string) error {
	return createCleanupMass(w, example, "mass_a", "x_a", "y_a")
}

func createCleanupMassB(w *world, example map[string]string) error {
	return createCleanupMass(w, example, "mass_b", "x_b", "y_b")
}

func createCleanupMass(w *world, example map[string]string, idKey, xKey, yKey string) error {
	id, err := intValue(example, idKey)
	if err != nil {
		return err
	}
	values, err := floatValues(example, xKey, yKey)
	if err != nil {
		return err
	}
	return ensureDomainWorld(w).AddMass(sim.Mass{
		ID:       id,
		Position: sim.Vec2{X: values[0], Y: values[1]},
		Mass:     1,
		Fixed:    true,
	})
}

func createCleanupSpring(w *world, example map[string]string) error {
	springID, err := intValue(example, "spring_id")
	if err != nil {
		return err
	}
	massA, massB, err := intPair(example, "spring_mass_a", "spring_mass_b")
	if err != nil {
		return err
	}
	return ensureDomainWorld(w).AddSpring(sim.Spring{ID: springID, MassA: massA, MassB: massB})
}

func advanceOffCanvasCleanup(w *world, _ map[string]string) error {
	ensureDomainWorld(w).Step(0)
	return nil
}

func assertCleanupMassCount(w *world, example map[string]string) error {
	expected, err := intValue(example, "expected_mass_count")
	if err != nil {
		return err
	}
	if got := len(ensureDomainWorld(w).Masses); got != expected {
		return fmt.Errorf("cleanup mass count = %d, expected %d", got, expected)
	}
	return nil
}

func assertCleanupSpringCount(w *world, example map[string]string) error {
	expected, err := intValue(example, "expected_spring_count")
	if err != nil {
		return err
	}
	if got := len(ensureDomainWorld(w).Springs); got != expected {
		return fmt.Errorf("cleanup spring count = %d, expected %d", got, expected)
	}
	return nil
}

func assertCleanupMassPresent(w *world, example map[string]string) error {
	id, err := intValue(example, "remaining_mass")
	if err != nil {
		return err
	}
	if _, ok := ensureDomainWorld(w).MassByID(id); !ok {
		return fmt.Errorf("cleanup mass %d was deleted", id)
	}
	return nil
}
