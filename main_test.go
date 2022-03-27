package main

import (
	"encoding/json"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/gitmann/b9schema-golang/common/util"
	"github.com/gitmann/b9schema-golang/fixtures"
	"github.com/gitmann/b9schema-golang/reflector"
	"github.com/gitmann/b9schema-golang/renderer"
	json2 "github.com/gitmann/b9schema-golang/renderer/json"
	"github.com/gitmann/b9schema-golang/renderer/openapi"
	"github.com/gitmann/b9schema-golang/renderer/simple"
	"strings"
	"testing"
	"time"
	"unsafe"
)

var allTests = [][]fixtures.TestCase{
	rootJSONTests,
	rootGoTests,
	typeTests,
	listTests,
	compoundTests,
	referenceTests,
	cycleTests,
	jsonTagTests,

	// structTests,
	// pointerTests,
}

// *** All reflect types ***

// rootTests validate that the top-level element is either a struct or Reference.
var rootJSONTests = []fixtures.TestCase{
	{
		Name:         "json-null",
		Value:        fromJSON([]byte(`null`)),
		RefStrings:   []string{"RootID.!invalid:nil! ERROR:kind not supported"},
		DerefStrings: []string{"RootID.!invalid:nil! ERROR:kind not supported"},
	},
	{
		Name:         "json-string",
		Value:        fromJSON([]byte(`"Hello"`)),
		RefStrings:   []string{"RootID.!string! ERROR:root type must be a struct"},
		DerefStrings: []string{"RootID.!string! ERROR:root type must be a struct"},
	},
	{
		Name:         "json-int",
		Value:        fromJSON([]byte(`123`)),
		RefStrings:   []string{"RootID.!float! ERROR:root type must be a struct"},
		DerefStrings: []string{"RootID.!float! ERROR:root type must be a struct"},
	},
	{
		Name:         "json-float",
		Value:        fromJSON([]byte(`234.345`)),
		RefStrings:   []string{"RootID.!float! ERROR:root type must be a struct"},
		DerefStrings: []string{"RootID.!float! ERROR:root type must be a struct"},
	},
	{
		Name:         "json-bool",
		Value:        fromJSON([]byte(`true`)),
		RefStrings:   []string{"RootID.!boolean! ERROR:root type must be a struct"},
		DerefStrings: []string{"RootID.!boolean! ERROR:root type must be a struct"},
	},
	{
		Name:         "json-list-empty",
		Value:        fromJSON([]byte(`[]`)),
		RefStrings:   []string{"RootID.![]! ERROR:root type must be a struct"},
		DerefStrings: []string{"RootID.![]! ERROR:root type must be a struct"},
	},
	{
		Name:         "json-list",
		Value:        fromJSON([]byte(`[1,2,3]`)),
		RefStrings:   []string{"RootID.![]! ERROR:root type must be a struct"},
		DerefStrings: []string{"RootID.![]! ERROR:root type must be a struct"},
	},
	{
		Name:         "json-object-empty",
		Value:        fromJSON([]byte(`{}`)),
		RefStrings:   []string{"RootID.!map{}! ERROR:root type must be a struct"},
		DerefStrings: []string{"RootID.!map{}! ERROR:root type must be a struct"},
	},
	{
		Name:  "json-object",
		Value: fromJSON([]byte(`{"key1":"Hello"}`)),
		RefStrings: []string{
			"RootID.{}",
			"RootID.{}.Key1:string",
		},
		DerefStrings: []string{
			"RootID.{}",
			"RootID.{}.Key1:string",
		},
	},
}

var rootGoTests = []fixtures.TestCase{
	{
		Name:         "golang-nil",
		Value:        nil,
		RefStrings:   []string{"RootID.!invalid:nil! ERROR:kind not supported"},
		DerefStrings: []string{"RootID.!invalid:nil! ERROR:kind not supported"},
	},
	{
		Name:         "golang-string",
		Value:        "Hello",
		RefStrings:   []string{"RootID.!string! ERROR:root type must be a struct"},
		DerefStrings: []string{"RootID.!string! ERROR:root type must be a struct"},
	},
	{
		Name:         "golang-int",
		Value:        123,
		RefStrings:   []string{"RootID.!integer! ERROR:root type must be a struct"},
		DerefStrings: []string{"RootID.!integer! ERROR:root type must be a struct"},
	},
	{
		Name:         "golang-float",
		Value:        234.345,
		RefStrings:   []string{"RootID.!float! ERROR:root type must be a struct"},
		DerefStrings: []string{"RootID.!float! ERROR:root type must be a struct"},
	},
	{
		Name:         "golang-bool",
		Value:        true,
		RefStrings:   []string{"RootID.!boolean! ERROR:root type must be a struct"},
		DerefStrings: []string{"RootID.!boolean! ERROR:root type must be a struct"},
	},
	{
		Name:         "golang-array-0",
		Value:        [0]string{},
		RefStrings:   []string{"RootID.![]! ERROR:root type must be a struct"},
		DerefStrings: []string{"RootID.![]! ERROR:root type must be a struct"},
	},
	{
		Name:         "golang-array-3",
		Value:        [3]string{},
		RefStrings:   []string{"RootID.![]! ERROR:root type must be a struct"},
		DerefStrings: []string{"RootID.![]! ERROR:root type must be a struct"},
	},
	{
		Name:         "golang-slice-nil",
		Value:        func() interface{} { var s []string; return s }(),
		RefStrings:   []string{"RootID.![]! ERROR:root type must be a struct"},
		DerefStrings: []string{"RootID.![]! ERROR:root type must be a struct"},
	},
	{
		Name:         "golang-slice-0",
		Value:        []string{},
		RefStrings:   []string{"RootID.![]! ERROR:root type must be a struct"},
		DerefStrings: []string{"RootID.![]! ERROR:root type must be a struct"},
	},
	{
		Name:         "golang-slice-3",
		Value:        make([]string, 3),
		RefStrings:   []string{"RootID.![]! ERROR:root type must be a struct"},
		DerefStrings: []string{"RootID.![]! ERROR:root type must be a struct"},
	},
	{
		Name: "golang-struct-empty", Value: func() interface{} { var s struct{}; return s }(),
		RefStrings:   []string{"RootID.!{}! ERROR:empty struct not supported"},
		DerefStrings: []string{"RootID.!{}! ERROR:empty struct not supported"},
	},
	{
		Name:  "golang-struct-noinit",
		Value: func() interface{} { var s StringStruct; return s }(),
		RefStrings: []string{
			`TypeRefID.StringStruct:{}`,
			`TypeRefID.StringStruct:{}.Value:string`,
			`RootID.{}:StringStruct`,
		},
		DerefStrings: []string{
			`RootID.{}`,
			`RootID.{}.Value:string`,
		},
	},
	{
		Name:  "golang-struct-init",
		Value: StringStruct{},
		RefStrings: []string{
			`TypeRefID.StringStruct:{}`,
			`TypeRefID.StringStruct:{}.Value:string`,
			`RootID.{}:StringStruct`,
		},
		DerefStrings: []string{
			`RootID.{}`,
			`RootID.{}.Value:string`,
		},
	},
	{
		Name:  "golang-struct-private",
		Value: PrivateStruct{},
		RefStrings: []string{
			`TypeRefID.!PrivateStruct:{}! ERROR:struct has no exported fields`,
			`RootID.!{}:PrivateStruct! ERROR:struct has no exported fields`,
		},
		DerefStrings: []string{
			`RootID.!{}! ERROR:struct has no exported fields`,
		},
	},

	{
		Name:  "golang-interface-struct-noinit",
		Value: func() interface{} { var s interface{} = StringStruct{}; return s }(),
		RefStrings: []string{
			`TypeRefID.StringStruct:{}`,
			`TypeRefID.StringStruct:{}.Value:string`,
			`RootID.{}:StringStruct`,
		},
		DerefStrings: []string{
			`RootID.{}`,
			`RootID.{}.Value:string`,
		},
	},
	{
		Name:  "golang-pointer-struct-noinit",
		Value: func() interface{} { var s *StringStruct; return s }(),
		RefStrings: []string{
			`TypeRefID.StringStruct:{}`,
			`TypeRefID.StringStruct:{}.Value:string`,
			`RootID.{}:StringStruct`,
		},
		DerefStrings: []string{
			`RootID.{}`,
			`RootID.{}.Value:string`,
		},
	},
	{
		Name:  "golang-pointer-struct-init",
		Value: &StringStruct{},
		RefStrings: []string{
			`TypeRefID.StringStruct:{}`,
			`TypeRefID.StringStruct:{}.Value:string`,
			`RootID.{}:StringStruct`,
		},
		DerefStrings: []string{
			`RootID.{}`,
			`RootID.{}.Value:string`,
		},
	},
}

