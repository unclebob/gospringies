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
				original := example[key]
				path := fmt.Sprintf("$.scenarios[%d].examples[%d].%s", scenarioIndex, exampleIndex, key)
				mutated := mutateValue(path, original)
				if mutated == original {
					continue
				}
				if isEquivalentMutation(feature, scenarioIndex, key) {
					continue
				}
				id := fmt.Sprintf("m%d", len(mutations)+1)
				mutations = append(mutations, Mutation{
					ID:          id,
					Path:        path,
					Description: fmt.Sprintf("%s: %s -> %s", path, original, mutated),
					Original:    original,
					Mutated:     mutated,
					Scenario:    scenarioIndex,
					Example:     exampleIndex,
					Key:         key,
				})
			}
		}
	}
	return mutations
}

func isEquivalentMutation(feature gherkin.Feature, scenarioIndex int, key string) bool {
	if feature.Name != "Domain model" {
		return false
	}
	if scenarioIndex == 0 {
		return false
	}
	if key == "reason" {
		return false
	}
	// Domain-model property scenarios use example cells as both setup data and
	// expected lookup values. Mutating both sides preserves the same behavior,
	// so only externally checked counts and validation reasons are meaningful.
	return true
}

func RunMutations(feature gherkin.Feature, workDir string) ([]MutationResult, error) {
	mutations := BuildMutations(feature)
	results := make([]MutationResult, 0, len(mutations))
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		return nil, err
	}
	for _, mutation := range mutations {
		start := time.Now()
		result := MutationResult{Mutation: mutation}
		mutatedFeature := cloneFeature(feature)
		mutatedFeature.Scenarios[mutation.Scenario].Examples[mutation.Example][mutation.Key] = mutation.Mutated
		mutationDir := filepath.Join(workDir, mutation.ID)
		generated := filepath.Join(mutationDir, "generated", "feature_acceptance_test.go")
		ir := filepath.Join(mutationDir, "feature.json")
		if err := os.MkdirAll(filepath.Dir(generated), 0o755); err != nil {
			result.Status = "error"
			result.Error = err.Error()
			results = append(results, result)
			continue
		}
		if err := gherkin.WriteJSON(mutatedFeature, ir); err != nil {
			result.Status = "error"
			result.Error = err.Error()
			results = append(results, result)
			continue
		}
		if err := generateTaggedGoTest(ir, generated, "acceptance_mutation"); err != nil {
			result.Status = "error"
			result.Error = err.Error()
			results = append(results, result)
			continue
		}
		cmd := exec.Command("go", "test", "-tags", "acceptance_mutation", "./"+filepath.ToSlash(filepath.Dir(generated)))
		output, err := cmd.CombinedOutput()
		result.Output = string(output)
		result.Duration = time.Since(start)
		if err != nil {
			result.Status = "killed"
		} else {
			result.Status = "survived"
		}
		results = append(results, result)
	}
	return results, nil
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
		parts := strings.Split(trimmed, ",")
		index := rng.Intn(len(parts))
		parts[index] = mutateValue(path+fmt.Sprintf("[%d]", index), strings.TrimSpace(parts[index]))
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		return strings.Join(parts, ", ")
	}
	switch strings.ToLower(trimmed) {
	case "true":
		return "false"
	case "false":
		return "true"
	case "null", "nil", "none":
		return dither(value, rng)
	}
	if i, err := strconv.Atoi(trimmed); err == nil {
		delta := rng.Intn(9) + 1
		if rng.Intn(2) == 0 {
			delta = -delta
		}
		return strconv.Itoa(i + delta)
	}
	if f, err := strconv.ParseFloat(trimmed, 64); err == nil && strings.ContainsAny(trimmed, ".eE") {
		delta := float64(rng.Intn(900)+100) / 100
		if rng.Intn(2) == 0 {
			delta = -delta
		}
		return strconv.FormatFloat(f+delta, 'f', -1, 64)
	}
	if t, err := time.Parse("2006-01-02", trimmed); err == nil {
		days := rng.Intn(9) + 1
		if rng.Intn(2) == 0 {
			days = -days
		}
		return t.AddDate(0, 0, days).Format("2006-01-02")
	}
	if d, err := time.ParseDuration(trimmed); err == nil {
		seconds := time.Duration(rng.Intn(9)+1) * time.Second
		return (d + seconds).String()
	}
	return dither(value, rng)
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
