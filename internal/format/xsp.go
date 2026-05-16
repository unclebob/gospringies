package format

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"springs/internal/sim"
)

var (
	ErrUnsupportedMarker   = errors.New("unsupported format marker")
	ErrMissingFinalNewline = errors.New("missing final newline")
)

func LoadXSP(text string) (*sim.Simulation, error) {
	if !strings.HasSuffix(text, "\n") {
		return nil, ErrMissingFinalNewline
	}
	lines := strings.Split(strings.TrimSuffix(text, "\n"), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "#1.0" {
		return nil, ErrUnsupportedMarker
	}
	world := sim.NewWorld()
	for index, line := range lines[1:] {
		if err := loadXSPLine(world, line); err != nil {
			return nil, fmt.Errorf("line %d: %w", index+2, err)
		}
	}
	return world, nil
}

func loadXSPLine(world *sim.Simulation, line string) error {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return nil
	}
	loaders := map[string]func(*sim.Simulation, []string) error{
		"cmas": func(w *sim.Simulation, f []string) error { return setParameterLine(w, f, "current mass") },
		"elas": func(w *sim.Simulation, f []string) error { return setParameterLine(w, f, "elasticity") },
		"kspr": func(w *sim.Simulation, f []string) error { return setParameterLine(w, f, "spring constant") },
		"kdmp": func(w *sim.Simulation, f []string) error { return setParameterLine(w, f, "damping") },
		"frce": loadForceLine,
		"wall": loadWallLine,
		"mass": loadMassLine,
		"spng": loadSpringLine,
	}
	loader, ok := loaders[fields[0]]
	if !ok {
		return fmt.Errorf("unsupported command %q", fields[0])
	}
	return loader(world, fields)
}

func setParameterLine(world *sim.Simulation, fields []string, name string) error {
	if len(fields) != 2 {
		return fmt.Errorf("%s expects one value", fields[0])
	}
	world.Parameters.Set(name, fields[1])
	return nil
}

func loadForceLine(world *sim.Simulation, fields []string) error {
	if len(fields) < 3 {
		return fmt.Errorf("frce expects name and enabled state")
	}
	values, err := forceValues(fields[3:])
	if err != nil {
		return err
	}
	force, _ := world.Parameters.Force(fields[1])
	force.Enabled = fields[2]
	if force.Values == nil {
		force.Values = map[string]string{}
	}
	for key, value := range values {
		force.Values[key] = value
	}
	world.Parameters.Forces[fields[1]] = force
	return nil
}

func forceValues(fields []string) (map[string]string, error) {
	values := map[string]string{}
	for _, field := range fields {
		key, value, ok := strings.Cut(field, "=")
		if !ok {
			return nil, fmt.Errorf("force value %q must be key=value", field)
		}
		values[key] = value
	}
	return values, nil
}

func loadWallLine(world *sim.Simulation, fields []string) error {
	if len(fields) != 3 {
		return fmt.Errorf("wall expects name and enabled state")
	}
	world.Parameters.Walls[fields[1]] = fields[2] == "true"
	return nil
}

func loadMassLine(world *sim.Simulation, fields []string) error {
	if len(fields) != 6 {
		return fmt.Errorf("mass expects id x y mass elasticity")
	}
	id, err := intField(fields[1], "mass id")
	if err != nil {
		return err
	}
	x, y, massValue, elasticity, err := massNumericFields(fields)
	if err != nil {
		return err
	}
	fixed := massValue < 0
	if fixed {
		massValue = -massValue
	}
	return world.AddMass(sim.Mass{
		ID:         id,
		Position:   sim.Vec2{X: x, Y: y},
		Mass:       massValue,
		Elasticity: elasticity,
		Fixed:      fixed,
	})
}

func massNumericFields(fields []string) (float64, float64, float64, float64, error) {
	x, err := floatField(fields[2], "mass x")
	if err != nil {
		return 0, 0, 0, 0, err
	}
	y, err := floatField(fields[3], "mass y")
	if err != nil {
		return 0, 0, 0, 0, err
	}
	massValue, err := floatField(fields[4], "mass value")
	if err != nil {
		return 0, 0, 0, 0, err
	}
	elasticity, err := floatField(fields[5], "mass elasticity")
	if err != nil {
		return 0, 0, 0, 0, err
	}
	return x, y, massValue, elasticity, nil
}