// Check all types from reflect package.
type BoolTypes struct {
	Bool bool
}

type IntegerTypes struct {
	Int     int
	Int8    int8
	Int16   int16
	Int32   int32
	Int64   int64
	Uint    uint
	Uint8   uint8
	Uint16  uint16
	Uint32  uint32
	Uint64  uint64
	Uintptr uintptr
}

type FloatTypes struct {
	Float32 float32
	Float64 float64
}

type StringTypes struct {
	String string
}

type InvalidTypes struct {
	Complex64  complex64
	Complex128 complex128

	Chan          chan int
	Func          func()
	UnsafePointer unsafe.Pointer
}

type CompoundTypes struct {
	Array0 [0]string
	Array3 [3]string

	Interface  interface{}
	Map        map[int]int
	Ptr        *StringStruct
	PrivatePtr *PrivateStruct
	Slice      []interface{}
	Struct     struct{}
}

// Special types from protobuf: https://developers.google.com/protocol-buffers/docs/reference/google.protobuf
type SpecialTypes struct {
	DateTime time.Time
}

var typeTests = []fixtures.TestCase{
	{
		Name:  "boolean",
		Value: BoolTypes{},
		RefStrings: []string{
			`TypeRefID.BoolTypes:{}`,
			`TypeRefID.BoolTypes:{}.Bool:boolean`,
			`RootID.{}:BoolTypes`,
		},
		DerefStrings: []string{
			`RootID.{}`,
			`RootID.{}.Bool:boolean`,
		},
		OpenAPIStrings: []string{
			`openapi: 3.0.0`,
			`components:`,
			`  schemas:`,
			`    BoolTypes:`,
			`      type: object`,
			`      additionalProperties: false`,
			`      properties:`,
			`        Bool:`,
			`          type: boolean`,
			`paths:`,
			`  /test/path:`,
			`    get:`,
			`      summary: Return data.`,
			`      responses:`,
			`        '200':`,
			`          description: Success`,
			`          content:`,
			`            application/json:`,
			`              schema:`,
			`                $ref: '#/components/schemas/BoolTypes'`,
		},
	},
	{
		Name:  "integer",
		Value: IntegerTypes{},
		RefStrings: []string{
			`TypeRefID.IntegerTypes:{}`,
			`TypeRefID.IntegerTypes:{}.Int:integer`,
			`TypeRefID.IntegerTypes:{}.Int16:integer`,
			`TypeRefID.IntegerTypes:{}.Int32:integer`,
			`TypeRefID.IntegerTypes:{}.Int64:integer`,
			`TypeRefID.IntegerTypes:{}.Int8:integer`,
			`TypeRefID.IntegerTypes:{}.Uint:integer`,
			`TypeRefID.IntegerTypes:{}.Uint16:integer`,
			`TypeRefID.IntegerTypes:{}.Uint32:integer`,
			`TypeRefID.IntegerTypes:{}.Uint64:integer`,
			`TypeRefID.IntegerTypes:{}.Uint8:integer`,
			`TypeRefID.IntegerTypes:{}.Uintptr:integer`,
			`RootID.{}:IntegerTypes`,
		},
		DerefStrings: []string{
			`RootID.{}`,
			`RootID.{}.Int:integer`,
			`RootID.{}.Int16:integer`,
			`RootID.{}.Int32:integer`,
			`RootID.{}.Int64:integer`,
			`RootID.{}.Int8:integer`,
			`RootID.{}.Uint:integer`,
			`RootID.{}.Uint16:integer`,
			`RootID.{}.Uint32:integer`,
			`RootID.{}.Uint64:integer`,
			`RootID.{}.Uint8:integer`,
			`RootID.{}.Uintptr:integer`,
		},
		OpenAPIStrings: []string{
			`openapi: 3.0.0`,
			`components:`,
			`  schemas:`,
			`    IntegerTypes:`,
			`      type: object`,
			`      additionalProperties: false`,
			`      properties:`,
			`        Int:`,
			`          type: integer`,
			`        Int16:`,
			`          type: integer`,
			`        Int32:`,
			`          type: integer`,
			`        Int64:`,
			`          type: integer`,
			`          format: int64`,
			`        Int8:`,
			`          type: integer`,
			`        Uint:`,
			`          type: integer`,
			`        Uint16:`,
			`          type: integer`,
			`        Uint32:`,
			`          type: integer`,
			`        Uint64:`,
			`          type: integer`,
			`          format: int64`,
			`        Uint8:`,
			`          type: integer`,
			`        Uintptr:`,
			`          type: integer`,
			`paths:`,
			`  /test/path:`,
			`    get:`,
			`      summary: Return data.`,
			`      responses:`,
			`        '200':`,
			`          description: Success`,
			`          content:`,
			`            application/json:`,
			`              schema:`,
			`                $ref: '#/components/schemas/IntegerTypes'`,
		},
	},
	{
		Name:  `float`,
		Value: FloatTypes{},
		RefStrings: []string{
			`TypeRefID.FloatTypes:{}`,
			`TypeRefID.FloatTypes:{}.Float32:float`,
			`TypeRefID.FloatTypes:{}.Float64:float`,
			`RootID.{}:FloatTypes`,
		},
		DerefStrings: []string{
			`RootID.{}`,
			`RootID.{}.Float32:float`,
			`RootID.{}.Float64:float`,
		},
		OpenAPIStrings: []string{
			`openapi: 3.0.0`,
			`components:`,
			`  schemas:`,
			`    FloatTypes:`,
			`      type: object`,
			`      additionalProperties: false`,
			`      properties:`,
			`        Float32:`,
			`          type: number`,
			`        Float64:`,
			`          type: number`,
			`          format: double`,
			`paths:`,
			`  /test/path:`,
			`    get:`,
			`      summary: Return data.`,
			`      responses:`,
			`        '200':`,
			`          description: Success`,
			`          content:`,
			`            application/json:`,
			`              schema:`,
			`                $ref: '#/components/schemas/FloatTypes'`,
		},
	},
	{
		Name:  "string",
		Value: StringTypes{},
		RefStrings: []string{
			`TypeRefID.StringTypes:{}`,
			`TypeRefID.StringTypes:{}.String:string`,
			`RootID.{}:StringTypes`,
		},
		DerefStrings: []string{
			`RootID.{}`,
			`RootID.{}.String:string`,
		},
		OpenAPIStrings: []string{
			`openapi: 3.0.0`,
			`components:`,
			`  schemas:`,
			`    StringTypes:`,
			`      type: object`,
			`      additionalProperties: false`,
			`      properties:`,
			`        String:`,
			`          type: string`,
			`paths:`,
			`  /test/path:`,
			`    get:`,
			`      summary: Return data.`,
			`      responses:`,
			`        '200':`,
			`          description: Success`,
			`          content:`,
			`            application/json:`,
			`              schema:`,
			`                $ref: '#/components/schemas/StringTypes'`,
		},
	},
	{
		Name:  "invalid",
		Value: InvalidTypes{},
		RefStrings: []string{
			`TypeRefID.InvalidTypes:{}`,
			`TypeRefID.InvalidTypes:{}.!Chan:invalid:chan! ERROR:kind not supported`,
			`TypeRefID.InvalidTypes:{}.!Complex128:invalid:complex128! ERROR:kind not supported`,
			`TypeRefID.InvalidTypes:{}.!Complex64:invalid:complex64! ERROR:kind not supported`,
			`TypeRefID.InvalidTypes:{}.!Func:invalid:func! ERROR:kind not supported`,
			`TypeRefID.InvalidTypes:{}."!UnsafePointer:invalid:unsafe.Pointer!" ERROR:kind not supported`,
			`RootID.{}:InvalidTypes`,
		},
		DerefStrings: []string{
			`RootID.{}`,
			`RootID.{}.!Chan:invalid:chan! ERROR:kind not supported`,
			`RootID.{}.!Complex128:invalid:complex128! ERROR:kind not supported`,
			`RootID.{}.!Complex64:invalid:complex64! ERROR:kind not supported`,
			`RootID.{}.!Func:invalid:func! ERROR:kind not supported`,
			`RootID.{}."!UnsafePointer:invalid:unsafe.Pointer!" ERROR:kind not supported`,
		},
		OpenAPIStrings: []string{
			`openapi: 3.0.0`,
			`components:`,
			`  schemas:`,
			`    InvalidTypes:`,
			`      type: object`,
			`      additionalProperties: false`,
			`      properties:`,
			`        Chan:`,
			`          type: invalid:chan`,
			`          error: kind not supported`,
			`        Complex128:`,
			`          type: invalid:complex128`,
			`          error: kind not supported`,
			`        Complex64:`,
			`          type: invalid:complex64`,
			`          error: kind not supported`,
			`        Func:`,
			`          type: invalid:func`,
			`          error: kind not supported`,
			`        UnsafePointer:`,
			`          type: invalid:unsafe.Pointer`,
			`          error: kind not supported`,
			`paths:`,
			`  /test/path:`,
			`    get:`,
			`      summary: Return data.`,
			`      responses:`,
			`        '200':`,
			`          description: Success`,
			`          content:`,
			`            application/json:`,
			`              schema:`,
			`                $ref: '#/components/schemas/InvalidTypes'`,
		},
	},
	{
		Name:  "compound",
		Value: CompoundTypes{},
		RefStrings: []string{
			`TypeRefID.CompoundTypes:{}`,
			`TypeRefID.CompoundTypes:{}.Array0:[]`,
			`TypeRefID.CompoundTypes:{}.Array0:[].string`,
			`TypeRefID.CompoundTypes:{}.Array3:[]`,
			`TypeRefID.CompoundTypes:{}.Array3:[].string`,
			`TypeRefID.CompoundTypes:{}.!Interface:invalid! ERROR:interface element is nil`,
			`TypeRefID.CompoundTypes:{}.!Map:map{}! ERROR:map key type must be string`,
			`TypeRefID.CompoundTypes:{}.PrivatePtr:{}:PrivateStruct`,
			`TypeRefID.CompoundTypes:{}.Ptr:{}:StringStruct`,
			`TypeRefID.CompoundTypes:{}.Slice:[]`,
			`TypeRefID.CompoundTypes:{}.Slice:[].!invalid! ERROR:interface element is nil`,
			`TypeRefID.CompoundTypes:{}.!Struct:{}! ERROR:empty struct not supported`,
			`TypeRefID.!PrivateStruct:{}! ERROR:struct has no exported fields`,
			`TypeRefID.StringStruct:{}`,
			`TypeRefID.StringStruct:{}.Value:string`,
			`RootID.{}:CompoundTypes`,
		},
		DerefStrings: []string{
			`RootID.{}`,
			`RootID.{}.Array0:[]`,
			`RootID.{}.Array0:[].string`,
			`RootID.{}.Array3:[]`,
			`RootID.{}.Array3:[].string`,
			`RootID.{}.!Interface:invalid! ERROR:interface element is nil`,
			`RootID.{}.!Map:map{}! ERROR:map key type must be string`,
			`RootID.{}.!PrivatePtr:{}! ERROR:struct has no exported fields`,
			`RootID.{}.Ptr:{}`,
			`RootID.{}.Ptr:{}.Value:string`,
			`RootID.{}.Slice:[]`,
			`RootID.{}.Slice:[].!invalid! ERROR:interface element is nil`,
			`RootID.{}.!Struct:{}! ERROR:empty struct not supported`,
		},
		OpenAPIStrings: []string{
			`openapi: 3.0.0`,
			`components:`,
			`  schemas:`,
			`    CompoundTypes:`,
			`      type: object`,
			`      additionalProperties: false`,
			`      properties:`,
			`        Array0:`,
			`          type: array`,
			`          items:`,
			`            type: string`,
			`        Array3:`,
			`          type: array`,
			`          items:`,
			`            type: string`,
			`        Interface:`,
			`          type: invalid`,
			`          error: interface element is nil`,
			`        Map:`,
			`          type: object`,
			`          additionalProperties: true`,
			`          properties:`,
			`            error: map key type must be string`,
			`        PrivatePtr:`,
			`          $ref: '#/components/schemas/PrivateStruct'`,
			`        Ptr:`,
			`          $ref: '#/components/schemas/StringStruct'`,
			`        Slice:`,
			`          type: array`,
			`          items:`,
			`            type: invalid`,
			`            error: interface element is nil`,
			`        Struct:`,
			`          type: object`,
			`          additionalProperties: false`,
			`          properties:`,
			`            error: empty struct not supported`,
			`    PrivateStruct:`,
			`      type: object`,
			`      additionalProperties: false`,
			`      properties:`,
			`        error: struct has no exported fields`,
			`    StringStruct:`,
			`      type: object`,
			`      additionalProperties: false`,
			`      properties:`,
			`        Value:`,
			`          type: string`,
			`paths:`,
			`  /test/path:`,
			`    get:`,
			`      summary: Return data.`,
			`      responses:`,
			`        '200':`,
			`          description: Success`,
			`          content:`,
			`            application/json:`,
			`              schema:`,
			`                $ref: '#/components/schemas/CompoundTypes'`,
		},
	},
	{
		Name:  "special",
		Value: SpecialTypes{},
		RefStrings: []string{
			`TypeRefID.SpecialTypes:{}`,
			`TypeRefID.SpecialTypes:{}.DateTime:datetime`,
			`RootID.{}:SpecialTypes`,
		},
		DerefStrings: []string{
			`RootID.{}`,
			`RootID.{}.DateTime:datetime`,
		},
		OpenAPIStrings: []string{
			`openapi: 3.0.0`,
			`components:`,
			`  schemas:`,
			`    SpecialTypes:`,
			`      type: object`,
			`      additionalProperties: false`,
			`      properties:`,
			`        DateTime:`,
			`          type: string`,
			`          format: date-time`,
			`paths:`,
			`  /test/path:`,
			`    get:`,
			`      summary: Return data.`,
			`      responses:`,
			`        '200':`,
			`          description: Success`,
			`          content:`,
			`            application/json:`,
			`              schema:`,
			`                $ref: '#/components/schemas/SpecialTypes'`,
		},
	},
}

