package acceptance

import "springs/internal/acceptancegen"

func GenerateGoTest(jsonIRPath, outputPath string) error {
	return acceptancegen.GenerateGoTest(jsonIRPath, outputPath)
}

func generateTaggedGoTest(jsonIRPath, outputPath, tag string) error {
	return acceptancegen.GenerateTaggedGoTest(jsonIRPath, outputPath, tag)
}
