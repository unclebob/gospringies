package acceptancemutation

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"springs/internal/gherkin"
)

const (
	ScenarioManifestVersion   = 1
	ScenarioManifestBegin     = "# acceptance-mutation-manifest-begin"
	ScenarioManifestEnd       = "# acceptance-mutation-manifest-end"
	DefaultImplementationHash = "acceptance-mutation-v1"
)

type ScenarioManifest struct {
	Version            int                     `json:"version"`
	TestedAt           string                  `json:"tested_at"`
	FeatureName        string                  `json:"feature_name"`
	FeaturePath        string                  `json:"feature_path"`
	BackgroundHash     string                  `json:"background_hash"`
	ImplementationHash string                  `json:"implementation_hash"`
	Scenarios          []ScenarioManifestEntry `json:"scenarios"`
}

type ScenarioManifestEntry struct {
	Index         int             `json:"index"`
	Name          string          `json:"name"`
	ScenarioHash  string          `json:"scenario_hash"`
	MutationCount int             `json:"mutation_count"`
	Result        MutationSummary `json:"result"`
	TestedAt      string          `json:"tested_at,omitempty"`
}

type ScenarioSkipPlan struct {
	SkipScenarios    map[int]bool
	SkippedScenarios int
	SkippedMutations int
}

func ParseScenarioManifest(content string) (ScenarioManifest, bool, error) {
	block, ok := scenarioManifestBlock(content)
	if !ok {
		return ScenarioManifest{}, false, nil
	}
	var manifest ScenarioManifest
	if err := json.Unmarshal([]byte(block), &manifest); err != nil {
		return ScenarioManifest{}, true, err
	}
	return manifest, true, nil
}

func scenarioManifestBlock(content string) (string, bool) {
	lines := strings.Split(content, "\n")
	inBlock := false
	var body []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		switch trimmed {
		case ScenarioManifestBegin:
			inBlock = true
			continue
		case ScenarioManifestEnd:
			return strings.Join(body, "\n"), true
		}
		if inBlock {
			body = append(body, strings.TrimSpace(strings.TrimPrefix(trimmed, "#")))
		}
	}
	return "", false
}

func RemoveScenarioManifest(content string) string {
	lines := strings.SplitAfter(content, "\n")
	var out strings.Builder
	inBlock := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == ScenarioManifestBegin {
			inBlock = true
			continue
		}
		if inBlock {
			if trimmed == ScenarioManifestEnd {
				inBlock = false
			}
			continue
		}
		out.WriteString(line)
	}
	return out.String()
}

func ScenarioSkipPlanFor(feature gherkin.Feature, featurePath string, manifest ScenarioManifest, implementationHash string) ScenarioSkipPlan {
	plan := ScenarioSkipPlan{SkipScenarios: map[int]bool{}}
	if !scenarioManifestMatches(feature, featurePath, manifest, implementationHash) {
		return plan
	}
	entries := scenarioManifestEntriesByIndex(manifest)
	for index, scenario := range feature.Scenarios {
		if entry, ok := skippableScenarioEntry(entries, index, scenario); ok {
			plan.SkipScenarios[index] = true
			plan.SkippedScenarios++
			plan.SkippedMutations += entry.MutationCount
		}
	}
	return plan
}

func scenarioManifestMatches(feature gherkin.Feature, featurePath string, manifest ScenarioManifest, implementationHash string) bool {
	return manifest.Version == ScenarioManifestVersion &&
		manifest.FeatureName == feature.Name &&
		manifest.FeaturePath == featurePath &&
		manifest.BackgroundHash == BackgroundHash(feature) &&
		manifest.ImplementationHash == implementationHash
}

func scenarioManifestEntriesByIndex(manifest ScenarioManifest) map[int]ScenarioManifestEntry {
	entries := map[int]ScenarioManifestEntry{}
	for _, entry := range manifest.Scenarios {
		entries[entry.Index] = entry
	}
	return entries
}

func skippableScenarioEntry(entries map[int]ScenarioManifestEntry, index int, scenario gherkin.Scenario) (ScenarioManifestEntry, bool) {
	entry, ok := entries[index]
	if !ok || entry.Name != scenario.Name || entry.ScenarioHash != ScenarioHash(scenario) {
		return ScenarioManifestEntry{}, false
	}
	return entry, entry.Result.Survived == 0 && entry.Result.Errors == 0
}

func WriteScenarioManifestFile(path string, feature gherkin.Feature, previous ScenarioManifest, plan ScenarioSkipPlan, results []MutationResult, implementationHash string, now time.Time) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	manifest := BuildScenarioManifest(path, feature, previous, plan, results, implementationHash, now)
	data, err := json.Marshal(manifest)
	if err != nil {
		return err
	}
	var block strings.Builder
	block.WriteString(ScenarioManifestBegin + "\n")
	block.WriteString("# " + string(data) + "\n")
	block.WriteString(ScenarioManifestEnd + "\n")
	without := strings.TrimLeft(RemoveScenarioManifest(string(content)), "\n")
	return os.WriteFile(path, []byte(block.String()+without), 0o644)
}

