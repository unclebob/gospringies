package edit

import (
	"fmt"
	"strconv"

	"springs/internal/sim"
)

var controlChangeHandlers = map[string]func(*Editor, string) error{
	"mass": func(e *Editor, value string) error {
		return e.changeMassFloat(value, "current mass", func(mass *sim.Mass, parsed float64) { mass.Mass = parsed })
	},
	"elasticity": func(e *Editor, value string) error {
		return e.changeMassFloat(value, "elasticity", func(mass *sim.Mass, parsed float64) { mass.Elasticity = parsed })
	},
	"fixed": (*Editor).changeMassFixed,
	"Kspring": func(e *Editor, value string) error {
		return e.changeSpringFloat(value, "spring constant", setSpringConstant)
	},
	"Kdamp": func(e *Editor, value string) error {
		return e.changeSpringFloat(value, "damping", func(spring *sim.Spring, parsed float64) { spring.Damping = parsed })
	},
}

func (e *Editor) ChangeControl(control string, value string) error {
	change, ok := controlChangeHandlers[control]
	if !ok {
		return fmt.Errorf("unsupported control %q", control)
	}
	return change(e, value)
}

func (e *Editor) SetRestLength() error {
	for i := range e.World.Springs {
		if !e.SelectedSprings[e.World.Springs[i].ID] {
			continue
		}
		length, err := e.currentSpringLength(e.World.Springs[i])
		if err != nil {
			return err
		}
		e.World.Springs[i].RestLength = length
	}
	return nil
}

func (e *Editor) changeMassFloat(value string, defaultParameter string, update func(*sim.Mass, float64)) error {
	return e.changeFloat(value, defaultParameter, e.hasSelectedMass(), func(parsed float64) {
		for i := range e.World.Masses {
			if e.SelectedMasses[e.World.Masses[i].ID] {
				update(&e.World.Masses[i], parsed)
			}
		}
	})
}

func (e *Editor) changeMassFixed(value string) error {
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fmt.Errorf("invalid fixed value %q", value)
	}
	for i := range e.World.Masses {
		if e.SelectedMasses[e.World.Masses[i].ID] {
			e.World.Masses[i].Fixed = parsed
		}
	}
	return nil
}

func (e *Editor) changeSpringFloat(value string, defaultParameter string, update func(*sim.Spring, float64)) error {
	return e.changeFloat(value, defaultParameter, e.hasSelectedSpring(), func(parsed float64) {
		for i := range e.World.Springs {
			if e.SelectedSprings[e.World.Springs[i].ID] {
				update(&e.World.Springs[i], parsed)
			}
		}
	})
}

func (e *Editor) changeFloat(value string, defaultParameter string, compatibleSelection bool, update func(float64)) error {
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fmt.Errorf("invalid %s value %q", defaultParameter, value)
	}
	if !compatibleSelection {
		e.World.Parameters.Set(defaultParameter, value)
		return nil
	}
	update(parsed)
	return nil
}

func setSpringConstant(spring *sim.Spring, value float64) {
	spring.SpringConstant = value
	spring.Stiffness = value
}

func (e *Editor) currentSpringLength(spring sim.Spring) (float64, error) {
	a, okA := e.World.MassByID(spring.MassA)
	b, okB := e.World.MassByID(spring.MassB)
	if !okA || !okB {
		return 0, fmt.Errorf("missing spring endpoint")
	}
	return distance(a.Position, b.Position), nil
}

func (e *Editor) hasSelectedMass() bool {
	return selectedObjectExists(e.World.Masses, func(mass sim.Mass) bool { return e.SelectedMasses[mass.ID] })
}

func (e *Editor) hasSelectedSpring() bool {
	return selectedObjectExists(e.World.Springs, func(spring sim.Spring) bool { return e.SelectedSprings[spring.ID] })
}

func selectedObjectExists[T any](objects []T, selected func(T) bool) bool {
	for _, object := range objects {
		if selected(object) {
			return true
		}
	}
	return false
}
