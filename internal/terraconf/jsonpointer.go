package terraconf

import (
	"fmt"
	"strings"
)

// Represents a JSON Pointer as defined in RFC 6901. The raw string is stored,
// and methods are provided to manipulate it. The pointer is expected to start
// with a forward slash (/). The root pointer can be represented as an empty
// string or a single slash. Trailing slashes are not allowed.
type JsonPointer struct {
	raw string
}

// NewJsonPointerRoot returns a JsonPointer representing the root path ("/").
func NewJsonPointerRoot() JsonPointer {
	return JsonPointer{raw: "/"}
}

// NewJsonPointer creates a new JsonPointer from the specified raw pointer
// path. It ensures that a leading slash is always present and that a trailing
// slash is not present. This method should be used on strings that come
// from the JsonPointer.Raw() method.
func NewJsonPointer(raw string) JsonPointer {
	trimmedPointer := strings.TrimPrefix(strings.TrimSuffix(raw, "/"), "/")
	return JsonPointer{raw: "/" + trimmedPointer}
}

// Raw returns the raw json pointer string encoded by this object. It produces
// a string for use with the NewJsonPointer method.
func (p JsonPointer) Raw() string {
	return p.raw
}

// IsRoot checks if the pointer represents the root path.
func (p JsonPointer) IsRoot() bool {
	return p.raw == "" || p.raw == "/"
}

// Append appends a new path element to the pointer.
func (p JsonPointer) Append(innerPointer string) JsonPointer {
	if p.IsRoot() {
		return JsonPointer{
			raw: "/" + escapeJSONPointer(innerPointer),
		}
	} else {
		return JsonPointer{
			raw: p.raw + "/" + escapeJSONPointer(innerPointer),
		}
	}
}

// Join appends the current Json pointer to the provided one, joining them
// together into one Json pointer.
func (p JsonPointer) Join(other JsonPointer) JsonPointer {
	return JsonPointer{raw: p.raw + other.raw}
}

// Truncate removes the first path element from the pointer and returns it along
// with the remaining pointer. It is the opposite of Append. If the pointer is
// root, it returns an error since there no elements to truncate.
func (p JsonPointer) Truncate() (string, JsonPointer, error) {
	if p.IsRoot() {
		return "", p, fmt.Errorf("root path cannot be truncated")
	}

	pointerNoLeadingSlash := p.raw[1:]
	elementEnd := strings.Index(pointerNoLeadingSlash, "/")
	if elementEnd == -1 {
		return unescapeJSONPointer(pointerNoLeadingSlash), JsonPointer{raw: "/"}, nil
	}

	return unescapeJSONPointer(pointerNoLeadingSlash[:elementEnd]), JsonPointer{raw: pointerNoLeadingSlash[elementEnd:]}, nil
}

// escapeJSONPointer escapes a reference token per RFC 6901.
func escapeJSONPointer(s string) string {
	s = strings.ReplaceAll(s, "~", "~0")
	s = strings.ReplaceAll(s, "/", "~1")
	return s
}

// removes the escaping from a JSON Pointer token
func unescapeJSONPointer(s string) string {
	s = strings.ReplaceAll(s, "~1", "/")
	s = strings.ReplaceAll(s, "~0", "~")
	return s
}
