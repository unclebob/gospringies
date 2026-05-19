package app

import (
	"springs/internal/appcore"
	"springs/internal/sim"
)

func DefaultStartupScenePath() string {
	return appcore.DefaultStartupScenePath()
}

func newDefaultStartupWorld() *sim.Simulation {
	return appcore.NewDefaultStartupWorld(appBounds())
}

func loadDefaultStartupWorld() (*sim.Simulation, error) {
	return appcore.LoadDefaultStartupWorld(appBounds())
}

func defaultStartupSceneCandidates() []string {
	return appcore.DefaultStartupSceneCandidates()
}

// mutate4go-manifest-begin
// {"version":1,"tested_at":"2026-05-19T12:12:19-05:00","module_hash":"2453fabed6511a3c48d5f7719f5a5d5c0206557d42eeb2bf4e0f54285b500025","functions":[{"id":"func/DefaultStartupScenePath","name":"DefaultStartupScenePath","line":8,"end_line":10,"hash":"145c5b8dc334c29695568aa443d747e3e84c98c4d07f9a97115947d95cfc0ee7"},{"id":"func/newDefaultStartupWorld","name":"newDefaultStartupWorld","line":12,"end_line":14,"hash":"f2efe10ee1abea228a27949627829678abc0793773d8de68556c68b2e16bd735"},{"id":"func/loadDefaultStartupWorld","name":"loadDefaultStartupWorld","line":16,"end_line":18,"hash":"49a14c7a41b03e89bfd264c4c54fb6c0f39f58e121bb815c66e0c33c952f6ae5"},{"id":"func/defaultStartupSceneCandidates","name":"defaultStartupSceneCandidates","line":20,"end_line":22,"hash":"6dbc3d604823ec0472e204decfce74dc7eb1272f84f81f4d218211da50131a14"}]}
// mutate4go-manifest-end
