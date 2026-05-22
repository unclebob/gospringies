package acceptance

import (
	"errors"

	"springs/internal/app"
	"springs/internal/sim"
)

type editorScreen = app.EditorScreen
type renderResult = app.RenderResult
type drawFrameReport = app.DrawFrameReport

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

func startApplicationDriver(w *world) *app.Game {
	game := newApplicationDriverGame()
	w.appGame = game
	return game
}

func newApplicationDriverGame() *app.Game {
	return app.NewGame()
}

func newApplicationDriverWorld() *sim.Simulation {
	return newApplicationDriverGame().World()
}

func applicationWindowResizable() bool {
	return app.DefaultWindowConfig().Resizable
}

func ensureConcreteApplicationDriver(w *world) (*app.Game, error) {
	if w.appGame == nil {
		startApplicationDriver(w)
	}
	return concreteApplicationDriver(w)
}

func concreteApplicationDriver(w *world) (*app.Game, error) {
	return concreteApplicationDriverWithMessage(w, "application is not running")
}

func concreteApplicationDriverWithMessage(w *world, message string) (*app.Game, error) {
	game, ok := w.appGame.(*app.Game)
	if !ok || game == nil {
		return nil, errors.New(message)
	}
	return game, nil
}

func optionalConcreteApplicationDriver(w *world) (*app.Game, bool) {
	game, ok := w.appGame.(*app.Game)
	return game, ok && game != nil
}

func applicationDriverGame(w *world) (appGame, error) {
	if w.appGame == nil {
		return nil, errors.New("application was not started")
	}
	return w.appGame, nil
}
