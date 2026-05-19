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

// mutate4go-manifest-begin
// {"version":1,"tested_at":"2026-05-19T07:18:20-05:00","module_hash":"36204486c3e2e689e6b6ba829efe34c120e09f59be32006172de581607af4ebb","functions":[{"id":"func/assertDemoFileAdded","name":"assertDemoFileAdded","line":13,"end_line":24,"hash":"d5ef86b751c2f42049ffc0954b5908abe9fb2f2c95d4763087cb2b9da2c9f99d"},{"id":"func/assertDemoFileValid","name":"assertDemoFileValid","line":26,"end_line":33,"hash":"253eaa92eeea108633d36dfe45473e61697e65711a41ee47ffeb9d942c8cd5be"},{"id":"func/assertDemoFileHumanReadable","name":"assertDemoFileHumanReadable","line":35,"end_line":44,"hash":"40d9c0ee90f25429fb8858c4f2dd39b3d09b39a17e55c5443d910cb0cbfe1002"},{"id":"func/assertDemoEndsWithNewline","name":"assertDemoEndsWithNewline","line":46,"end_line":51,"hash":"3f970c047b2d5e48fe5482f7ab63d77a51783c04ec18799d334a2113c9db9af2"},{"id":"func/assertDemoLinesReadable","name":"assertDemoLinesReadable","line":53,"end_line":63,"hash":"0990dc89ac14b79286f5b6209909434726787637e0c9fa49cc383e9e2888969d"},{"id":"func/assertDemoFileExists","name":"assertDemoFileExists","line":65,"end_line":68,"hash":"6b4cbad51417ca86723539f3c765ea3b3b7840346552d718c89532f055ae8dc2"},{"id":"func/loadDemoFile","name":"loadDemoFile","line":70,"end_line":81,"hash":"239ef9d5e9c84540d7190f5a8994211b8e7a7ebee51e729b059c490f1368d466"},{"id":"func/assertDemoLoadedFeature","name":"assertDemoLoadedFeature","line":83,"end_line":96,"hash":"3c10e5ef5088c12aa3569d45a7fd6c3d8cca85d101063c19c71fc44a225dae46"},{"id":"func/readDemoFile","name":"readDemoFile","line":98,"end_line":108,"hash":"2af265b11f9f054cb6139b3bb1b215025fc3ccabc77c8bd7df89152fa8af33c7"},{"id":"func/demoFilePath","name":"demoFilePath","line":110,"end_line":120,"hash":"09c3347929c291714e2671eb6872882a890ca4bf7fca5f908a6c3633c437bfd0"},{"id":"func/demoFileNames","name":"demoFileNames","line":122,"end_line":124,"hash":"aac2468cfdc6d04f06716de4f718e9d509e77a3dcb03cbcbb770d63a8467b30c"},{"id":"func/assertDemoHasFixedMass","name":"assertDemoHasFixedMass","line":126,"end_line":136,"hash":"f6f1a5f436f01c64b39f9ec7cd69b443befe173be4baaf1fdb9ca1fa6696ef57"},{"id":"func/assertDemoHasMultipleSprings","name":"assertDemoHasMultipleSprings","line":138,"end_line":146,"hash":"3970ce4716f6789af90f07eb09f4a9d030882636804cc916fa3173ccb4a4fb6e"}]}
// mutate4go-manifest-end
