package acceptancemutation

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"springs/internal/gherkin"
)

func TestScenarioManifestParseAndRemove(t *testing.T) {
	content := "# acceptance-mutation-manifest-begin\n# {\"version\":1,\"feature_name\":\"F\"}\n# acceptance-mutation-manifest-end\nFeature: F\n"

	manifest, ok, err := ParseScenarioManifest(content)
	if err != nil || !ok {
		t.Fatalf("parse ok=%t err=%v", ok, err)
	}
	if manifest.Version != 1 || manifest.FeatureName != "F" {
		t.Fatalf("manifest = %#v", manifest)
	}
	if got := RemoveScenarioManifest(content); strings.Contains(got, "acceptance-mutation") || !strings.Contains(got, "Feature: F") {
		t.Fatalf("removed content = %q", got)
	}
}

func TestScenarioAndBackgroundHashesChangeWithContent(t *testing.T) {
	base := manifestFeature()
	changedScenario := manifestFeature()
	changedScenario.Scenarios[0].Examples[0]["value"] = "2"
	changedBackground := manifestFeature()
	changedBackground.Background[0].Text = "different"

	if ScenarioHash(base.Scenarios[0]) == ScenarioHash(changedScenario.Scenarios[0]) {
		t.Fatal("scenario hash did not change")
	}
	if BackgroundHash(base) == BackgroundHash(changedBackground) {
		t.Fatal("background hash did not change")
	}
}

func TestScenarioSkipPlanSkipsOnlyValidCleanEntries(t *testing.T) {
	feature := manifestFeature()
	manifest := BuildScenarioManifest("f.feature", feature, ScenarioManifest{}, ScenarioSkipPlan{SkipScenarios: map[int]bool{}}, []MutationResult{
		{Mutation: Mutation{Scenario: 0}, Status: MutationKilled},
		{Mutation: Mutation{Scenario: 1}, Status: MutationKilled},
	}, "impl", time.Unix(1, 0).UTC())
	manifest.Scenarios[1].Result.Survived = 1

	plan := ScenarioSkipPlanFor(feature, "f.feature", manifest, "impl")
	if !plan.SkipScenarios[0] || plan.SkipScenarios[1] || plan.SkippedScenarios != 1 {
		t.Fatalf("plan = %#v", plan)
	}

	if stale := ScenarioSkipPlanFor(feature, "f.feature", manifest, "different"); stale.SkippedScenarios != 0 {
		t.Fatalf("stale implementation skipped: %#v", stale)
	}

	manifest.Scenarios[0].Result.Survived = 1
	if failed := ScenarioSkipPlanFor(feature, "f.feature", manifest, "impl"); failed.SkippedScenarios != 0 {
		t.Fatalf("failed scenario skipped: %#v", failed)
	}
}

func TestScenarioSkipPlanRerunsChangedScenarioOnly(t *testing.T) {
	feature := manifestFeature()
	manifest := BuildScenarioManifest("f.feature", feature, ScenarioManifest{}, ScenarioSkipPlan{SkipScenarios: map[int]bool{}}, []MutationResult{
		{Mutation: Mutation{Scenario: 0}, Status: MutationKilled},
		{Mutation: Mutation{Scenario: 1}, Status: MutationKilled},
	}, "impl", time.Unix(1, 0).UTC())
	changed := manifestFeature()
	changed.Scenarios[1].Steps[0].Text = "different <name>"

	plan := ScenarioSkipPlanFor(changed, "f.feature", manifest, "impl")
	if !plan.SkipScenarios[0] || plan.SkipScenarios[1] || plan.SkippedScenarios != 1 {
		t.Fatalf("plan = %#v", plan)
	}
}

func TestScenarioSkipPlanInvalidatesAllScenariosWhenBackgroundChanges(t *testing.T) {
	feature := manifestFeature()
	manifest := BuildScenarioManifest("f.feature", feature, ScenarioManifest{}, ScenarioSkipPlan{SkipScenarios: map[int]bool{}}, []MutationResult{
		{Mutation: Mutation{Scenario: 0}, Status: MutationKilled},
		{Mutation: Mutation{Scenario: 1}, Status: MutationKilled},
	}, "impl", time.Unix(1, 0).UTC())
	changed := manifestFeature()
	changed.Background[0].Text = "different"

	plan := ScenarioSkipPlanFor(changed, "f.feature", manifest, "impl")
	if plan.SkippedScenarios != 0 {
		t.Fatalf("changed background skipped: %#v", plan)
	}
}

