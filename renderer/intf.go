package renderer

import (
	"github.com/gitmann/b9schema-golang/common/types"
)

type Renderer interface {
	// ProcessResult starts the render process on a Schema and returns a slice of strings.
	// - settings is one or more strings with options for specific processors.
	ProcessSchema(schema *types.Schema, settings ...string) ([]string, error)

	// DeReference returns true if schema references should be replaced with inline types.
	DeReference() bool

	// Indent returns the current indent value.
	Indent() int

	// SetIndent sets the indent to a given value.
	SetIndent(value int)

	// Prefix returns a prefix string with the current indent.
	Prefix() string

	// NativeType returns the native type for the renderer.
	NativeType(t *types.TypeNode) *types.NativeType

	// Pre and Post return strings before/after a type element's children are processed.
	Pre(t *types.TypeNode) []string
	Post(t *types.TypeNode) []string

	// Path is a function that builds a path string from a TypeNode.
	Path(t *types.TypeNode) []string
}
