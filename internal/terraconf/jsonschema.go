package terraconf

import (
	"fmt"
	"path/filepath"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// Represents an error found when validating a YAML document against a JSON
// schema. Message is the message that we want to display to the user and
// Pointer points to the location of the error.
type SchemaError struct {
	Message string
	Pointer JsonPointer
}

// CheckWithSchema validates the given value against the JSON schema at the
// given path and returns a list of SchemaErrors representing the errors found
// during validation. If there is an error reading or compiling the schema, or
// if the error returned by the jsonschema library is not a ValidationError,
// an error is returned.
func CheckWithSchema(value any, schemaPath string) ([]SchemaError, error) {
	path, err := filepath.Abs(schemaPath)
	if err != nil {
		return nil, fmt.Errorf("getting absolute path of schema: %w", err)
	}

	compiler := jsonschema.NewCompiler()
	schema, err := compiler.Compile(path)
	if err != nil {
		return nil, fmt.Errorf("compiling schema: %w", err)
	}

	err = schema.Validate(value)
	if err == nil {
		return nil, nil
	}

	validationError, ok := err.(*jsonschema.ValidationError)
	if !ok {
		return nil, fmt.Errorf("found an error that is not a validation error: %w", err)
	}

	schemaErrors := []SchemaError{}
	for _, leaf := range validationError.DetailedOutput().Errors {
		schemaError, err := findError(leaf)
		if err != nil {
			return nil, fmt.Errorf("finding error message in schema validation error output: %w", err)
		}

		schemaErrors = append(schemaErrors, schemaError)
	}

	return schemaErrors, nil
}

// Finds the error message that we want to display for a specific OutputUnit
// from the jsonschema library.
func findError(unit jsonschema.OutputUnit) (SchemaError, error) {
	if unit.Error != nil {
		return SchemaError{
			Message: unit.Error.String(),
			Pointer: NewJsonPointer(unit.InstanceLocation),
		}, nil
	}

	if len(unit.Errors) > 0 {
		// FIXME: We should probably notify the user that there are other
		//  things they could do to conform to the schema instead of just
		//  showing the first error of a oneOf/anyOf combination.
		//  See Testcase: "test wrong array oneOf none"
		errorMessage, err := findError(unit.Errors[0])
		if err != nil {
			return SchemaError{}, err
		}

		return errorMessage, nil
	}

	return SchemaError{}, fmt.Errorf("no error found")
}
