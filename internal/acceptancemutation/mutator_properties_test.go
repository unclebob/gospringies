//go:build property

package acceptancemutation

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"testing/quick"
	"time"

	"springs/internal/gherkin"
)

func TestPropertyMutationBuildersAndSummariesAreDeterministic(t *testing.T) {
	checkProperty(t, 1, 300, mutationBuildersAndSummariesAreDeterministic)
}

func TestPropertyScenarioManifestRoundTripsAndSkipPlansAreStable(t *testing.T) {
	checkProperty(t, 2, 300, scenarioManifestRoundTripsAndSkipPlansAreStable)
}

func checkProperty(t *testing.T, seed int64, maxCount int, property any) {
	t.Helper()
	config := &quick.Config{MaxCount: maxCount, Rand: rand.New(rand.NewSource(seed))}
	if err := quick.Check(property, config); err != nil {
		t.Fatal(err)
	}
}

func mutationBuildersAndSummariesAreDeterministic(input float64) bool {
	value := strconv.Itoa(int(propertyFloat(input, 1, 100000)))
	feature := propertyFeature(value)
	first := BuildMutations(feature)
	second := BuildMutations(feature)
	if !reflect.DeepEqual(first, second) {
		panic("BuildMutations is not deterministic")
	}
	if len(first) == 0 {
		panic("expected mutations")
	}
	for _, mutation := range first {
		if mutation.Mutated == mutation.Original || mutation.ID == "" || mutation.Path == "" {
			panic(fmt.Sprintf("invalid mutation: %#v", mutation))
		}
		if mutateValue(mutation.Path, mutation.Original) != mutation.Mutated {
			panic("mutateValue is not deterministic")
		}
	}
	filtered := filterMutations(first, func(m Mutation) bool { return strings.HasSuffix(m.ID, "1") || m.Key == "flag" })
	for _, mutation := range filtered {
		if !(strings.HasSuffix(mutation.ID, "1") || mutation.Key == "flag") {
			panic("filterMutations kept wrong mutation")
		}
	}
	if mutationWorkerCount(0, 0) != 0 || mutationWorkerCount(100, len(first)) != len(first) {
		panic("mutationWorkerCount bounds failed")
	}
	results := []MutationResult{
		{Mutation: first[0], Status: MutationKilled},
		{Mutation: first[0], Status: MutationSurvived},
		{Mutation: first[0], Status: MutationError},
	}
	summary := Summarize(results)
	if summary.Total != 3 || summary.Killed != 1 || summary.Survived != 1 || summary.Errors != 1 {
		panic(fmt.Sprintf("summary mismatch: %#v", summary))
	}
	cloned := cloneFeature(feature)
	cloned.Scenarios[0].Examples[0]["number"] = "changed"
	if feature.Scenarios[0].Examples[0]["number"] == "changed" {
		panic("cloneFeature aliases examples")
	}
	if mutated, ok := mutateNumber(value, deterministicRand("number", value)); !ok || mutated == value {
		panic("mutateNumber failed")
	}
	if mutated, ok := mutateKeyword("true", "true", deterministicRand("keyword")); !ok || mutated != "false" {
		panic("mutateKeyword failed")
	}
	if mutated := mutateList("list", "1, 2, 3", deterministicRand("list")); mutated == "1, 2, 3" || !strings.Contains(mutated, ", ") {
		panic("mutateList failed")
	}
	if signedIntDelta(deterministicRand("int")) == 0 || signedFloatDelta(deterministicRand("float")) == 0 {
		panic("signed deltas must be non-zero")
	}
	if mutated, ok := mutateDate("2026-05-23", deterministicRand("date")); !ok || mutated == "2026-05-23" {
		panic("mutateDate failed")
	}
	if mutated, ok := mutateDuration("10s", deterministicRand("duration")); !ok || mutated == "10s" {
		panic("mutateDuration failed")
	}
	if dither("abc", deterministicRand("dither")) == "abc" || dither("", deterministicRand("empty")) != "x" {
		panic("dither failed")
	}
	return true
}

func scenarioManifestRoundTripsAndSkipPlansAreStable(input float64) bool {
	value := strconv.Itoa(int(propertyFloat(input, 1, 100000)))
	feature := propertyFeature(value)
	results := []MutationResult{
		{Mutation: Mutation{Scenario: 0}, Status: MutationKilled},
	}
	now := time.Unix(int64(propertyFloat(input, 1, 100000)), 0).UTC()
	manifest := BuildScenarioManifest("feature.feature", feature, ScenarioManifest{}, ScenarioSkipPlan{SkipScenarios: map[int]bool{}}, results, "impl", now)
	block := ScenarioManifestBegin + "\n# " + mustJSONManifest(manifest) + "\n" + ScenarioManifestEnd + "\nFeature: F\n"
	parsed, ok, err := ParseScenarioManifest(block)
	if err != nil || !ok || !reflect.DeepEqual(parsed, manifest) {
		panic(fmt.Sprintf("manifest parse mismatch: %#v ok=%v err=%v", parsed, ok, err))
	}
	removed := RemoveScenarioManifest(block)
	if strings.Contains(removed, ScenarioManifestBegin) || !strings.Contains(removed, "Feature: F") {
		panic("RemoveScenarioManifest failed")
	}
	plan := ScenarioSkipPlanFor(feature, "feature.feature", manifest, "impl")
	if !plan.SkipScenarios[0] || plan.SkippedScenarios != 1 {
		panic(fmt.Sprintf("skip plan mismatch: %#v", plan))
	}
	softPlan := ScenarioSkipPlanForMode(feature, "feature.feature", manifest, "different", ScenarioManifestSoft)
	if !softPlan.SkipScenarios[0] {
		panic("soft skip plan should ignore implementation hash")
	}
	fullPlan := ScenarioSkipPlanForMode(feature, "feature.feature", manifest, "impl", ScenarioManifestFull)
	if len(fullPlan.SkipScenarios) != 0 {
		panic("full skip plan should skip nothing")
	}
	rebuilt := BuildScenarioManifest("feature.feature", feature, manifest, plan, nil, "impl2", now.Add(time.Second))
	if !reflect.DeepEqual(rebuilt.Scenarios[0], manifest.Scenarios[0]) {
		panic("BuildScenarioManifest did not preserve skipped scenario entry")
	}
	if BackgroundHash(feature) != BackgroundHash(feature) || ScenarioHash(feature.Scenarios[0]) != ScenarioHash(feature.Scenarios[0]) {
		panic("hash helpers are not deterministic")
	}
	return true
}

func propertyFeature(value string) gherkin.Feature {
	return gherkin.Feature{
		Name:       "Property Feature",
		Background: []gherkin.Step{{Keyword: "Given", Text: "setup"}},
		Scenarios: []gherkin.Scenario{{
			Name:  "scenario",
			Steps: []gherkin.Step{{Keyword: "Then", Text: "value <number>"}},
			Examples: []map[string]string{{
				"number":   value,
				"flag":     "true",
				"date":     "2026-05-23",
				"duration": "10s",
				"list":     "1, 2, 3",
			}},
		}},
	}
}

func mustJSONManifest(manifest ScenarioManifest) string {
	data, err := jsonMarshal(manifest)
	if err != nil {
		panic(err)
	}
	return string(data)
}

var jsonMarshal = func(v any) ([]byte, error) {
	return json.Marshal(v)
}

func propertyFloat(input float64, minimum float64, maximum float64) float64 {
	if math.IsNaN(input) || math.IsInf(input, 0) {
		return minimum
	}
	return minimum + math.Mod(math.Abs(input), maximum-minimum)
}
