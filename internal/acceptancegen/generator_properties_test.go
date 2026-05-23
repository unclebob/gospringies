//go:build property

package acceptancegen

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"testing"
	"testing/quick"
	"unicode"
)

func TestPropertyGeneratedTestNameIsDeterministicAndValid(t *testing.T) {
	checkProperty(t, 1, 500, generatedTestNameIsDeterministicAndValid)
}

func checkProperty(t *testing.T, seed int64, maxCount int, property any) {
	t.Helper()
	config := &quick.Config{MaxCount: maxCount, Rand: rand.New(rand.NewSource(seed))}
	if err := quick.Check(property, config); err != nil {
		t.Fatal(err)
	}
}

func generatedTestNameIsDeterministicAndValid(input float64) bool {
	n := int(propertyFloat(input, 1, 100000))
	path := fmt.Sprintf("target/some-feature_%d.acceptance-test.go", n)
	first := generatedTestName(path)
	second := generatedTestName(path)
	if first != second {
		panic("generated test name is not deterministic")
	}
	if !strings.HasPrefix(first, "TestGeneratedAcceptance") {
		panic(fmt.Sprintf("generated test name prefix mismatch: %q", first))
	}
	for _, r := range strings.TrimPrefix(first, "TestGeneratedAcceptance") {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			panic(fmt.Sprintf("generated test name contains non-identifier rune: %q", first))
		}
	}
	if generatedTestName("a-b.go") == generatedTestName("ab.go") {
		panic("common separated path variants collided")
	}
	return true
}

func propertyFloat(input float64, minimum float64, maximum float64) float64 {
	if math.IsNaN(input) || math.IsInf(input, 0) {
		return minimum
	}
	return minimum + math.Mod(math.Abs(input), maximum-minimum)
}
