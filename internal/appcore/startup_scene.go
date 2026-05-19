package appcore

import (
	"os"
	"path/filepath"

	xspfmt "springs/internal/format"
	"springs/internal/sim"
)

const defaultStartupScenePath = "demos/pendulum.xsp"

func DefaultStartupScenePath() string {
	return defaultStartupScenePath
}

func NewDefaultStartupWorld(bounds sim.Bounds) *sim.Simulation {
	world, err := LoadDefaultStartupWorld(bounds)
	if err == nil {
		return world
	}
	world = sim.NewDemoSimulation()
	ApplyBounds(world, bounds)
	return world
}

func LoadDefaultStartupWorld(bounds sim.Bounds) (*sim.Simulation, error) {
	var lastErr error
	for _, path := range DefaultStartupSceneCandidates() {
		content, err := os.ReadFile(path)
		if err != nil {
			lastErr = err
			continue
		}
		world, err := xspfmt.LoadXSP(string(content))
		if err != nil {
			lastErr = err
			continue
		}
		ApplyBounds(world, bounds)
		return world, nil
	}
	return nil, lastErr
}

func DefaultStartupSceneCandidates() []string {
	return []string{
		defaultStartupScenePath,
		filepath.Join("..", "..", defaultStartupScenePath),
	}
}

func ApplyBounds(world *sim.Simulation, bounds sim.Bounds) {
	world.Bounds = bounds
}

// mutate4go-manifest-begin
// {"version":1,"tested_at":"2026-05-19T10:17:55-05:00","module_hash":"b10025a18d0358f95094432ca39328b89999007da1ab31a4122b83ed0e256541","functions":[{"id":"func/DefaultStartupScenePath","name":"DefaultStartupScenePath","line":13,"end_line":15,"hash":"d2e19b195cb80748a095ea044a67b8732738c4a82ca7883760eeb46ff4441c4a"},{"id":"func/NewDefaultStartupWorld","name":"NewDefaultStartupWorld","line":17,"end_line":25,"hash":"0b47a6982aaf4cf1421f142df273bfe84b6acd18ad5a60e2e5ee28e8691dfdec"},{"id":"func/LoadDefaultStartupWorld","name":"LoadDefaultStartupWorld","line":27,"end_line":44,"hash":"e3be10d423845412aeb9325754dd5dda0b02b36de9587ae7261d4c307fd38cbb"},{"id":"func/DefaultStartupSceneCandidates","name":"DefaultStartupSceneCandidates","line":46,"end_line":51,"hash":"fb76ed7328d1b240d31152395b7448b650340481463c345238963f24ec11d827"},{"id":"func/ApplyBounds","name":"ApplyBounds","line":53,"end_line":55,"hash":"ee7b7740bf446308fc8d19b0b63b1e0da7c708aa9b2f0d55856c41e68eb3134b"}]}
// mutate4go-manifest-end
