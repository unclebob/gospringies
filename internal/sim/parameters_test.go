package sim

import "testing"

func TestNewWorldStartsWithEditableDefaults(t *testing.T) {
	world := NewWorld()

	for _, name := range []string{
		"current mass",
		"elasticity",
		"spring constant",
		"damping",
		"viscosity",
		"stickiness",
		"timestep",
		"precision",
		"grid snap",
		"show springs",
	} {
		if !world.Parameters.Has(name) {
			t.Fatalf("missing parameter %q", name)
		}
	}
}

func TestNewWorldStartsWithEditableForceAndWallConfiguration(t *testing.T) {
	world := NewWorld()

	for _, name := range []string{"gravity", "center attraction", "center of mass attraction", "wall repulsion"} {
		force, ok := world.Parameters.Force(name)
		if !ok {
			t.Fatalf("missing force %q", name)
		}
		if len(force.Values) == 0 {
			t.Fatalf("force %q has no editable values", name)
		}
	}

	for _, name := range []string{"top", "left", "right", "bottom"} {
		if _, ok := world.Parameters.WallEnabled(name); !ok {
			t.Fatalf("missing wall %q", name)
		}
	}
}

func TestResetRestoresDefaultParameters(t *testing.T) {
	world := NewWorld()
	world.Parameters.Set("viscosity", "custom")

	world.Reset()

	if got := world.Parameters.Value("viscosity"); got != DefaultParameters().Value("viscosity") {
		t.Fatalf("viscosity = %q", got)
	}
}

func TestLoadReplacesParameters(t *testing.T) {
	world := NewWorld()
	world.Parameters.Set("timestep", "custom")
	loaded := NewWorld()
	loaded.Parameters.Set("timestep", "loaded")

	world.LoadFrom(loaded)

	if got := world.Parameters.Value("timestep"); got != "loaded" {
		t.Fatalf("timestep = %q", got)
	}
}

func TestInsertPreservesExistingParameters(t *testing.T) {
	world := NewWorld()
	world.Parameters.Set("damping", "custom")
	inserted := NewWorld()
	inserted.Parameters.Set("damping", "inserted")

	world.InsertFrom(inserted)

	if got := world.Parameters.Value("damping"); got != "custom" {
		t.Fatalf("damping = %q", got)
	}
}
