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
		"fixm 1",
		"shws 0",
		"cent -1",
		"frce gravity 2 magnitude=10 direction=90",
		"visc 0.2",
		"stck 0.3",
		"step 0.01",
		"prec 0.0001",
		"adpt 1",
		"gsnp 5",
		"wall left 1",
		"mass 1 10 20 -3.0 0.4",
		"mass 2 30 40 2.0 0.5",
		"spng 7 1 2 12.5 0.7 15",
	}, "\n") + "\n")
	if err != nil {
		t.Fatalf("LoadXSP returned error: %v", err)
	}

	assertParameter(t, world, "current mass", "3.0")
	assertParameter(t, world, "elasticity", "0.4")
	assertParameter(t, world, "spring constant", "12.5")
	assertParameter(t, world, "damping", "0.7")
	assertParameter(t, world, "fixed mass", "true")
	assertParameter(t, world, "show springs", "false")
	assertParameter(t, world, "center mass", "-1")
	assertParameter(t, world, "viscosity", "0.2")
	assertParameter(t, world, "stickiness", "0.3")
	assertParameter(t, world, "timestep", "0.01")
	assertParameter(t, world, "precision", "0.0001")
	assertParameter(t, world, "adaptive timestep", "true")
	assertParameter(t, world, "grid snap", "5")
	if force, _ := world.Parameters.Force("gravity"); force.Enabled != "true" || force.Values["magnitude"] != "10" || force.Values["direction"] != "90" {
		t.Fatalf("gravity = %#v", force)
	}
	if enabled, _ := world.Parameters.WallEnabled("left"); !enabled {
		t.Fatal("left wall was not enabled")
	}
	mass, _ := world.MassByID(1)
	if !mass.Fixed || mass.Position.X != 10 || mass.Position.Y != 20 || mass.Mass != 3.0 || mass.Elasticity != 0.4 {
		t.Fatalf("mass 1 = %#v", mass)
	}
	mass, _ = world.MassByID(2)
	if mass.Fixed || mass.Position.X != 30 || mass.Position.Y != 40 || mass.Mass != 2.0 || mass.Elasticity != 0.5 {
		t.Fatalf("mass 2 = %#v", mass)
	}
	spring, _ := world.SpringByID(7)
	if spring.MassA != 1 || spring.MassB != 2 || spring.RestLength != 15 || spring.SpringConstant != 12.5 || spring.Stiffness != 12.5 || spring.Damping != 0.7 {
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
		{"blank line", "#1.0\n\n", ErrBlankLine},
		{"non-positive mass id", "#1.0\nmass 0 0 0 1 0.8\n", ErrNonPositiveID},
		{"non-positive spring id", "#1.0\nmass 1 0 0 1 0.8\nmass 2 1 1 1 0.8\nspng 0 1 2 1 1 0\n", ErrNonPositiveID},
		{"non-positive center id", "#1.0\ncent 0\n", ErrNonPositiveID},
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
		"#1.0\nfixm\n",
		"#1.0\nfixm maybe\n",
		"#1.0\ncent x\n",
		"#1.0\nfrce gravity true magnitude\n",
		"#1.0\nfrce gravity\n",
		"#1.0\nfrce gravity maybe\n",
		"#1.0\nfrce bad 1 10 90\n",
		"#1.0\nfrce -1 1 10 90\n",
		"#1.0\nfrce 5 1 10 90\n",
		"#1.0\nfrce 0 maybe 10 90\n",
		"#1.0\nfrce 0 1 bad 90\n",
		"#1.0\nfrce 0 1 10 bad\n",
		"#1.0\nwall left maybe\n",
		"#1.0\nmass bad 0 0 1 0.8\n",
		"#1.0\nmass 1 bad 0 1 0.8\n",
		"#1.0\nmass 1 0 bad 1 0.8\n",
		"#1.0\nmass 1 0 0 bad 0 1 0.8\n",
		"#1.0\nmass 1 0 0 1 bad 1 0.8\n",
		"#1.0\nmass 1 0 0 bad 0.8\n",
		"#1.0\nmass 1 0 0 1 bad\n",
		"#1.0\nmass 1 0 0 1\n",
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

func TestLoadXSPReportsFailingLineNumber(t *testing.T) {
	assertLoadXSPErrorContains(t, "#1.0\ncmas 1\nmass bad 0 0 1 0.8\n", "line 3:")
}

func TestLoadXSPReportsBooleanCommandInError(t *testing.T) {
	assertLoadXSPErrorContains(t, "#1.0\nfixm maybe\n", "fixm:")
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

func TestLoadXSPAcceptsMultiWordForceNames(t *testing.T) {
	input := strings.Join([]string{
		"#1.0",
		"frce center attraction false magnitude=10 exponent=0",
		"frce center of mass attraction false magnitude=5 damping=2",
		"frce wall repulsion false magnitude=10000 exponent=1",
		"frce mass collision false",
	}, "\n") + "\n"

	world, err := LoadXSP(input)
	if err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"center attraction", "center of mass attraction", "wall repulsion", "mass collision"} {
		force, ok := world.Parameters.Force(name)
		if !ok || force.Enabled != "false" {
			t.Fatalf("force %q = %#v, %t", name, force, ok)
		}
	}
	if force, _ := world.Parameters.Force("wall repulsion"); force.Values["magnitude"] != "10000" || force.Values["exponent"] != "1" {
		t.Fatalf("wall repulsion values = %#v", force)
	}
	if force, _ := world.Parameters.Force("center of mass attraction"); force.Values["magnitude"] != "5" || force.Values["damping"] != "2" {
		t.Fatalf("center of mass values = %#v", force)
	}
}

func TestLoadXSPMapsStableForceTokens(t *testing.T) {
	input := strings.Join([]string{
		"#1.0",
		"frce center-attraction false magnitude=10 exponent=0",
		"frce center-of-mass-attraction false magnitude=5 damping=2",
		"frce wall-repulsion false magnitude=10000 exponent=1",
		"frce mass-collision false",
	}, "\n") + "\n"

	world, err := LoadXSP(input)
	if err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"center attraction", "center of mass attraction", "wall repulsion", "mass collision"} {
		force, ok := world.Parameters.Force(name)
		if !ok || force.Enabled != "false" {
			t.Fatalf("force %q = %#v, %t", name, force, ok)
		}
	}
	if _, ok := world.Parameters.Force("center-attraction"); ok {
		t.Fatal("hyphenated force token was stored as an internal force name")
	}
	if force, _ := world.Parameters.Force("wall repulsion"); force.Values["magnitude"] != "10000" || force.Values["exponent"] != "1" {
		t.Fatalf("wall repulsion values = %#v", force)
	}
	if force, _ := world.Parameters.Force("center of mass attraction"); force.Values["magnitude"] != "5" || force.Values["damping"] != "2" {
		t.Fatalf("center of mass values = %#v", force)
	}
}

