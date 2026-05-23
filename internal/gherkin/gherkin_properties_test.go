//go:build property

package gherkin

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"testing/quick"
)

func TestPropertyParseJSONTableAndStepRoundTrips(t *testing.T) {
	checkProperty(t, 1, 300, parseJSONTableAndStepRoundTrips)
}

func checkProperty(t *testing.T, seed int64, maxCount int, property any) {
	t.Helper()
	config := &quick.Config{MaxCount: maxCount, Rand: rand.New(rand.NewSource(seed))}
	if err := quick.Check(property, config); err != nil {
		t.Fatal(err)
	}
}

func parseJSONTableAndStepRoundTrips(nameInput, valueInput float64) bool {
	name := fmt.Sprintf("sample-%d", int(propertyFloat(nameInput, 1, 1000)))
	value := fmt.Sprintf("%d", int(propertyFloat(valueInput, 1, 1000)))
	text := fmt.Sprintf(`Feature: %s
Background:
  Given setup <id>
Scenario Outline: example
  When value <value>
  Then result <id>
Examples:
  | id | value |
  | %s | %s |
`, name, value, value)
	feature, err := Parse(strings.NewReader(text))
	if err != nil {
		panic(err)
	}
	if feature.Name != name || len(feature.Background) != 1 || len(feature.Scenarios) != 1 {
		panic(fmt.Sprintf("parsed feature mismatch: %#v", feature))
	}
	step := parseStep("Given setup <id>")
	if !isStep("Given setup <id>") || step.Keyword != "Given" || step.Text != "setup <id>" || !reflect.DeepEqual(step.Parameters, []string{"id"}) {
		panic(fmt.Sprintf("step parse mismatch: %#v", step))
	}
	row := parseTableRow("| id | value |")
	if !reflect.DeepEqual(row, []string{"id", "value"}) {
		panic(fmt.Sprintf("table row mismatch: %#v", row))
	}
	dir, err := os.MkdirTemp("", "gherkin-property-*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)
	featurePath := filepath.Join(dir, "feature.feature")
	jsonPath := filepath.Join(dir, "feature.json")
	if err := os.WriteFile(featurePath, []byte(text), 0o644); err != nil {
		panic(err)
	}
	fromFile, err := ReadFile(featurePath)
	if err != nil {
		panic(err)
	}
	if !reflect.DeepEqual(fromFile, feature) {
		panic("ReadFile did not match Parse")
	}
	if err := WriteJSON(feature, jsonPath); err != nil {
		panic(err)
	}
	fromJSON, err := ReadJSON(jsonPath)
	if err != nil {
		panic(err)
	}
	if !reflect.DeepEqual(fromJSON, feature) {
		panic(fmt.Sprintf("json round trip mismatch: %#v != %#v", fromJSON, feature))
	}
	return true
}

func propertyFloat(input float64, minimum float64, maximum float64) float64 {
	if math.IsNaN(input) || math.IsInf(input, 0) {
		return minimum
	}
	return minimum + math.Mod(math.Abs(input), maximum-minimum)
}