type ArrayStruct struct {
	Array0   [0]string
	Array3   [3]string
	Array2_3 [2][3]string
}

type SliceStruct struct {
	Slice  []string
	Array2 [][]string
}

var jsonArrayTest = `
{
	"Array0": [],
	"Array3": ["a","b","c"],
	"Array2_3": [
		[1,2,3],
		[2,3,4]
	]
}
`

// Array tests.
var listTests = []fixtures.TestCase{
	{
		Name:  "arrays",
		Value: &ArrayStruct{},
		RefStrings: []string{
			`TypeRefID.ArrayStruct:{}`,
			`TypeRefID.ArrayStruct:{}.Array0:[]`,
			`TypeRefID.ArrayStruct:{}.Array0:[].string`,
			`TypeRefID.ArrayStruct:{}.Array2_3:[]`,
			`TypeRefID.ArrayStruct:{}.Array2_3:[].[]`,
			`TypeRefID.ArrayStruct:{}.Array2_3:[].[].string`,
			`TypeRefID.ArrayStruct:{}.Array3:[]`,
			`TypeRefID.ArrayStruct:{}.Array3:[].string`,
			`RootID.{}:ArrayStruct`,
		},
		DerefStrings: []string{
			`RootID.{}`,
			`RootID.{}.Array0:[]`,
			`RootID.{}.Array0:[].string`,
			`RootID.{}.Array2_3:[]`,
			`RootID.{}.Array2_3:[].[]`,
			`RootID.{}.Array2_3:[].[].string`,
			`RootID.{}.Array3:[]`,
			`RootID.{}.Array3:[].string`,
		},
		OpenAPIStrings: []string{
			`openapi: 3.0.0`,
			`components:`,
			`  schemas:`,
			`    ArrayStruct:`,
			`      type: object`,
			`      additionalProperties: false`,
			`      properties:`,
			`        Array0:`,
			`          type: array`,
			`          items:`,
			`            type: string`,
			`        Array2_3:`,
			`          type: array`,
			`          items:`,
			`            type: array`,
			`            items:`,
			`              type: string`,
			`        Array3:`,
			`          type: array`,
			`          items:`,
			`            type: string`,
			`paths:`,
			`  /test/path:`,
			`    get:`,
			`      summary: Return data.`,
			`      responses:`,
			`        '200':`,
			`          description: Success`,
			`          content:`,
			`            application/json:`,
			`              schema:`,
			`                $ref: '#/components/schemas/ArrayStruct'`,
		},
	},
	{
		Name:  "json-array",
		Value: fromJSON([]byte(jsonArrayTest)),
		RefStrings: []string{
			`RootID.{}`,
			`RootID.{}.Array0:[]`,
			`RootID.{}.Array0:[].!invalid! ERROR:interface element is nil`,
			`RootID.{}.Array2_3:[]`,
			`RootID.{}.Array2_3:[].[]`,
			`RootID.{}.Array2_3:[].[].float`,
			`RootID.{}.Array3:[]`,
			`RootID.{}.Array3:[].string`,
		},
		DerefStrings: []string{
			`RootID.{}`,
			`RootID.{}.Array0:[]`,
			`RootID.{}.Array0:[].!invalid! ERROR:interface element is nil`,
			`RootID.{}.Array2_3:[]`,
			`RootID.{}.Array2_3:[].[]`,
			`RootID.{}.Array2_3:[].[].float`,
			`RootID.{}.Array3:[]`,
			`RootID.{}.Array3:[].string`,
		},
		OpenAPIStrings: []string{
			`openapi: 3.0.0`,
			`paths:`,
			`  /test/path:`,
			`    get:`,
			`      summary: Return data.`,
			`      responses:`,
			`        '200':`,
			`          description: Success`,
			`          content:`,
			`            application/json:`,
			`              schema:`,
			`                type: object`,
			`                additionalProperties: false`,
			`                properties:`,
			`                  Array0:`,
			`                    type: array`,
			`                    items:`,
			`                      type: invalid`,
			`                      error: interface element is nil`,
			`                  Array2_3:`,
			`                    type: array`,
			`                    items:`,
			`                      type: array`,
			`                      items:`,
			`                        type: number`,
			`                        format: double`,
			`                  Array3:`,
			`                    type: array`,
			`                    items:`,
			`                      type: string`,
		},
	},
	{
		Name:  "slices",
		Value: &SliceStruct{},
		RefStrings: []string{
			`TypeRefID.SliceStruct:{}`,
			`TypeRefID.SliceStruct:{}.Array2:[]`,
			`TypeRefID.SliceStruct:{}.Array2:[].[]`,
			`TypeRefID.SliceStruct:{}.Array2:[].[].string`,
			`TypeRefID.SliceStruct:{}.Slice:[]`,
			`TypeRefID.SliceStruct:{}.Slice:[].string`,
			`RootID.{}:SliceStruct`,
		},
		DerefStrings: []string{
			`RootID.{}`,
			`RootID.{}.Array2:[]`,
			`RootID.{}.Array2:[].[]`,
			`RootID.{}.Array2:[].[].string`,
			`RootID.{}.Slice:[]`,
			`RootID.{}.Slice:[].string`,
		},
		OpenAPIStrings: []string{
			`openapi: 3.0.0`,
			`components:`,
			`  schemas:`,
			`    SliceStruct:`,
			`      type: object`,
			`      additionalProperties: false`,
			`      properties:`,
			`        Array2:`,
			`          type: array`,
			`          items:`,
			`            type: array`,
			`            items:`,
			`              type: string`,
			`        Slice:`,
			`          type: array`,
			`          items:`,
			`            type: string`,
			`paths:`,
			`  /test/path:`,
			`    get:`,
			`      summary: Return data.`,
			`      responses:`,
			`        '200':`,
			`          description: Success`,
			`          content:`,
			`            application/json:`,
			`              schema:`,
			`                $ref: '#/components/schemas/SliceStruct'`,
		},
	},
}

