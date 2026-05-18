package format

import (
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"springs/internal/sim"
)

var (
	ErrUnsupportedMarker   = errors.New("unsupported format marker")
	ErrMissingFinalNewline = errors.New("missing final newline")
	ErrBlankLine           = errors.New("blank lines not allowed")
	ErrNonPositiveID       = errors.New("ids must be positive")
)

var xspParameterLines = []struct {
	command string
	name    string
}{
	{"cmas", "current mass"},
	{"elas", "elasticity"},
	{"kspr", "spring constant"},
	{"kdmp", "damping"},
	{"fixm", "fixed mass"},
	{"shws", "show springs"},
	{"cent", "center mass"},
	{"visc", "viscosity"},
	{"stck", "stickiness"},
	{"step", "timestep"},
	{"prec", "precision"},
	{"adpt", "adaptive timestep"},
	{"gsnp", "grid snap"},
}

var xspForceNames = []string{"gravity", "center attraction", "center of mass attraction", "wall repulsion", "mass collision"}
var xspForceValueKeys = []string{"magnitude", "direction", "exponent", "damping"}
var xspWallNames = []string{"top", "left", "right", "bottom"}

const originalXSpringiesMarker = "#1.0 *** XSpringies data file"

func UsesOriginalXSpringiesCoordinates(text string) bool {
	line, _, _ := strings.Cut(text, "\n")
	return strings.Contains(line, "XSpringies data file")
}

