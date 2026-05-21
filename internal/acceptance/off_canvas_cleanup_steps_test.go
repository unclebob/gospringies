package acceptance

import "testing"

func TestOffCanvasCleanupStepHandlersRunScenario(t *testing.T) {
	if _, ok := stepHandlers["cleanup spring <spring_id> connects mass <spring_mass_a> to mass <spring_mass_b>"]; !ok {
		t.Fatal("cleanup spring step was not registered")
	}

	w := &world{}
	example := map[string]string{
		"canvas_width":          "200",
		"canvas_height":         "100",
		"mass_a":                "1",
		"x_a":                   "50",
		"y_a":                   "50",
		"mass_b":                "2",
		"x_b":                   "50",
		"y_b":                   "201",
		"spring_id":             "1",
		"spring_mass_a":         "1",
		"spring_mass_b":         "2",
		"remaining_mass":        "1",
		"expected_mass_count":   "1",
		"expected_spring_count": "0",
	}

	for _, step := range []stepHandler{
		createCleanupCanvas,
		createCleanupMassA,
		createCleanupMassB,
		createCleanupSpring,
		advanceOffCanvasCleanup,
		assertCleanupMassCount,
		assertCleanupSpringCount,
		assertCleanupMassPresent,
	} {
		if err := step(w, example); err != nil {
			t.Fatal(err)
		}
	}
}