type MapTestsStruct struct {
	MapOK struct {
		StringVal string
		IntVal    float64
		FloatVal  float32
		BoolVal   bool
		ListVal   []float64
		MapVal    struct {
			Key1 string
			Key2 struct {
				DeepKey1 string
				DeepKey2 float64
			}
		}
	}
}

var jsonMapTests = `
{
	"MapOK": {
		"StringVal": "Hello",
		"IntVal": 123,
		"FloatVal": 234.345,
		"BoolVal": true,
		"ListVal": [2,3,4,5],
		"MapVal": {
			"Key1": "Hey",
			"Key2": {
				"DeepKey1": "Hi",
				"DeepKey2": 234
			}
		}
	}
}
`

var compoundTests = []fixtures.TestCase{
	{
		Name:  "golang-map",
		Value: MapTestsStruct{},
		RefStrings: []string{
			`TypeRefID.MapTestsStruct:{}`,
			`TypeRefID.MapTestsStruct:{}.MapOK:{}`,
			`TypeRefID.MapTestsStruct:{}.MapOK:{}.BoolVal:boolean`,
			`TypeRefID.MapTestsStruct:{}.MapOK:{}.FloatVal:float`,
			`TypeRefID.MapTestsStruct:{}.MapOK:{}.IntVal:float`,
			`TypeRefID.MapTestsStruct:{}.MapOK:{}.ListVal:[]`,
			`TypeRefID.MapTestsStruct:{}.MapOK:{}.ListVal:[].float`,
			`TypeRefID.MapTestsStruct:{}.MapOK:{}.MapVal:{}`,
			`TypeRefID.MapTestsStruct:{}.MapOK:{}.MapVal:{}.Key1:string`,
			`TypeRefID.MapTestsStruct:{}.MapOK:{}.MapVal:{}.Key2:{}`,
			`TypeRefID.MapTestsStruct:{}.MapOK:{}.MapVal:{}.Key2:{}.DeepKey1:string`,
			`TypeRefID.MapTestsStruct:{}.MapOK:{}.MapVal:{}.Key2:{}.DeepKey2:float`,
			`TypeRefID.MapTestsStruct:{}.MapOK:{}.StringVal:string`,
			`RootID.{}:MapTestsStruct`,
		},
		DerefStrings: []string{
			`RootID.{}`,
			`RootID.{}.MapOK:{}`,
			`RootID.{}.MapOK:{}.BoolVal:boolean`,
			`RootID.{}.MapOK:{}.FloatVal:float`,
			`RootID.{}.MapOK:{}.IntVal:float`,
			`RootID.{}.MapOK:{}.ListVal:[]`,
			`RootID.{}.MapOK:{}.ListVal:[].float`,
			`RootID.{}.MapOK:{}.MapVal:{}`,
			`RootID.{}.MapOK:{}.MapVal:{}.Key1:string`,
			`RootID.{}.MapOK:{}.MapVal:{}.Key2:{}`,
			`RootID.{}.MapOK:{}.MapVal:{}.Key2:{}.DeepKey1:string`,
			`RootID.{}.MapOK:{}.MapVal:{}.Key2:{}.DeepKey2:float`,
			`RootID.{}.MapOK:{}.StringVal:string`,
		},
		OpenAPIStrings: []string{
			`openapi: 3.0.0`,
			`components:`,
			`  schemas:`,
			`    MapTestsStruct:`,
			`      type: object`,
			`      additionalProperties: false`,
			`      properties:`,
			`        MapOK:`,
			`          type: object`,
			`          additionalProperties: false`,
			`          properties:`,
			`            BoolVal:`,
			`              type: boolean`,
			`            FloatVal:`,
			`              type: number`,
			`            IntVal:`,
			`              type: number`,
			`              format: double`,
			`            ListVal:`,
			`              type: array`,
			`              items:`,
			`                type: number`,
			`                format: double`,
			`            MapVal:`,
			`              type: object`,
			`              additionalProperties: false`,
			`              properties:`,
			`                Key1:`,
			`                  type: string`,
			`                Key2:`,
			`                  type: object`,
			`                  additionalProperties: false`,
			`                  properties:`,
			`                    DeepKey1:`,
			`                      type: string`,
			`                    DeepKey2:`,
			`                      type: number`,
			`                      format: double`,
			`            StringVal:`,
			`              type: string`,
			`paths:`,
			`  /test/path:`,
			`    get:`,
			`      summary: Return data.`,
			`      responses:`,
			`        '200':`,
			`          description: Success`,
			`          content:`,
			`            application/json:`,
			`              schema:`,
			`                $ref: '#/components/schemas/MapTestsStruct'`,
		},
	},
	{
		Name:  "json-map",
		Value: fromJSON([]byte(jsonMapTests)),
		RefStrings: []string{
			`RootID.{}`,
			`RootID.{}.MapOK:{}`,
			`RootID.{}.MapOK:{}.BoolVal:boolean`,
			`RootID.{}.MapOK:{}.FloatVal:float`,
			`RootID.{}.MapOK:{}.IntVal:float`,
			`RootID.{}.MapOK:{}.ListVal:[]`,
			`RootID.{}.MapOK:{}.ListVal:[].float`,
			`RootID.{}.MapOK:{}.MapVal:{}`,
			`RootID.{}.MapOK:{}.MapVal:{}.Key1:string`,
			`RootID.{}.MapOK:{}.MapVal:{}.Key2:{}`,
			`RootID.{}.MapOK:{}.MapVal:{}.Key2:{}.DeepKey1:string`,
			`RootID.{}.MapOK:{}.MapVal:{}.Key2:{}.DeepKey2:float`,
			`RootID.{}.MapOK:{}.StringVal:string`,
		},
		DerefStrings: []string{
			`RootID.{}`,
			`RootID.{}.MapOK:{}`,
			`RootID.{}.MapOK:{}.BoolVal:boolean`,
			`RootID.{}.MapOK:{}.FloatVal:float`,
			`RootID.{}.MapOK:{}.IntVal:float`,
			`RootID.{}.MapOK:{}.ListVal:[]`,
			`RootID.{}.MapOK:{}.ListVal:[].float`,
			`RootID.{}.MapOK:{}.MapVal:{}`,
			`RootID.{}.MapOK:{}.MapVal:{}.Key1:string`,
			`RootID.{}.MapOK:{}.MapVal:{}.Key2:{}`,
			`RootID.{}.MapOK:{}.MapVal:{}.Key2:{}.DeepKey1:string`,
			`RootID.{}.MapOK:{}.MapVal:{}.Key2:{}.DeepKey2:float`,
			`RootID.{}.MapOK:{}.StringVal:string`,
		},
		OpenAPIStrings: []string{
			`openapi: 3.0.0`,
			`paths:`,
			`  /test/path:`,
			`    get:`,
			`      summary: Return data.`,
			`      responses:`,
			`        '200':`,
			`          description: Success`,
			`          content:`,
			`            application/json:`,
			`              schema:`,
			`                type: object`,
			`                additionalProperties: false`,
			`                properties:`,
			`                  MapOK:`,
			`                    type: object`,
			`                    additionalProperties: false`,
			`                    properties:`,
			`                      BoolVal:`,
			`                        type: boolean`,
			`                      FloatVal:`,
			`                        type: number`,
			`                        format: double`,
			`                      IntVal:`,
			`                        type: number`,
			`                        format: double`,
			`                      ListVal:`,
			`                        type: array`,
			`                        items:`,
			`                          type: number`,
			`                          format: double`,
			`                      MapVal:`,
			`                        type: object`,
			`                        additionalProperties: false`,
			`                        properties:`,
			`                          Key1:`,
			`                            type: string`,
			`                          Key2:`,
			`                            type: object`,
			`                            additionalProperties: false`,
			`                            properties:`,
			`                              DeepKey1:`,
			`                                type: string`,
			`                              DeepKey2:`,
			`                                type: number`,
			`                                format: double`,
			`                      StringVal:`,
			`                        type: string`,
		},
	},
}

