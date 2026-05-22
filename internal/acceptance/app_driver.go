package acceptance

import (
	"errors"

	"springs/internal/app"
	"springs/internal/sim"
)

type editorScreen = app.EditorScreen
type renderResult = app.RenderResult
type drawFrameReport = app.DrawFrameReport
type numericSettingReport = app.NumericSettingFrame
type driverGame = app.Game

type appGame interface {
	Update() error
	RenderFrame()
	RenderWorld() renderResult
	World() *sim.Simulation
	SetPaused(bool)
	EditorScreen() editorScreen
	SetSelected(bool)
	SetDirty(bool)
	HandleShortcut(string) bool
	LastCommand() string
	DrawFrameReport() drawFrameReport
	InputActive() bool
	RenderingActive() bool
	Close() error
	Closed() bool
}

func startApplicationDriver(w *world) *driverGame {
	game := newApplicationDriverGame()
	w.appGame = game
	return game
}

func newApplicationDriverGame() *driverGame {
	return app.NewGame()
}

func newApplicationDriverWorld() *sim.Simulation {
	return newApplicationDriverGame().World()
}

func applicationWindowResizable() bool {
	return app.DefaultWindowConfig().Resizable
}

func defaultStartupScenePath() string {
	return app.DefaultStartupScenePath()
}

func ensureConcreteApplicationDriver(w *world) (*driverGame, error) {
	if w.appGame == nil {
		startApplicationDriver(w)
	}
	return concreteApplicationDriver(w)
}

func concreteApplicationDriver(w *world) (*driverGame, error) {
	return concreteApplicationDriverWithMessage(w, "application is not running")
}

func concreteApplicationDriverWithMessage(w *world, message string) (*driverGame, error) {
	game, ok := w.appGame.(*app.Game)
	if !ok || game == nil {
		return nil, errors.New(message)
	}
	return game, nil
}

func optionalConcreteApplicationDriver(w *world) (*driverGame, bool) {
	game, ok := w.appGame.(*app.Game)
	return game, ok && game != nil
}

func applicationDriverGame(w *world) (appGame, error) {
	if w.appGame == nil {
		return nil, errors.New("application was not started")
	}
	return w.appGame, nil
}
