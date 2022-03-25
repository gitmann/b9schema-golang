package types

// Schema is the result of parsing types.
type Schema struct {
	// Root is a list of types in the order found.
	Root *TypeElement

	// TypeRefs holds a map of named types by name.
	TypeRefs *TypeElement
}
