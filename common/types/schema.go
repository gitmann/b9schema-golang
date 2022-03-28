package types

// Schema is the result of parsing types.
type Schema struct {
	// RootID is the node ID of the root of types in the order found.
	RootID string

	// TypeRefID is the node ID that holds a map of named types by name.
	TypeRefID string

	// NodePool is the pool of all TypeNodes.
	NodePool *NodePool
}

// NewSchema initializes a new schema with root nodes.
func NewSchema(nativeDialect string) *Schema {
	pool := NewNodePool()

	schema := &Schema{
		RootID:    pool.NewRootNode("Root", nativeDialect).ID,
		TypeRefID: pool.NewRootNode("TypeRef", nativeDialect).ID,

		NodePool: pool,
	}

	return schema
}

func (schema *Schema) RootNode() *TypeNode {
	return schema.NodePool.Nodes[schema.RootID]
}

func (schema *Schema) TypeRefNode() *TypeNode {
	return schema.NodePool.Nodes[schema.TypeRefID]
}
