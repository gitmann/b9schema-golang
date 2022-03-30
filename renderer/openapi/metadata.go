package openapi

import (
	"errors"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/gitmann/b9schema-golang/common/util"
	"net/mail"
	"net/url"
	"strings"
)

// This file defines structs for metadata in the OpenAPI spec:
// https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.0.3.md

const OPENAPI_VERSION = "3.0.0"

type MetaData struct {
	// REQUIRED. This string MUST be the semantic version number of the OpenAPI Specification version that
	// the OpenAPI document uses. The openapi field SHOULD be used by tooling specifications and clients
	// to interpret the OpenAPI document. This is not related to the API info.version string.
	OpenAPI string `json:"openapi"`

	// REQUIRED. Provides metadata about the API. The metadata MAY be used by tooling as required.
	Info *InfoObject `json:"info,omitempty"`

	// An array of Server Objects, which provide connectivity information to a target server. If the servers
	// property is not provided, or is an empty array, the default value would be a Server Object with a url value of /.
	Servers []*ServerObject `json:"servers,omitempty"`

	// Additional external documentation.
	ExternalDocs *ExternalDocumentationObject `json:"externalDocs,omitempty"`
}

// NewMetaData returns an empty metadata struct with the default version.
func NewMetaData(title, version string) *MetaData {
	if title == "" {
		title = "default title"
	}
	if version == "" {
		version = "default version"
	}

	return &MetaData{
		OpenAPI: OPENAPI_VERSION,
		Info: &InfoObject{
			Title:   title,
			Version: version,
		},
	}
}

// MarshalYAML builds YAML strings in a specific key order.
func (m *MetaData) MarshalYAML(prefix string) ([]byte, error) {
	outLines := []string{}

	// OpenAPI
	if b, err := yaml.Marshal(m.OpenAPI); err != nil {
		return nil, err
	} else {
		out := fmt.Sprintf(`openapi: %s`, strings.TrimSpace(string(b)))
		outLines = append(outLines, out)
	}

	// Info
	if b, err := m.Info.MarshalYAML(prefix); err != nil {
		return nil, err
	} else {
		outLines = append(outLines, `info:`)
		outLines = util.AppendStrings(outLines, []string{string(b)}, prefix)
	}

	// ExternalDocs
	if m.ExternalDocs != nil {
		if b, err := yaml.Marshal(m.ExternalDocs); err != nil {
			return nil, err
		} else {
			outLines = append(outLines, `externalDocs:`)
			outLines = util.AppendStrings(outLines, []string{string(b)}, prefix)
		}
	}

	// Servers
	if m.Servers != nil {
		if b, err := yaml.Marshal(m.Servers); err != nil {
			return nil, err
		} else {
			outLines = append(outLines, `servers:`)
			outLines = util.AppendStrings(outLines, []string{string(b)}, prefix)
		}
	}

	outLines = append(outLines, "")
	finalOut := strings.Join(outLines, "\n")

	return []byte(finalOut), nil
}

// Validate checks that metadata contains required fields.
func (m *MetaData) Validate() error {
	if !strings.HasPrefix(m.OpenAPI, "3.0") {
		return fmt.Errorf("invalid 'openapi' value %q", m.OpenAPI)
	}

	if m.Info == nil {
		return errors.New("missing 'info' object")
	} else if err := m.Info.Validate(); err != nil {
		return err
	}

	if m.ExternalDocs != nil {
		if err := m.ExternalDocs.Validate(); err != nil {
			return err
		}
	}

	for _, srv := range m.Servers {
		if err := srv.Validate(); err != nil {
			return err
		}
	}

	return nil
}

type InfoObject struct {
	// REQUIRED. The title of the API.
	Title string `json:"title" yaml:"title"`
	// REQUIRED. The version of the OpenAPI document (which is distinct from the OpenAPI Specification version or the API implementation version).
	Version string `json:"version"`

	// A short description of the API. CommonMark syntax MAY be used for rich text representation.
	Description string `json:"description,omitempty"`
	// A URL to the Terms of Service for the API. MUST be in the format of a URL.
	TermsOfService string `json:"termsOfService,omitempty"`

	// The contact information for the exposed API.
	Contact *ContactObject `json:"contact,omitempty"`

	// The license information for the exposed API.
	License *LicenseObject `json:"license,omitempty"`
}