func TestSaveXSPWritesStableForceTokens(t *testing.T) {
	world := sim.NewWorld()
	world.Parameters.Forces["center attraction"] = sim.ForceConfig{Enabled: "false", Values: map[string]string{"magnitude": "0", "exponent": "2"}}
	world.Parameters.Forces["center of mass attraction"] = sim.ForceConfig{Enabled: "false", Values: map[string]string{"magnitude": "0", "damping": "0"}}
	world.Parameters.Forces["wall repulsion"] = sim.ForceConfig{Enabled: "false", Values: map[string]string{"magnitude": "0", "exponent": "2"}}
	world.Parameters.Forces["mass collision"] = sim.ForceConfig{Enabled: "false", Values: map[string]string{}}

	output := SaveXSP(world)

	for _, line := range []string{
		"frce center-attraction false magnitude=0 exponent=2\n",
		"frce center-of-mass-attraction false magnitude=0 damping=0\n",
		"frce wall-repulsion false magnitude=0 exponent=2\n",
		"frce mass-collision false\n",
	} {
		if !strings.Contains(output, "\n"+line) {
			t.Fatalf("saved output missing %q:\n%s", line, output)
		}
	}
	if strings.Contains(output, "\nfrce center attraction ") {
		t.Fatalf("saved output used spaced force name:\n%s", output)
	}
}

func TestLoadXSPHandlesForcesWithoutValues(t *testing.T) {
	world, err := LoadXSP("#1.0\nfrce gravity true\n")
	if err != nil {
		t.Fatal(err)
	}
	force, _ := world.Parameters.Force("gravity")
	if force.Enabled != "true" || force.Values["magnitude"] != "0" {
		t.Fatalf("gravity = %#v", force)
	}
	if _, ok := world.Parameters.Forces["frce"]; ok {
		t.Fatalf("unexpected command-name force = %#v", world.Parameters.Forces["frce"])
	}
}