func loadSpringLine(world *sim.Simulation, fields []string) error {
	if len(fields) != 7 {
		return fmt.Errorf("spng expects id mass_a mass_b rest_length spring_constant damping")
	}
	spring, err := springFromFields(fields)
	if err != nil {
		return err
	}
	return world.AddSpring(spring)
}

func springFromFields(fields []string) (sim.Spring, error) {
	ids, floats, err := springFieldGroups(fields)
	if err != nil {
		return sim.Spring{}, err
	}
	return sim.Spring{
		ID:             ids[0],
		MassA:          ids[1],
		MassB:          ids[2],
		RestLength:     floats[0],
		SpringConstant: floats[1],
		Stiffness:      floats[1],
		Damping:        floats[2],
	}, nil
}

func springFieldGroups(fields []string) ([3]int, [3]float64, error) {
	intNames := []string{"spring id", "spring mass a", "spring mass b"}
	floatNames := []string{"spring rest length", "spring constant", "spring damping"}
	var ids [3]int
	var values [3]float64
	for i, name := range intNames {
		value, err := intField(fields[i+1], name)
		if err != nil {
			return ids, values, err
		}
		ids[i] = value
	}
	for i, name := range floatNames {
		value, err := floatField(fields[i+4], name)
		if err != nil {
			return ids, values, err
		}
		values[i] = value
	}
	return ids, values, nil
}

func SaveXSP(world *sim.Simulation) string {
	var builder strings.Builder
	builder.WriteString("#1.0\n")
	writeParameterLines(&builder, world)
	writeForceLines(&builder, world)
	writeWallLines(&builder, world)
	for _, mass := range world.Masses {
		builder.WriteString(fmt.Sprintf("mass %d %s %s %s %s\n",
			mass.ID,
			formatFloat(mass.Position.X),
			formatFloat(mass.Position.Y),
			formatFloat(fileMassValue(mass)),
			formatFloat(mass.Elasticity),
		))
	}
	for _, spring := range world.Springs {
		builder.WriteString(fmt.Sprintf("spng %d %d %d %s %s %s\n",
			spring.ID,
			spring.MassA,
			spring.MassB,
			formatFloat(spring.RestLength),
			formatFloat(spring.SpringConstant),
			formatFloat(spring.Damping),
		))
	}
	return builder.String()
}

func writeParameterLines(builder *strings.Builder, world *sim.Simulation) {
	lines := []struct {
		command string
		name    string
	}{
		{"cmas", "current mass"},
		{"elas", "elasticity"},
		{"kspr", "spring constant"},
		{"kdmp", "damping"},
	}
	for _, line := range lines {
		builder.WriteString(fmt.Sprintf("%s %s\n", line.command, world.Parameters.Value(line.name)))
	}
}

func writeForceLines(builder *strings.Builder, world *sim.Simulation) {
	if force, ok := world.Parameters.Force("gravity"); ok && force.Enabled == "true" {
		builder.WriteString(fmt.Sprintf("frce gravity true magnitude=%s direction=%s\n", force.Values["magnitude"], force.Values["direction"]))
	}
}

func writeWallLines(builder *strings.Builder, world *sim.Simulation) {
	for _, name := range []string{"top", "left", "right", "bottom"} {
		if enabled, _ := world.Parameters.WallEnabled(name); enabled {
			builder.WriteString(fmt.Sprintf("wall %s true\n", name))
		}
	}
}

func fileMassValue(mass sim.Mass) float64 {
	if mass.Fixed {
		return -mass.Mass
	}
	return mass.Mass
}

func intField(value, name string) (int, error) {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", name, err)
	}
	return parsed, nil
}

func floatField(value, name string) (float64, error) {
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", name, err)
	}
	return parsed, nil
}

func formatFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}
