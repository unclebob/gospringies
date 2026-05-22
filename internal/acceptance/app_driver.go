package acceptance

import (
	"errors"

	"springs/internal/app"
)

type editorScreen = app.EditorScreen
type renderResult = app.RenderResult
type drawFrameReport = app.DrawFrameReport

func startApplicationDriver(w *world) *app.Game {
	game := newApplicationDriverGame()
	w.appGame = game
	return game
}

func newApplicationDriverGame() *app.Game {
	return app.NewGame()
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
