package openapi

import (
	"errors"
	"fmt"
	"github.com/ghodss/yaml"
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
	MetaData *MetaData
	Options  *renderer.Options
}

func NewOpenAPIRenderer(metadata *MetaData, opt *renderer.Options) *OpenAPIRenderer {
	if opt == nil {
		opt = renderer.NewOptions()
	}

	opt.Prefix = "  "

	return &OpenAPIRenderer{
		MetaData: metadata,
		Options:  opt,
	}
}

func (r *OpenAPIRenderer) ProcessSchema(schema *types.Schema, settings ...string) ([]string, error) {
	out := []string{}

	if r.MetaData == nil {
		return out, errors.New("missing metadata")
	} else if err := r.MetaData.Validate(); err != nil {
		return out, err
	}

	// Header
	if b, err := yaml.Marshal(r.MetaData); err != nil {
		return out, err
	} else {
		out = append(out, string(b))
	}

	out = util.AppendStrings(out, renderer.RenderSchema(schema, r), "")

	// Footer

	return out, nil
}

func (r *OpenAPIRenderer) DeReference() bool {
	return r.Options.DeReference
}

func (r *OpenAPIRenderer) Indent() int {
	return r.Options.Indent
}

func (r *OpenAPIRenderer) SetIndent(value int) {
	r.Options.Indent = value
}

func (r *OpenAPIRenderer) Prefix() string {
	if r.Options.Prefix == "" {
		return ""
	}
	return strings.Repeat(r.Options.Prefix, r.Options.Indent)
}

func (r *OpenAPIRenderer) NativeType(t *types.TypeNode) *types.NativeType {
	return t.GetNativeType("json")
}

func (r *OpenAPIRenderer) Pre(t *types.TypeNode) []string {
	jsonType := r.NativeType(t)
	if jsonType.Include == threeflag.False {
		// Skip this element.
		return []string{}
	}

	// Special handling for root elements.
	if t.Type == generictype.Root.String() {
		if t.Name == types.ROOT_NAME {
			// Build an API path.
			out := []string{r.Prefix() + `paths:`}
			r.SetIndent(r.Indent() + 1)
			return out
		} else if t.Name == types.TYPEREF_NAME {
			// Store TypeRef under the SCHEMA_PATH key.
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

	out := []string{}

	// Start PathItem block if current element parent is Root.
	if t.Parent.Name == types.ROOT_NAME {
		urlPath := "/unknown/path"
		if t.MetaKey != "" {
			urlPath = t.MetaKey
		}

		// Path must start with "/"
		if !strings.HasPrefix(urlPath, "/") {
			urlPath = "/" + urlPath
		}
		out = append(out, r.Prefix()+urlPath+":")

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
	} else if t.Parent.Type == generictype.Map.String() {
		// Map child only exists when map has no known keys. In order to build a valid OpenAPI
		// schema, make a fake property with the name "unknownKey".
		jsonType.Name = "valueType"
	}

	if jsonType.Name != "" {
		out = append(out, fmt.Sprintf("%s%s:", r.Prefix(), jsonType.Name))
		r.SetIndent(r.Indent() + 1)
	}

	if !r.Options.DeReference && jsonType.TypeRef != "" {
		out = append(out, fmt.Sprintf(`%s$ref: '#/%s/%s'`, r.Prefix(), SCHEMA_PATH, jsonType.TypeRef))
	} else {
		// Build description field.
		descriptionTokens := []string{}
		if r.Options.DeReference && jsonType.TypeRef != "" {
			descriptionTokens = append(descriptionTokens, fmt.Sprintf(`From $ref: #/%s/%s`, SCHEMA_PATH, jsonType.TypeRef))
		}
		if t.Error != "" {
			descriptionTokens = append(descriptionTokens, fmt.Sprintf("ERROR=%s", t.Error))
			if strings.HasPrefix(t.Type, generictype.Invalid.String()) {
				if t.Type != generictype.Invalid.String() {
					// Add specific type error to description.
					descriptionTokens = append(descriptionTokens, fmt.Sprintf("Kind=%s", t.Type))
				}
			}
		}
		if len(descriptionTokens) > 0 {
			out = append(out, fmt.Sprintf("%sdescription: '%s'", r.Prefix(), strings.Join(descriptionTokens, ";")))
		}

		switch t.Type {
		case generictype.Struct.String():
			out = append(out,
				r.Prefix()+"type: object",
				r.Prefix()+"additionalProperties: false",
			)
			if len(t.Children) > 0 {
				out = append(out, r.Prefix()+"properties:")
			}
			r.SetIndent(r.Indent() + 1)
		case generictype.Map.String():
			out = append(out,
				r.Prefix()+"type: object",
			)
			if len(t.Children) > 0 {
				out = append(out,
					r.Prefix()+"additionalProperties: true",
					r.Prefix()+"properties:",
				)
			} else {
				out = append(out, r.Prefix()+"additionalProperties: false")
			}
			r.SetIndent(r.Indent() + 1)
		case generictype.List.String():
			out = append(out,
				r.Prefix()+"type: array",
				r.Prefix()+"items:",
			)
			r.SetIndent(r.Indent() + 1)
		case generictype.Boolean.String():
			out = append(out,
				r.Prefix()+"type: boolean",
			)
		case generictype.Integer.String():
			out = append(out,
				r.Prefix()+"type: integer",
			)
			if nativeType.Type == "int64" || nativeType.Type == "uint64" {
				out = append(out,
					r.Prefix()+"format: int64",
				)
			}
		case generictype.Float.String():
			out = append(out,
				r.Prefix()+"type: number",
			)
			if nativeType.Type == "float64" {
				out = append(out,
					r.Prefix()+"format: double",
				)
			}
		case generictype.String.String():
			out = append(out,
				r.Prefix()+"type: string",
			)
		case generictype.DateTime.String():
			out = append(out,
				r.Prefix()+"type: string",
				r.Prefix()+"format: date-time",
			)
		default:
			if strings.HasPrefix(t.Type, generictype.Invalid.String()) {
				// Use "string" type for invalid elements so that OpenAPI schema is valid.
				out = append(out, r.Prefix()+"type: string")
			} else {
				// What else could this be? Let OpenAPI figure it out.
				out = append(out, r.Prefix()+"type: "+t.Type)
			}
		}
	}

	return out
}

func (r *OpenAPIRenderer) Post(t *types.TypeNode) []string {
	return []string{}
}

// Path is a function that builds a path string from a TypeNode.
func (r *OpenAPIRenderer) Path(t *types.TypeNode) []string {
	return []string{}
}
