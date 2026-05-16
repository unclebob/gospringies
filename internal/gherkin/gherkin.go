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
	parser := &lineParser{}

	scanner := bufio.NewScanner(r)
	parser.lineNo = 0
	for scanner.Scan() {
		parser.lineNo++
		line := strings.TrimSpace(scanner.Text())
		if err := parser.parseLine(line); err != nil {
			return Feature{}, err
		}
	}
	if err := scanner.Err(); err != nil {
		return Feature{}, err
	}
	if parser.feature.Name == "" {
		return Feature{}, fmt.Errorf("missing feature declaration")
	}
	return parser.feature, nil
}

type lineParser struct {
	feature         Feature
	currentScenario *Scenario
	inBackground    bool
	inExamples      bool
	headers         []string
	lineNo          int
}

type lineHandler func(*lineParser, string) (bool, error)

func (p *lineParser) parseLine(line string) error {
	for _, handler := range lineHandlers {
		handled, err := handler(p, line)
		if handled || err != nil {
			return err
		}
	}
	return nil
}

var lineHandlers = []lineHandler{
	ignoreBlankOrComment,
	parseFeatureLine,
	parseBackgroundLine,
	parseScenarioOutlineLine,
	parseScenarioLine,
	parseExamplesLine,
	parseExampleRowLine,
	parseStepLine,
}

func ignoreBlankOrComment(_ *lineParser, line string) (bool, error) {
	return line == "" || strings.HasPrefix(line, "#"), nil
}

func parseFeatureLine(p *lineParser, line string) (bool, error) {
	if !strings.HasPrefix(line, "Feature:") {
		return false, nil
	}
	p.startFeature(line)
	return true, nil
}

func parseBackgroundLine(p *lineParser, line string) (bool, error) {
	if line != "Background:" {
		return false, nil
	}
	p.startBackground()
	return true, nil
}

func parseScenarioOutlineLine(p *lineParser, line string) (bool, error) {
	if !strings.HasPrefix(line, "Scenario Outline:") {
		return false, nil
	}
	p.startScenario(strings.TrimSpace(strings.TrimPrefix(line, "Scenario Outline:")))
	return true, nil
}

func parseScenarioLine(p *lineParser, line string) (bool, error) {
	if !strings.HasPrefix(line, "Scenario:") {
		return false, nil
	}
	p.startScenario(strings.TrimSpace(strings.TrimPrefix(line, "Scenario:")))
	return true, nil
}

func parseExamplesLine(p *lineParser, line string) (bool, error) {
	if line != "Examples:" {
		return false, nil
	}
	return true, p.startExamples()
}

func parseExampleRowLine(p *lineParser, line string) (bool, error) {
	if !p.inExamples || !strings.HasPrefix(line, "|") {
		return false, nil
	}
	return true, p.addExampleRow(line)
}

func parseStepLine(p *lineParser, line string) (bool, error) {
	if !isStep(line) {
		return false, nil
	}
	return true, p.addStep(line)
}

func (p *lineParser) startFeature(line string) {
	p.feature.Name = strings.TrimSpace(strings.TrimPrefix(line, "Feature:"))
	p.currentScenario = nil
	p.inBackground = false
	p.inExamples = false
}

func (p *lineParser) startBackground() {
	p.inBackground = true
	p.currentScenario = nil
	p.inExamples = false
}

func (p *lineParser) startScenario(name string) {
	p.feature.Scenarios = append(p.feature.Scenarios, Scenario{
		Name:     name,
		Examples: []map[string]string{},
	})
	p.currentScenario = &p.feature.Scenarios[len(p.feature.Scenarios)-1]
	p.inBackground = false
	p.inExamples = false
	p.headers = nil
}

func (p *lineParser) startExamples() error {
	if p.currentScenario == nil {
		return fmt.Errorf("line %d: examples outside scenario", p.lineNo)
	}
	p.inExamples = true
	p.headers = nil
	return nil
}

func (p *lineParser) addExampleRow(line string) error {
	cells := parseTableRow(line)
	if p.headers == nil {
		p.headers = cells
		return nil
	}
	if len(cells) != len(p.headers) {
		return fmt.Errorf("line %d: examples row has %d cells, expected %d", p.lineNo, len(cells), len(p.headers))
	}
	p.currentScenario.Examples = append(p.currentScenario.Examples, exampleRow(p.headers, cells))
	return nil
}

func exampleRow(headers, cells []string) map[string]string {
	row := map[string]string{}
	for i, header := range headers {
		row[header] = cells[i]
	}
	return row
}

func (p *lineParser) addStep(line string) error {
	step := parseStep(line)
	if p.inBackground {
		p.feature.Background = append(p.feature.Background, step)
	} else if p.currentScenario != nil {
		p.currentScenario.Steps = append(p.currentScenario.Steps, step)
	} else {
		return fmt.Errorf("line %d: step outside background or scenario", p.lineNo)
	}
	p.inExamples = false
	return nil
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
