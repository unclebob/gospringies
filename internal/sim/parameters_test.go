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

	for _, name := range []string{"gravity", "center attraction", "center of mass attraction", "wall repulsion", "mass collision"} {
		force, ok := world.Parameters.Force(name)
		if !ok {
			t.Fatalf("missing force %q", name)
		}
		if force.Enabled != "false" {
			t.Fatalf("force %q enabled by default: %#v", name, force)
		}
		if name != "mass collision" && len(force.Values) == 0 {
			t.Fatalf("force %q has no editable values", name)
		}
	}

	for _, name := range []string{"top", "left", "right", "bottom"} {
		enabled, ok := world.Parameters.WallEnabled(name)
		if !ok {
			t.Fatalf("missing wall %q", name)
		}
		if enabled {
			t.Fatalf("wall %q enabled by default", name)
		}
	}
}

func TestEnableForcePreservesExistingValues(t *testing.T) {
	parameters := DefaultParameters()

	parameters.EnableForce("gravity", nil)

	force, _ := parameters.Force("gravity")
	if force.Enabled != "true" || force.Values["magnitude"] != "0" || force.Values["direction"] != "90" {
		t.Fatalf("gravity after enable = %#v", force)
	}
	if parameters.ActiveForce != "gravity" {
		t.Fatalf("active force = %q", parameters.ActiveForce)
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
