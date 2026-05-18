package app

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

func newDefaultStartupWorld() *sim.Simulation {
	world, err := loadDefaultStartupWorld()
	if err == nil {
		return world
	}
	return sim.NewDemoSimulation()
}

func loadDefaultStartupWorld() (*sim.Simulation, error) {
	var lastErr error
	for _, path := range defaultStartupSceneCandidates() {
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
		setAppBounds(world)
		return world, nil
	}
	return nil, lastErr
}

func defaultStartupSceneCandidates() []string {
	return []string{
		defaultStartupScenePath,
		filepath.Join("..", "..", defaultStartupScenePath),
	}
}