type ReferenceTestsStruct struct {
	InterfaceVal interface{}
	PtrVal       *BasicStruct
	PtrPtrVal    **BasicStruct
}

var referenceTests = []fixtures.TestCase{
	{
		Name:  "reference-tests-empty",
		Value: ReferenceTestsStruct{},
		RefStrings: []string{
			`TypeRefID.BasicStruct:{}`,
			`TypeRefID.BasicStruct:{}.BoolVal:boolean`,
			`TypeRefID.BasicStruct:{}.Float64Val:float`,
			`TypeRefID.BasicStruct:{}.IntVal:integer`,
			`TypeRefID.BasicStruct:{}.StringVal:string`,
			`TypeRefID.ReferenceTestsStruct:{}`,
			`TypeRefID.ReferenceTestsStruct:{}.!InterfaceVal:invalid! ERROR:interface element is nil`,
			`TypeRefID.ReferenceTestsStruct:{}.PtrPtrVal:{}:BasicStruct`,
			`TypeRefID.ReferenceTestsStruct:{}.PtrVal:{}:BasicStruct`,
			`RootID.{}:ReferenceTestsStruct`,
		},
		DerefStrings: []string{
			`RootID.{}`,
			`RootID.{}.!InterfaceVal:invalid! ERROR:interface element is nil`,
			`RootID.{}.PtrPtrVal:{}`,
			`RootID.{}.PtrPtrVal:{}.BoolVal:boolean`,
			`RootID.{}.PtrPtrVal:{}.Float64Val:float`,
			`RootID.{}.PtrPtrVal:{}.IntVal:integer`,
			`RootID.{}.PtrPtrVal:{}.StringVal:string`,
			`RootID.{}.PtrVal:{}`,
			`RootID.{}.PtrVal:{}.BoolVal:boolean`,
			`RootID.{}.PtrVal:{}.Float64Val:float`,
			`RootID.{}.PtrVal:{}.IntVal:integer`,
			`RootID.{}.PtrVal:{}.StringVal:string`,
		},
	},
	{
		Name:  "reference-tests-init",
		Value: ReferenceTestsStruct{InterfaceVal: &BasicStruct{}},
		RefStrings: []string{
			`TypeRefID.BasicStruct:{}`,
			`TypeRefID.BasicStruct:{}.BoolVal:boolean`,
			`TypeRefID.BasicStruct:{}.Float64Val:float`,
			`TypeRefID.BasicStruct:{}.IntVal:integer`,
			`TypeRefID.BasicStruct:{}.StringVal:string`,
			`TypeRefID.ReferenceTestsStruct:{}`,
			`TypeRefID.ReferenceTestsStruct:{}.InterfaceVal:{}:BasicStruct`,
			`TypeRefID.ReferenceTestsStruct:{}.PtrPtrVal:{}:BasicStruct`,
			`TypeRefID.ReferenceTestsStruct:{}.PtrVal:{}:BasicStruct`,
			`RootID.{}:ReferenceTestsStruct`,
		},
		DerefStrings: []string{
			`RootID.{}`,
			`RootID.{}.InterfaceVal:{}`,
			`RootID.{}.InterfaceVal:{}.BoolVal:boolean`,
			`RootID.{}.InterfaceVal:{}.Float64Val:float`,
			`RootID.{}.InterfaceVal:{}.IntVal:integer`,
			`RootID.{}.InterfaceVal:{}.StringVal:string`,
			`RootID.{}.PtrPtrVal:{}`,
			`RootID.{}.PtrPtrVal:{}.BoolVal:boolean`,
			`RootID.{}.PtrPtrVal:{}.Float64Val:float`,
			`RootID.{}.PtrPtrVal:{}.IntVal:integer`,
			`RootID.{}.PtrPtrVal:{}.StringVal:string`,
			`RootID.{}.PtrVal:{}`,
			`RootID.{}.PtrVal:{}.BoolVal:boolean`,
			`RootID.{}.PtrVal:{}.Float64Val:float`,
			`RootID.{}.PtrVal:{}.IntVal:integer`,
			`RootID.{}.PtrVal:{}.StringVal:string`,
		},
		OpenAPIStrings: []string{
			`openapi: 3.0.0`,
			`components:`,
			`  schemas:`,
			`    BasicStruct:`,
			`      type: object`,
			`      additionalProperties: false`,
			`      properties:`,
			`        BoolVal:`,
			`          type: boolean`,
			`        Float64Val:`,
			`          type: number`,
			`          format: double`,
			`        IntVal:`,
			`          type: integer`,
			`        StringVal:`,
			`          type: string`,
			`    ReferenceTestsStruct:`,
			`      type: object`,
			`      additionalProperties: false`,
			`      properties:`,
			`        InterfaceVal:`,
			`          $ref: '#/components/schemas/BasicStruct'`,
			`        PtrPtrVal:`,
			`          $ref: '#/components/schemas/BasicStruct'`,
			`        PtrVal:`,
			`          $ref: '#/components/schemas/BasicStruct'`,
			`paths:`,
			`  /test/path:`,
			`    get:`,
			`      summary: Return data.`,
			`      responses:`,
			`        '200':`,
			`          description: Success`,
			`          content:`,
			`            application/json:`,
			`              schema:`,
			`                $ref: '#/components/schemas/ReferenceTestsStruct'`,
		},
	},
}

