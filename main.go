package main

import (
	"fmt"
	"github.com/gitmann/b9schema-golang/reflector"
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

	//// Print schema as JSON.
	//if b, err := json.MarshalIndent(schema, "", "  "); err != nil {
	//	fmt.Printf("error marshalling schema: %s\n", err)
	//} else {
	//	fmt.Println(string(b))
	//}

	swagger := openapi.NewOpenAPIRenderer("unknown", nil)
	outLines, _ := swagger.ProcessSchema(schema)
	fmt.Println(strings.Join(outLines, "\n"))
}
