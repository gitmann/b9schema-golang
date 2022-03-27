package simple

import (
	"fmt"
	"github.com/gitmann/b9schema-golang/common/enum/generictype"
	"github.com/gitmann/b9schema-golang/common/enum/typecategory"
	"github.com/gitmann/b9schema-golang/common/types"
	"github.com/gitmann/b9schema-golang/renderer"
	"strings"
)

// SimpleRenderer provides a simple string renderer.
type SimpleRenderer struct {
	opt *renderer.Options
}

func NewSimpleRenderer(opt *renderer.Options) *SimpleRenderer {
	if opt == nil {
		opt = renderer.NewOptions()
	}

	return &SimpleRenderer{opt: opt}
}

func (r *SimpleRenderer) ProcessSchema(schema *types.Schema, settings ...string) ([]string, error) {
	// Header
	return renderer.RenderSchema(schema, r), nil
	// Footer
}

func (r *SimpleRenderer) DeReference() bool {
	return r.opt.DeReference
}

func (r *SimpleRenderer) Indent() int {
	return r.opt.Indent
}

func (r *SimpleRenderer) SetIndent(value int) {
	r.opt.Indent = value
}

func (r *SimpleRenderer) Prefix() string {
	if r.opt.Prefix == "" {
		return ""
	}
	return strings.Repeat(r.opt.Prefix, r.opt.Indent)
}

func (r *SimpleRenderer) Pre(t *types.TypeNode) []string {
	if t.Type == generictype.Root.String() {
		return []string{}
	}

	path := r.Path(t)
	out := strings.Join(path, ".")

	if t.Error != "" {
		out += " ERROR:" + t.Error
	}

	return []string{out}
}

func (r *SimpleRenderer) Post(t *types.TypeNode) []string {
	return []string{}
}

// Path is a function that builds a path string from a TypeNode.
// Format is: [<Name>:]<Type>[:<TypeRef>]
// - If Name is set, prefix with "Name", otherwise "-"
// - If TypeRef is set, suffix with "TypeRef", otherwise "-"
// - If Error is set, wrap entire string with "!"
func (r *SimpleRenderer) Path(t *types.TypeNode) []string {
	if t.Parent == "" {
		// RootID element. Start a new path.
		return []string{t.Name}
	}

	namePart := t.Name
	if namePart != "" {
		namePart += ":"
	}

	// Type.
	var typePart string
	if t.TypeCategory == typecategory.Invalid.String() {
		typePart = t.Type
	} else {
		typePart = generictype.PathDefaultOfType(t.Type)
	}

	// Add TypeRef suffix if set but not if de-referencing.
	refPart := ""
	if !r.DeReference() {
		refPart = t.NativeDefault().TypeRef
	} else if r.DeReference() && t.Error == types.CyclicalReferenceErr {
		// Keep reference if it's a cyclical error.
		refPart = t.NativeDefault().TypeRef
	}
	if refPart != "" {
		refPart = ":" + refPart
	}

	// Build path.
	path := namePart + typePart + refPart

	// Wrap type in "!" if current element is an error.
	if t.Error != "" {
		path = fmt.Sprintf("!%s!", path)
	}

	// Add quotes if path contains "."
	if strings.Contains(path, ".") {
		path = fmt.Sprintf("%q", path)
	}

	return append(r.Path(t.Node(t.Parent)), path)
}
