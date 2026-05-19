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
	enabledIndex, enabled, err := forceEnabledField(fields)
	if err != nil {
		return err
	}
	name := strings.Join(fields[1:enabledIndex], " ")
	values, err := forceValues(fields[enabledIndex+1:])
	if err != nil {
		return err
	}
	setForceValues(world, name, enabled, values)
	return nil
}

func forceEnabledField(fields []string) (int, string, error) {
	for i := 2; i < len(fields); i++ {
		enabled, err := booleanField(fields[i], "frce enabled")
		if err == nil {
			return i, enabled, nil
		}
	}
	_, err := booleanField(fields[2], "frce enabled")
	return 0, "", err
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

// mutate4go-manifest-begin
// {"version":1,"tested_at":"2026-05-19T10:16:14-05:00","module_hash":"8acc7e3ab2a7d38fac8395ab5b1ca010c9c991d3ce48a30e6d8435bc2005b020","functions":[{"id":"func/UsesOriginalXSpringiesCoordinates","name":"UsesOriginalXSpringiesCoordinates","line":45,"end_line":48,"hash":"631a3e8f94ed4b60aee7d857d65701c2ac1345d62971237540a002c4b70b2067"},{"id":"func/LoadXSP","name":"LoadXSP","line":50,"end_line":65,"hash":"010d221a8c28d6a0e17dd66999cfb7d3ee3d40f24f234c14934146c9fb7b5299"},{"id":"func/loadXSPLine","name":"loadXSPLine","line":67,"end_line":77,"hash":"f5e5ad3d5a286f0e6438fa6a64f9b3b6cb1e24c56889b4a5319cb47c9defe330"},{"id":"func/parameterLineLoader","name":"parameterLineLoader","line":99,"end_line":103,"hash":"d682b9eb2e5c2c1f46da113c804be274f1738a3cd26bb1573eeda3d59727af95"},{"id":"func/booleanParameterLineLoader","name":"booleanParameterLineLoader","line":105,"end_line":117,"hash":"2f8bf8263e56f86d2d7d6d8bbd7e59e0ee5b0a60cd5131137f66067a185b63b7"},{"id":"func/setParameterLine","name":"setParameterLine","line":119,"end_line":125,"hash":"ad02832a85916f779b13abbd160c76361f8cdc5b2d824c28191f6c6e99465a17"},{"id":"func/loadGridSnapLine","name":"loadGridSnapLine","line":127,"end_line":133,"hash":"907f1fc3fc0766bc004c8f75ab8597f6cc0dd0c095537887177131280bd3bafe"},{"id":"func/loadCenterLine","name":"loadCenterLine","line":135,"end_line":148,"hash":"a34f106330c60322381039afa6ee51057893576d0661fe849c9a94c9ea3dae5e"},{"id":"func/loadForceLine","name":"loadForceLine","line":150,"end_line":167,"hash":"29c6dd8c69e34cebabcbef55976c38105f719621e655e10f94740e972faea901"},{"id":"func/legacyForceLine","name":"legacyForceLine","line":169,"end_line":175,"hash":"ba0ab0c32a17f574798b577d1ee0e9d4059b6f028ff499042b3061d4dad7ef4b"},{"id":"func/setForceValues","name":"setForceValues","line":177,"end_line":187,"hash":"1da5dc08cb49bf54bbe9c9bd7939093ddfe662055d934b5134a4d6883ef06de6"},{"id":"func/loadLegacyForceLine","name":"loadLegacyForceLine","line":189,"end_line":199,"hash":"94a33996f5e18abf1c55ef52125d86acf822954646e599cd8954d3ec785fc5dd"},{"id":"func/legacyForceFields","name":"legacyForceFields","line":201,"end_line":220,"hash":"2a1a71a6be4eb12ba859cc7260891062e1c1fdf6022869e41c22b450178bdd19"},{"id":"func/legacyForceName","name":"legacyForceName","line":222,"end_line":227,"hash":"429343111d2b192606d5252bb33554a52560403f1aaa61a60f293a99a4dbf1d5"},{"id":"func/legacyForceValues","name":"legacyForceValues","line":231,"end_line":242,"hash":"eb78ae827b8e6048e0f811a98ef2f909dc0b18a7a76f9bd2bbcc284e864fafb8"},{"id":"func/forceValues","name":"forceValues","line":244,"end_line":254,"hash":"abdffd599f0dd0289adab16dec91c006955cbb3aaa07dcb4c00b1becb9799026"},{"id":"func/loadWallLine","name":"loadWallLine","line":256,"end_line":269,"hash":"d08999ce94c902f1cc0865aab144441f04115003825c1bb672e8fd86d58bf865"},{"id":"func/loadLegacyWallLine","name":"loadLegacyWallLine","line":271,"end_line":280,"hash":"ff9d5c4c6910f5659b24313cc9ccfdee5c5e6dbc0c6a62d63a6924cbf553e968"},{"id":"func/loadMassLine","name":"loadMassLine","line":282,"end_line":303,"hash":"6f2160e6cafe0671e119bd7d1e4d109c9cd4917883734d10d2eba852d944d210"},{"id":"func/positiveIDField","name":"positiveIDField","line":305,"end_line":314,"hash":"256ff6014a767d5b0063f8bf171647a126eb06f06eb9c05ec32c9f02b9ce7892"},{"id":"func/parsedMassValue","name":"parsedMassValue","line":316,"end_line":321,"hash":"2c21c5ad6661abe90cdef579eae60c5fa6d51b5754910c4cc3818f84c9e905de"},{"id":"func/massNumericFields","name":"massNumericFields","line":323,"end_line":337,"hash":"8b6ebe5509fa0d8e40cb4e4e27a4f42cc5586dbf8a6950bd8d5641f8f82b0a4e"},{"id":"func/massVelocityFields","name":"massVelocityFields","line":339,"end_line":346,"hash":"4a6a3d19a05e968c472679361d776bc59893e5b89b7c4f040b8c8fec57157a4d"},{"id":"func/vectorFields","name":"vectorFields","line":348,"end_line":358,"hash":"32fd2fb6a024d849c43ddefce7b50197d10b52c12a7ba7278d3db03dae931132"},{"id":"func/massValueFields","name":"massValueFields","line":360,"end_line":367,"hash":"41a938b287791b6fd1ef01257bf9fc21b9b393122593b90c8fe2373010780a77"},{"id":"func/loadSpringLine","name":"loadSpringLine","line":369,"end_line":378,"hash":"e06481e7e0ea0cad1b94b66a21760cc29cf594b02262586e2d5104a671034bc8"},{"id":"func/springFromFields","name":"springFromFields","line":380,"end_line":394,"hash":"64db5cef1164b1f30fe3677a5b3c01c6d04707d90347ee936ac8647a18674f9c"},{"id":"func/springFieldGroups","name":"springFieldGroups","line":396,"end_line":419,"hash":"403b1181c823846dfef2187c1216e4e72170e00f0aad66354c1509959f9c2aec"},{"id":"func/SaveXSP","name":"SaveXSP","line":421,"end_line":430,"hash":"33b87704de4646eddf0a704be03ab18f6e6a57a83a6003eeb46f6c572049489d"},{"id":"func/writeMassLines","name":"writeMassLines","line":432,"end_line":442,"hash":"38520f4279c3a1c1a1b29670cbd3ec1be1a0d0506b592bf76d8aa7819be08ca1"},{"id":"func/writeSpringLines","name":"writeSpringLines","line":444,"end_line":455,"hash":"2eb24782b66488f3eb4f1a0bdbd76b11d8d87aafa8cf6f5a454b8f14c56787f0"},{"id":"func/writeParameterLines","name":"writeParameterLines","line":457,"end_line":461,"hash":"c6c885befbaab1fd6ca204717c1e52276131d08dd20622b969e3d93a966a65b9"},{"id":"func/writeForceLines","name":"writeForceLines","line":463,"end_line":471,"hash":"345942293531ffc62af418260be40e42ffba7db3d084cb2d5bc9224588e7d4c3"},{"id":"func/forceValueSuffix","name":"forceValueSuffix","line":473,"end_line":484,"hash":"f00a1a1e2da7402a70ea29004dc36c5aeb6a6ceeaefc5058928e1f61d187d274"},{"id":"func/writeWallLines","name":"writeWallLines","line":486,"end_line":492,"hash":"32059b9f917cd33cf398626f3e1c459c7a1978c78aede42b2a70fde3cd639230"},{"id":"func/fileMassValue","name":"fileMassValue","line":494,"end_line":499,"hash":"54a608a07200ba96ee55b2502ba6e94bc3a2b2f480414500cfb492ae75313558"},{"id":"func/intField","name":"intField","line":501,"end_line":507,"hash":"f6a9374fb647c1dfce45b40048a5f4b943d40232fde0c17e56e6e6b757ca3053"},{"id":"func/floatField","name":"floatField","line":509,"end_line":515,"hash":"db715e847c94f8eb956cd844d5b86948c8794554a5f5853b0b115fd57fe2f148"},{"id":"func/formatFloat","name":"formatFloat","line":517,"end_line":519,"hash":"425b19bddab66f570e8c1211910b663faf30eb2251df59b49e521a4a1168b1af"},{"id":"func/booleanField","name":"booleanField","line":521,"end_line":537,"hash":"2962b0956be1d1734d97b12f6e3986c2b2926cb27f65d560e5f10b9498176dc0"},{"id":"func/ResolveXSPFilename","name":"ResolveXSPFilename","line":539,"end_line":548,"hash":"8e85e837d31293a7e7add34c6fcbb27ccda1e11a3a55acd690cb9de090fa0269"}]}
// mutate4go-manifest-end
