package terraconf

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

// LoadYAML reads the YAML file at the given path and unmarshals it into a Go
// data structure. It also normalizes the YAML structure to ensure that all maps
// have string keys, which is more convenient to work with in Go/Terraform.
func LoadYAML(filepath string) (any, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", filepath, err)
	}

	var value any
	if err := yaml.Unmarshal(data, &value); err != nil {
		return nil, fmt.Errorf("unmarshal %s: %w", filepath, err)
	}

	value, err = normalizeYAML(value, NewJsonPointerRoot())
	if err != nil {
		return nil, fmt.Errorf("could not normalize yaml document for loading: %w", err)
	}

	return value, nil
}

// normalizeYAML converts maps with non-string keys (returned by yaml) into
// map[string]any recursively so it's easier to work with.
func normalizeYAML(input any, pointer JsonPointer) (any, error) {
	switch input := input.(type) {
	case map[string]any:
		for key, value := range input {
			normalizedValue, err := normalizeYAML(value, pointer.Append(key))
			if err != nil {
				return nil, err
			}

			input[key] = normalizedValue
		}

		return input, nil
	case map[any]any:
		output := make(map[string]any, len(input))

		for key, value := range input {
			yamlKey, err := normalizeYAMLKey(key)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", pointer.Raw(), err)
			}

			normalizedValue, err := normalizeYAML(value, pointer.Append(yamlKey))
			if err != nil {
				return nil, err
			}

			output[yamlKey] = normalizedValue
		}

		return output, nil
	case []any:
		for i, value := range input {
			normalizedValue, err := normalizeYAML(value, pointer.Append(strconv.Itoa(i)))
			if err != nil {
				return nil, err
			}

			input[i] = normalizedValue
		}

		return input, nil
	default:
		return input, nil
	}
}

// Converts YAML map keys to strings. YAML allows non-string keys, but we
// want to work with map[string]any for simplicity. We support scalar keys
// (string, bool, int, float, null) and convert them to their string
// representation. We don't support complex keys (maps or sequences)
// since they can't be represented as JSON Pointer tokens and throw an error
// in that case.
func normalizeYAMLKey(key any) (string, error) {
	switch key := key.(type) {
	case string:
		return key, nil
	case bool, int, int64, float32, float64, nil:
		return fmt.Sprintf("%v", key), nil
	default:
		return "", fmt.Errorf("unsupported non-scalar YAML map key: %T", key)
	}
}
