package acceptance

import (
	"fmt"

	"springs/internal/sim"
)

func assertParameterDefault(w *world, example map[string]string) error {
	world, parameter, err := parameterFromExample(w, example)
	if err != nil {
		return err
	}
	value, err := stringValue(example, "value")
	if err != nil {
		return err
	}
	if err := requireSetMarker("expected default", value); err != nil {
		return err
	}
	return assertParameterHasDefault(world, parameter)
}

func assertForceEnabledState(w *world, example map[string]string) error {
	force, err := forceFromExample(w, example)
	if err != nil {
		return err
	}
	enabled, err := stringValue(example, "enabled")
	if err != nil {
		return err
	}
	return assertEnabledStateConfigured("force", enabled, force.Enabled != "")
}

func assertForceEditableParameters(w *world, example map[string]string) error {
	force, err := forceFromExample(w, example)
	if err != nil {
		return err
	}
	if len(force.Values) == 0 {
		return fmt.Errorf("force has no editable parameters")
	}
	return nil
}

func assertWallEnabledState(w *world, example map[string]string) error {
	world, err := domainWorld(w)
	if err != nil {
		return err
	}
	wall, err := stringValue(example, "wall")
	if err != nil {
		return err
	}
	enabled, err := stringValue(example, "enabled")
	if err != nil {
		return err
	}
	if _, ok := world.Parameters.WallEnabled(wall); !ok {
		return fmt.Errorf("wall %q enabled state is not set", wall)
	}
	return requireSetMarker("enabled", enabled)
}

func changeWorldParameter(w *world, example map[string]string) error {
	world := ensureDomainWorld(w)
	parameter, value, err := parameterChange(example)
	if err != nil {
		return err
	}
	world.Parameters.Set(parameter, value)
	return nil
}

func performWorldOperation(w *world, example map[string]string) error {
	world, err := domainWorld(w)
	if err != nil {
		return err
	}
	operation, err := stringValue(example, "operation")
	if err != nil {
		return err
	}
	return performNamedWorldOperation(world, operation, example)
}

func performNamedWorldOperation(world *sim.Simulation, operation string, example map[string]string) error {
	switch operation {
	case "reset":
		world.Reset()
		return nil
	case "load file":
		return loadWorldWithChangedParameter(world, example)
	case "insert file":
		return insertWorldWithChangedParameter(world, example)
	default:
		return fmt.Errorf("unsupported operation %q", operation)
	}
}

func loadWorldWithChangedParameter(world *sim.Simulation, example map[string]string) error {
	return applyWorldWithParameter(world, example, "loaded", (*sim.Simulation).LoadFrom)
}

func insertWorldWithChangedParameter(world *sim.Simulation, example map[string]string) error {
	return applyWorldWithParameter(world, example, "inserted", (*sim.Simulation).InsertFrom)
}

func applyWorldWithParameter(world *sim.Simulation, example map[string]string, value string, apply func(*sim.Simulation, *sim.Simulation)) error {
	other, err := worldWithParameter(example, value)
	if err != nil {
		return err
	}
	apply(world, other)
	return nil
}

func worldWithParameter(example map[string]string, value string) (*sim.Simulation, error) {
	world := sim.NewWorld()
	parameter, err := stringValue(example, "parameter")
	if err != nil {
		return nil, err
	}
	world.Parameters.Set(parameter, value)
	return world, nil
}

func assertParameterSource(w *world, example map[string]string) error {
	world, parameter, source, err := parameterSourceFields(w, example)
	if err != nil {
		return err
	}
	expected, err := expectedParameterValue(parameter, source, example)
	if err != nil {
		return err
	}
	return assertParameterValue(world, parameter, expected)
}

func expectedParameterValue(parameter, source string, example map[string]string) (string, error) {
	switch source {
	case "default value":
		return sim.DefaultParameters().Value(parameter), nil
	case "value from loaded file":
		return "loaded", nil
	case "existing world value":
		return stringValue(example, "changed_value")
	default:
		return "", fmt.Errorf("unsupported expected value source %q", source)
	}
}

func forceFromExample(w *world, example map[string]string) (sim.ForceConfig, error) {
	world, err := domainWorld(w)
	if err != nil {
		return sim.ForceConfig{}, err
	}
	name, err := stringValue(example, "force")
	if err != nil {
		return sim.ForceConfig{}, err
	}
	force, ok := world.Parameters.Force(name)
	if !ok {
		return sim.ForceConfig{}, fmt.Errorf("force %q not configured", name)
	}
	return force, nil
}

func parameterFromExample(w *world, example map[string]string) (*sim.Simulation, string, error) {
	world, err := domainWorld(w)
	if err != nil {
		return nil, "", err
	}
	parameter, err := stringValue(example, "parameter")
	if err != nil {
		return nil, "", err
	}
	return world, parameter, nil
}

func parameterSourceFields(w *world, example map[string]string) (*sim.Simulation, string, string, error) {
	world, parameter, err := parameterFromExample(w, example)
	if err != nil {
		return nil, "", "", err
	}
	source, err := stringValue(example, "expected_value_source")
	if err != nil {
		return nil, "", "", err
	}
	return world, parameter, source, nil
}

func parameterChange(example map[string]string) (string, string, error) {
	return stringPair(example, "parameter", "changed_value")
}

func requireSetMarker(label, value string) error {
	if value != "set" {
		return fmt.Errorf("unsupported %s marker %q", label, value)
	}
	return nil
}

func assertEnabledStateConfigured(label, marker string, configured bool) error {
	if err := requireSetMarker("enabled", marker); err != nil {
		return err
	}
	if !configured {
		return fmt.Errorf("%s enabled state is not set", label)
	}
	return nil
}

func assertParameterHasDefault(world *sim.Simulation, parameter string) error {
	if !world.Parameters.Has(parameter) || world.Parameters.Value(parameter) == "" {
		return fmt.Errorf("parameter %q has no default value", parameter)
	}
	return nil
}

func assertParameterValue(world *sim.Simulation, parameter, expected string) error {
	if got := world.Parameters.Value(parameter); got != expected {
		return fmt.Errorf("expected parameter %q to be %q, got %q", parameter, expected, got)
	}
	return nil
}
