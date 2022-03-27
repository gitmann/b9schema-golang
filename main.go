package main

import (
	"encoding/json"
	"fmt"
	"github.com/gitmann/b9schema-golang/reflector"
)

// Hello, World for b9schema.
type HelloStruct struct {
	Hello string
	World float64
}

func main() {
	var h *HelloStruct

	// Derive schema.
	r := reflector.NewReflector()
	schema := r.DeriveSchema(h)

	// Print schema as JSON.
	if b, err := json.MarshalIndent(schema, "", "  "); err != nil {
		fmt.Printf("error marshalling schema: %s\n", err)
	} else {
		fmt.Println(string(b))
	}
}
