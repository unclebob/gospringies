package acceptance

import (
	"fmt"
	"reflect"
	"slices"

	"springs/internal/sim"
)

func createMemoryWorldState(w *world, example map[string]string) error {
	state, err := stringValue(example, "saved_state")
	if err != nil {
		state, err = stringValue(example, "memory_state")
	}
	if err != nil {
		return err
	}
	return setApplicationStateWorld(w, state)
}

func saveApplicationState(w *world, _ map[string]string) error {
	return withConcreteGame(w, func(game *driverGame) error {
		game.SaveState()
		return nil
	})
}

func changeApplicationState(w *world, example map[string]string) error {
	state, err := stringValue(example, "changed_state")
	if err != nil {
		return err
	}
	return replaceApplicationWorld(w, state)
}

func restoreApplicationStateTimes(w *world, example map[string]string) error {
	count, err := intValue(example, "restore_count")
	if err != nil {
		return err
	}
	return restoreApplicationState(w, count)
}

func assertApplicationStateWorld(w *world, example map[string]string) error {
	state, err := stringValue(example, "saved_state")
	if err != nil {
		state, err = stringValue(example, "memory_state")
	}
	if err != nil {
		return err
	}
	return withConcreteGame(w, func(game *driverGame) error {
		expected, err := applicationStateWorld(state)
		if err != nil {
			return err
		}
		if !simulationStateEqual(game.World(), expected) {
			return fmt.Errorf("world state = %#v, want %s", game.World(), state)
		}
		return nil
	})
}

func createNoSavedApplicationState(w *world, _ map[string]string) error {
	startApplicationDriver(w)
	return nil
}

func changeFromInitialApplicationState(w *world, _ map[string]string) error {
	return replaceApplicationWorld(w, "B")
}

func restoreApplicationStateOnce(w *world, _ map[string]string) error {
	return restoreApplicationState(w, 1)
}

func assertInitialApplicationState(w *world, _ map[string]string) error {
	return withConcreteGame(w, func(game *driverGame) error {
		expected := newApplicationDriverGame().World()
		if !simulationStateEqual(game.World(), expected) {
			return fmt.Errorf("world state = %#v, want initial state", game.World())
		}
		return nil
	})
}

func runStateFileOperation(w *world, example map[string]string) error {
	operation, err := stringValue(example, "file_operation")
	if err != nil {
		return err
	}
	return withConcreteGame(w, func(game *driverGame) error {
		switch operation {
		case "save file":
			w.xspSavedFirst = game.SaveXSP()
		case "load file":
			return game.LoadXSP(stateFileXSP())
		default:
			return fmt.Errorf("unsupported file operation %q", operation)
		}
		return nil
	})
}

func setApplicationStateWorld(w *world, state string) error {
	game := newApplicationDriverGame()
	world, err := applicationStateWorld(state)
	if err != nil {
		return err
	}
	game.ReplaceWorld(world)
	w.appGame = game
	return nil
}

func replaceApplicationWorld(w *world, state string) error {
	return withConcreteGame(w, func(game *driverGame) error {
		world, err := applicationStateWorld(state)
		if err != nil {
			return err
		}
		game.ReplaceWorld(world)
		return nil
	})
}

func restoreApplicationState(w *world, count int) error {
	return withConcreteGame(w, func(game *driverGame) error {
		for i := 0; i < count; i++ {
			game.RestoreState()
		}
		return nil
	})
}

func applicationStateWorld(state string) (*sim.Simulation, error) {
	switch state {
	case "A":
		return stateAWorld(), nil
	case "B":
		return stateBWorld(), nil
	default:
		return nil, fmt.Errorf("unsupported state %q", state)
	}
}

func stateAWorld() *sim.Simulation {
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 1, Position: sim.Vec2{X: 10, Y: 20}, Mass: 2, Elasticity: 0.6, Fixed: true})
	_ = world.AddMass(sim.Mass{ID: 2, Position: sim.Vec2{X: 40, Y: 20}, Mass: 3, Elasticity: 0.7})
	_ = world.AddSpring(sim.Spring{ID: 3, MassA: 1, MassB: 2, RestLength: 30, SpringConstant: 8, Damping: 0.4})
	world.Parameters.Set("current mass", "state-a")
	world.Parameters.EnableWall("left")
	return world
}

func stateBWorld() *sim.Simulation {
	world := sim.NewWorld()
	_ = world.AddMass(sim.Mass{ID: 7, Position: sim.Vec2{X: 70, Y: 80}, Mass: 4})
	world.Parameters.Set("current mass", "state-b")
	return world
}

func stateFileXSP() string {
	return "#1.0\ncmas file-loaded\nmass 9 90 90 1 0\n"
}

func simulationStateEqual(actual *sim.Simulation, expected *sim.Simulation) bool {
	return slices.Equal(actual.Masses, expected.Masses) &&
		slices.Equal(actual.Springs, expected.Springs) &&
		reflect.DeepEqual(actual.Parameters, expected.Parameters)
}
