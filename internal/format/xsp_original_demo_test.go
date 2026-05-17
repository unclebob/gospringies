package format

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadXSPAcceptsOriginalXSpringiesLineForms(t *testing.T) {
	world, err := LoadXSP(strings.Join([]string{
		"#1.0 *** XSpringies data file",
		"cmas 0.4",
		"elas 1.0",
		"kspr 10.0",
		"kdmp 5.0",
		"fixm 0",
		"shws 1",
		"cent -1",
		"frce 0 1 10.0 90.0",
		"frce 1 0 5.0 2.0",
		"frce 2 0 3.0 0.4",
		"frce 3 0 100.0 1.0",
		"visc 0.0",
		"stck 0.0",
		"step 0.05",
		"prec 1.0",
		"adpt 0",
		"gsnp 20.0 0",
		"wall 1 0 1 0",
		"mass 1 10.0 20.0 1.5 -2.5 -1.0 0.8",
		"mass 2 30.0 20.0 0.0 0.0 1.0 0.7",
		"spng 1 1 2 100.0 5.0 20.0",
		"",
	}, "\n"))
	if err != nil {
		t.Fatalf("LoadXSP returned error: %v", err)
	}
	if len(world.Masses) != 2 || len(world.Springs) != 1 {
		t.Fatalf("world = %#v", world)
	}
	if mass := world.Masses[0]; !mass.Fixed || mass.Velocity.X != 1.5 || mass.Velocity.Y != -2.5 {
		t.Fatalf("legacy mass = %#v", mass)
	}
	if enabled, _ := world.Parameters.WallEnabled("top"); !enabled {
		t.Fatalf("walls = %#v", world.Parameters.Walls)
	}
}

func TestOriginalDemoCorpusLoads(t *testing.T) {
	entries, err := os.ReadDir(originalDemoCorpusDir())
	if err != nil {
		t.Fatalf("read demo corpus: %v", err)
	}
	demoCount := 0
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".xsp" {
			demoCount++
		}
	}
	if demoCount != 67 {
		t.Fatalf("demo count = %d, want 67", demoCount)
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".xsp" {
			continue
		}
		content, err := os.ReadFile(filepath.Join(originalDemoCorpusDir(), entry.Name()))
		if err != nil {
			t.Fatalf("read %s: %v", entry.Name(), err)
		}
		if _, err := LoadXSP(string(content)); err != nil {
			t.Fatalf("load %s: %v", entry.Name(), err)
		}
	}
}

func originalDemoCorpusDir() string {
	return filepath.Join("..", "..", "demos", "original")
}
