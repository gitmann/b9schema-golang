package openapi

import (
	"fmt"
	"github.com/gitmann/b9schema-golang/common/enum/generictype"
	"github.com/gitmann/b9schema-golang/common/enum/threeflag"
	"github.com/gitmann/b9schema-golang/common/types"
	"github.com/gitmann/b9schema-golang/common/util"
	"github.com/gitmann/b9schema-golang/renderer"
	"strings"
)

// Default location for schema references without leading or training "/".
const SCHEMA_PATH = "components/schemas"

// OpenAPIRenderer provides a simple string renderer.
type OpenAPIRenderer struct {
	// Path
	URLPath string

	opt *renderer.Options
}

func NewOpenAPIRenderer(urlPath string, opt *renderer.Options) *OpenAPIRenderer {
	if opt == nil {
		opt = renderer.NewOptions()
	}

	opt.Prefix = "  "

	return &OpenAPIRenderer{
		URLPath: urlPath,
		opt:     opt,
	}
}

func (r *OpenAPIRenderer) ProcessSchema(schema *types.Schema, settings ...string) ([]string, error) {
	out := []string{}

	// Header
	out = append(out, `openapi: 3.0.0`)

	out = util.AppendStrings(out, renderer.RenderSchema(schema, r))

	// Footer

	return out, nil
}

func (r *OpenAPIRenderer) DeReference() bool {
	return r.opt.DeReference
}

func (r *OpenAPIRenderer) Indent() int {
	return r.opt.Indent
}

func (r *OpenAPIRenderer) SetIndent(value int) {
	r.opt.Indent = value
}

func (r *OpenAPIRenderer) Prefix() string {
	if r.opt.Prefix == "" {
		return ""
	}
	return strings.Repeat(r.opt.Prefix, r.opt.Indent)
}

func (r *OpenAPIRenderer) Pre(t *types.TypeNode) []string {
	jsonType := t.GetNativeType("json")
	if jsonType.Include == threeflag.False {
		// Skip this element.
		return []string{}
	}

	// Special handling for root elements.
	if t.Type == generictype.Root.String() {
		if t.Name == "RootID" {
			// Build an API path.
			out := []string{r.Prefix() + `paths:`}

			r.SetIndent(r.Indent() + 1)
			out = append(out, r.Prefix()+r.URLPath+":")

			r.SetIndent(r.Indent() + 1)
			out = append(out, r.Prefix()+`get:`)

			r.SetIndent(r.Indent() + 1)
			out = append(out, r.Prefix()+`summary: Return data.`)
			out = append(out, r.Prefix()+`responses:`)

			r.SetIndent(r.Indent() + 1)
			out = append(out, r.Prefix()+`'200':`)

			r.SetIndent(r.Indent() + 1)
			out = append(out, r.Prefix()+`description: Success`)
			out = append(out, r.Prefix()+`content:`)

			r.SetIndent(r.Indent() + 1)
			out = append(out, r.Prefix()+`application/json:`)

			r.SetIndent(r.Indent() + 1)
			out = append(out, r.Prefix()+`schema:`)

			r.SetIndent(r.Indent() + 1)
			return out
		} else if t.Name == "TypeRefID" {
			// Store TypeRefID under the SCHEMA_PATH key.
			tokens := strings.Split(SCHEMA_PATH, "/")

			out := []string{}
			for _, t := range tokens {
				out = append(out, r.Prefix()+t+":")
				r.SetIndent(r.Indent() + 1)
			}
			return out
		}
	}

	nativeType := t.NativeDefault()

	outLines := []string{}

	if jsonType.Name != "" {
		outLines = append(outLines, fmt.Sprintf("%s%s:", r.Prefix(), jsonType.Name))
		r.SetIndent(r.Indent() + 1)
	}

	if jsonType.TypeRef != "" {
		outLines = append(outLines, fmt.Sprintf(`%s$ref: '#/%s/%s'`, r.Prefix(), SCHEMA_PATH, jsonType.TypeRef))
	} else {
		switch t.Type {
		case generictype.Struct.String():
			outLines = append(outLines,
				r.Prefix()+"type: object",
				r.Prefix()+"additionalProperties: false",
				r.Prefix()+"properties:",
			)
			r.SetIndent(r.Indent() + 1)
		case generictype.Map.String():
			outLines = append(outLines,
				r.Prefix()+"type: object",
				r.Prefix()+"additionalProperties: true",
				r.Prefix()+"properties:",
			)
			r.SetIndent(r.Indent() + 1)
		case generictype.List.String():
			outLines = append(outLines,
				r.Prefix()+"type: array",
				r.Prefix()+"items:",
			)
			r.SetIndent(r.Indent() + 1)
		case generictype.Boolean.String():
			outLines = append(outLines,
				r.Prefix()+"type: boolean",
			)
		case generictype.Integer.String():
			outLines = append(outLines,
				r.Prefix()+"type: integer",
			)
			if nativeType.Type == "int64" || nativeType.Type == "uint64" {
				outLines = append(outLines,
					r.Prefix()+"format: int64",
				)
			}
		case generictype.Float.String():
			outLines = append(outLines,
				r.Prefix()+"type: number",
			)
			if nativeType.Type == "float64" {
				outLines = append(outLines,
					r.Prefix()+"format: double",
				)
			}
		case generictype.String.String():
			outLines = append(outLines,
				r.Prefix()+"type: string",
			)
		case generictype.DateTime.String():
			outLines = append(outLines,
				r.Prefix()+"type: string",
				r.Prefix()+"format: date-time",
			)
		default:
			outLines = append(outLines,
				r.Prefix()+"type: "+t.Type,
			)
		}
	}

	if t.Error != "" {
		outLines = append(outLines,
			r.Prefix()+"error: "+t.Error,
		)
	}

	return outLines
}

func (r *OpenAPIRenderer) Post(t *types.TypeNode) []string {
	return []string{}
}

// Path is a function that builds a path string from a TypeNode.
func (r *OpenAPIRenderer) Path(t *types.TypeNode) []string {
	return []string{}
}