// Test cyclical relationships:
// A --> B --> C --> A
type AStruct struct {
	AName  string   `json:"aName,omitempty"`
	AChild *BStruct `json:"aChild"`
}

type BStruct struct {
	BName  string   `json:"bName"`
	BChild *CStruct `json:"bChild"`
}

type CStruct struct {
	CName  string   `json:"cName"`
	CChild *AStruct `json:"cChild"`
}

type BadType interface{}

type CycleTest struct {
	Level  int      `json:"-"`
	CycleA AStruct  `json:"cycleA"`
	CycleB *BStruct `json:"cycleB"`
	CycleC struct {
		C CStruct `json:"c"`
	}
}

var cycleTests = []fixtures.TestCase{
	{
		Name:  "cycle-test",
		Value: &CycleTest{},
		RefStrings: []string{
			`TypeRefID.AStruct:{}`,
			`TypeRefID.AStruct:{}.AChild:{}:BStruct`,
			`TypeRefID.AStruct:{}.AName:string`,
			`TypeRefID.BStruct:{}`,
			`TypeRefID.BStruct:{}.BChild:{}:CStruct`,
			`TypeRefID.BStruct:{}.BName:string`,
			`TypeRefID.CStruct:{}`,
			`TypeRefID.CStruct:{}.CChild:{}:AStruct`,
			`TypeRefID.CStruct:{}.CName:string`,
			`TypeRefID.CycleTest:{}`,
			`TypeRefID.CycleTest:{}.CycleA:{}:AStruct`,
			`TypeRefID.CycleTest:{}.CycleB:{}:BStruct`,
			`TypeRefID.CycleTest:{}.CycleC:{}`,
			`TypeRefID.CycleTest:{}.CycleC:{}.C:{}:CStruct`,
			`TypeRefID.CycleTest:{}.Level:integer`,
			`RootID.{}:CycleTest`,
		},
		DerefStrings: []string{
			`RootID.{}`,
			`RootID.{}.CycleA:{}`,
			`RootID.{}.CycleA:{}.AChild:{}`,
			`RootID.{}.CycleA:{}.AChild:{}.BChild:{}`,
			`RootID.{}.CycleA:{}.AChild:{}.BChild:{}.!CChild:{}:AStruct! ERROR:cyclical reference`,
			`RootID.{}.CycleA:{}.AChild:{}.BChild:{}.CName:string`,
			`RootID.{}.CycleA:{}.AChild:{}.BName:string`,
			`RootID.{}.CycleA:{}.AName:string`,
			`RootID.{}.CycleB:{}`,
			`RootID.{}.CycleB:{}.BChild:{}`,
			`RootID.{}.CycleB:{}.BChild:{}.CChild:{}`,
			`RootID.{}.CycleB:{}.BChild:{}.CChild:{}.!AChild:{}:BStruct! ERROR:cyclical reference`,
			`RootID.{}.CycleB:{}.BChild:{}.CChild:{}.AName:string`,
			`RootID.{}.CycleB:{}.BChild:{}.CName:string`,
			`RootID.{}.CycleB:{}.BName:string`,
			`RootID.{}.CycleC:{}`,
			`RootID.{}.CycleC:{}.C:{}`,
			`RootID.{}.CycleC:{}.C:{}.CChild:{}`,
			`RootID.{}.CycleC:{}.C:{}.CChild:{}.AChild:{}`,
			`RootID.{}.CycleC:{}.C:{}.CChild:{}.AChild:{}.!BChild:{}:CStruct! ERROR:cyclical reference`,
			`RootID.{}.CycleC:{}.C:{}.CChild:{}.AChild:{}.BName:string`,
			`RootID.{}.CycleC:{}.C:{}.CChild:{}.AName:string`,
			`RootID.{}.CycleC:{}.C:{}.CName:string`,
			`RootID.{}.Level:integer`,
		},
		JSONStrings: []string{
			`definitions.cycleA:{}`,
			`definitions.cycleA:{}.aChild:{}:BStruct`,
			`definitions.cycleA:{}.aName:string`,
			`definitions.aChild:{}`,
			`definitions.aChild:{}.bChild:{}:CStruct`,
			`definitions.aChild:{}.bName:string`,
			`definitions.bChild:{}`,
			`definitions.bChild:{}.cChild:{}:AStruct`,
			`definitions.bChild:{}.cName:string`,
			`definitions.CycleTest:{}`,
			`definitions.CycleTest:{}.cycleA:{}:AStruct`,
			`definitions.CycleTest:{}.cycleB:{}:BStruct`,
			`definitions.CycleTest:{}.CycleC:{}`,
			`definitions.CycleTest:{}.CycleC:{}.c:{}:CStruct`,
			`$.{}:CycleTest`,
		},
		OpenAPIStrings: []string{
			`openapi: 3.0.0`,
			`components:`,
			`  schemas:`,
			`    cycleA:`,
			`      type: object`,
			`      additionalProperties: false`,
			`      properties:`,
			`        aChild:`,
			`          $ref: '#/components/schemas/BStruct'`,
			`        aName:`,
			`          type: string`,
			`    aChild:`,
			`      type: object`,
			`      additionalProperties: false`,
			`      properties:`,
			`        bChild:`,
			`          $ref: '#/components/schemas/CStruct'`,
			`        bName:`,
			`          type: string`,
			`    bChild:`,
			`      type: object`,
			`      additionalProperties: false`,
			`      properties:`,
			`        cChild:`,
			`          $ref: '#/components/schemas/AStruct'`,
			`        cName:`,
			`          type: string`,
			`    CycleTest:`,
			`      type: object`,
			`      additionalProperties: false`,
			`      properties:`,
			`        cycleA:`,
			`          $ref: '#/components/schemas/AStruct'`,
			`        cycleB:`,
			`          $ref: '#/components/schemas/BStruct'`,
			`        CycleC:`,
			`          type: object`,
			`          additionalProperties: false`,
			`          properties:`,
			`            c:`,
			`              $ref: '#/components/schemas/CStruct'`,
			`paths:`,
			`  /test/path:`,
			`    get:`,
			`      summary: Return data.`,
			`      responses:`,
			`        '200':`,
			`          description: Success`,
			`          content:`,
			`            application/json:`,
			`              schema:`,
			`                $ref: '#/components/schemas/CycleTest'`,
		},
	},
}

