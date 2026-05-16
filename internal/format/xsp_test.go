package format

import (
	"errors"
	"strings"
	"testing"

	"springs/internal/sim"
)

func TestLoadXSPRequiresSupportedMarker(t *testing.T) {
	if _, err := LoadXSP("#1.0\n"); err != nil {
		t.Fatalf("LoadXSP returned error: %v", err)
	}
	if _, err := LoadXSP("mass 1 0 0 1 0.8\n"); !errors.Is(err, ErrUnsupportedMarker) {
		t.Fatalf("marker error = %v", err)
	}
}

func TestLoadXSPSupportsParametersForcesWallsMassesAndSprings(t *testing.T) {
	world, err := LoadXSP(strings.Join([]string{
		"#1.0",
		"cmas 3.0",
		"elas 0.4",
		"kspr 12.5",
		"kdmp 0.7",
		"frce gravity true magnitude=10 direction=90",
		"wall left true",
		"mass 1 10 20 -3.0 0.4",
		"mass 2 30 40 2.0 0.5",
		"spng 7 1 2 15 12.5 0.7",
		"",
	}, "\n"))
	if err != nil {
		t.Fatalf("LoadXSP returned error: %v", err)
	}

	assertParameter(t, world, "current mass", "3.0")
	assertParameter(t, world, "elasticity", "0.4")
	assertParameter(t, world, "spring constant", "12.5")
	assertParameter(t, world, "damping", "0.7")
	if force, _ := world.Parameters.Force("gravity"); force.Enabled != "true" || force.Values["magnitude"] != "10" {
		t.Fatalf("gravity = %#v", force)
	}
	if enabled, _ := world.Parameters.WallEnabled("left"); !enabled {
		t.Fatal("left wall was not enabled")
	}
	mass, _ := world.MassByID(1)
	if !mass.Fixed || mass.Mass != 3.0 || mass.Elasticity != 0.4 {
		t.Fatalf("mass 1 = %#v", mass)
	}
	spring, _ := world.SpringByID(7)
	if spring.MassA != 1 || spring.MassB != 2 || spring.RestLength != 15 || spring.SpringConstant != 12.5 || spring.Damping != 0.7 {
		t.Fatalf("spring = %#v", spring)
	}
}

func TestSaveXSPIsDeterministicAndEndsWithNewline(t *testing.T) {
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 10, Y: 20}, Mass: 1, Elasticity: 0.8})

	first := SaveXSP(world)
	second := SaveXSP(world)

	if first != second {
		t.Fatalf("save output changed:\n%s\n---\n%s", first, second)
	}
	if !strings.HasSuffix(first, "\n") {
		t.Fatalf("saved output missing final newline: %q", first)
	}
}

func TestFixedMassesRoundTripThroughNegativeFileMass(t *testing.T) {
	input := "#1.0\nmass 1 10 20 -3.0 0.8\n"

	world, err := LoadXSP(input)
	if err != nil {
		t.Fatal(err)
	}
	output := SaveXSP(world)

	mass, _ := world.MassByID(1)
	if !mass.Fixed || mass.Mass != 3.0 {
		t.Fatalf("mass = %#v", mass)
	}
	if !strings.Contains(output, "mass 1 10 20 -3 0.8\n") {
		t.Fatalf("saved output = %q", output)
	}
}

func TestLoadXSPReportsMalformedInput(t *testing.T) {
	cases := []struct {
		name string
		text string
		want error
	}{
		{"duplicate mass id", "#1.0\nmass 1 0 0 1 0.8\nmass 1 1 1 1 0.8\n", sim.ErrDuplicateID},
		{"duplicate spring id", "#1.0\nmass 1 0 0 1 0.8\nmass 2 1 1 1 0.8\nspng 1 1 2 1 1 0\nspng 1 1 2 1 1 0\n", sim.ErrDuplicateID},
		{"missing spring endpoint", "#1.0\nmass 1 0 0 1 0.8\nspng 1 1 2 1 1 0\n", sim.ErrMissingSpringEndpoint},
		{"missing final newline", "#1.0", ErrMissingFinalNewline},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := LoadXSP(tc.text); !errors.Is(err, tc.want) {
				t.Fatalf("error = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestLoadXSPReportsBadCommandFields(t *testing.T) {
	cases := []string{
		"#1.0\nbogus\n",
		"#1.0\ncmas\n",
		"#1.0\nfrce gravity true magnitude\n",
		"#1.0\nfrce gravity\n",
		"#1.0\nmass bad 0 0 1 0.8\n",
		"#1.0\nmass 1 bad 0 1 0.8\n",
		"#1.0\nmass 1 0 bad 1 0.8\n",
		"#1.0\nmass 1 0 0 bad 0.8\n",
		"#1.0\nmass 1 0 0 1 bad\n",
		"#1.0\nspng 1 1 2 bad 1 0\n",
	}
	for _, text := range cases {
		t.Run(text, func(t *testing.T) {
			if _, err := LoadXSP(text); err == nil {
				t.Fatal("expected malformed command error")
			}
		})
	}
}

func TestLoadXSPPreservesDisabledForceValues(t *testing.T) {
	world, err := LoadXSP("#1.0\nfrce gravity false magnitude=5 direction=180\n")
	if err != nil {
		t.Fatal(err)
	}
	force, _ := world.Parameters.Force("gravity")
	if force.Enabled != "false" || force.Values["magnitude"] != "5" || force.Values["direction"] != "180" {
		t.Fatalf("gravity = %#v", force)
	}
}

func assertParameter(t *testing.T, world *sim.Simulation, name, value string) {
	t.Helper()
	if got := world.Parameters.Value(name); got != value {
		t.Fatalf("%s = %q", name, got)
	}
}
