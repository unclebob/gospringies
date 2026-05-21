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
	"Wall": (*Editor).changeSpringWall,
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

func (e *Editor) changeSpringWall(value string) error {
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fmt.Errorf("invalid wall value %q", value)
	}
	for i := range e.World.Springs {
		if e.SelectedSprings[e.World.Springs[i].ID] {
			e.World.Springs[i].Wall = parsed
		}
	}
	return nil
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

// mutate4go-manifest-begin
// {"version":1,"tested_at":"2026-05-19T10:12:18-05:00","module_hash":"fe05505c2dfeb8e19ba30e906a7a656c0046abbe88dd6e3857997ef6e630fd9c","functions":[{"id":"func/Editor.ChangeControl","name":"Editor.ChangeControl","line":26,"end_line":32,"hash":"0838f6df828f9b4043b7c7cdfbd1d98320c4d68a653995fee1f3c0c89622c095"},{"id":"func/Editor.SetRestLength","name":"Editor.SetRestLength","line":34,"end_line":46,"hash":"aa3f25eb24be2f7379b69712e0fd727681a6eb0828e2c760f804f6cf458cd0e6"},{"id":"func/Editor.changeMassFloat","name":"Editor.changeMassFloat","line":48,"end_line":56,"hash":"23e24e53d8f497da8d34e806f990489cfd723cdf4be1c2349533dd183a022975"},{"id":"func/Editor.changeMassFixed","name":"Editor.changeMassFixed","line":58,"end_line":69,"hash":"66d265c6a840a8efa0257d3c9c80319c3d06c8f65d4cca81d732b321885f4609"},{"id":"func/Editor.changeSpringFloat","name":"Editor.changeSpringFloat","line":71,"end_line":79,"hash":"a09bbb1ccc63cd5e00e485fdfec3cda12c8225d7f9929d4b3c609c47c0ecbe60"},{"id":"func/Editor.changeFloat","name":"Editor.changeFloat","line":81,"end_line":92,"hash":"e93f1aa763a084b41ede524f7e2fbc8280e59cfa9655583c5e1fb761a2c1e7b7"},{"id":"func/setSpringConstant","name":"setSpringConstant","line":94,"end_line":97,"hash":"bfd9c2f7ed6176a9abbb4ecac512f7a13081000ce360d54947d6f336ea10957a"},{"id":"func/Editor.currentSpringLength","name":"Editor.currentSpringLength","line":99,"end_line":106,"hash":"36b3bfa414e56241e0e5bd71d5abc657a264977a63b6faef1f944a09c1792005"},{"id":"func/Editor.hasSelectedMass","name":"Editor.hasSelectedMass","line":108,"end_line":110,"hash":"e11f9a884a0e1adb66dc747159173553245d10bd3ac49000b8aa95f3747bc212"},{"id":"func/Editor.hasSelectedSpring","name":"Editor.hasSelectedSpring","line":112,"end_line":114,"hash":"643520703da3e053924ef0be43de528c9360985082d497baf0b63ed38760f4e7"},{"id":"func/selectedObjectExists","name":"selectedObjectExists","line":116,"end_line":123,"hash":"387d173ea7410f70d6aea4818f38653d73d789de753a50e5e34ad8baf9ccc73a"}]}
// mutate4go-manifest-end
