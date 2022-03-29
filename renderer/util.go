package renderer

import (
	"github.com/gitmann/b9schema-golang/common/enum/threeflag"
	"github.com/gitmann/b9schema-golang/common/types"
	"github.com/gitmann/b9schema-golang/common/util"
)

// RenderStrings builds a string representation of a type result using the given pre, path, and post functions.
func RenderSchema(schema *types.Schema, r Renderer) []string {
	// Build output outLines.
	out := []string{}

	// Print type refs.
	if !r.DeReference() {
		if len(schema.TypeRef.Children) > 0 {
			rendered := RenderType(schema.TypeRef, r)
			for _, r := range rendered {
				if r != "" {
					out = append(out, r)
				}
			}
		}
	}

	//	Print types.
	if len(schema.Root.Children) > 0 {
		rendered := RenderType(schema.Root, r)
		for _, r := range rendered {
			if r != "" {
				out = append(out, r)
			}
		}
	}

	//	Return strings.
	return out
}

// RenderType builds strings for a TypeNode and its children.
func RenderType(t *types.TypeNode, r Renderer) []string {
	// Capture initial indent and restore on exit.
	originalIndent := r.Indent()

	out := []string{}

	// Process element with preFunc.
	out = util.AppendStrings(out, r.Pre(t), "")

	// Process children.
	if !r.DeReference() && t.TypeRef != "" {
		// Skip children.
	} else {
		// Always process children in alphabetical order.
		typeRefMap := t.ChildMap()
		typeRefKeys := t.ChildKeys(typeRefMap)

		// Capture indent before children.
		childIndent := r.Indent()

		for _, childName := range typeRefKeys {
			childNode := typeRefMap[childName]
			childNative := r.NativeType(childNode)
			if childNative.Include == threeflag.False {
				continue
			}

			// Reset indent before each child.
			r.SetIndent(childIndent)
			out = util.AppendStrings(out, RenderType(childNode, r), "")
		}
	}

	// Restore original indent.
	r.SetIndent(originalIndent)

	// Process element with postFunc.
	out = util.AppendStrings(out, r.Post(t), "")

	// Restore original indent.
	r.SetIndent(originalIndent)

	return out
}
