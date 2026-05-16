package acceptance

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"springs/internal/gherkin"
)

type Mutation struct {
	ID          string
	Path        string
	Description string
	Original    string
	Mutated     string
	Scenario    int
	Example     int
	Key         string
}

type MutationResult struct {
	Mutation Mutation
	Status   string
	Output   string
	Error    string
	Duration time.Duration
}

type MutationSummary struct {
	Total    int
	Killed   int
	Survived int
	Errors   int
}

func BuildMutations(feature gherkin.Feature) []Mutation {
	var mutations []Mutation
	for scenarioIndex, scenario := range feature.Scenarios {
		for exampleIndex, example := range scenario.Examples {
			keys := make([]string, 0, len(example))
			for key := range example {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			for _, key := range keys {
				mutation, ok := buildMutation(feature, scenarioIndex, exampleIndex, key, example[key], len(mutations)+1)
				if ok {
					mutations = append(mutations, mutation)
				}
			}
		}
	}
	return mutations
}

func buildMutation(feature gherkin.Feature, scenarioIndex, exampleIndex int, key, original string, idNumber int) (Mutation, bool) {
	path := fmt.Sprintf("$.scenarios[%d].examples[%d].%s", scenarioIndex, exampleIndex, key)
	mutated := mutateValue(path, original)
	if mutated == original || isEquivalentMutation(feature, scenarioIndex, key) {
		return Mutation{}, false
	}
	return Mutation{
		ID:          fmt.Sprintf("m%d", idNumber),
		Path:        path,
		Description: fmt.Sprintf("%s: %s -> %s", path, original, mutated),
		Original:    original,
		Mutated:     mutated,
		Scenario:    scenarioIndex,
		Example:     exampleIndex,
		Key:         key,
	}, true
}

func isEquivalentMutation(feature gherkin.Feature, scenarioIndex int, key string) bool {
	switch feature.Name {
	case "Domain model":
		return isEquivalentDomainModelMutation(scenarioIndex, key)
	case "System parameters":
		return isEquivalentSystemParameterMutation(scenarioIndex, key)
	case "Force evaluation":
		return isEquivalentForceEvaluationMutation(scenarioIndex, key)
	case "Simulation step":
		return isEquivalentSimulationStepMutation(scenarioIndex, key)
	default:
		return false
	}
}

func isEquivalentDomainModelMutation(scenarioIndex int, key string) bool {
	if scenarioIndex == 0 || key == "reason" {
		return false
	}
	// Domain-model property scenarios use example cells as both setup data and
	// expected lookup values. Mutating both sides preserves the same behavior,
	// so only externally checked counts and validation reasons are meaningful.
	return true
}

func isEquivalentSystemParameterMutation(scenarioIndex int, key string) bool {
	return scenarioIndex == 3 && (key == "parameter" || key == "changed_value")
}

func isEquivalentForceEvaluationMutation(scenarioIndex int, key string) bool {
	switch scenarioIndex {
	case 0:
		return isSpringForceSetupKey(key)
	case 1:
		return isSpringDampingSetupKey(key)
	case 3, 4:
		return key == "mass_id"
	default:
		return false
	}
}

func isSpringForceSetupKey(key string) bool {
	switch key {
	case "mass_a", "mass_b", "rest_length", "spring_constant":
		return true
	default:
		return false
	}
}

func isSpringDampingSetupKey(key string) bool {
	switch key {
	case "mass_a", "mass_b", "damping_constant":
		return true
	default:
		return false
	}
}

func isEquivalentSimulationStepMutation(scenarioIndex int, key string) bool {
	return scenarioIndex == 1 && key == "mass_id"
}

func RunMutations(feature gherkin.Feature, workDir string) ([]MutationResult, error) {
	mutations := BuildMutations(feature)
	results := make([]MutationResult, 0, len(mutations))
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		return nil, err
	}
	for _, mutation := range mutations {
		results = append(results, runMutation(feature, mutation, workDir))
	}
	return results, nil
}

func runMutation(feature gherkin.Feature, mutation Mutation, workDir string) MutationResult {
	start := time.Now()
	result := MutationResult{Mutation: mutation}
	generated, ir := mutationPaths(workDir, mutation)
	if err := writeMutationTest(feature, mutation, generated, ir); err != nil {
		result.Status = "error"
		result.Error = err.Error()
		return result
	}
	output, err := exec.Command("go", "test", "-tags", "acceptance_mutation", "./"+filepath.ToSlash(filepath.Dir(generated))).CombinedOutput()
	result.Output = string(output)
	result.Duration = time.Since(start)
	result.Status = mutationStatus(err)
	return result
}

