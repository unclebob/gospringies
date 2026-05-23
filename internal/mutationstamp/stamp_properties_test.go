//go:build property

package mutationstamp

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"testing"
	"testing/quick"
)

func TestPropertySplitHashAndStampFormattingRoundTrip(t *testing.T) {
	checkProperty(t, 1, 500, splitHashAndStampFormattingRoundTrip)
}

func checkProperty(t *testing.T, seed int64, maxCount int, property any) {
	t.Helper()
	config := &quick.Config{MaxCount: maxCount, Rand: rand.New(rand.NewSource(seed))}
	if err := quick.Check(property, config); err != nil {
		t.Fatal(err)
	}
}

func splitHashAndStampFormattingRoundTrip(input float64) bool {
	body := fmt.Sprintf("Feature: F\nScenario: S\nThen value %d\n", int(propertyFloat(input, 1, 100000)))
	hash := Hash(body)
	stamped := Prefix + formatStamp(hash) + " extra metadata\n" + body
	stamp, unstamped := Split(stamped)
	if unstamped != body {
		panic("Split did not remove only the stamp line")
	}
	if stampHash(stamp) != hash {
		panic("stampHash(formatStamp(hash)) did not recover hash")
	}
	if Hash(unstamped) != hash {
		panic("Hash changed after Split")
	}
	_, second := Split(unstamped)
	if second != body {
		panic("Split without stamp changed content")
	}
	if Hash(body+"x") == hash {
		panic("Hash did not change for changed content")
	}
	if strings.Contains(unstamped, Prefix) {
		panic("unstamped content still contains stamp prefix")
	}
	return true
}

func propertyFloat(input float64, minimum float64, maximum float64) float64 {
	if math.IsNaN(input) || math.IsInf(input, 0) {
		return minimum
	}
	return minimum + math.Mod(math.Abs(input), maximum-minimum)
}
