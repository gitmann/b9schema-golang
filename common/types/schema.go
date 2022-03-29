package types

const (
	ROOT_NAME    = "Root"
	TYPEREF_NAME = "TypeRef"
)

// Schema is the result of parsing types.
type Schema struct {
	// Root is the node ID of the root of types in the order found.
	Root *TypeNode

	// TypeRef is the node ID that holds a map of named types by name.
	TypeRef *TypeNode
}

// NewSchema initializes a new schema with root nodes.
func NewSchema(nativeDialect string) *Schema {
	schema := &Schema{
		Root:    NewRootNode(ROOT_NAME, nativeDialect),
		TypeRef: NewRootNode(TYPEREF_NAME, nativeDialect),
	}

	return schema
}

// CopyWithoutNative removes all native dialects for the minimal schema.
func (schema *Schema) CopyWithoutNative() *Schema {
	return &Schema{
		Root:    schema.Root.CopyWithoutNative(),
		TypeRef: schema.TypeRef.CopyWithoutNative(),
	}
}