func TestLoadXSPInitializesUnknownForceValues(t *testing.T) {
	world, err := LoadXSP("#1.0\nfrce custom true magnitude=7\n")
	if err != nil {
		t.Fatal(err)
	}
	force, ok := world.Parameters.Force("custom")
	if !ok || force.Enabled != "true" || force.Values["magnitude"] != "7" {
		t.Fatalf("custom force = %#v, %v", force, ok)
	}
}

func TestLoadXSPTreatsZeroMassAsMovable(t *testing.T) {
	world, err := LoadXSP("#1.0\nmass 1 0 0 0 0.8\n")
	if err != nil {
		t.Fatal(err)
	}
	mass, _ := world.MassByID(1)
	if mass.Fixed || mass.Mass != 0 {
		t.Fatalf("mass = %#v", mass)
	}
}

func TestSaveXSPWritesDocumentedCommands(t *testing.T) {
	world, err := LoadXSP(strings.Join([]string{
		"#1.0",
		"cmas 3",
		"elas 0.4",
		"kspr 12",
		"kdmp 0.7",
		"fixm true",
		"shws false",
		"cent -1",
		"frce gravity true magnitude=10 direction=90",
		"visc 0.2",
		"stck 0.3",
		"step 0.01",
		"prec 0.001",
		"adpt true",
		"gsnp 5",
		"wall left true",
		"mass 1 10 20 1 0.8",
		"mass 2 20 20 1 0.8",
		"spng 7 1 2 12.5 0.7 15",
	}, "\n") + "\n")
	if err != nil {
		t.Fatal(err)
	}
	output := SaveXSP(world)
	for _, command := range []string{"cmas", "elas", "kspr", "kdmp", "fixm", "shws", "cent", "frce", "visc", "stck", "step", "prec", "adpt", "gsnp", "wall", "mass", "spng"} {
		if !strings.Contains(output, "\n"+command+" ") {
			t.Fatalf("saved output missing %s:\n%s", command, output)
		}
	}
	if !strings.Contains(output, "\nfrce gravity true magnitude=10 direction=90\n") {
		t.Fatalf("saved output missing force values:\n%s", output)
	}
	if !strings.Contains(output, "\nspng 7 1 2 12.5 0.7 15\n") {
		t.Fatalf("saved output should use original spring order:\n%s", output)
	}
}

func TestForceValueSuffix(t *testing.T) {
	if got := forceValueSuffix(nil); got != "" {
		t.Fatalf("empty suffix = %q", got)
	}
	got := forceValueSuffix(map[string]string{
		"ignored":   "x",
		"damping":   "0.5",
		"magnitude": "10",
	})
	if got != " magnitude=10 damping=0.5" {
		t.Fatalf("suffix = %q", got)
	}
}

