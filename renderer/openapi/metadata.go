package openapi

// This file defines structs for metadata in the OpenAPI spec:
// https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.0.3.md

const OPENAPI_VERSION = "3.0.0"

type MetaData struct {
	OpenAPI      string                       `json:"openapi"`
	Info         *InfoObject                  `json:"info,omitempty"`
	Servers      []*ServerObject              `json:"servers,omitempty"`
	ExternalDocs *ExternalDocumentationObject `json:"externalDocs,omitempty"`
}

// NewMetaData returns an empty metadata struct with the default version.
func NewMetaData() *MetaData {
	return &MetaData{
		OpenAPI: OPENAPI_VERSION,
	}
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
	Contact *ContactObject `yaml"contact,omitempty"`

	// The license information for the exposed API.
	License *LicenseObject `json:"license,omitempty"`
}

type ContactObject struct {
	//The identifying name of the contact person/organization.
	Name string `json:"name"`
	// The URL pointing to the contact information. MUST be in the format of a URL.
	URL string `json:"url,omitempty"`
	//email	string	The email address of the contact person/organization. MUST be in the format of an email address.
	Email string `json:"email,omitempty"`
}

type LicenseObject struct {
	// REQUIRED. The license name used for the API.
	Name string `json:"name"`
	// A URL to the license used for the API. MUST be in the format of a URL.
	URL string `json:"url,omitempty"`
}

type ServerObject struct {
	// REQUIRED. A URL to the target host. This URL supports Server Variables and MAY be relative, to indicate that the host location is relative to the location where the OpenAPI document is being served. Variable substitutions will be made when a variable is named in {brackets}.
	URL string `json:"url"`
	// An optional string describing the host designated by the URL. CommonMark syntax MAY be used for rich text representation.
	Description string `json:"description,omitempty"`

	// variables	Map[string, Server Variable Object]	A map between a variable name and its value. The value is used for substitution in the server's URL template.
	// NOTE: Variables is omitted here!!!
}

type ExternalDocumentationObject struct {
	// REQUIRED. The URL for the target documentation. Value MUST be in the format of a URL.
	URL string `json:"url"`
	//description	string	A short description of the target documentation. CommonMark syntax MAY be used for rich text representation.
	Description string `json:"description,omitempty"`
}