func LoadXSP(text string) (*sim.Simulation, error) {
	if !strings.HasSuffix(text, "\n") {
		return nil, ErrMissingFinalNewline
	}
	lines := strings.Split(strings.TrimSuffix(text, "\n"), "\n")
	if len(lines) == 0 || !strings.HasPrefix(strings.TrimSpace(lines[0]), "#1.0") {
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
	if strings.TrimSpace(line) == "" {
		return ErrBlankLine
	}
	fields := strings.Fields(line)
	loader, ok := xspLoaders[fields[0]]
	if !ok {
		return fmt.Errorf("unsupported command %q", fields[0])
	}
	return loader(world, fields)
}

var xspLoaders = map[string]func(*sim.Simulation, []string) error{
	"cmas": parameterLineLoader("current mass"),
	"elas": parameterLineLoader("elasticity"),
	"kspr": parameterLineLoader("spring constant"),
	"kdmp": parameterLineLoader("damping"),
	"fixm": booleanParameterLineLoader("fixed mass"),
	"shws": booleanParameterLineLoader("show springs"),
	"cent": loadCenterLine,
	"frce": loadForceLine,
	"visc": parameterLineLoader("viscosity"),
	"stck": parameterLineLoader("stickiness"),
	"step": parameterLineLoader("timestep"),
	"prec": parameterLineLoader("precision"),
	"adpt": booleanParameterLineLoader("adaptive timestep"),
	"gsnp": loadGridSnapLine,
	"wall": loadWallLine,
	"mass": loadMassLine,
	"spng": loadSpringLine,
}

func parameterLineLoader(name string) func(*sim.Simulation, []string) error {
	return func(world *sim.Simulation, fields []string) error {
		return setParameterLine(world, fields, name)
	}
}

func booleanParameterLineLoader(name string) func(*sim.Simulation, []string) error {
	return func(world *sim.Simulation, fields []string) error {
		if len(fields) != 2 {
			return fmt.Errorf("%s expects one value", fields[0])
		}
		value, err := booleanField(fields[1], fields[0])
		if err != nil {
			return err
		}
		world.Parameters.Set(name, value)
		return nil
	}
}

func setParameterLine(world *sim.Simulation, fields []string, name string) error {
	if len(fields) != 2 {
		return fmt.Errorf("%s expects one value", fields[0])
	}
	world.Parameters.Set(name, fields[1])
	return nil
}

func loadGridSnapLine(world *sim.Simulation, fields []string) error {
	if len(fields) != 2 && len(fields) != 3 {
		return fmt.Errorf("gsnp expects grid snap value")
	}
	world.Parameters.Set("grid snap", fields[1])
	return nil
}

func loadCenterLine(world *sim.Simulation, fields []string) error {
	if len(fields) != 2 {
		return fmt.Errorf("cent expects one value")
	}
	id, err := intField(fields[1], "center mass id")
	if err != nil {
		return err
	}
	if id == 0 || id < -1 {
		return ErrNonPositiveID
	}
	world.Parameters.Set("center mass", strconv.Itoa(id))
	return nil
}

func loadForceLine(world *sim.Simulation, fields []string) error {
	if legacyForceLine(fields) {
		return loadLegacyForceLine(world, fields)
	}
	if len(fields) < 3 {
		return fmt.Errorf("frce expects name and enabled state")
	}
	enabled, err := booleanField(fields[2], "frce enabled")
	if err != nil {
		return err
	}
	values, err := forceValues(fields[3:])
	if err != nil {
		return err
	}
	setForceValues(world, fields[1], enabled, values)
	return nil
}

func legacyForceLine(fields []string) bool {
	if len(fields) != 5 {
		return false
	}
	_, err := intField(fields[1], "force id")
	return err == nil
}

func setForceValues(world *sim.Simulation, name string, enabled string, values map[string]string) {
	force, _ := world.Parameters.Force(name)
	force.Enabled = enabled
	if force.Values == nil {
		force.Values = map[string]string{}
	}
	for key, value := range values {
		force.Values[key] = value
	}
	world.Parameters.Forces[name] = force
}

func loadLegacyForceLine(world *sim.Simulation, fields []string) error {
	name, enabled, first, second, err := legacyForceFields(fields)
	if err != nil {
		return err
	}
	force, _ := world.Parameters.Force(name)
	force.Enabled = enabled
	force.Values = legacyForceValues(name, first, second)
	world.Parameters.Forces[name] = force
	return nil
}

func legacyForceFields(fields []string) (string, string, float64, float64, error) {
	forceID, err := intField(fields[1], "force id")
	if err != nil {
		return "", "", 0, 0, err
	}
	name, err := legacyForceName(forceID)
	if err != nil {
		return "", "", 0, 0, err
	}
	enabled, err := booleanField(fields[2], "frce enabled")
	if err != nil {
		return "", "", 0, 0, err
	}
	first, err := floatField(fields[3], "force first parameter")
	if err != nil {
		return "", "", 0, 0, err
	}
	second, err := floatField(fields[4], "force second parameter")
	return name, enabled, first, second, err
}

func legacyForceName(forceID int) (string, error) {
	if forceID < 0 || forceID >= len(legacyForceNames) {
		return "", fmt.Errorf("unsupported force id %d", forceID)
	}
	return legacyForceNames[forceID], nil
}

var legacyForceNames = []string{"gravity", "center attraction", "center of mass attraction", "wall repulsion", "mass collision"}

func legacyForceValues(name string, first, second float64) map[string]string {
	values := map[string]string{"magnitude": formatFloat(first)}
	switch name {
	case "gravity":
		values["direction"] = formatFloat(second)
	case "center of mass attraction":
		values["damping"] = formatFloat(second)
	default:
		values["exponent"] = formatFloat(second)
	}
	return values
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
	if len(fields) == 5 {
		return loadLegacyWallLine(world, fields)
	}
	if len(fields) != 3 {
		return fmt.Errorf("wall expects name and enabled state")
	}
	enabled, err := booleanField(fields[2], "wall enabled")
	if err != nil {
		return err
	}
	world.Parameters.Walls[fields[1]] = enabled == "true"
	return nil
}

func loadLegacyWallLine(world *sim.Simulation, fields []string) error {
	for i, name := range []string{"top", "left", "right", "bottom"} {
		enabled, err := booleanField(fields[i+1], "wall enabled")
		if err != nil {
			return err
		}
		world.Parameters.Walls[name] = enabled == "true"
	}
	return nil
}

func loadMassLine(world *sim.Simulation, fields []string) error {
	if len(fields) != 6 && len(fields) != 8 {
		return fmt.Errorf("mass expects id x y mass elasticity")
	}
	id, err := positiveIDField(fields[1], "mass id")
	if err != nil {
		return err
	}
	position, velocity, massValue, elasticity, err := massNumericFields(fields)
	if err != nil {
		return err
	}
	massValue, fixed := parsedMassValue(massValue)
	return world.AddMass(sim.Mass{
		ID:         id,
		Position:   position,
		Velocity:   velocity,
		Mass:       massValue,
		Elasticity: elasticity,
		Fixed:      fixed,
	})
}

func positiveIDField(field string, name string) (int, error) {
	id, err := intField(field, name)
	if err != nil {
		return 0, err
	}
	if id <= 0 {
		return 0, ErrNonPositiveID
	}
	return id, nil
}

func parsedMassValue(massValue float64) (float64, bool) {
	if massValue < 0 {
		return -massValue, true
	}
	return massValue, false
}

func massNumericFields(fields []string) (sim.Vec2, sim.Vec2, float64, float64, error) {
	position, err := vectorFields(fields[2:4], "mass x", "mass y")
	if err != nil {
		return sim.Vec2{}, sim.Vec2{}, 0, 0, err
	}
	velocity, massIndex, err := massVelocityFields(fields)
	if err != nil {
		return sim.Vec2{}, sim.Vec2{}, 0, 0, err
	}
	massValue, elasticity, err := massValueFields(fields[massIndex : massIndex+2])
	if err != nil {
		return sim.Vec2{}, sim.Vec2{}, 0, 0, err
	}
	return position, velocity, massValue, elasticity, nil
}

func massVelocityFields(fields []string) (sim.Vec2, int, error) {
	massIndex := 4
	if len(fields) != 8 {
		return sim.Vec2{}, massIndex, nil
	}
	velocity, err := vectorFields(fields[4:6], "mass velocity x", "mass velocity y")
	return velocity, 6, err
}

func vectorFields(fields []string, xName string, yName string) (sim.Vec2, error) {
	x, err := floatField(fields[0], xName)
	if err != nil {
		return sim.Vec2{}, err
	}
	y, err := floatField(fields[1], yName)
	if err != nil {
		return sim.Vec2{}, err
	}
	return sim.Vec2{X: x, Y: y}, nil
}

func massValueFields(fields []string) (float64, float64, error) {
	massValue, err := floatField(fields[0], "mass value")
	if err != nil {
		return 0, 0, err
	}
	elasticity, err := floatField(fields[1], "mass elasticity")
	return massValue, elasticity, err
}

func loadSpringLine(world *sim.Simulation, fields []string) error {
	if len(fields) != 7 {
		return fmt.Errorf("spng expects id mass_a mass_b spring_constant damping rest_length")
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
		RestLength:     floats[2],
		SpringConstant: floats[0],
		Stiffness:      floats[0],
		Damping:        floats[1],
	}, nil
}

func springFieldGroups(fields []string) ([3]int, [3]float64, error) {
	intNames := []string{"spring id", "spring mass a", "spring mass b"}
	floatNames := []string{"spring constant", "spring damping", "spring rest length"}
	var ids [3]int
	var values [3]float64
	for i, name := range intNames {
		value, err := intField(fields[i+1], name)
		if err != nil {
			return ids, values, err
		}
		if value <= 0 {
			return ids, values, ErrNonPositiveID
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
	builder.WriteString(originalXSpringiesMarker + "\n")
	writeParameterLines(&builder, world)
	writeForceLines(&builder, world)
	writeWallLines(&builder, world)
	writeMassLines(&builder, world)
	writeSpringLines(&builder, world)
	return builder.String()
}

func writeMassLines(builder *strings.Builder, world *sim.Simulation) {
	for _, mass := range world.Masses {
		builder.WriteString(fmt.Sprintf("mass %d %s %s %s %s\n",
			mass.ID,
			formatFloat(mass.Position.X),
			formatFloat(mass.Position.Y),
			formatFloat(fileMassValue(mass)),
			formatFloat(mass.Elasticity),
		))
	}
}

func writeSpringLines(builder *strings.Builder, world *sim.Simulation) {
	for _, spring := range world.Springs {
		builder.WriteString(fmt.Sprintf("spng %d %d %d %s %s %s\n",
			spring.ID,
			spring.MassA,
			spring.MassB,
			formatFloat(spring.SpringConstant),
			formatFloat(spring.Damping),
			formatFloat(spring.RestLength),
		))
	}
}

func writeParameterLines(builder *strings.Builder, world *sim.Simulation) {
	for _, line := range xspParameterLines {
		builder.WriteString(fmt.Sprintf("%s %s\n", line.command, world.Parameters.Value(line.name)))
	}
}

func writeForceLines(builder *strings.Builder, world *sim.Simulation) {
	for _, name := range xspForceNames {
		force, ok := world.Parameters.Force(name)
		if !ok {
			continue
		}
		builder.WriteString(fmt.Sprintf("frce %s %s%s\n", name, force.Enabled, forceValueSuffix(force.Values)))
	}
}

func forceValueSuffix(values map[string]string) string {
	var parts []string
	for _, key := range xspForceValueKeys {
		if value, ok := values[key]; ok {
			parts = append(parts, key+"="+value)
		}
	}
	if len(parts) == 0 {
		return ""
	}
	return " " + strings.Join(parts, " ")
}

func writeWallLines(builder *strings.Builder, world *sim.Simulation) {
	for _, name := range xspWallNames {
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

func booleanField(value, name string) (string, error) {
	switch value {
	case "true", "1":
		return "true", nil
	case "false", "0":
		return "false", nil
	default:
		number, err := strconv.Atoi(value)
		if err != nil {
			return "", fmt.Errorf("%s: %w", name, err)
		}
		if number == 0 {
			return "false", nil
		}
		return "true", nil
	}
}

func ResolveXSPFilename(filename string, springDir string) string {
	resolved := filename
	if filepath.Ext(resolved) == "" {
		resolved += ".xsp"
	}
	if springDir != "" && !filepath.IsAbs(resolved) {
		resolved = filepath.Join(springDir, resolved)
	}
	return resolved
}