type JSONTagTests struct {
	NoTag      string
	ExcludeTag string `json:"-"`
	RenameOne  string `json:"renameOne"`
	RenameTwo  string `json:"something"`
}

var jsonTagTests = []fixtures.TestCase{
	{
		Name:  "json-tags",
		Value: JSONTagTests{},
		RefStrings: []string{
			`TypeRefID.JSONTagTests:{}`,
			`TypeRefID.JSONTagTests:{}.ExcludeTag:string`,
			`TypeRefID.JSONTagTests:{}.NoTag:string`,
			`TypeRefID.JSONTagTests:{}.RenameOne:string`,
			`TypeRefID.JSONTagTests:{}.RenameTwo:string`,
			`RootID.{}:JSONTagTests`,
		},
		DerefStrings: []string{
			`RootID.{}`,
			`RootID.{}.ExcludeTag:string`,
			`RootID.{}.NoTag:string`,
			`RootID.{}.RenameOne:string`,
			`RootID.{}.RenameTwo:string`,
		},
		JSONStrings: []string{
			`definitions.JSONTagTests:{}`,
			`definitions.JSONTagTests:{}.NoTag:string`,
			`definitions.JSONTagTests:{}.renameOne:string`,
			`definitions.JSONTagTests:{}.something:string`,
			`$.{}:JSONTagTests`,
		},
		OpenAPIStrings: []string{
			`openapi: 3.0.0`,
			`components:`,
			`  schemas:`,
			`    JSONTagTests:`,
			`      type: object`,
			`      additionalProperties: false`,
			`      properties:`,
			`        NoTag:`,
			`          type: string`,
			`        renameOne:`,
			`          type: string`,
			`        something:`,
			`          type: string`,
			`paths:`,
			`  /test/path:`,
			`    get:`,
			`      summary: Return data.`,
			`      responses:`,
			`        '200':`,
			`          description: Success`,
			`          content:`,
			`            application/json:`,
			`              schema:`,
			`                $ref: '#/components/schemas/JSONTagTests'`,
		},
	},
}

var structTests = []fixtures.TestCase{
	// {Name: "struct-empty", Value: func() interface{} { var g struct{}; return g }()},
	// {Name: "PrivateStruct-nil", Value: func() interface{} { var g PrivateStruct; return g }()},
	{Name: "BasicStruct-nil", Value: func() interface{} { var g BasicStruct; return g }()},
	// {Name: "CompoundStruct-nil", Value: func() interface{} { var g CompoundStruct; return g }()},
	// {Name: "CycleTest-nil", Value: func() interface{} { var g CycleTest; return g }()},
}

//
//{Name: "makeJSON, value", Value: makeJSON(nil)},

var testCases = []fixtures.TestCase{
	{Name: "GoodEntity, var", Value: func() interface{} { var g GoodEntity; return g }()},
	{Name: "GoodEntity, empty", Value: GoodEntity{}},
	{Name: "GoodEntity, values", Value: GoodEntity{
		Message: "hello",
		IntVal:  123,
		Same:    true,
		secret:  "shh",
	}},

	{Name: "map[string]bool, values", Value: map[string]bool{"trueVal": true, "falseVal": false}},

	{Name: "[]*MainStruct, nil", Value: []*MainStruct{}},
	{Name: "[0]*MainStruct, nil", Value: [0]*MainStruct{}},
	{Name: "[1]*MainStruct, nil", Value: [1]*MainStruct{}},

	{Name: "*GoodEntity, var", Value: func() interface{} { var g *GoodEntity; return g }()},
	{Name: "*GoodEntity, empty", Value: &GoodEntity{}},
	{Name: "*GoodEntity, values", Value: &GoodEntity{
		Message: "hello",
		IntVal:  123,
		Same:    true,
		secret:  "shh",
	}},

	{Name: "*OtherEntity, var", Value: func() interface{} { var g *OtherEntity; return g }()},
	{Name: "*OtherEntity, empty", Value: &OtherEntity{}},
	{Name: "*OtherEntity, values", Value: &OtherEntity{
		Status:   "ok",
		IntVal:   123,
		FloatVal: 234.345,
		Same:     true,
		MapVal:   make(map[string]int64),
		Good:     GoodEntity{},
	}},

	{Name: "NamedEntity, empty", Value: &NamedEntity{}},
}

