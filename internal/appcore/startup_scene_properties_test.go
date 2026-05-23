//go:build property

package appcore

import (
	"math"
	"math/rand"
	"path/filepath"
	"testing"
	"testing/quick"

	"springs/internal/sim"
)

func TestPropertyStartupSceneHelpersAreDeterministicAndApplyBounds(t *testing.T) {
	config := &quick.Config{MaxCount: 300, Rand: rand.New(rand.NewSource(1))}
	if err := quick.Check(startupSceneHelpersAreDeterministicAndApplyBounds, config); err != nil {
		t.Fatal(err)
	}
}

func startupSceneHelpersAreDeterministicAndApplyBounds(widthInput, heightInput float64) bool {
	bounds := sim.Bounds{Width: propertyFloat(widthInput, 1, 2000), Height: propertyFloat(heightInput, 1, 2000)}
	world := sim.NewWorld()
	ApplyBounds(world, bounds)
	if world.Bounds != bounds {
		panic("ApplyBounds did not set exact bounds")
	}
	if DefaultStartupScenePath() != "demos/pendulum.xsp" {
		panic("default startup scene path changed")
	}
	candidates := DefaultStartupSceneCandidates()
	if len(candidates) != 2 || candidates[0] != DefaultStartupScenePath() || candidates[1] != filepath.Join("..", "..", DefaultStartupScenePath()) {
		panic("startup scene candidates are not deterministic")
	}
	defaultWorld := NewDefaultStartupWorld(bounds)
	if defaultWorld.Bounds != bounds {
		panic("default startup world did not apply bounds")
	}
	return true
}

func propertyFloat(input float64, minimum float64, maximum float64) float64 {
	if math.IsNaN(input) || math.IsInf(input, 0) {
		return minimum
	}
	return minimum + math.Mod(math.Abs(input), maximum-minimum)
}