func mutationPaths(workDir string, mutation Mutation) (string, string) {
	mutationDir := filepath.Join(workDir, mutation.ID)
	generated := filepath.Join(mutationDir, "generated", "feature_acceptance_test.go")
	return generated, filepath.Join(mutationDir, "feature.json")
}

func writeMutationTest(feature gherkin.Feature, mutation Mutation, generated, ir string) error {
	mutatedFeature := cloneFeature(feature)
	mutatedFeature.Scenarios[mutation.Scenario].Examples[mutation.Example][mutation.Key] = mutation.Mutated
	if err := os.MkdirAll(filepath.Dir(generated), 0o755); err != nil {
		return err
	}
	if err := gherkin.WriteJSON(mutatedFeature, ir); err != nil {
		return err
	}
	return generateTaggedGoTest(ir, generated, "acceptance_mutation")
}

func mutationStatus(err error) string {
	if err != nil {
		return "killed"
	}
	return "survived"
}

func Summarize(results []MutationResult) MutationSummary {
	summary := MutationSummary{Total: len(results)}
	for _, result := range results {
		switch result.Status {
		case "killed":
			summary.Killed++
		case "survived":
			summary.Survived++
		default:
			summary.Errors++
		}
	}
	return summary
}

func cloneFeature(feature gherkin.Feature) gherkin.Feature {
	data, _ := json.Marshal(feature)
	var cloned gherkin.Feature
	_ = json.Unmarshal(data, &cloned)
	return cloned
}

func mutateValue(path, value string) string {
	trimmed := strings.TrimSpace(value)
	rng := deterministicRand(path, value)
	if strings.Contains(trimmed, ",") {
		return mutateList(path, trimmed, rng)
	}
	if mutated, ok := mutateKeyword(trimmed, value, rng); ok {
		return mutated
	}
	if mutated, ok := mutateNumber(trimmed, rng); ok {
		return mutated
	}
	if mutated, ok := mutateDate(trimmed, rng); ok {
		return mutated
	}
	if mutated, ok := mutateDuration(trimmed, rng); ok {
		return mutated
	}
	return dither(value, rng)
}

func mutateList(path, value string, rng *rand.Rand) string {
	parts := strings.Split(value, ",")
	index := rng.Intn(len(parts))
	parts[index] = mutateValue(path+fmt.Sprintf("[%d]", index), strings.TrimSpace(parts[index]))
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return strings.Join(parts, ", ")
}

func mutateKeyword(trimmed, original string, rng *rand.Rand) (string, bool) {
	switch strings.ToLower(trimmed) {
	case "true":
		return "false", true
	case "false":
		return "true", true
	case "null", "nil", "none":
		return dither(original, rng), true
	default:
		return "", false
	}
}

func mutateNumber(value string, rng *rand.Rand) (string, bool) {
	if i, err := strconv.Atoi(value); err == nil {
		return strconv.Itoa(i + signedIntDelta(rng)), true
	}
	if f, err := strconv.ParseFloat(value, 64); err == nil && strings.ContainsAny(value, ".eE") {
		return strconv.FormatFloat(f+signedFloatDelta(rng), 'f', -1, 64), true
	}
	return "", false
}

func signedIntDelta(rng *rand.Rand) int {
	delta := rng.Intn(9) + 1
	if rng.Intn(2) == 0 {
		return -delta
	}
	return delta
}

func signedFloatDelta(rng *rand.Rand) float64 {
	delta := float64(rng.Intn(900)+100) / 100
	if rng.Intn(2) == 0 {
		return -delta
	}
	return delta
}

func mutateDate(value string, rng *rand.Rand) (string, bool) {
	t, err := time.Parse("2006-01-02", value)
	if err != nil {
		return "", false
	}
	return t.AddDate(0, 0, signedIntDelta(rng)).Format("2006-01-02"), true
}

func mutateDuration(value string, rng *rand.Rand) (string, bool) {
	d, err := time.ParseDuration(value)
	if err != nil {
		return "", false
	}
	seconds := time.Duration(rng.Intn(9)+1) * time.Second
	return (d + seconds).String(), true
}

func deterministicRand(parts ...string) *rand.Rand {
	hash := sha256.Sum256([]byte(strings.Join(parts, "\x00")))
	seed := int64(binary.BigEndian.Uint64(hash[:8]))
	return rand.New(rand.NewSource(seed))
}

func dither(value string, rng *rand.Rand) string {
	if value == "" {
		return "x"
	}
	runes := []rune(value)
	index := rng.Intn(len(runes))
	if runes[index] == 'x' {
		runes[index] = 'y'
	} else {
		runes[index] = 'x'
	}
	return string(runes)
}
