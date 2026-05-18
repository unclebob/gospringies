package acceptance

import (
	"strings"
	"testing"

	"springs/internal/app"
	"springs/internal/sim"
)

func TestClickableVisibleControlActiveAssertion(t *testing.T) {
	w := &world{appGame: app.NewGame()}

	if err := assertVisibleControlActive(w, map[string]string{"control": "Run"}); err != nil {
		t.Fatal(err)
	}
	if err := assertVisibleControlActive(w, map[string]string{"control": "Missing"}); err == nil {
		t.Fatal("expected missing active control error")
	}
}

func TestDragAppMassUpdatesWorld(t *testing.T) {
	game := app.NewGame()
	domain := sim.NewWorld()
	_ = domain.AddMass(sim.Mass{ID: 7, Position: sim.Vec2{}, Mass: 1})
	w := &world{domainWorld: domain}

	if err := dragAppMass(w, game, 7, sim.Vec2{X: 3, Y: 4}); err != nil {
		t.Fatal(err)
	}
	mass, ok := w.domainWorld.MassByID(7)
	if !ok || mass.Position != (sim.Vec2{X: 3, Y: 4}) {
		t.Fatalf("dragged mass = %#v ok=%t", mass, ok)
	}
	if err := dragAppMass(w, game, 99, sim.Vec2{}); err == nil {
		t.Fatal("expected missing mass drag error")
	}
}

func TestMouseMassExpectedIDAssertion(t *testing.T) {
	domain := sim.NewWorld()
	_ = domain.AddMass(sim.Mass{ID: 3, Mass: 1})
	w := &world{domainWorld: domain}

	if err := assertMouseMassExpectedID(w, map[string]string{"mass_id": "3", "expected_mass_id": "3"}); err != nil {
		t.Fatal(err)
	}
	if err := assertMouseMassExpectedID(w, map[string]string{"mass_id": "3", "expected_mass_id": "4"}); err == nil {
		t.Fatal("expected mismatched mass id")
	}
}

func TestCollisionMassPropertiesSetter(t *testing.T) {
	domain := sim.NewWorld()
	_ = domain.AddMass(sim.Mass{ID: 2, Mass: 1})
	w := &world{domainWorld: domain}
	example := map[string]string{
		"mass":       "2",
		"mass_value": "5",
		"elasticity": "0.25",
		"fixed":      "true",
	}

	if err := setCollisionMassProperties(w, example, "mass", "mass_value", "elasticity", "fixed"); err != nil {
		t.Fatal(err)
	}
	mass, _ := domain.MassByID(2)
	if mass.Mass != 5 || mass.Elasticity != 0.25 || !mass.Fixed {
		t.Fatalf("mass properties = %#v", mass)
	}
	example["mass"] = "99"
	if err := setCollisionMassProperties(w, example, "mass", "mass_value", "elasticity", "fixed"); err == nil {
		t.Fatal("expected missing mass error")
	}
}

func TestDocumentedCommandRunnerChecksPrerequisites(t *testing.T) {
	if err := runDocumentedCommand(&world{}, map[string]string{"command": "run"}); err == nil {
		t.Fatal("expected clean checkout prerequisite error")
	}
	w := &world{cleanCheckout: true}
	if err := runDocumentedCommand(w, map[string]string{"command": "run"}); err != nil {
		t.Fatal(err)
	}
	if w.documentedCommand != "run" || w.documentedCommandErr != nil {
		t.Fatalf("documented command state = %q %v", w.documentedCommand, w.documentedCommandErr)
	}
	if err := runDocumentedCommand(&world{cleanCheckout: true}, map[string]string{"command": "unknown"}); err == nil || !strings.Contains(err.Error(), "unsupported") {
		t.Fatalf("expected unsupported command error, got %v", err)
	}
}