// StringStruct has one string field.
type StringStruct struct {
	Value string
}

// Private Struct only has private fields.
type PrivateStruct struct {
	boolVal    bool
	intVal     int
	float64Val float64
	stringVal  string
}

// BasicStruct has one field for each basic type.
type BasicStruct struct {
	BoolVal    bool
	IntVal     int
	Float64Val float64
	StringVal  string
}

// CompoundStruct has fields with compound types.
type CompoundStruct struct {
	//	Array
	ZeroArrayVal  [0]string
	ThreeArrayVal [3]string

	//	Slice
	SliceVal []string

	//	Map
	MapVal map[string]string

	//	Struct
	StructVal        StringStruct
	PrivateStructVal PrivateStruct
}

/*
Only consider basic types:
- string, int, float, bool
- slices, arrays
- structs, maps

*/
type MainStruct struct {
	StringVal string `json:"stringVal,omitempty"`
	IntVal    int    `json:"intVal" datastore:",noindex"`
	FloatVal  float64
	BoolVal   bool

	SliceVal []int

	InterfaceVal interface{}

	StructPtr *GoodEntity
	StructVal OtherEntity

	StringPtr *string

	// Test duplicate JSON keys when capitalized.
	DuplicateOne string
	DuplicateTwo string `json:"duplicateOne"`

	privateVal string
}

// define a struct for data storage
type GoodEntity struct {
	Message string
	IntVal  int64
	Same    bool

	secret string
}

// Test named and un-named types.
type SimpleString string
type SimpleInt int64
type SimpleFloat float64
type SimpleBool bool
type SimpleInterface interface{}
type SimpleSlice []string
type SimpleMap map[string]int64
type SimpleStruct GoodEntity
type SimpleStructSlice []GoodEntity
type SimplePtr *GoodEntity
type SimplePtrSlice []*GoodEntity

type NamedEntity struct {
	NamedString SimpleString `json:"namedString,omitempty"`
	RealString  string

	NamedInt SimpleInt
	RealInt  int64

	NamedFloat SimpleFloat
	RealFloat  float64

	NamedBool SimpleBool
	RealBool  bool

	NamedInterface SimpleInterface
	RealInterface  interface{}

	NamedSlice SimpleSlice
	RealSlice  []string

	NamedMap SimpleMap
	RealMap  map[string]int64

	NamedStruct SimpleStruct
	RealStruct  GoodEntity

	NamedStructSlice SimpleStructSlice
	RealStructSlice  []GoodEntity

	NamedPtr SimplePtr
	RealPtr  *GoodEntity

	NamedPtrSlice SimplePtrSlice
	RealPtrSlice  []*GoodEntity
}

// define a different struct to test mismatched structs
type OtherEntity struct {
	Status   string
	IntVal   int64
	FloatVal float64
	Same     bool
	Simple   SimpleInt

	MapNil map[string]int64
	MapVal map[string]int64

	Good         GoodEntity
	GoodPtr      *GoodEntity
	GoodSlice    []GoodEntity
	GoodPtrSlice []*GoodEntity

	AnonStruct struct {
		FieldOne   string
		FieldTwo   int32
		FieldThree float32
	}
}

// fromJSON converts a JSON string into an interface.
func fromJSON(in []byte) interface{} {
	var out interface{}

	if err := json.Unmarshal(in, &out); err != nil {
		err = fmt.Errorf("ERROR json.Unmarshal: %s\n%s", err, string(in))
		fmt.Println(err)
		return err
	}

	// // DEBUGXXXXX Print indented JSON string.
	// if out != nil {
	// 	if b, err := json.MarshalIndent(out, "", "  "); err == nil {
	// 		fmt.Println(string(b))
	// 	}
	// }

	return out
}

// makeJSON converts an interface to JSON.
func makeJSON(x interface{}) interface{} {
	var s = "hey"

	x = &MainStruct{
		StringVal: "hello",
		IntVal:    123,
		FloatVal:  234.345,
		BoolVal:   true,
		SliceVal:  []int{1, 2, 3},
		StructPtr: &GoodEntity{
			Message: "hi",
			IntVal:  234,
			Same:    true,
			secret:  "eyes only",
		},
		StructVal: OtherEntity{
			Status:   "ok",
			IntVal:   789,
			FloatVal: 789.123,
			Same:     true,
			MapVal:   map[string]int64{"one": 234, "two": 345, "three": 456},
			Good: GoodEntity{
				Message: "",
				IntVal:  0,
				Same:    false,
				secret:  "",
			},
			GoodPtr: &GoodEntity{
				Message: "hi",
				IntVal:  234,
				Same:    true,
				secret:  "eyes only",
			},
			GoodSlice:    []GoodEntity{},
			GoodPtrSlice: []*GoodEntity{},
		},
		StringPtr: &s,

		DuplicateOne: "one",
		DuplicateTwo: "two",

		privateVal: "shh",
	}

	if b, err := json.Marshal(x); err != nil {
		return fmt.Errorf("ERROR json.Marshal: %s", err)
	} else {
		return fromJSON(b)
	}
}

func runTests(t *testing.T, testCases []fixtures.TestCase) {
	r := reflector.NewReflector()

	for _, test := range testCases {
		r.Reset()
		//r.Label = test.name

		gotResult := r.DeriveSchema(test.Value)

		// if b, err := json.MarshalIndent(gotResult, "", "  "); err != nil {
		// 	t.Errorf("TEST_FAIL %s: json.Marshal err=%s", test.name, err)
		// } else {
		// 	fmt.Println(string(b))
		// }

		for i := 0; i < 2; i++ {
			opt := renderer.NewOptions()
			opt.DeReference = i == 1

			r := simple.NewSimpleRenderer(opt)
			gotStrings, _ := r.ProcessSchema(gotResult)

			var wantStrings []string
			if opt.DeReference {
				wantStrings = test.DerefStrings
			} else {
				wantStrings = test.RefStrings
			}

			testName := fmt.Sprintf("%s: deref=%t", test.Name, opt.DeReference)
			util.CompareStrings(t, testName, gotStrings, wantStrings)
		}

		// Test json dialect.
		if len(test.JSONStrings) > 0 {
			opt := renderer.NewOptions()
			opt.DeReference = false

			r := json2.NewJSONRenderer(opt)
			gotStrings, _ := r.ProcessSchema(gotResult)
			wantStrings := test.JSONStrings

			testName := fmt.Sprintf("%s: dialect=json", test.Name)
			util.CompareStrings(t, testName, gotStrings, wantStrings)
		}

		// Test OpenAPI schema.
		if len(test.OpenAPIStrings) > 0 {
			opt := renderer.NewOptions()
			opt.DeReference = false
			opt.Indent = 0

			r := openapi.NewOpenAPIRenderer("/test/path", opt)
			gotStrings, _ := r.ProcessSchema(gotResult)
			wantStrings := test.OpenAPIStrings

			testName := fmt.Sprintf("%s: dialect=openapi", test.Name)
			util.CompareStrings(t, testName, gotStrings, wantStrings)

			// Verify that YAML is valid.
			yamlStr := strings.Join(gotStrings, "\n")
			var yamlOut interface{}
			if err := yaml.Unmarshal([]byte(yamlStr), &yamlOut); err != nil {
				t.Errorf("TEST_FAIL %s: yaml err=%s", test.Name, err)
			} else {
				t.Logf("TEST_OK %s: yaml", test.Name)
			}
		}
	}
}

func TestReflector_AllTests(t *testing.T) {
	for _, testCases := range allTests {
		runTests(t, testCases)
	}
}
