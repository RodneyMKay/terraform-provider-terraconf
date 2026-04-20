package terraconf

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestScenario represents a test case from the schema.yaml file.
type TestScenario struct {
	Name                 string `yaml:"name"`
	YamlPattern          string `yaml:"yamlPattern"`
	SchemaFile           string `yaml:"schemaFile"`
	ExpectedErrorMessage string `yaml:"expectedErrorMessage"`
}

func TestSchemaValidation(t *testing.T) {
	// Load test scenarios from schema.yaml
	schemaYamlPath := filepath.Join("testdata", "schema.yaml")
	scenariosData, err := os.ReadFile(schemaYamlPath)
	if err != nil {
		t.Fatalf("Failed to read schema.yaml: %v", err)
	}

	var scenarios []TestScenario
	if err := yaml.Unmarshal(scenariosData, &scenarios); err != nil {
		t.Fatalf("Failed to parse schema.yaml: %v", err)
	}

	for _, scenario := range scenarios {
		t.Run(scenario.Name, func(t *testing.T) {
			// Find all files matching the yaml pattern
			files, err := FindGlobFiles(scenario.YamlPattern)
			if err != nil {
				t.Fatalf("Failed to find files with pattern %s: %v", scenario.YamlPattern, err)
			}

			if len(files) == 0 {
				t.Fatalf("No files found matching pattern: %s", scenario.YamlPattern)
			}

			// Process each matched file
			for _, yamlFile := range files {
				// Load YAML file
				yamlData, err := LoadYAML(yamlFile)
				if err != nil {
					t.Fatalf("LoadYAML failed for %s: %v", yamlFile, err)
				}

				// Check against schema BEFORE adding annotations
				schemaErrors, err := CheckWithSchema(yamlData, scenario.SchemaFile)
				if err != nil {
					t.Fatalf("CheckWithSchema failed: %v", err)
				}

				// Add annotations for tracing AFTER schema validation
				err = AddAnnotations(yamlData, yamlFile)
				if err != nil {
					t.Fatalf("AddAnnotations failed: %v", err)
				}

				// Verify test input
				hasErrors := len(schemaErrors) > 0
				expectsErrors := scenario.ExpectedErrorMessage != ""

				if !hasErrors && expectsErrors {
					t.Fatalf("Expected schema validation errors but got none")
				}
				if !hasErrors && !expectsErrors {
					return // test passed
				}

				// Trace the error to get formatted output
				var errorDetails []string
				for _, schemaError := range schemaErrors {
					detail, err := TraceError(yamlFile, schemaError.Pointer, schemaError.Message)
					if err != nil {
						t.Fatalf("TraceError failed: %v", err)
					}

					errorDetails = append(errorDetails, detail)
				}

				// Join all error details
				actualError := strings.Join(errorDetails, "\n")

				// Compare with expected error message
				expectedError := strings.TrimSpace(scenario.ExpectedErrorMessage)
				actualError = strings.TrimSpace(actualError)

				if actualError != expectedError {
					t.Errorf("Error message mismatch:\nExpected:\n%s\n\nActual:\n%s", expectedError, actualError)
				}
			}
		})
	}
}