func TestBuildScenarioManifestKeepsSkippedAndRemovesDeletedScenarios(t *testing.T) {
	feature := manifestFeature()
	previous := BuildScenarioManifest("f.feature", feature, ScenarioManifest{}, ScenarioSkipPlan{SkipScenarios: map[int]bool{}}, []MutationResult{
		{Mutation: Mutation{Scenario: 0}, Status: MutationKilled},
		{Mutation: Mutation{Scenario: 1}, Status: MutationKilled},
	}, "impl", time.Unix(1, 0).UTC())

	shortened := feature
	shortened.Scenarios = shortened.Scenarios[:1]
	next := BuildScenarioManifest("f.feature", shortened, previous, ScenarioSkipPlan{SkipScenarios: map[int]bool{0: true}}, nil, "impl", time.Unix(2, 0).UTC())

	if len(next.Scenarios) != 1 {
		t.Fatalf("scenario count = %d", len(next.Scenarios))
	}
	if next.Scenarios[0].TestedAt != previous.Scenarios[0].TestedAt {
		t.Fatalf("skipped scenario timestamp changed: %#v", next.Scenarios[0])
	}
}

func TestWriteScenarioManifestFileWritesDeterministicCommentManifest(t *testing.T) {
	path := filepath.Join(t.TempDir(), "manifest.feature")
	feature := manifestFeature()
	previous := BuildScenarioManifest(path, feature, ScenarioManifest{}, ScenarioSkipPlan{SkipScenarios: map[int]bool{}}, []MutationResult{
		{Mutation: Mutation{Scenario: 0}, Status: MutationKilled},
		{Mutation: Mutation{Scenario: 1}, Status: MutationKilled},
	}, "impl", time.Unix(1, 0).UTC())
	content := "# acceptance-mutation-manifest-begin\n# {\"version\":0}\n# acceptance-mutation-manifest-end\n\nFeature: F\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	err := WriteScenarioManifestFile(path, feature, previous, ScenarioSkipPlan{SkipScenarios: map[int]bool{0: true}}, []MutationResult{
		{Mutation: Mutation{Scenario: 1}, Status: MutationKilled},
	}, "impl", time.Unix(2, 0).UTC())
	if err != nil {
		t.Fatal(err)
	}

	written, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(written)
	if strings.Count(text, "acceptance-mutation-manifest-begin") != 1 || strings.Contains(text, "\"version\":0") {
		t.Fatalf("manifest block not replaced:\n%s", text)
	}
	manifest, ok, err := ParseScenarioManifest(text)
	if err != nil || !ok {
		t.Fatalf("parse ok=%t err=%v", ok, err)
	}
	if manifest.Scenarios[0].TestedAt != previous.Scenarios[0].TestedAt {
		t.Fatalf("skipped scenario timestamp changed: %#v", manifest.Scenarios[0])
	}
	if manifest.Scenarios[1].TestedAt != time.Unix(2, 0).UTC().Format(time.RFC3339) {
		t.Fatalf("executed scenario timestamp = %q", manifest.Scenarios[1].TestedAt)
	}
}

func TestImplementationHashHelpersUseProductionSources(t *testing.T) {
	hash, err := CurrentImplementationHash()
	if err != nil {
		t.Fatal(err)
	}
	if hash == "" || hash == DefaultImplementationHash {
		t.Fatalf("implementation hash = %q", hash)
	}

	root := t.TempDir()
	dir := filepath.Join(root, "internal", "acceptancemutation")
	if err := os.MkdirAll(filepath.Join(dir, "nested"), 0o755); err != nil {
		t.Fatal(err)
	}
	for path, content := range map[string]string{
		filepath.Join(dir, "a.go"):           "package acceptancemutation\n",
		filepath.Join(dir, "a_test.go"):      "package acceptancemutation\n",
		filepath.Join(dir, "notes.txt"):      "ignore",
		filepath.Join(dir, "nested", "b.go"): "package acceptancemutation\n",
	} {
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	files, err := goSourceFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 {
		t.Fatalf("go source files = %#v", files)
	}
	hashed, err := hashImplementationFiles(root, files)
	if err != nil {
		t.Fatal(err)
	}
	if hashed == "" || hashed == DefaultImplementationHash {
		t.Fatalf("hashed implementation files = %q", hashed)
	}
}

func manifestFeature() gherkin.Feature {
	return gherkin.Feature{
		Name:       "F",
		Background: []gherkin.Step{{Keyword: "Given", Text: "background"}},
		Scenarios: []gherkin.Scenario{
			{Name: "first", Steps: []gherkin.Step{{Keyword: "Then", Text: "value <value>"}}, Examples: []map[string]string{{"value": "1"}}},
			{Name: "second", Steps: []gherkin.Step{{Keyword: "Then", Text: "name <name>"}}, Examples: []map[string]string{{"name": "Ada"}}},
		},
	}
}
