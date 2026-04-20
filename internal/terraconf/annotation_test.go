package terraconf

import (
	"fmt"
	"path/filepath"
	"testing"
)

func TestLoadAndAnnotateYAML(t *testing.T) {
	relPath := filepath.Join("testdata", "tracing-complex", "tracing.yaml")
	absPath, err := filepath.Abs(relPath)
	if err != nil {
		t.Fatalf("abspath: %v", err)
	}

	t.Run("relative", func(t *testing.T) {
		v, err := LoadAndAnnotateYAML(relPath)
		if err != nil {
			t.Fatalf("LoadAndTagYAML(rel): %v", err)
		}
		testLoadedValue(t, v, relPath)
	})

	t.Run("absolute", func(t *testing.T) {
		v, err := LoadAndAnnotateYAML(absPath)
		if err != nil {
			t.Fatalf("LoadAndTagYAML(abs): %v", err)
		}
		testLoadedValue(t, v, relPath)
	})
}

func TestLoadAddRemoveAnnotations(t *testing.T) {
	path := filepath.Join("testdata", "tracing-complex", "tracing.yaml")

	// Load, annotate, then remove annotations
	value, err := LoadAndAnnotateYAML(path)
	if err != nil {
		t.Fatalf("LoadAndAnnotateYAML: %v", err)
	}

	RemoveAnnotations(value)

	// Load original data again for reference
	expected, err := LoadYAML(path)
	if err != nil {
		t.Fatalf("LoadYAML for expected: %v", err)
	}

	// Compare both
	if err := compareValues(expected, value, NewJsonPointerRoot()); err != nil {
		t.Fatalf("mismatch after removing annotations: %v", err)
	}
}

func testLoadedValue(t *testing.T, v any, expectedFilepath string) {
	expectedPrefix := "v1|" + expectedFilepath + "|"
	// build expected structure
	expected := map[string]any{
		"name":    "example",
		"enabled": true,
		"count":   3,
		"1":       "one",
		"false":   "foo",
		"items": []any{
			map[string]any{
				"id":         "foo",
				"val":        1,
				"_terraconf": expectedPrefix + "/items/0",
			},
			map[string]any{
				"id":         "bar",
				"val":        2,
				"_terraconf": expectedPrefix + "/items/1",
			},
		},
		"primitives": []any{10, "foobar"},
		"object": []any{
			map[string]any{
				"name":   "I am a complex object",
				"region": "europe",
				"attributes": []any{
					map[string]any{
						"name":       "awesome",
						"value":      true,
						"_terraconf": expectedPrefix + "/object/0/attributes/0",
					},
					map[string]any{
						"name":       "thetruth",
						"value":      42,
						"_terraconf": expectedPrefix + "/object/0/attributes/1",
					},
					map[string]any{
						"name":       "nothing",
						"value":      nil,
						"_terraconf": expectedPrefix + "/object/0/attributes/2",
					},
					map[string]any{
						"name":       "empty",
						"value":      "",
						"_terraconf": expectedPrefix + "/object/0/attributes/3",
					},
				},
				"some":       1,
				"other":      2,
				"properties": 3,
				"_terraconf": expectedPrefix + "/object/0",
			},
		},
		"inlineJsonForGoodMeasure": []any{
			map[string]any{
				"foo":        "bar",
				"baz":        false,
				"_terraconf": expectedPrefix + "/inlineJsonForGoodMeasure/0",
			},
		},
		"_terraconf": expectedPrefix + "/",
	}

	if err := compareValues(expected, v, NewJsonPointerRoot()); err != nil {
		t.Fatalf("mismatch: %v", err)
	}
}

func compareValues(expected any, actual any, path JsonPointer) error {
	switch exp := expected.(type) {
	case map[string]any:
		act, ok := actual.(map[string]any)
		if !ok {
			return fmt.Errorf("%s: expected map, got %T", path.Raw(), actual)
		}
		if len(exp) != len(act) {
			return fmt.Errorf("%s: map size mismatch: expected %d keys, got %d", path.Raw(), len(exp), len(act))
		}

		for k, ev := range exp {
			av, ok := act[k]
			if !ok {
				return fmt.Errorf("%s/%s: missing key", path.Raw(), k)
			}
			if err := compareValues(ev, av, path.Append(k)); err != nil {
				return err
			}
		}

		return nil
	case []any:
		act, ok := actual.([]any)
		if !ok {
			return fmt.Errorf("%s: expected array, got %T", path.Raw(), actual)
		}

		if len(exp) != len(act) {
			return fmt.Errorf("%s: array length mismatch: expected %d, got %d", path.Raw(), len(exp), len(act))
		}

		for i := range exp {
			if err := compareValues(exp[i], act[i], path.Append(fmt.Sprintf("%d", i))); err != nil {
				return err
			}
		}

		return nil
	default:
		// compare scalar via string representation to tolerate numeric widths
		if fmt.Sprintf("%v", exp) != fmt.Sprintf("%v", actual) {
			return fmt.Errorf("%s: value mismatch: expected %v (%T), got %v (%T)", path.Raw(), exp, exp, actual, actual)
		}
		return nil
	}
}
