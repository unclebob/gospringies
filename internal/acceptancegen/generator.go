package acceptancegen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

func GenerateGoTest(jsonIRPath, outputPath string) error {
	return generateGoTest(jsonIRPath, outputPath, "")
}

func GenerateTaggedGoTest(jsonIRPath, outputPath, tag string) error {
	return generateGoTest(jsonIRPath, outputPath, tag)
}

func generateGoTest(jsonIRPath, outputPath, buildTag string) error {
	feature, err := readFeatureForGeneration(jsonIRPath)
	if err != nil {
		return err
	}
	source, err := generatedTestSource(feature, outputPath, buildTag)
	if err != nil {
		return err
	}
	return writeFormattedGo(outputPath, source)
}

func generatedTestSource(feature any, outputPath, buildTag string) ([]byte, error) {
	embedded, err := json.MarshalIndent(feature, "\t", "\t")
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if buildTag != "" {
		fmt.Fprintf(&buf, "//go:build %s\n\n", buildTag)
	}
	fmt.Fprintf(&buf, "package generated\n\n")
	fmt.Fprintf(&buf, "import (\n\t\"encoding/json\"\n\t\"testing\"\n\n\t\"springs/internal/acceptance\"\n\t\"springs/internal/gherkin\"\n)\n\n")
	fmt.Fprintf(&buf, "func %s(t *testing.T) {\n", generatedTestName(outputPath))
	fmt.Fprintf(&buf, "\tvar feature gherkin.Feature\n")
	fmt.Fprintf(&buf, "\tdata := []byte(`%s`)\n", string(embedded))
	fmt.Fprintf(&buf, "\tif err := json.Unmarshal(data, &feature); err != nil {\n\t\tt.Fatal(err)\n\t}\n")
	fmt.Fprintf(&buf, "\tif err := acceptance.RunFeature(feature); err != nil {\n\t\tt.Fatal(err)\n\t}\n")
	fmt.Fprintf(&buf, "}\n")
	return buf.Bytes(), nil
}

func writeFormattedGo(outputPath string, source []byte) error {
	formatted, err := format.Source(source)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(outputPath, formatted, 0o644)
}

func generatedTestName(outputPath string) string {
	name := strings.TrimSuffix(filepath.Base(outputPath), filepath.Ext(outputPath))
	var builder strings.Builder
	builder.WriteString("TestGeneratedAcceptance")
	capitalizeNext := true
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if capitalizeNext {
				r = unicode.ToUpper(r)
				capitalizeNext = false
			}
			builder.WriteRune(r)
			continue
		}
		capitalizeNext = true
	}
	return builder.String()
}

func readFeatureForGeneration(path string) (any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var feature any
	if err := json.Unmarshal(data, &feature); err != nil {
		return nil, err
	}
	return feature, nil
}