func (i *InfoObject) Validate() error {
	if i.Title == "" {
		return errors.New("'info.title' is required")
	}
	if i.Version == "" {
		return errors.New("'info.version' is required")
	}

	if i.TermsOfService != "" {
		if _, err := url.ParseRequestURI(i.TermsOfService); err != nil {
			return errors.New("'info.termsOfService' is not a valid URL")
		}
	}

	if i.Contact != nil {
		if err := i.Contact.Validate(); err != nil {
			return err
		}
	}

	if i.License != nil {
		if err := i.License.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func (i *InfoObject) MarshalYAML(prefix string) ([]byte, error) {
	outLines := []string{}

	// Title
	if b, err := yaml.Marshal(i.Title); err != nil {
		return nil, err
	} else {
		outLines = append(outLines, fmt.Sprintf(`title: %s`, strings.TrimSpace(string(b))))
	}

	// Version
	if b, err := yaml.Marshal(i.Version); err != nil {
		return nil, err
	} else {
		outLines = append(outLines, fmt.Sprintf(`version: %s`, strings.TrimSpace(string(b))))
	}

	// Description
	if i.Description != "" {
		if b, err := yaml.Marshal(i.Description); err != nil {
			return nil, err
		} else {
			outLines = append(outLines, fmt.Sprintf(`description: %s`, strings.TrimSpace(string(b))))
		}
	}

	// TermsOfService
	if i.TermsOfService != "" {
		if b, err := yaml.Marshal(i.TermsOfService); err != nil {
			return nil, err
		} else {
			outLines = append(outLines, fmt.Sprintf(`termsOfService: %s`, strings.TrimSpace(string(b))))
		}
	}

	// Contact
	if i.Contact != nil {
		if b, err := yaml.Marshal(i.Contact); err != nil {
			return nil, err
		} else {
			outLines = append(outLines, `contact:`)
			outLines = util.AppendStrings(outLines, []string{string(b)}, prefix)
		}
	}

	// License
	if i.License != nil {
		if b, err := yaml.Marshal(i.License); err != nil {
			return nil, err
		} else {
			outLines = append(outLines, `license:`)
			outLines = util.AppendStrings(outLines, []string{string(b)}, prefix)
		}
	}

	finalOut := strings.Join(outLines, "\n")
	return []byte(finalOut), nil
}

type ContactObject struct {
	//The identifying name of the contact person/organization.
	Name string `json:"name"`
	// The URL pointing to the contact information. MUST be in the format of a URL.
	URL string `json:"url,omitempty"`
	//email	string	The email address of the contact person/organization. MUST be in the format of an email address.
	Email string `json:"email,omitempty"`
}

func (c *ContactObject) Validate() error {
	if c.URL != "" {
		if _, err := url.ParseRequestURI(c.URL); err != nil {
			return errors.New("'contact.url' is not a valid URL")
		}
	}

	if c.Email != "" {
		if _, err := mail.ParseAddress(c.Email); err != nil {
			return errors.New("`contact.email is not a valid email address")
		}
	}

	return nil
}

type LicenseObject struct {
	// REQUIRED. The license name used for the API.
	Name string `json:"name"`
	// A URL to the license used for the API. MUST be in the format of a URL.
	URL string `json:"url,omitempty"`
}

func (lic *LicenseObject) Validate() error {
	if lic.Name == "" {
		return errors.New("'license.name' is required")
	}

	if lic.URL != "" {
		if _, err := url.ParseRequestURI(lic.URL); err != nil {
			return errors.New("'license.url' is not a valid URL")
		}
	}

	return nil
}

type ServerObject struct {
	// REQUIRED. A URL to the target host. This URL supports Server Variables and MAY be relative, to indicate that the host location is relative to the location where the OpenAPI document is being served. Variable substitutions will be made when a variable is named in {brackets}.
	URL string `json:"url"`

	// An optional string describing the host designated by the URL. CommonMark syntax MAY be used for rich text representation.
	Description string `json:"description,omitempty"`

	// variables	Map[string, Server Variable Object]	A map between a variable name and its value. The value is used for substitution in the server's URL template.
	// NOTE: Variables is omitted here!!!
}

func (s *ServerObject) Validate() error {
	if _, err := url.ParseRequestURI(s.URL); err != nil {
		return errors.New("'server.url' is not a valid URL")
	}

	return nil
}

type ExternalDocumentationObject struct {
	// REQUIRED. The URL for the target documentation. Value MUST be in the format of a URL.
	URL string `json:"url"`
	//description	string	A short description of the target documentation. CommonMark syntax MAY be used for rich text representation.
	Description string `json:"description,omitempty"`
}

func (d *ExternalDocumentationObject) Validate() error {
	if _, err := url.ParseRequestURI(d.URL); err != nil {
		return errors.New("'externalDocs.url' is not a valid URL")
	}

	return nil
}

type PathsObject map[string]*PathItemObject

type PathItemObject struct {
	//summary	string	An optional, string summary, intended to apply to all operations in this path.
	Summary string `json:"summary,omitempty"`
	//description	string	An optional, string description, intended to apply to all operations in this path. CommonMark syntax MAY be used for rich text representation.
	Description string `json:"description,omitempty"`
	//get	Operation Object	A definition of a GET operation on this path.
	Get *OperationObject `json:"get,omitempty"`
	//put	Operation Object	A definition of a PUT operation on this path.
	Put *OperationObject `json:"put,omitempty"`
	//post	Operation Object	A definition of a POST operation on this path.
	Post *OperationObject `json:"post,omitempty"`
	//delete	Operation Object	A definition of a DELETE operation on this path.
	Delete *OperationObject `json:"delete,omitempty"`
	//options	Operation Object	A definition of a OPTIONS operation on this path.
	Options *OperationObject `json:"options,omitempty"`
	//head	Operation Object	A definition of a HEAD operation on this path.
	Head *OperationObject `json:"head,omitempty"`
	//patch	Operation Object	A definition of a PATCH operation on this path.
	Patch *OperationObject `json:"patch,omitempty"`
	//trace	Operation Object	A definition of a TRACE operation on this path.
	Trace *OperationObject `json:"trace,omitempty"`
	//servers	[Server Object]	An alternative server array to service all operations in this path.
	Servers []*ServerObject `json:"servers,omitempty"`
	//parameters	[Parameter Object | Reference Object]	A list of parameters that are applicable for all the operations described under this path. These parameters can be overridden at the operation level, but cannot be removed there. The list MUST NOT include duplicated parameters. A unique parameter is defined by a combination of a name and location. The list can use the Reference Object to link to parameters that are defined at the OpenAPI Object's components/parameters.
	Parameters []*ParameterObject `json:"parameters,omitempty"`

	// OMITTED FIELDS
	//$ref	string	Allows for an external definition of this path item. The referenced structure MUST be in the format of a Path Item Object. In case a Path Item Object field appears both in the defined object and the referenced object, the behavior is undefined.
}

type OperationObject struct {
	//tags	[string]	A list of tags for API documentation control. Tags can be used for logical grouping of operations by resources or any other qualifier.
	Tags []string `json:"tags,omitempty"`
	//summary	string	A short summary of what the operation does.
	Summary string `json:"summary,omitempty"`
	//description	string	A verbose explanation of the operation behavior. CommonMark syntax MAY be used for rich text representation.
	Description string `json:"description,omitempty"`
	//externalDocs	External Documentation Object	Additional external documentation for this operation.
	ExternalDocs *ExternalDocumentationObject `json:"externalDocs,omitempty"`
	//operationId	string	Unique string used to identify the operation. The id MUST be unique among all operations described in the API. The operationId value is case-sensitive. Tools and libraries MAY use the operationId to uniquely identify an operation, therefore, it is RECOMMENDED to follow common programming naming conventions.
	OperationId string `json:"operationId,omitempty"`
	//parameters	[Parameter Object | Reference Object]	A list of parameters that are applicable for this operation. If a parameter is already defined at the Path Item, the new definition will override it but can never remove it. The list MUST NOT include duplicated parameters. A unique parameter is defined by a combination of a name and location. The list can use the Reference Object to link to parameters that are defined at the OpenAPI Object's components/parameters.
	Parameters []*ParameterObject `json:"parameters,omitempty"`
	//responses	Responses Object	REQUIRED. The list of possible responses as they are returned from executing this operation.
	Responses map[string]*ResponseObject `json:"responses,omitempty"`
	//deprecated	boolean	Declares this operation to be deprecated. Consumers SHOULD refrain from usage of the declared operation. Default value is false.
	Deprecated bool `json:"deprecated,omitempty"`
	//servers	[Server Object]	An alternative server array to service this operation. If an alternative server object is specified at the Path Item Object or Root level, it will be overridden by this value.
	Servers []*ServerObject `json:"servers,omitempty"`

	// OMITTED FIELDS
	//security	[Security Requirement Object]	A declaration of which security mechanisms can be used for this operation. The list of values includes alternative security requirement objects that can be used. Only one of the security requirement objects need to be satisfied to authorize a request. To make security optional, an empty security requirement ({}) can be included in the array. This definition overrides any declared top-level security. To remove a top-level security declaration, an empty array can be used.
	//requestBody	Request Body Object | Reference Object	The request body applicable for this operation. The requestBody is only supported in HTTP methods where the HTTP 1.1 specification RFC7231 has explicitly defined semantics for request bodies. In other cases where the HTTP spec is vague, requestBody SHALL be ignored by consumers.
	//callbacks	Map[string, Callback Object | Reference Object]	A map of possible out-of band callbacks related to the parent operation. The key is a unique identifier for the Callback Object. Each value in the map is a Callback Object that describes a request that may be initiated by the API provider and the expected responses.
}

type ParameterObject struct {
	//name	string	REQUIRED. The name of the parameter. Parameter names are case sensitive.
	//If in is "path", the name field MUST correspond to a template expression occurring within the path field in the Paths Object. See Path Templating for further information.
	//If in is "header" and the name field is "Accept", "Content-Type" or "Authorization", the parameter definition SHALL be ignored.
	//For all other cases, the name corresponds to the parameter name used by the in property.
	Name string `json:"name"`
	//in	string	REQUIRED. The location of the parameter. Possible values are "query", "header", "path" or "cookie".
	In string `json:"in"`
	//description	string	A brief description of the parameter. This could contain examples of use. CommonMark syntax MAY be used for rich text representation.
	Description string `json:"description,omitempty"`
	//required	boolean	Determines whether this parameter is mandatory. If the parameter location is "path", this property is REQUIRED and its value MUST be true. Otherwise, the property MAY be included and its default value is false.
	Required bool `json:"required,omitempty"`
	//deprecated	boolean	Specifies that a parameter is deprecated and SHOULD be transitioned out of usage. Default value is false.
	Deprecated bool `json:"deprecated,omitempty"`
	//allowEmptyValue	boolean	Sets the ability to pass empty-valued parameters. This is valid only for query parameters and allows sending a parameter with an empty value. Default value is false. If style is used, and if behavior is n/a (cannot be serialized), the value of allowEmptyValue SHALL be ignored. Use of this property is NOT RECOMMENDED, as it is likely to be removed in a later revision.}
	AllowEmptyValue bool `json:"allowEmptyValue,omitempty"`
	//schema Schema Object The schema defining the type used for the parameter.
	//NOTE: This is just a placeholder using a map. Actual SchemaObject is much more complex.
	Schema *SimpleSchemaObject `json:"schema,omitempty"`
}

// SimpleSchemaObject is a lightweight representation of the SchemaObject.
type SimpleSchemaObject struct {
	Type      string `json:"type,omitempty"`
	Reference string `json:"$ref,omitempty"`
}

type ResponseObject struct {
	//description	string	REQUIRED. A short description of the response. CommonMark syntax MAY be used for rich text representation.
	Description string `json:"description"`
	//headers	Map[string, Header Object | Reference Object]	Maps a header name to its definition. RFC7230 states header names are case insensitive. If a response header is defined with the name "Content-Type", it SHALL be ignored.
	Headers map[string]*HeaderObject `json:"headers,omitempty"`
	//content	Map[string, Media Type Object]	A map containing descriptions of potential response payloads. The key is a media type or media type range and the value describes it. For responses that match multiple keys, only the most specific key is applicable. e.g. text/plain overrides text/*
	Content map[string]*MediaTypeObject `json:"content,omitempty"`

	// OMITTED FIELDS
	//links	Map[string, Link Object | Reference Object]	A map of operations links that can be followed from the response. The key of the map is a short name for the link, following the naming constraints of the names for Component Objects.
}

type HeaderObject struct {
	//description	string	A brief description of the parameter. This could contain examples of use. CommonMark syntax MAY be used for rich text representation.
	Description string `json:"description,omitempty"`
	//required	boolean	Determines whether this parameter is mandatory. If the parameter location is "path", this property is REQUIRED and its value MUST be true. Otherwise, the property MAY be included and its default value is false.
	Required bool `json:"required,omitempty"`
	//deprecated	boolean	Specifies that a parameter is deprecated and SHOULD be transitioned out of usage. Default value is false.
	Deprecated bool `json:"deprecated,omitempty"`
	//allowEmptyValue	boolean	Sets the ability to pass empty-valued parameters. This is valid only for query parameters and allows sending a parameter with an empty value. Default value is false. If style is used, and if behavior is n/a (cannot be serialized), the value of allowEmptyValue SHALL be ignored. Use of this property is NOT RECOMMENDED, as it is likely to be removed in a later revision.}
	AllowEmptyValue bool `json:"allowEmptyValue,omitempty"`
}

type MediaTypeObject struct {
	//schema	Schema Object | Reference Object	The schema defining the content of the request, response, or parameter.
	//example	Any	Example of the media type. The example object SHOULD be in the correct format as specified by the media type. The example field is mutually exclusive of the examples field. Furthermore, if referencing a schema which contains an example, the example value SHALL override the example provided by the schema.
	//examples	Map[ string, Example Object | Reference Object]	Examples of the media type. Each example object SHOULD match the media type and specified schema if present. The examples field is mutually exclusive of the example field. Furthermore, if referencing a schema which contains an example, the examples value SHALL override the example provided by the schema.
	//encoding	Map[string, Encoding Object]	A map between a property name and its encoding information. The key, being the property name, MUST exist in the schema as a property. The encoding object SHALL only apply to requestBody objects when the media type is multipart or application/x-www-form-urlencoded.
}
