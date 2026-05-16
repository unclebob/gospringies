package acceptance

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	xspfmt "springs/internal/format"
	"springs/internal/sim"
)

func assertDemoFileAdded(_ *world, example map[string]string) error {
	name, err := stringValue(example, "demo_file")
	if err != nil {
		return err
	}
	for _, demo := range demoFileNames() {
		if name == demo {
			return nil
		}
	}
	return fmt.Errorf("unknown demo file %q", name)
}

func assertDemoFileValid(_ *world, example map[string]string) error {
	content, err := readDemoFile(example)
	if err != nil {
		return err
	}
	_, err = xspfmt.LoadXSP(content)
	return err
}

func assertDemoFileHumanReadable(_ *world, example map[string]string) error {
	content, err := readDemoFile(example)
	if err != nil {
		return err
	}
	if err := assertDemoEndsWithNewline(content); err != nil {
		return err
	}
	return assertDemoLinesReadable(content)
}

func assertDemoEndsWithNewline(content string) error {
	if !strings.HasSuffix(content, "\n") {
		return fmt.Errorf("demo file must end with newline")
	}
	return nil
}

func assertDemoLinesReadable(content string) error {
	for i, line := range strings.Split(strings.TrimSuffix(content, "\n"), "\n") {
		if strings.TrimSpace(line) == "" {
			return fmt.Errorf("demo file contains blank line %d", i+1)
		}
		if strings.TrimSpace(line) != line {
			return fmt.Errorf("demo file line %d has surrounding whitespace", i+1)
		}
	}
	return nil
}

func assertDemoFileExists(_ *world, example map[string]string) error {
	_, err := readDemoFile(example)
	return err
}

func loadDemoFile(w *world, example map[string]string) error {
	content, err := readDemoFile(example)
	if err != nil {
		return err
	}
	loaded, err := xspfmt.LoadXSP(content)
	if err != nil {
		return err
	}
	w.xspWorld = loaded
	return nil
}

func assertDemoLoadedFeature(w *world, example map[string]string) error {
	feature, err := stringValue(example, "required_feature")
	if err != nil {
		return err
	}
	switch feature {
	case "fixed mass":
		return assertDemoHasFixedMass(w.xspWorld)
	case "multiple springs":
		return assertDemoHasMultipleSprings(w.xspWorld)
	default:
		return fmt.Errorf("unsupported demo feature %q", feature)
	}
}

func readDemoFile(example map[string]string) (string, error) {
	path, err := demoFilePath(example)
	if err != nil {
		return "", err
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func demoFilePath(example map[string]string) (string, error) {
	name, err := stringValue(example, "demo_file")
	if err != nil {
		return "", err
	}
	root, err := repoRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "demos", name), nil
}

func demoFileNames() []string {
	return []string{"pendulum.xsp", "spring-chain.xsp", "small-mesh.xsp"}
}

func assertDemoHasFixedMass(world *sim.Simulation) error {
	if world == nil {
		return fmt.Errorf("demo world was not loaded")
	}
	for _, mass := range world.Masses {
		if mass.Fixed {
			return nil
		}
	}
	return fmt.Errorf("demo world has no fixed mass")
}

func assertDemoHasMultipleSprings(world *sim.Simulation) error {
	if world == nil {
		return fmt.Errorf("demo world was not loaded")
	}
	if len(world.Springs) < 2 {
		return fmt.Errorf("demo world has %d springs, want multiple", len(world.Springs))
	}
	return nil
}
