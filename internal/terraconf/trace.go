package terraconf

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type errorInfo struct {
	Filename     string
	ErrorMessage string

	LineStart int
	LineEnd   int

	HighlightedLine   int
	HighlightedColumn int

	ShowHerePointer bool
}

// TraceError reads the _terraconf value from errorObject, finds the filename
// and the pointer to the errorObject within that file, appends the innerPointer
// to the pointer and returns a new error with a message containing the
// errorMessage, the filename and if the pointer is found a reference to where
// in the file the error occurred formatted to be readable.
// If the pointer points to an object, the content of the object should be
// included in the error message as well, but only up to 10 lines to prevent
// excessively long error messages.
func TraceError(filepath string, pointer JsonPointer, errorMessage string) (string, error) {
	// Read the affected file
	fileContent, err := os.ReadFile(filepath)
	if err != nil {
		return "", fmt.Errorf("reading file %s: %w", filepath, err)
	}

	var rootNode yaml.Node
	if err := yaml.Unmarshal(fileContent, &rootNode); err != nil {
		return "", fmt.Errorf("decoding YAML: %w", err)
	}

	// Find correct nodes in yaml
	targetNode, nextNode, err := findPointerContext(&rootNode, pointer)
	if err != nil {
		return "", fmt.Errorf("pointer %s not found: %w", pointer.Raw(), err)
	}

	// Build the error
	errorInfo, err := computeErrorInfo(targetNode, nextNode, filepath, errorMessage)
	if err != nil {
		return "", fmt.Errorf("computing error info: %w", err)
	}

	return renderError(errorInfo, string(fileContent)), nil
}

// findPointerContext navigates through a YAML node tree following the given
// JSON Pointer. Returns the target node, the next sibling node (if any),
// and an error if the pointer is invalid or the path cannot be traversed.
func findPointerContext(node *yaml.Node, pointer JsonPointer) (*yaml.Node, *yaml.Node, error) {
	// Handle "pseudo" nodes: Just recurse into them if possible
	if node.Kind == yaml.DocumentNode {
		if len(node.Content) != 1 {
			return nil, nil, fmt.Errorf("document did not have exactly one root node")
		}

		return findPointerContext(node.Content[0], pointer)
	}

	if node.Kind == yaml.AliasNode {
		if node.Alias == nil {
			return nil, nil, fmt.Errorf("alias node with nil target")
		}

		return findPointerContext(node.Alias, pointer)
	}

	// If we've reached the end of the pointer, return the current node
	if pointer.IsRoot() {
		return node, nil, nil
	}

	// Otherwise, we need to navigate further. Truncate the pointer to get the
	// current element and the remaining pointer for the next steps.
	pointerElement, remainingPointer, err := pointer.Truncate()
	if err != nil {
		return nil, nil, fmt.Errorf("cannot decode next path element: %w", err)
	}

	// traverse the node tree
	switch node.Kind {
	case yaml.MappingNode:
		for i := 0; i < len(node.Content)-1; i += 2 {
			keyNode := node.Content[i]
			valueNode := node.Content[i+1]

			if keyNode.Value == pointerElement {
				// We've found the element
				target, next, err := findPointerContext(valueNode, remainingPointer)

				// ...but we might need to populate "next"
				if err != nil {
					return nil, nil, err
				}

				if next == nil && i+2 < len(node.Content) {
					next = node.Content[i+2]
				}

				return target, next, nil
			}
		}

		return nil, nil, fmt.Errorf("key %s not found", pointerElement)
	case yaml.SequenceNode:
		// Decode index
		i, err := strconv.Atoi(pointerElement)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid sequence index: %w", err)
		}
		if i < 0 || i >= len(node.Content) {
			return nil, nil, fmt.Errorf("sequence index out of bounds: %d", i)
		}

		// We've found the element
		target, next, err := findPointerContext(node.Content[i], remainingPointer)

		// ...but we might need to populate "next"
		if err != nil {
			return nil, nil, err
		}

		if next == nil && i+1 < len(node.Content) {
			next = node.Content[i+1]
		}

		return target, next, nil
	case yaml.ScalarNode:
		return nil, nil, fmt.Errorf("found scalar, but was expected to navigate further: %s", pointer.Raw())
	default:
		return nil, nil, fmt.Errorf("unsupported YAML node kind: %d", node.Kind)
	}
}

