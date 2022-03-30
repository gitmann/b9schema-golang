package main

import (
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/gitmann/b9schema-golang/reflector"
	"github.com/gitmann/b9schema-golang/renderer"
	"github.com/gitmann/b9schema-golang/renderer/openapi"
	"strings"
)

// Hello, World for b9schema.
type HelloStruct struct {
	Hello string
}

type GoodbyeStruct struct {
	Bye float64
}

type MorningStruct struct {
	Morning HelloStruct
}

func main() {
	// Derive schema.
	r := reflector.NewReflector()

	r.DeriveSchema(HelloStruct{}, "/path/to/hello")
	r.DeriveSchema(GoodbyeStruct{}, "/path/to/goodbye")
	r.DeriveSchema(MorningStruct{}, "/path/to/morning")

	schema := r.Schema

	// Print schema as YAML.
	fmt.Println("********** YAML")
	if b, err := yaml.Marshal(schema); err != nil {
		fmt.Printf("error marshalling schema: %s\n", err)
	} else {
		fmt.Println(string(b))
	}

	fmt.Println("********** YAML (min)")
	minSchema := schema.CopyWithoutNative()
	if b, err := yaml.Marshal(minSchema); err != nil {
		fmt.Printf("error marshalling schema: %s\n", err)
	} else {
		fmt.Println(string(b))
	}

	fmt.Println("********** OpenAPI")
	opt := renderer.NewOptions()
	opt.DeReference = true

	meta := openapi.NewMetaData("main.go demo", "v1.0.0")
	meta.Info.Description = "Demonstration of OpenAPI rendering."
	meta.Info.TermsOfService = "https://some.site/tos"
	meta.Info.Contact = &openapi.ContactObject{
		Name:  "Contact Name",
		URL:   "https://some.site/contact",
		Email: "contact@some.site",
	}
	meta.Info.License = &openapi.LicenseObject{
		Name: "This is the license.",
		URL:  "https://license.server/license",
	}
	meta.ExternalDocs = &openapi.ExternalDocumentationObject{
		URL:         "https://some.site/docs",
		Description: "This is the documentation site.",
	}
	meta.Servers = []*openapi.ServerObject{
		{
			URL:         "https://a.b.com/",
			Description: "Main server",
		},
		{
			URL:         "https://a.b.c.com/",
			Description: "Backup server",
		},
	}

	swagger := openapi.NewOpenAPIRenderer(meta, opt)
	outLines, err := swagger.ProcessSchema(schema)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(strings.Join(outLines, "\n"))
}
