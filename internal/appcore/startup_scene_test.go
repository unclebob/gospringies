package appcore

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	xspfmt "springs/internal/format"
	"springs/internal/sim"
)

var testBounds = sim.Bounds{Width: 1700, Height: 1000}

func TestNewDefaultStartupWorldLoadsPendulumDemo(t *testing.T) {
	world := NewDefaultStartupWorld(testBounds)
	expected := loadAppCoreTestXSP(t, filepath.Join("..", "..", DefaultStartupScenePath()))

	if !reflect.DeepEqual(world, expected) {
		t.Fatalf("startup world = %#v, want %#v", world, expected)
	}
}

func TestNewDefaultStartupWorldStartsWithNonblankStarterWorld(t *testing.T) {
	world := NewDefaultStartupWorld(testBounds)

	if len(world.Masses) == 0 {
		t.Fatal("starter world should include masses")
	}
	if len(world.Springs) == 0 {
		t.Fatal("starter world should include springs")
	}
}

func TestStartupPendulumEnablesGravity(t *testing.T) {
	world := NewDefaultStartupWorld(testBounds)

	force, ok := world.Parameters.Force("gravity")
	if !ok {
		t.Fatal("missing gravity force")
	}
	if force.Enabled != "true" || force.Values["magnitude"] != "10" || force.Values["direction"] != "0" {
		t.Fatalf("gravity force = %#v", force)
	}
}

func TestApplyBoundsSetsSimulationBounds(t *testing.T) {
	world := sim.NewWorld()
	ApplyBounds(world, testBounds)

	if world.Bounds != testBounds {
		t.Fatalf("bounds = %#v, want %#v", world.Bounds, testBounds)
	}
}

func TestDefaultStartupSceneCandidatesIncludesRepoAndPackagePaths(t *testing.T) {
	got := DefaultStartupSceneCandidates()
	want := []string{
		DefaultStartupScenePath(),
		filepath.Join("..", "..", DefaultStartupScenePath()),
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("candidates = %#v, want %#v", got, want)
	}
}

func loadAppCoreTestXSP(t *testing.T, path string) *sim.Simulation {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	world, err := xspfmt.LoadXSP(string(content))
	if err != nil {
		t.Fatalf("load %s: %v", path, err)
	}
	ApplyBounds(world, testBounds)
	return world
}