func CurrentImplementationHash() (string, error) {
	_, sourceFile, _, ok := runtime.Caller(0)
	if !ok {
		return DefaultImplementationHash, nil
	}
	root := filepath.Dir(filepath.Dir(filepath.Dir(sourceFile)))
	files, err := implementationHashFiles(root)
	if err != nil {
		return "", err
	}
	return hashImplementationFiles(root, files)
}

func implementationHashFiles(root string) ([]string, error) {
	dirs := implementationHashDirs(root)
	var files []string
	for _, dir := range dirs {
		dirFiles, err := goSourceFiles(dir)
		if err != nil {
			return nil, err
		}
		files = append(files, dirFiles...)
	}
	sort.Strings(files)
	return files, nil
}

func implementationHashDirs(root string) []string {
	return []string{
		filepath.Join(root, "cmd", "gherkin-mutator"),
		filepath.Join(root, "internal", "acceptance"),
		filepath.Join(root, "internal", "acceptancemutation"),
		filepath.Join(root, "internal", "gherkin"),
	}
}

func goSourceFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		files = append(files, path)
		return nil
	})
	return files, err
}

func hashImplementationFiles(root string, files []string) (string, error) {
	parts := []string{DefaultImplementationHash}
	for _, path := range files {
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return "", err
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		parts = append(parts, relative, string(content))
	}
	return hashStrings(parts...), nil
}

func BuildScenarioManifest(path string, feature gherkin.Feature, previous ScenarioManifest, plan ScenarioSkipPlan, results []MutationResult, implementationHash string, now time.Time) ScenarioManifest {
	resultByScenario := summarizeByScenario(results)
	mutationsByScenario := mutationCountByScenario(feature)
	testedAt := now.Format(time.RFC3339)
	if len(results) == 0 && previous.TestedAt != "" {
		testedAt = previous.TestedAt
	}
	previousByScenario := map[int]ScenarioManifestEntry{}
	for _, entry := range previous.Scenarios {
		previousByScenario[entry.Index] = entry
	}
	entries := make([]ScenarioManifestEntry, 0, len(feature.Scenarios))
	for index, scenario := range feature.Scenarios {
		if plan.SkipScenarios[index] {
			entries = append(entries, previousByScenario[index])
			continue
		}
		entries = append(entries, ScenarioManifestEntry{
			Index:         index,
			Name:          scenario.Name,
			ScenarioHash:  ScenarioHash(scenario),
			MutationCount: mutationsByScenario[index],
			Result:        resultByScenario[index],
			TestedAt:      now.Format(time.RFC3339),
		})
	}
	return ScenarioManifest{
		Version:            ScenarioManifestVersion,
		TestedAt:           testedAt,
		FeatureName:        feature.Name,
		FeaturePath:        path,
		BackgroundHash:     BackgroundHash(feature),
		ImplementationHash: implementationHash,
		Scenarios:          entries,
	}
}

func summarizeByScenario(results []MutationResult) map[int]MutationSummary {
	summaries := map[int]MutationSummary{}
	for _, result := range results {
		summary := summaries[result.Mutation.Scenario]
		summary.Total++
		switch result.Status {
		case MutationKilled:
			summary.Killed++
		case MutationSurvived:
			summary.Survived++
		default:
			summary.Errors++
		}
		summaries[result.Mutation.Scenario] = summary
	}
	return summaries
}

func mutationCountByScenario(feature gherkin.Feature) map[int]int {
	counts := map[int]int{}
	for _, mutation := range BuildMutations(feature) {
		counts[mutation.Scenario]++
	}
	return counts
}

func BackgroundHash(feature gherkin.Feature) string {
	parts := make([]string, 0, len(feature.Background))
	for _, step := range feature.Background {
		parts = append(parts, step.Keyword+" "+step.Text)
	}
	return hashStrings(parts...)
}

func ScenarioHash(scenario gherkin.Scenario) string {
	var parts []string
	parts = append(parts, scenario.Name)
	for _, step := range scenario.Steps {
		parts = append(parts, step.Keyword+" "+step.Text)
	}
	headers := exampleHeaders(scenario)
	parts = append(parts, strings.Join(headers, "|"))
	for _, example := range scenario.Examples {
		for _, header := range headers {
			parts = append(parts, header+"="+example[header])
		}
	}
	return hashStrings(parts...)
}

func exampleHeaders(scenario gherkin.Scenario) []string {
	seen := map[string]bool{}
	for _, example := range scenario.Examples {
		for key := range example {
			seen[key] = true
		}
	}
	headers := make([]string, 0, len(seen))
	for key := range seen {
		headers = append(headers, key)
	}
	sort.Strings(headers)
	return headers
}

func hashStrings(parts ...string) string {
	hash := sha256.Sum256([]byte(strings.Join(parts, "\x00")))
	return hex.EncodeToString(hash[:])
}