func TestXSPParseHelperErrorValues(t *testing.T) {
	if name, enabled, first, second, err := legacyForceFields([]string{"frce", "bad", "1", "10", "90"}); err == nil || name != "" || enabled != "" || first != 0 || second != 0 {
		t.Fatalf("legacy force id parse = %q, %q, %v, %v, %v", name, enabled, first, second, err)
	}
	if name, enabled, first, second, err := legacyForceFields([]string{"frce", "5", "1", "10", "90"}); err == nil || name != "" || enabled != "" || first != 0 || second != 0 {
		t.Fatalf("legacy force name parse = %q, %q, %v, %v, %v", name, enabled, first, second, err)
	}
	if name, enabled, first, second, err := legacyForceFields([]string{"frce", "0", "maybe", "10", "90"}); err == nil || name != "" || enabled != "" || first != 0 || second != 0 {
		t.Fatalf("legacy force enabled parse = %q, %q, %v, %v, %v", name, enabled, first, second, err)
	}
	if name, enabled, first, second, err := legacyForceFields([]string{"frce", "0", "1", "bad", "90"}); err == nil || name != "" || enabled != "" || first != 0 || second != 0 {
		t.Fatalf("legacy force first parse = %q, %q, %v, %v, %v", name, enabled, first, second, err)
	}
	if position, velocity, mass, elasticity, err := massNumericFields([]string{"mass", "1", "bad", "2", "3", "4"}); err == nil || position != (sim.Vec2{}) || velocity != (sim.Vec2{}) || mass != 0 || elasticity != 0 {
		t.Fatalf("mass x parse = %v, %v, %v, %v, %v", position, velocity, mass, elasticity, err)
	}
	if position, velocity, mass, elasticity, err := massNumericFields([]string{"mass", "1", "1", "bad", "3", "4"}); err == nil || position != (sim.Vec2{}) || velocity != (sim.Vec2{}) || mass != 0 || elasticity != 0 {
		t.Fatalf("mass y parse = %v, %v, %v, %v, %v", position, velocity, mass, elasticity, err)
	}
	if position, velocity, mass, elasticity, err := massNumericFields([]string{"mass", "1", "1", "2", "bad", "0", "3", "4"}); err == nil || position != (sim.Vec2{}) || velocity != (sim.Vec2{}) || mass != 0 || elasticity != 0 {
		t.Fatalf("mass velocity x parse = %v, %v, %v, %v, %v", position, velocity, mass, elasticity, err)
	}
	if position, velocity, mass, elasticity, err := massNumericFields([]string{"mass", "1", "1", "2", "0", "bad", "3", "4"}); err == nil || position != (sim.Vec2{}) || velocity != (sim.Vec2{}) || mass != 0 || elasticity != 0 {
		t.Fatalf("mass velocity y parse = %v, %v, %v, %v, %v", position, velocity, mass, elasticity, err)
	}
	if position, velocity, mass, elasticity, err := massNumericFields([]string{"mass", "1", "1", "2", "bad", "4"}); err == nil || position != (sim.Vec2{}) || velocity != (sim.Vec2{}) || mass != 0 || elasticity != 0 {
		t.Fatalf("mass value parse = %v, %v, %v, %v, %v", position, velocity, mass, elasticity, err)
	}
	if position, velocity, mass, elasticity, err := massNumericFields([]string{"mass", "1", "1", "2", "3", "bad"}); err == nil || position != (sim.Vec2{}) || velocity != (sim.Vec2{}) || mass != 0 || elasticity != 0 {
		t.Fatalf("mass elasticity parse = %v, %v, %v, %v, %v", position, velocity, mass, elasticity, err)
	}
	if mass, elasticity, err := massValueFields([]string{"bad", "4"}); err == nil || mass != 0 || elasticity != 0 {
		t.Fatalf("mass value fields parse = %v, %v, %v", mass, elasticity, err)
	}
	if id, err := positiveIDField("bad", "id"); err == nil || id != 0 {
		t.Fatalf("positive id parse = %v, %v", id, err)
	}
	if id, err := positiveIDField("0", "id"); err == nil || id != 0 {
		t.Fatalf("positive id value = %v, %v", id, err)
	}
	if value, err := intField("bad", "id"); err == nil || value != 0 {
		t.Fatalf("int parse = %v, %v", value, err)
	}
	if value, err := floatField("bad", "float"); err == nil || value != 0 {
		t.Fatalf("float parse = %v, %v", value, err)
	}
	if value, err := booleanField("bad", "bool"); err == nil || value != "" {
		t.Fatalf("boolean parse = %q, %v", value, err)
	}
	if value, err := booleanField("00", "bool"); err != nil || value != "false" {
		t.Fatalf("numeric false parse = %q, %v", value, err)
	}
}

func TestResolveXSPFilenameAddsExtensionAndSpringDir(t *testing.T) {
	cases := []struct {
		name      string
		filename  string
		springDir string
		want      string
	}{
		{"extension", "demo", "", "demo.xsp"},
		{"existing extension", "demo.xsp", "", "demo.xsp"},
		{"springdir", "demo", "examples", "examples/demo.xsp"},
		{"springdir existing extension", "demo.xsp", "examples", "examples/demo.xsp"},
		{"absolute path", "/tmp/demo", "examples", "/tmp/demo.xsp"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ResolveXSPFilename(tc.filename, tc.springDir); got != tc.want {
				t.Fatalf("ResolveXSPFilename = %q, want %q", got, tc.want)
			}
		})
	}
}

func assertParameter(t *testing.T, world *sim.Simulation, name, value string) {
	t.Helper()
	if got := world.Parameters.Value(name); got != value {
		t.Fatalf("%s = %q", name, got)
	}
}

func assertLoadXSPErrorContains(t *testing.T, text, want string) {
	t.Helper()
	_, err := LoadXSP(text)
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("error = %v, want %s context", err, want)
	}
}
