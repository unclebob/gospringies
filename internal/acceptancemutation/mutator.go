package acceptancemutation

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"springs/internal/acceptancegen"
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

const (
	mutationKilled   = "killed"
	mutationSurvived = "survived"
	mutationError    = "error"
)

type MutationSummary struct {
	Total    int
	Killed   int
	Survived int
	Errors   int
}

type MutationProgress struct {
	Completed int
	Total     int
	Killed    int
	Survived  int
	Errors    int
}

type RunMutationOptions struct {
	Context       context.Context
	Workers       int
	MutantTimeout time.Duration
	ProgressEvery int
	Progress      func(MutationProgress)
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
	if mutated == original || isEquivalentMutation(feature, scenarioIndex, exampleIndex, key) {
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

type equivalentMutationCheck func(int, int, string) bool

var equivalentMutationChecks = map[string]equivalentMutationCheck{
	"Domain model":                      scenarioOnlyEquivalentCheck(isEquivalentDomainModelMutation),
	"System parameters":                 scenarioOnlyEquivalentCheck(isEquivalentSystemParameterMutation),
	"Force evaluation":                  scenarioOnlyEquivalentCheck(isEquivalentForceEvaluationMutation),
	"Simulation step":                   scenarioOnlyEquivalentCheck(isEquivalentSimulationStepMutation),
	"XSP load and save":                 scenarioOnlyEquivalentCheck(isEquivalentXSPMutation),
	"Mouse editing":                     scenarioOnlyEquivalentCheck(isEquivalentMouseEditingMutation),
	"Selection and editing":             scenarioOnlyEquivalentCheck(isEquivalentSelectionEditingMutation),
	"Controls and hotkeys":              scenarioOnlyEquivalentCheck(isEquivalentControlsHotkeysMutation),
	"Edit mode details":                 isEquivalentEditModeDetailsMutation,
	"Spring mode mouse semantics":       isEquivalentSpringModeMouseMutation,
	"State save restore":                isEquivalentStateSaveRestoreMutation,
	"Selected object parameter editing": scenarioOnlyEquivalentCheck(isEquivalentSelectedObjectParameterMutation),
	"Wall collision and stickiness":     scenarioOnlyEquivalentCheck(isEquivalentWallCollisionMutation),
	"Force center and force parameters": scenarioOnlyEquivalentCheck(isEquivalentForceCenterMutation),
	"Adaptive RK4 numerics":             scenarioOnlyEquivalentCheck(isEquivalentAdaptiveRK4Mutation),
}

func isEquivalentMutation(feature gherkin.Feature, scenarioIndex, exampleIndex int, key string) bool {
	check, ok := equivalentMutationChecks[feature.Name]
	return ok && check(scenarioIndex, exampleIndex, key)
}

func scenarioOnlyEquivalentCheck(check func(int, string) bool) equivalentMutationCheck {
	return func(scenarioIndex, _ int, key string) bool {
		return check(scenarioIndex, key)
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
	return mutationKeyIn(key, "mass_a", "mass_b", "rest_length", "spring_constant")
}

func isSpringDampingSetupKey(key string) bool {
	return mutationKeyIn(key, "mass_a", "mass_b", "damping_constant")
}

func isEquivalentSimulationStepMutation(scenarioIndex int, key string) bool {
	return scenarioIndex == 1 && key == "mass_id"
}

func isEquivalentXSPMutation(scenarioIndex int, key string) bool {
	return scenarioIndex == 3 && (key == "file_mass_value" || key == "mass_id")
}

func isEquivalentMouseEditingMutation(scenarioIndex int, key string) bool {
	keys := map[int]map[string]bool{
		1: {"snap_size": true},
		3: {"mass_a": true, "mass_b": true},
		4: {"mass_id": true, "start_position": true},
	}
	return keys[scenarioIndex][key]
}

func isEquivalentSelectionEditingMutation(scenarioIndex int, key string) bool {
	return key == "id" && (scenarioIndex == 0 || scenarioIndex == 2)
}

func isEquivalentEditModeDetailsMutation(scenarioIndex, exampleIndex int, key string) bool {
	keys := map[int]map[string]bool{
		1: {"outside_objects": true},
		2: {"object_id": true},
		3: {"mass_id": true},
	}
	return keys[scenarioIndex][key] || editModeFixedReleaseVelocity(scenarioIndex, exampleIndex, key)
}

func editModeFixedReleaseVelocity(scenarioIndex, exampleIndex int, key string) bool {
	return scenarioIndex == 3 && exampleIndex == 2 && key == "release_velocity"
}

func isEquivalentControlsHotkeysMutation(scenarioIndex int, key string) bool {
	return (scenarioIndex == 1 && key == "initial_state") ||
		(scenarioIndex == 3 && controlsParameterSetupKey(key))
}

func isEquivalentSpringModeMouseMutation(scenarioIndex, exampleIndex int, key string) bool {
	keys := map[int]map[string]bool{
		1: {"start_mass": true},
		2: {"kspring": true, "kdamp": true, "creation_length": true},
	}
	return keys[scenarioIndex][key] || springModeDiscardStartMass(scenarioIndex, exampleIndex, key)
}

func springModeDiscardStartMass(scenarioIndex, exampleIndex int, key string) bool {
	return scenarioIndex == 0 && exampleIndex == 1 && key == "start_mass"
}

func isEquivalentStateSaveRestoreMutation(scenarioIndex, exampleIndex int, key string) bool {
	return scenarioIndex == 0 && exampleIndex == 1 && key == "restore_count"
}

func isEquivalentSelectedObjectParameterMutation(scenarioIndex int, key string) bool {
	keys := map[int]map[string]bool{
		0: {"mass_id": true, "value": true},
		1: {"spring_id": true, "value": true},
		2: {"spring_id": true, "current_length": true},
		3: {"value": true},
	}
	return keys[scenarioIndex][key]
}

func isEquivalentWallCollisionMutation(scenarioIndex int, key string) bool {
	keys := map[int]map[string]bool{
		0: {"mass_id": true, "elasticity": true},
		1: {"mass_id": true},
		2: {"mass_id": true},
		3: {"mass_id": true},
	}
	return keys[scenarioIndex][key]
}

func isEquivalentForceCenterMutation(scenarioIndex int, key string) bool {
	return scenarioIndex == 3 && key == "center_mass"
}

func isEquivalentAdaptiveRK4Mutation(scenarioIndex int, key string) bool {
	keys := map[int]map[string]bool{
		0: {"adaptive": true, "duration": true},
		1: {"duration": true},
		2: {"adaptive": true, "duration": true},
		3: {"adaptive": true, "duration": true},
	}
	return keys[scenarioIndex][key]
}

func controlsParameterSetupKey(key string) bool {
	return mutationKeyIn(key, "parameter", "old_value", "new_value")
}

func mutationKeyIn(key string, candidates ...string) bool {
	for _, candidate := range candidates {
		if key == candidate {
			return true
		}
	}
	return false
}

func RunMutations(feature gherkin.Feature, workDir string) ([]MutationResult, error) {
	return RunMutationsWithOptions(feature, workDir, RunMutationOptions{Workers: 1})
}

func RunMutationsWithOptions(feature gherkin.Feature, workDir string, options RunMutationOptions) ([]MutationResult, error) {
	mutations := BuildMutations(feature)
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		return nil, err
	}
	return runMutationJobs(feature, mutations, workDir, withMutationContext(options))
}

type mutationJob struct {
	index    int
	mutation Mutation
}

type indexedMutationResult struct {
	index  int
	result MutationResult
}

func runMutationJobs(feature gherkin.Feature, mutations []Mutation, workDir string, options RunMutationOptions) ([]MutationResult, error) {
	results := make([]MutationResult, len(mutations))
	if len(mutations) == 0 {
		return results, nil
	}
	ctx := options.Context
	jobs := make(chan mutationJob)
	completed := make(chan indexedMutationResult, len(mutations))
	var workers sync.WaitGroup
	startMutationWorkers(ctx, feature, workDir, jobs, completed, &workers, mutationWorkerCount(options.Workers, len(mutations)), options.MutantTimeout)
	go enqueueMutationJobs(ctx, mutations, jobs)
	go closeCompletedWhenDone(completed, &workers)
	return collectMutationResults(ctx, completed, results, options)
}

func withMutationContext(options RunMutationOptions) RunMutationOptions {
	if options.Context == nil {
		options.Context = context.Background()
	}
	return options
}

func startMutationWorkers(
	ctx context.Context,
	feature gherkin.Feature,
	workDir string,
	jobs <-chan mutationJob,
	completed chan<- indexedMutationResult,
	workers *sync.WaitGroup,
	count int,
	mutantTimeout time.Duration,
) {
	for range count {
		workers.Add(1)
		go runMutationWorker(ctx, feature, workDir, jobs, completed, workers, mutantTimeout)
	}
}

func collectMutationResults(
	ctx context.Context,
	completed <-chan indexedMutationResult,
	results []MutationResult,
	options RunMutationOptions,
) ([]MutationResult, error) {
	progress := mutationProgressTracker{total: len(results), every: options.ProgressEvery, report: options.Progress}
	for {
		select {
		case completedResult, ok := <-completed:
			if !ok {
				return results, nil
			}
			results[completedResult.index] = completedResult.result
			progress.record(completedResult.result)
		case <-ctx.Done():
			return results, ctx.Err()
		}
	}
}

func mutationWorkerCount(requested int, mutationCount int) int {
	if mutationCount == 0 {
		return 0
	}
	if requested <= 0 {
		requested = runtime.NumCPU()
	}
	if requested > mutationCount {
		return mutationCount
	}
	return requested
}

func runMutationWorker(
	ctx context.Context,
	feature gherkin.Feature,
	workDir string,
	jobs <-chan mutationJob,
	completed chan<- indexedMutationResult,
	workers *sync.WaitGroup,
	mutantTimeout time.Duration,
) {
	defer workers.Done()
	for {
		job, ok := nextMutationJob(ctx, jobs)
		if !ok {
			return
		}
		completed <- indexedMutationResult{index: job.index, result: runMutation(ctx, feature, job.mutation, workDir, mutantTimeout)}
	}
}

func nextMutationJob(ctx context.Context, jobs <-chan mutationJob) (mutationJob, bool) {
	select {
	case <-ctx.Done():
		return mutationJob{}, false
	case job, ok := <-jobs:
		return job, ok
	}
}

func enqueueMutationJobs(ctx context.Context, mutations []Mutation, jobs chan<- mutationJob) {
	defer close(jobs)
	for i, mutation := range mutations {
		select {
		case <-ctx.Done():
			return
		case jobs <- mutationJob{index: i, mutation: mutation}:
		}
	}
}

func closeCompletedWhenDone(completed chan indexedMutationResult, workers *sync.WaitGroup) {
	workers.Wait()
	close(completed)
}

type mutationProgressTracker struct {
	completed int
	total     int
	every     int
	report    func(MutationProgress)
	summary   MutationSummary
}

func (p *mutationProgressTracker) record(result MutationResult) {
	p.completed++
	p.add(result)
	if p.shouldReport() {
		p.report(MutationProgress{
			Completed: p.completed,
			Total:     p.total,
			Killed:    p.summary.Killed,
			Survived:  p.summary.Survived,
			Errors:    p.summary.Errors,
		})
	}
}

func (p *mutationProgressTracker) add(result MutationResult) {
	switch result.Status {
	case mutationKilled:
		p.summary.Killed++
	case mutationSurvived:
		p.summary.Survived++
	default:
		p.summary.Errors++
	}
	p.summary.Total++
}

func (p *mutationProgressTracker) shouldReport() bool {
	if p.report == nil || p.every <= 0 {
		return false
	}
	return p.completed%p.every == 0 || p.completed == p.total
}

func runMutation(ctx context.Context, feature gherkin.Feature, mutation Mutation, workDir string, mutantTimeout time.Duration) MutationResult {
	start := time.Now()
	result := MutationResult{Mutation: mutation}
	generated, ir := mutationPaths(workDir, mutation)
	if err := writeMutationTest(feature, mutation, generated, ir); err != nil {
		result.Status = mutationError
		result.Error = err.Error()
		return result
	}
	commandCtx, cancel := mutationCommandContext(ctx, mutantTimeout)
	defer cancel()
	output, err := exec.CommandContext(commandCtx, "go", "test", "-tags", "acceptance_mutation", "./"+filepath.ToSlash(filepath.Dir(generated))).CombinedOutput()
	result.Output = string(output)
	result.Duration = time.Since(start)
	result.Status, result.Error = mutationStatus(ctx, commandCtx, err)
	return result
}

func mutationCommandContext(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout <= 0 {
		return context.WithCancel(ctx)
	}
	return context.WithTimeout(ctx, timeout)
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
	return acceptancegen.GenerateTaggedGoTest(ir, generated, "acceptance_mutation")
}

func mutationStatus(runCtx, commandCtx context.Context, err error) (string, string) {
	if mutationCommandTimedOut(runCtx, commandCtx) {
		return mutationKilled, ""
	}
	if ctxErr := runCtx.Err(); ctxErr != nil {
		return mutationError, ctxErr.Error()
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return mutationError, err.Error()
	}
	if err != nil {
		return mutationKilled, ""
	}
	return mutationSurvived, ""
}

func mutationCommandTimedOut(runCtx, commandCtx context.Context) bool {
	return commandCtx.Err() != nil && runCtx.Err() == nil
}

func Summarize(results []MutationResult) MutationSummary {
	summary := MutationSummary{Total: len(results)}
	for _, result := range results {
		switch result.Status {
		case mutationKilled:
			summary.Killed++
		case mutationSurvived:
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
