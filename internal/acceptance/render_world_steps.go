package acceptance

import (
	"fmt"

	"springs/internal/app"
	"springs/internal/sim"
)

type renderResult = app.RenderResult

func createApplicationWorldState(w *world, example map[string]string) error {
	state, err := stringValue(example, "world_state")
	if err != nil {
		return err
	}
	game := app.NewGame()
	if state == "a non-empty world" {
		addRenderSpring(game.World())
	} else if state != "an empty world" {
		return fmt.Errorf("unsupported world state %q", state)
	}
	w.appGame = game
	return nil
}

func renderApplicationWorld(w *world, _ map[string]string) error {
	return updateApplicationGame(w, func(game appGame) { w.renderResult = game.RenderWorld() })
}

func assertRenderingComplete(w *world, _ map[string]string) error {
	return requirePrerequisite(w.renderResult.Completed, "rendering did not complete")
}

func createRenderableObject(w *world, example map[string]string) error {
	object, err := stringValue(example, "object")
	if err != nil {
		return err
	}
	game := app.NewGame()
	if err := addRenderableObject(game, object); err != nil {
		return err
	}
	w.appGame = game
	return nil
}

func createRenderableSpring(w *world, _ map[string]string) error {
	return createRenderWorld(w, addRenderSpring)
}

func assertVisibleRepresentation(w *world, example map[string]string) error {
	object, err := stringValue(example, "object")
	if err != nil {
		return err
	}
	message := fmt.Sprintf("%s did not have a visible representation", object)
	return requirePrerequisite(w.renderResult.HasVisibleRepresentation(object), message)
}

func setShowSprings(w *world, example map[string]string) error {
	showSprings, err := stringValue(example, "show_springs")
	if err != nil {
		return err
	}
	game, err := ensureApplicationGame(w)
	if err != nil {
		return err
	}
	game.World().Parameters.Set("show springs", showSprings)
	return nil
}

func assertSpringLineVisibility(w *world, example map[string]string) error {
	visibility, err := stringValue(example, "spring_visibility")
	if err != nil {
		return err
	}
	expected, ok := map[string]bool{"visible": true, "hidden": false}[visibility]
	if !ok {
		return fmt.Errorf("unsupported spring visibility %q", visibility)
	}
	if w.renderResult.SpringLinesVisible != expected {
		return fmt.Errorf("spring lines visible = %t, expected %t", w.renderResult.SpringLinesVisible, expected)
	}
	return nil
}

func assertMassesVisible(w *world, _ map[string]string) error {
	return requirePrerequisite(w.renderResult.MassesVisible, "masses were not visible")
}

func createFixedAndMovableMasses(w *world, _ map[string]string) error {
	return createRenderWorld(w, addFixedAndMovableMasses)
}

func createRenderWorld(w *world, addObjects func(*sim.Simulation)) error {
	game := app.NewGame()
	addObjects(game.World())
	w.appGame = game
	return nil
}

func assertFixedMassDistinguishable(w *world, _ map[string]string) error {
	return requirePrerequisite(w.renderResult.FixedMassDistinguishable, "fixed mass was not visually distinguishable")
}

func addRenderableObject(game appGame, object string) error {
	add, ok := renderableObjectSetups()[object]
	if !ok {
		return fmt.Errorf("unsupported renderable object %q", object)
	}
	add(game)
	return nil
}

func renderableObjectSetups() map[string]func(appGame) {
	return map[string]func(appGame){
		"movable mass": func(game appGame) {
			_ = game.World().AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 20, Y: 20}, Mass: 1})
		},
		"fixed mass": func(game appGame) {
			_ = game.World().AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 20, Y: 20}, Mass: 1, Fixed: true})
		},
		"spring":       func(game appGame) { addRenderSpring(game.World()) },
		"enabled wall": func(game appGame) { game.World().Parameters.EnableWall("left") },
		"selection": func(game appGame) {
			_ = game.World().AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 20, Y: 20}, Mass: 1})
			game.SetSelected(true)
		},
	}
}

func addRenderSpring(world *sim.Simulation) {
	addFixedAndMovableMasses(world)
	world.AddSpringBetween(0, 1, 20, 12)
}

func addFixedAndMovableMasses(world *sim.Simulation) {
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 20, Y: 20}, Mass: 1, Fixed: true})
	_ = world.AddMass(sim.Mass{ID: 2, Position: sim.Vec2{X: 40, Y: 20}, Mass: 1})
}
