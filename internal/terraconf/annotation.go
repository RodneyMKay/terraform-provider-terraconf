package terraconf

import (
	"fmt"
	"strconv"
	"strings"
)

var terraconfV1 string = "v1"

// Represents a decoded annotation.
type Annotation struct {
	Version  string
	Filename string
	Pointer  JsonPointer
}

// LoadAndAnnotateYAML combines both the LoadYAML() and AddAnnotations()
// methods.
func LoadAndAnnotateYAML(path string) (any, error) {
	value, err := LoadYAML(path)
	if err != nil {
		return nil, err
	}

	err = AddAnnotations(value, path)
	if err != nil {
		return nil, err
	}

	return value, nil
}

// AddAnnotations adds the _terraconf annotations to a YAML-parsed structure.
// The filename in the annotation is made relative to the current working
// directory. For each mapping (object), it adds a key "_terraconf" with the
// value "version|filename|json-pointer". The JSON pointer is a path to the
// object in the YAML file, starting with '/' for the root. This follows
// RFC 6901, where keys and array indices are appended as '/key' or '/0', for
// maps and arrays respectively. The version is currently fixed to "v1".
func AddAnnotations(value any, path string) error {
	filename, err := MakeRelative(path)
	if err != nil {
		return fmt.Errorf("making path relative: %w", err)
	}

	annotate(value, filename, NewJsonPointerRoot())
	return nil
}

// RemoveAnnotations removes all _terraconf annotations from the given
// structure.
func RemoveAnnotations(value any) {
	switch node := value.(type) {
	case map[string]any:
		delete(node, "_terraconf")

		for _, v := range node {
			RemoveAnnotations(v)
		}
	case []any:
		for _, v := range node {
			RemoveAnnotations(v)
		}
	default:
		// primitives: nothing to do
	}
}

// FindAnnotation searches for an annotation in the value tree, decodes it
// and returns the result. For this to work, the value tree must contain at
// least one object with an _terraconf annotation. This object may be in a
// list.
func FindAnnotation(value any) (Annotation, error) {
	return recurseIntoList(value)
}

// Recursively searches for an object in the value tree. If a list is 
// encountered, this method recursively searches for an object inside that
// list. If an object is found, the _terraconf annotation is decoded and
// returned. If _terraconf is not found or cannot be read, an error is returned.
// If this value tree contains no objects, an error is returned as well.
func recurseIntoList(value any) (Annotation, error) {
	switch value := value.(type) {
	case map[string]any:
		annotation, ok := value["_terraconf"]
		if !ok {
			return Annotation{}, fmt.Errorf("annotation not present in first encountered object")
		}

		return decodeAnnotation(annotation.(string))
	case []any:
		for _, v := range value {
			annotation, err := recurseIntoList(v)
			if err != nil {
				continue
			}

			_, pointer, err := annotation.Pointer.Truncate()
			if err != nil {
				return Annotation{}, fmt.Errorf("found an annotation in the value tree, which contains a json pointer that is shorter than its position in the value tree: %w", err)
			}

			return Annotation{
				Version:  annotation.Version,
				Filename: annotation.Filename,
				Pointer:  pointer,
			}, nil
		}

		return Annotation{}, fmt.Errorf("annotation not found in any of the list items")
	}

	return Annotation{}, fmt.Errorf("trying to find annotation in primitive value")
}

// decodeAnnotation decodes the _terraconf annotation value into its components.
// If decoding fails, an error is returned.
func decodeAnnotation(tag string) (Annotation, error) {
	parts := strings.SplitN(tag, "|", 3)
	if len(parts) != 3 {
		return Annotation{}, fmt.Errorf("invalid _terraconf format")
	}

	version := parts[0]
	if version != terraconfV1 {
		return Annotation{}, fmt.Errorf("unsupported _terraconf version: %s", version)
	}

	filename := parts[1]
	pointer := parts[2]

	return Annotation{
		Version:  version,
		Filename: filename,
		Pointer:  NewJsonPointer(pointer),
	}, nil
}

// annotate walks the YAML-parsed structure and for every mapping inserts the
// _terraconf key with the value "version|filename|json-pointer". Paths are
// JSON Pointer tokens starting with '/' for root (we represent the document
// root as '/'). Keys and array indices are appended as '/key' or '/0'.
func annotate(node any, filename string, pointer JsonPointer) {
	switch node := node.(type) {
	case map[string]any:
		for key, value := range node {
			annotate(value, filename, pointer.Append(key))
		}

		node["_terraconf"] = terraconfV1 + "|" + filename + "|" + pointer.Raw()
	case []any:
		for i, value := range node {
			annotate(value, filename, pointer.Append(strconv.Itoa(i)))
		}
	default:
		// primitives: nothing to do
	}
}