// Based on the gathered information, we decide what the error message should
// look like. If the error points to a scalar, we can just point to the line
// and column of the scalar. If it points to an object or array, we want to
// include a snippet of the file content around the error location, since
// objects and arrays can be large and it's not enough to just point to the
// line and column.
func computeErrorInfo(node *yaml.Node, nextNode *yaml.Node, filename string, errorMessage string) (errorInfo, error) {
	switch node.Kind {
	case yaml.ScalarNode:
		return errorInfo{
			Filename:          filename,
			ErrorMessage:      errorMessage,
			LineStart:         node.Line - 1,
			LineEnd:           node.Line + 2,
			HighlightedLine:   node.Line,
			HighlightedColumn: node.Column,
			ShowHerePointer:   true,
		}, nil
	case yaml.MappingNode, yaml.SequenceNode:
		startLine := node.Line
		// FIXME: Ideally, we should truncate the yaml document in an
		// intelligent way, so that all the keys of for example, an object,
		// are included in the traceback, since one of them might be the
		// source of the error.
		endLine := startLine + 10 

		if nextNode != nil && nextNode.Line > startLine && nextNode.Line < endLine {
			endLine = nextNode.Line
		}

		return errorInfo{
			Filename:          filename,
			ErrorMessage:      errorMessage,
			LineStart:         startLine,
			LineEnd:           endLine,
			HighlightedLine:   node.Line,
			HighlightedColumn: node.Column,
			ShowHerePointer:   false,
		}, nil
	default:
		return errorInfo{}, fmt.Errorf("unsupported node kind for error reporting: %d", node.Kind)
	}
}

// renderError takes the errorInfo and the file content and produces a
// human-readable error message. It includes the filename, the line and
// column of the error, the error message and a snippet of the file content
// around the error location. The snippet is trimmed to a maximum of 10 lines
// and common indentation is removed for better readability. If ShowHerePointer
// is true, a pointer (^) is included to indicate the exact column of the
// error on the highlighted line.
func renderError(info errorInfo, fileContent string) string {
	fileContent = strings.ReplaceAll(fileContent, "\r\n", "\n")
	fileContent = strings.ReplaceAll(fileContent, "\t", "    ")
	lines := strings.Split(fileContent, "\n")

	startLine := clamp(info.LineStart, 1, len(lines)+1)
	endLine := clamp(info.LineEnd, info.LineStart, len(lines)+1)
	errorLines := lines[startLine-1 : endLine-1]
	highlitedLine := clamp(info.HighlightedLine, startLine, endLine)

	lineNumberWidth := len(fmt.Sprintf("%d", endLine)) + 1

	commonIndent := findCommonIndentation(errorLines)
	highlitedIndent := info.HighlightedColumn - 1 - commonIndent

	var sb strings.Builder

	if info.ShowHerePointer {
		sb.WriteString(fmt.Sprintf("ERROR: In %s:%d:%d\n", info.Filename, info.HighlightedLine, info.HighlightedColumn))
	} else {
		sb.WriteString(fmt.Sprintf("ERROR: %s in %s:%d:%d\n", info.ErrorMessage, info.Filename, info.HighlightedLine, info.HighlightedColumn))
	}

	for i, line := range errorLines {
		lineNum := startLine + i

		if len(line) >= commonIndent {
			sb.WriteString(fmt.Sprintf("%*d │ %s\n", lineNumberWidth, lineNum, line[commonIndent:]))
		} else {
			sb.WriteString(fmt.Sprintf("%*d │\n", lineNumberWidth, lineNum))
		}

		if info.ShowHerePointer && lineNum == highlitedLine {
			sb.WriteString(fmt.Sprintf("%*s │ %*s^ %s\n", lineNumberWidth, " ", highlitedIndent, " ", info.ErrorMessage))
		}
	}

	return sb.String()
}

// clamp limits the provided value to be within min and max (inclusive)
func clamp(value int, min int, max int) int {
	if value < min {
		return min
	}

	if value > max {
		return max
	}

	return value
}

// findCommonIndentation takes a slice of lines and returns the number of
// leading spaces that are common to all non-empty lines. This is used to trim
// error message lines so that they are not excessively indented.
func findCommonIndentation(lines []string) int {
	minIndent := math.MaxInt32

	for _, line := range lines {
		trimmed := strings.TrimLeft(line, " ")

		if trimmed == "" {
			continue
		}

		indent := len(line) - len(trimmed)

		if indent < minIndent {
			minIndent = indent
		}
	}

	if minIndent == math.MaxInt32 {
		return 0
	}

	return minIndent
}
