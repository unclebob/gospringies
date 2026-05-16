package gherkin

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

type Feature struct {
	Name       string     `json:"name"`
	Background []Step     `json:"background,omitempty"`
	Scenarios  []Scenario `json:"scenarios"`
}

type Scenario struct {
	Name     string              `json:"name"`
	Steps    []Step              `json:"steps"`
	Examples []map[string]string `json:"examples"`
}

type Step struct {
	Keyword    string   `json:"keyword"`
	Text       string   `json:"text"`
	Parameters []string `json:"parameters,omitempty"`
}

var parameterPattern = regexp.MustCompile(`<([A-Za-z0-9_]+)>`)

func Parse(r io.Reader) (Feature, error) {
	var feature Feature
	var currentScenario *Scenario
	inBackground := false
	inExamples := false
	var headers []string

	scanner := bufio.NewScanner(r)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		switch {
		case strings.HasPrefix(line, "Feature:"):
			feature.Name = strings.TrimSpace(strings.TrimPrefix(line, "Feature:"))
			currentScenario = nil
			inBackground = false
			inExamples = false
		case line == "Background:":
			inBackground = true
			currentScenario = nil
			inExamples = false
		case strings.HasPrefix(line, "Scenario Outline:"):
			feature.Scenarios = append(feature.Scenarios, Scenario{
				Name:     strings.TrimSpace(strings.TrimPrefix(line, "Scenario Outline:")),
				Examples: []map[string]string{},
			})
			currentScenario = &feature.Scenarios[len(feature.Scenarios)-1]
			inBackground = false
			inExamples = false
			headers = nil
		case strings.HasPrefix(line, "Scenario:"):
			feature.Scenarios = append(feature.Scenarios, Scenario{
				Name:     strings.TrimSpace(strings.TrimPrefix(line, "Scenario:")),
				Examples: []map[string]string{},
			})
			currentScenario = &feature.Scenarios[len(feature.Scenarios)-1]
			inBackground = false
			inExamples = false
			headers = nil
		case line == "Examples:":
			if currentScenario == nil {
				return Feature{}, fmt.Errorf("line %d: examples outside scenario", lineNo)
			}
			inExamples = true
			headers = nil
		case inExamples && strings.HasPrefix(line, "|"):
			cells := parseTableRow(line)
			if headers == nil {
				headers = cells
				continue
			}
			if len(cells) != len(headers) {
				return Feature{}, fmt.Errorf("line %d: examples row has %d cells, expected %d", lineNo, len(cells), len(headers))
			}
			row := map[string]string{}
			for i, header := range headers {
				row[header] = cells[i]
			}
			currentScenario.Examples = append(currentScenario.Examples, row)
		case isStep(line):
			step := parseStep(line)
			if inBackground {
				feature.Background = append(feature.Background, step)
			} else if currentScenario != nil {
				currentScenario.Steps = append(currentScenario.Steps, step)
			} else {
				return Feature{}, fmt.Errorf("line %d: step outside background or scenario", lineNo)
			}
			inExamples = false
		default:
			continue
		}
	}
	if err := scanner.Err(); err != nil {
		return Feature{}, err
	}
	if feature.Name == "" {
		return Feature{}, fmt.Errorf("missing feature declaration")
	}
	return feature, nil
}

func ReadFile(path string) (Feature, error) {
	file, err := os.Open(path)
	if err != nil {
		return Feature{}, err
	}
	defer file.Close()
	return Parse(file)
}

func WriteJSON(feature Feature, path string) error {
	data, err := json.MarshalIndent(feature, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}

func ReadJSON(path string) (Feature, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Feature{}, err
	}
	var feature Feature
	if err := json.Unmarshal(data, &feature); err != nil {
		return Feature{}, err
	}
	return feature, nil
}

func isStep(line string) bool {
	for _, keyword := range []string{"Given ", "When ", "Then ", "And "} {
		if strings.HasPrefix(line, keyword) {
			return true
		}
	}
	return false
}

func parseStep(line string) Step {
	parts := strings.SplitN(line, " ", 2)
	text := strings.TrimSpace(parts[1])
	matches := parameterPattern.FindAllStringSubmatch(text, -1)
	parameters := make([]string, 0, len(matches))
	for _, match := range matches {
		parameters = append(parameters, match[1])
	}
	return Step{Keyword: parts[0], Text: text, Parameters: parameters}
}

func parseTableRow(line string) []string {
	line = strings.TrimSpace(line)
	line = strings.TrimPrefix(line, "|")
	line = strings.TrimSuffix(line, "|")
	parts := strings.Split(line, "|")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}
