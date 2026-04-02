package terraconf

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

type traceCase struct {
	Name                 string `yaml:"name"`
	File                 string `yaml:"file"`
	ErrorMessage         string `yaml:"errorMessage"`
	InternalPointer      string `yaml:"internalPointer"`
	ExpectedErrorMessage string `yaml:"expectedErrorMessage"`
}

func TestTraceCases_FromYAML(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "trace.yaml"))
	if err != nil {
		t.Fatalf("reading trace.yaml: %v", err)
	}

	var cases []traceCase
	if err := yaml.Unmarshal(data, &cases); err != nil {
		t.Fatalf("unmarshal trace.yaml: %v", err)
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			pointer := NewJsonPointer(tc.InternalPointer)
			got, err := TraceError(tc.File, pointer, tc.ErrorMessage)

			if err != nil {
				t.Fatalf("TraceError: %v", err)
			}

			if got != tc.ExpectedErrorMessage {
				t.Fatalf("unexpected output for %s:\nGOT:\n%s\nEXPECTED:\n%s", tc.Name, got, tc.ExpectedErrorMessage)
			}
		})
	}
}
