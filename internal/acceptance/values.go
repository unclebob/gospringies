package acceptance

import (
	"fmt"
	"strconv"
	"strings"
)

func stringValue(example map[string]string, key string) (string, error) {
	value, ok := example[key]
	if !ok {
		return "", fmt.Errorf("missing example value %s", key)
	}
	return value, nil
}

func stringPair(example map[string]string, firstKey, secondKey string) (string, string, error) {
	first, err := stringValue(example, firstKey)
	if err != nil {
		return "", "", err
	}
	second, err := stringValue(example, secondKey)
	if err != nil {
		return "", "", err
	}
	return first, second, nil
}

func intValue(example map[string]string, key string) (int, error) {
	value, err := stringValue(example, key)
	if err != nil {
		return 0, err
	}
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0, fmt.Errorf("invalid integer %s=%q", key, value)
	}
	return parsed, nil
}

func floatValue(example map[string]string, key string) (float64, error) {
	value, err := stringValue(example, key)
	if err != nil {
		return 0, err
	}
	parsed, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil {
		return 0, fmt.Errorf("invalid float %s=%q", key, value)
	}
	return parsed, nil
}

func boolValue(example map[string]string, key string) (bool, error) {
	value, err := stringValue(example, key)
	if err != nil {
		return false, err
	}
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, fmt.Errorf("invalid bool %s=%q", key, value)
	}
}
