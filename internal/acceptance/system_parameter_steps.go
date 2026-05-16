package acceptance

import (
	"fmt"

	"springs/internal/sim"
)

func assertParameterDefault(w *world, example map[string]string) error {
	world, err := domainWorld(w)
	if err != nil {
		return err
	}
	parameter, err := stringValue(example, "parameter")
	if err != nil {
		return err
	}
	value, err := stringValue(example, "value")
	if err != nil {
		return err
	}
	if value != "set" {
		return fmt.Errorf("unsupported expected default marker %q", value)
	}
	if !world.Parameters.Has(parameter) || world.Parameters.Value(parameter) == "" {
		return fmt.Errorf("parameter %q has no default value", parameter)
	}
	return nil
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
	if enabled != "set" {
		return fmt.Errorf("unsupported enabled marker %q", enabled)
	}
	if force.Enabled == "" {
		return fmt.Errorf("force enabled state is not set")
	}
	return nil
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
	if enabled != "set" {
		return fmt.Errorf("unsupported enabled marker %q", enabled)
	}
	if _, ok := world.Parameters.WallEnabled(wall); !ok {
		return fmt.Errorf("wall %q enabled state is not set", wall)
	}
	return nil
}

func changeWorldParameter(w *world, example map[string]string) error {
	world := ensureDomainWorld(w)
	parameter, err := stringValue(example, "parameter")
	if err != nil {
		return err
	}
	value, err := stringValue(example, "changed_value")
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
	switch operation {
	case "reset":
		world.Reset()
	case "load file":
		loaded := sim.NewWorld()
		parameter, err := stringValue(example, "parameter")
		if err != nil {
			return err
		}
		loaded.Parameters.Set(parameter, "loaded")
		world.LoadFrom(loaded)
	case "insert file":
		inserted := sim.NewWorld()
		parameter, err := stringValue(example, "parameter")
		if err != nil {
			return err
		}
		inserted.Parameters.Set(parameter, "inserted")
		world.InsertFrom(inserted)
	default:
		return fmt.Errorf("unsupported operation %q", operation)
	}
	return nil
}

func assertParameterSource(w *world, example map[string]string) error {
	world, err := domainWorld(w)
	if err != nil {
		return err
	}
	parameter, err := stringValue(example, "parameter")
	if err != nil {
		return err
	}
	source, err := stringValue(example, "expected_value_source")
	if err != nil {
		return err
	}
	var expected string
	switch source {
	case "default value":
		expected = sim.DefaultParameters().Value(parameter)
	case "value from loaded file":
		expected = "loaded"
	case "existing world value":
		expected, err = stringValue(example, "changed_value")
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported expected value source %q", source)
	}
	if got := world.Parameters.Value(parameter); got != expected {
		return fmt.Errorf("expected parameter %q to be %q, got %q", parameter, expected, got)
	}
	return nil
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
