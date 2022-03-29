package main

import (
	"encoding/json"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/gitmann/b9schema-golang/common/util"
	"github.com/gitmann/b9schema-golang/fixtures"
	"github.com/gitmann/b9schema-golang/reflector"
	"github.com/gitmann/b9schema-golang/renderer"
	"github.com/gitmann/b9schema-golang/renderer/openapi"
	"github.com/gitmann/b9schema-golang/renderer/simple"
	"sort"
	"strings"
	"testing"
	"time"
	"unsafe"
)

var allTests = map[string][]fixtures.TestCase{
	"01-root-jaon": rootJSONTests,
	"02-root-go":   rootGoTests,
	"03-type":      typeTests,
	"04-list":      listTests,
	"05-compound":  compoundTests,
	"06-reference": referenceTests,
	"07-cycle":     cycleTests,
	"08-json-tag":  jsonTagTests,
	"09-nested":    nestedTests,
}

// *** All reflect types ***

// rootTests validate that the top-level element is either a struct or Reference.
var rootJSONTests = []fixtures.TestCase{
	{
		Name:  "json-null",
		Value: fromJSON([]byte(`null`)),
		Want: map[string]fixtures.WantSet{
			"simple": map[bool][]string{
				false: []string{"Root.!invalid:nil! ERROR:kind not supported"},
				true:  []string{"Root.!invalid:nil! ERROR:kind not supported"},
			},
		},
	},
	{
		Name:  "json-string",
		Value: fromJSON([]byte(`"Hello"`)),
		Want: map[string]fixtures.WantSet{
			"simple": map[bool][]string{
				false: []string{"Root.!string! ERROR:root type must be a struct"},
				true:  []string{"Root.!string! ERROR:root type must be a struct"},
			},
		},
	},
	{
		Name:  "json-int",
		Value: fromJSON([]byte(`123`)),
		Want: map[string]fixtures.WantSet{
			"simple": map[bool][]string{
				false: []string{"Root.!float! ERROR:root type must be a struct"},
				true:  []string{"Root.!float! ERROR:root type must be a struct"},
			},
		},
	},
	{
		Name:  "json-float",
		Value: fromJSON([]byte(`234.345`)),
		Want: map[string]fixtures.WantSet{
			"simple": map[bool][]string{
				false: []string{"Root.!float! ERROR:root type must be a struct"},
				true:  []string{"Root.!float! ERROR:root type must be a struct"},
			},
		},
	},
	{
		Name:  "json-bool",
		Value: fromJSON([]byte(`true`)),
		Want: map[string]fixtures.WantSet{
			"simple": map[bool][]string{
				false: []string{"Root.!boolean! ERROR:root type must be a struct"},
				true:  []string{"Root.!boolean! ERROR:root type must be a struct"},
			},
		},
	},
	{
		Name:  "json-list-empty",
		Value: fromJSON([]byte(`[]`)),
		Want: map[string]fixtures.WantSet{
			"simple": map[bool][]string{
				false: []string{"Root.![]! ERROR:root type must be a struct"},
				true:  []string{"Root.![]! ERROR:root type must be a struct"},
			},
		},
	},
	{
		Name:  "json-list",
		Value: fromJSON([]byte(`[1,2,3]`)),
		Want: map[string]fixtures.WantSet{
			"simple": map[bool][]string{
				false: []string{"Root.![]! ERROR:root type must be a struct"},
				true:  []string{"Root.![]! ERROR:root type must be a struct"},
			},
		},
	},
	{
		Name:  "json-object-empty",
		Value: fromJSON([]byte(`{}`)),
		Want: map[string]fixtures.WantSet{
			"simple": map[bool][]string{
				false: []string{"Root.!map{}! ERROR:root type must be a struct"},
				true:  []string{"Root.!map{}! ERROR:root type must be a struct"},
			},
		},
	},
	{
		Name:  "json-object",
		Value: fromJSON([]byte(`{"key1":"Hello"}`)),
		Want: map[string]fixtures.WantSet{
			"simple": map[bool][]string{
				false: []string{
					"Root.{}",
					"Root.{}.Key1:string",
				},
				true: []string{
					"Root.{}",
					"Root.{}.Key1:string",
				},
			},
		},
	},
}

var rootGoTests = []fixtures.TestCase{
	{
		Name:  "golang-nil",
		Value: nil,
		Want: map[string]fixtures.WantSet{
			"simple": map[bool][]string{
				false: []string{"Root.!invalid:nil! ERROR:kind not supported"},
				true:  []string{"Root.!invalid:nil! ERROR:kind not supported"},
			},
		},
	},
	{
		Name:  "golang-string",
		Value: "Hello",
		Want: map[string]fixtures.WantSet{
			"simple": map[bool][]string{
				false: []string{"Root.!string! ERROR:root type must be a struct"},
				true:  []string{"Root.!string! ERROR:root type must be a struct"},
			},
		},
	},
	{
		Name:  "golang-int",
		Value: 123,
		Want: map[string]fixtures.WantSet{
			"simple": map[bool][]string{
				false: []string{"Root.!integer! ERROR:root type must be a struct"},
				true:  []string{"Root.!integer! ERROR:root type must be a struct"},
			},
		},
	},
	{
		Name:  "golang-float",
		Value: 234.345,
		Want: map[string]fixtures.WantSet{
			"simple": map[bool][]string{
				false: []string{"Root.!float! ERROR:root type must be a struct"},
				true:  []string{"Root.!float! ERROR:root type must be a struct"},
			},
		},
	},
	{
		Name:  "golang-bool",
		Value: true,
		Want: map[string]fixtures.WantSet{
			"simple": map[bool][]string{
				false: []string{"Root.!boolean! ERROR:root type must be a struct"},
				true:  []string{"Root.!boolean! ERROR:root type must be a struct"},
			},
		},
	},
	{
		Name:  "golang-array-0",
		Value: [0]string{},
		Want: map[string]fixtures.WantSet{
			"simple": map[bool][]string{
				false: []string{"Root.![]! ERROR:root type must be a struct"},
				true:  []string{"Root.![]! ERROR:root type must be a struct"},
			},
		},
	},
	{
		Name:  "golang-array-3",
		Value: [3]string{},
		Want: map[string]fixtures.WantSet{
			"simple": map[bool][]string{
				false: []string{"Root.![]! ERROR:root type must be a struct"},
				true:  []string{"Root.![]! ERROR:root type must be a struct"},
			},
		},
	},
	{
		Name:  "golang-slice-nil",
		Value: func() interface{} { var s []string; return s }(),
		Want: map[string]fixtures.WantSet{
			"simple": map[bool][]string{
				false: []string{"Root.![]! ERROR:root type must be a struct"},
				true:  []string{"Root.![]! ERROR:root type must be a struct"},
			},
		},
	},
	{
		Name:  "golang-slice-0",
		Value: []string{},
		Want: map[string]fixtures.WantSet{
			"simple": map[bool][]string{
				false: []string{"Root.![]! ERROR:root type must be a struct"},
				true:  []string{"Root.![]! ERROR:root type must be a struct"},
			},
		},
	},
	{
		Name:  "golang-slice-3",
		Value: make([]string, 3),
		Want: map[string]fixtures.WantSet{
			"simple": map[bool][]string{
				false: []string{"Root.![]! ERROR:root type must be a struct"},
				true:  []string{"Root.![]! ERROR:root type must be a struct"},
			},
		},
	},
	{
		Name: "golang-struct-empty", Value: func() interface{} { var s struct{}; return s }(),
		Want: map[string]fixtures.WantSet{
			"simple": map[bool][]string{
				false: []string{"Root.!{}! ERROR:empty struct not supported"},
				true:  []string{"Root.!{}! ERROR:empty struct not supported"},
			},
		},
	},
	{
		Name:  "golang-struct-noinit",
		Value: func() interface{} { var s StringStruct; return s }(),
		Want: map[string]fixtures.WantSet{
			"simple": map[bool][]string{
				false: []string{
					`TypeRef.StringStruct:{}`,
					`TypeRef.StringStruct:{}.Value:string`,
					`Root.{}:StringStruct`,
				},
				true: []string{
					`Root.{}`,
					`Root.{}.Value:string`,
				},
			},
		},
	},
	{
		Name:  "golang-struct-init",
		Value: StringStruct{},
		Want: map[string]fixtures.WantSet{
			"simple": map[bool][]string{
				false: []string{
					`TypeRef.StringStruct:{}`,
					`TypeRef.StringStruct:{}.Value:string`,
					`Root.{}:StringStruct`,
				},
				true: []string{
					`Root.{}`,
					`Root.{}.Value:string`,
				},
			},
		},
	},
	{
		Name:  "golang-struct-private",
		Value: PrivateStruct{},
		Want: map[string]fixtures.WantSet{
			"simple": map[bool][]string{
				false: []string{
					`TypeRef.!PrivateStruct:{}! ERROR:struct has no exported fields`,
					`Root.!{}:PrivateStruct! ERROR:struct has no exported fields`,
				},
				true: []string{
					`Root.!{}! ERROR:struct has no exported fields`,
				},
			},
		},
	},

	{
		Name:  "golang-interface-struct-noinit",
		Value: func() interface{} { var s interface{} = StringStruct{}; return s }(),
		Want: map[string]fixtures.WantSet{
			"simple": map[bool][]string{
				false: []string{
					`TypeRef.StringStruct:{}`,
					`TypeRef.StringStruct:{}.Value:string`,
					`Root.{}:StringStruct`,
				},
				true: []string{
					`Root.{}`,
					`Root.{}.Value:string`,
				},
			},
		},
	},
	{
		Name:  "golang-pointer-struct-noinit",
		Value: func() interface{} { var s *StringStruct; return s }(),
		Want: map[string]fixtures.WantSet{
			"simple": map[bool][]string{
				false: []string{
					`TypeRef.StringStruct:{}`,
					`TypeRef.StringStruct:{}.Value:string`,
					`Root.{}:StringStruct`,
				},
				true: []string{
					`Root.{}`,
					`Root.{}.Value:string`,
				},
			},
		},
	},
	{
		Name:  "golang-pointer-struct-init",
		Value: &StringStruct{},
		Want: map[string]fixtures.WantSet{
			"simple": map[bool][]string{
				false: []string{
					`TypeRef.StringStruct:{}`,
					`TypeRef.StringStruct:{}.Value:string`,
					`Root.{}:StringStruct`,
				},
				true: []string{
					`Root.{}`,
					`Root.{}.Value:string`,
				},
			},
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
		Want: map[string]fixtures.WantSet{
			"simple": map[bool][]string{
				false: []string{
					`TypeRef.BoolTypes:{}`,
					`TypeRef.BoolTypes:{}.Bool:boolean`,
					`Root.{}:BoolTypes`,
				},
				true: []string{
					`Root.{}`,
					`Root.{}.Bool:boolean`,
				},
			},
			"openapi": map[bool][]string{
				false: []string{
					`info:`,
					`  title: boolean`,
					`  version: v1.0.0`,
					`openapi: 3.0.0`,
					``,
					`components:`,
					`  schemas:`,
					`    BoolTypes:`,
					`      type: object`,
					`      additionalProperties: false`,
					`      properties:`,
					`        Bool:`,
					`          type: boolean`,
					`paths:`,
					`  boolean:`,
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
				true: []string{
					`info:`,
					`  title: boolean`,
					`  version: v1.0.0`,
					`openapi: 3.0.0`,
					``,
					`components:`,
					`  schemas:`,
					`    BoolTypes:`,
					`      type: object`,
					`      additionalProperties: false`,
					`      properties:`,
					`        Bool:`,
					`          type: boolean`,
					`paths:`,
					`  boolean:`,
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
		},
	},
	{
		Name:  "integer",
		Value: IntegerTypes{},
		Want: map[string]fixtures.WantSet{
			"simple": map[bool][]string{
				false: []string{
					`TypeRef.IntegerTypes:{}`,
					`TypeRef.IntegerTypes:{}.Int:integer`,
					`TypeRef.IntegerTypes:{}.Int16:integer`,
					`TypeRef.IntegerTypes:{}.Int32:integer`,
					`TypeRef.IntegerTypes:{}.Int64:integer`,
					`TypeRef.IntegerTypes:{}.Int8:integer`,
					`TypeRef.IntegerTypes:{}.Uint:integer`,
					`TypeRef.IntegerTypes:{}.Uint16:integer`,
					`TypeRef.IntegerTypes:{}.Uint32:integer`,
					`TypeRef.IntegerTypes:{}.Uint64:integer`,
					`TypeRef.IntegerTypes:{}.Uint8:integer`,
					`TypeRef.IntegerTypes:{}.Uintptr:integer`,
					`Root.{}:IntegerTypes`,
				},
				true: []string{
					`Root.{}`,
					`Root.{}.Int:integer`,
					`Root.{}.Int16:integer`,
					`Root.{}.Int32:integer`,
					`Root.{}.Int64:integer`,
					`Root.{}.Int8:integer`,
					`Root.{}.Uint:integer`,
					`Root.{}.Uint16:integer`,
					`Root.{}.Uint32:integer`,
					`Root.{}.Uint64:integer`,
					`Root.{}.Uint8:integer`,
					`Root.{}.Uintptr:integer`,
				},
			},
			"openapi": map[bool][]string{
				false: []string{
					`info:`,
					`  title: integer`,
					`  version: v1.0.0`,
					`openapi: 3.0.0`,
					``,
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
					`  integer:`,
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
				true: []string{
					`info:`,
					`  title: integer`,
					`  version: v1.0.0`,
					`openapi: 3.0.0`,
					``,
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
					`  integer:`,
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
		},
	},
	{
		Name:  `float`,
		Value: FloatTypes{},
		Want: map[string]fixtures.WantSet{
			"simple": map[bool][]string{
				false: []string{
					`TypeRef.FloatTypes:{}`,
					`TypeRef.FloatTypes:{}.Float32:float`,
					`TypeRef.FloatTypes:{}.Float64:float`,
					`Root.{}:FloatTypes`,
				},
				true: []string{
					`Root.{}`,
					`Root.{}.Float32:float`,
					`Root.{}.Float64:float`,
				},
			},
			"openapi": map[bool][]string{
				false: []string{
					`info:`,
					`  title: float`,
					`  version: v1.0.0`,
					`openapi: 3.0.0`,
					``,
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
					`  float:`,
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
				true: []string{
					`info:`,
					`  title: float`,
					`  version: v1.0.0`,
					`openapi: 3.0.0`,
					``,
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
					`  float:`,
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
		},
	},
	{
		Name:  "string",
		Value: StringTypes{},
		Want: map[string]fixtures.WantSet{
			"simple": map[bool][]string{
				false: []string{
					`TypeRef.StringTypes:{}`,
					`TypeRef.StringTypes:{}.String:string`,
					`Root.{}:StringTypes`,
				},
				true: []string{
					`Root.{}`,
					`Root.{}.String:string`,
				},
			},
			"openapi": map[bool][]string{
				false: []string{
					`info:`,
					`  title: string`,
					`  version: v1.0.0`,
					`openapi: 3.0.0`,
					``,
					`components:`,
					`  schemas:`,
					`    StringTypes:`,
					`      type: object`,
					`      additionalProperties: false`,
					`      properties:`,
					`        String:`,
					`          type: string`,
					`paths:`,
					`  string:`,
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
				true: []string{
					`info:`,
					`  title: string`,
					`  version: v1.0.0`,
					`openapi: 3.0.0`,
					``,
					`components:`,
					`  schemas:`,
					`    StringTypes:`,
					`      type: object`,
					`      additionalProperties: false`,
					`      properties:`,
					`        String:`,
					`          type: string`,
					`paths:`,
					`  string:`,
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
		},
	},
	{
		Name:  "invalid",
		Value: InvalidTypes{},
		Want: map[string]fixtures.WantSet{
			"simple": map[bool][]string{
				false: []string{
					`TypeRef.InvalidTypes:{}`,
					`TypeRef.InvalidTypes:{}.!Chan:invalid:chan! ERROR:kind not supported`,
					`TypeRef.InvalidTypes:{}.!Complex128:invalid:complex128! ERROR:kind not supported`,
					`TypeRef.InvalidTypes:{}.!Complex64:invalid:complex64! ERROR:kind not supported`,
					`TypeRef.InvalidTypes:{}.!Func:invalid:func! ERROR:kind not supported`,
					`TypeRef.InvalidTypes:{}."!UnsafePointer:invalid:unsafe.Pointer!" ERROR:kind not supported`,
					`Root.{}:InvalidTypes`,
				},
				true: []string{
					`Root.{}`,
					`Root.{}.!Chan:invalid:chan! ERROR:kind not supported`,
					`Root.{}.!Complex128:invalid:complex128! ERROR:kind not supported`,
					`Root.{}.!Complex64:invalid:complex64! ERROR:kind not supported`,
					`Root.{}.!Func:invalid:func! ERROR:kind not supported`,
					`Root.{}."!UnsafePointer:invalid:unsafe.Pointer!" ERROR:kind not supported`,
				},
			},
			"openapi": map[bool][]string{
				false: []string{
					`info:`,
					`  title: invalid`,
					`  version: v1.0.0`,
					`openapi: 3.0.0`,
					``,
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
					`  invalid:`,
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
				true: []string{
					`info:`,
					`  title: invalid`,
					`  version: v1.0.0`,
					`openapi: 3.0.0`,
					``,
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
					`  invalid:`,
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
		},
	},
	{
		Name:  "compound",
		Value: CompoundTypes{},
		Want: map[string]fixtures.WantSet{
			"simple": map[bool][]string{
				false: []string{
					`TypeRef.CompoundTypes:{}`,
					`TypeRef.CompoundTypes:{}.Array0:[]`,
					`TypeRef.CompoundTypes:{}.Array0:[].string`,
					`TypeRef.CompoundTypes:{}.Array3:[]`,
					`TypeRef.CompoundTypes:{}.Array3:[].string`,
					`TypeRef.CompoundTypes:{}.!Interface:invalid! ERROR:interface element is nil`,
					`TypeRef.CompoundTypes:{}.!Map:map{}! ERROR:map key type must be string`,
					`TypeRef.CompoundTypes:{}.PrivatePtr:{}:PrivateStruct`,
					`TypeRef.CompoundTypes:{}.Ptr:{}:StringStruct`,
					`TypeRef.CompoundTypes:{}.Slice:[]`,
					`TypeRef.CompoundTypes:{}.Slice:[].!invalid! ERROR:interface element is nil`,
					`TypeRef.CompoundTypes:{}.!Struct:{}! ERROR:empty struct not supported`,
					`TypeRef.!PrivateStruct:{}! ERROR:struct has no exported fields`,
					`TypeRef.StringStruct:{}`,
					`TypeRef.StringStruct:{}.Value:string`,
					`Root.{}:CompoundTypes`,
				},
				true: []string{
					`Root.{}`,
					`Root.{}.Array0:[]`,
					`Root.{}.Array0:[].string`,
					`Root.{}.Array3:[]`,
					`Root.{}.Array3:[].string`,
					`Root.{}.!Interface:invalid! ERROR:interface element is nil`,
					`Root.{}.!Map:map{}! ERROR:map key type must be string`,
					`Root.{}.!PrivatePtr:{}! ERROR:struct has no exported fields`,
					`Root.{}.Ptr:{}`,
					`Root.{}.Ptr:{}.Value:string`,
					`Root.{}.Slice:[]`,
					`Root.{}.Slice:[].!invalid! ERROR:interface element is nil`,
					`Root.{}.!Struct:{}! ERROR:empty struct not supported`,
				},
			},
			"openapi": map[bool][]string{
				false: []string{
					`info:`,
					`  title: compound`,
					`  version: v1.0.0`,
					`openapi: 3.0.0`,
					``,
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
					`  compound:`,
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
				true: []string{
					`info:`,
					`  title: compound`,
					`  version: v1.0.0`,
					`openapi: 3.0.0`,
					``,
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
					`  compound:`,
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
		},
	},
	{
		Name:  "special",
		Value: SpecialTypes{},
		Want: map[string]fixtures.WantSet{
			"simple": map[bool][]string{
				false: []string{
					`TypeRef.SpecialTypes:{}`,
					`TypeRef.SpecialTypes:{}.DateTime:datetime`,
					`Root.{}:SpecialTypes`,
				},
				true: []string{
					`Root.{}`,
					`Root.{}.DateTime:datetime`,
				},
			},
			"openapi": map[bool][]string{
				false: []string{
					`info:`,
					`  title: special`,
					`  version: v1.0.0`,
					`openapi: 3.0.0`,
					``,
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
					`  special:`,
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
				true: []string{
					`info:`,
					`  title: special`,
					`  version: v1.0.0`,
					`openapi: 3.0.0`,
					``,
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
					`  special:`,
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
		Want: map[string]fixtures.WantSet{
			"simple": map[bool][]string{
				false: []string{
					`TypeRef.ArrayStruct:{}`,
					`TypeRef.ArrayStruct:{}.Array0:[]`,
					`TypeRef.ArrayStruct:{}.Array0:[].string`,
					`TypeRef.ArrayStruct:{}.Array2_3:[]`,
					`TypeRef.ArrayStruct:{}.Array2_3:[].[]`,
					`TypeRef.ArrayStruct:{}.Array2_3:[].[].string`,
					`TypeRef.ArrayStruct:{}.Array3:[]`,
					`TypeRef.ArrayStruct:{}.Array3:[].string`,
					`Root.{}:ArrayStruct`,
				},
				true: []string{
					`Root.{}`,
					`Root.{}.Array0:[]`,
					`Root.{}.Array0:[].string`,
					`Root.{}.Array2_3:[]`,
					`Root.{}.Array2_3:[].[]`,
					`Root.{}.Array2_3:[].[].string`,
					`Root.{}.Array3:[]`,
					`Root.{}.Array3:[].string`,
				},
			},
			"openapi": map[bool][]string{
				false: []string{
					`info:`,
					`  title: arrays`,
					`  version: v1.0.0`,
					`openapi: 3.0.0`,
					``,
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
					`  arrays:`,
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
				true: []string{
					`info:`,
					`  title: arrays`,
					`  version: v1.0.0`,
					`openapi: 3.0.0`,
					``,
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
					`  arrays:`,
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
		},
	},
	{
		Name:  "json-array",
		Value: fromJSON([]byte(jsonArrayTest)),
		Want: map[string]fixtures.WantSet{
			"simple": map[bool][]string{
				false: []string{
					`Root.{}`,
					`Root.{}.Array0:[]`,
					`Root.{}.Array0:[].!invalid! ERROR:interface element is nil`,
					`Root.{}.Array2_3:[]`,
					`Root.{}.Array2_3:[].[]`,
					`Root.{}.Array2_3:[].[].float`,
					`Root.{}.Array3:[]`,
					`Root.{}.Array3:[].string`,
				},
				true: []string{
					`Root.{}`,
					`Root.{}.Array0:[]`,
					`Root.{}.Array0:[].!invalid! ERROR:interface element is nil`,
					`Root.{}.Array2_3:[]`,
					`Root.{}.Array2_3:[].[]`,
					`Root.{}.Array2_3:[].[].float`,
					`Root.{}.Array3:[]`,
					`Root.{}.Array3:[].string`,
				},
			},
			"openapi": map[bool][]string{
				false: []string{
					`info:`,
					`  title: json-array`,
					`  version: v1.0.0`,
					`openapi: 3.0.0`,
					``,
					`paths:`,
					`  json-array:`,
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
				true: []string{
					`info:`,
					`  title: json-array`,
					`  version: v1.0.0`,
					`openapi: 3.0.0`,
					``,
					`paths:`,
					`  json-array:`,
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
		},
	},
	{
		Name:  "slices",
		Value: &SliceStruct{},
		Want: map[string]fixtures.WantSet{
			"simple": map[bool][]string{
				false: []string{
					`TypeRef.SliceStruct:{}`,
					`TypeRef.SliceStruct:{}.Array2:[]`,
					`TypeRef.SliceStruct:{}.Array2:[].[]`,
					`TypeRef.SliceStruct:{}.Array2:[].[].string`,
					`TypeRef.SliceStruct:{}.Slice:[]`,
					`TypeRef.SliceStruct:{}.Slice:[].string`,
					`Root.{}:SliceStruct`,
				},
				true: []string{
					`Root.{}`,
					`Root.{}.Array2:[]`,
					`Root.{}.Array2:[].[]`,
					`Root.{}.Array2:[].[].string`,
					`Root.{}.Slice:[]`,
					`Root.{}.Slice:[].string`,
				},
			},
			"openapi": map[bool][]string{
				false: []string{
					`info:`,
					`  title: slices`,
					`  version: v1.0.0`,
					`openapi: 3.0.0`,
					``,
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
					`  slices:`,
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
				true: []string{
					`info:`,
					`  title: slices`,
					`  version: v1.0.0`,
					`openapi: 3.0.0`,
					``,
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
					`  slices:`,
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
		Want: map[string]fixtures.WantSet{
			"simple": map[bool][]string{
				false: []string{
					`TypeRef.MapTestsStruct:{}`,
					`TypeRef.MapTestsStruct:{}.MapOK:{}`,
					`TypeRef.MapTestsStruct:{}.MapOK:{}.BoolVal:boolean`,
					`TypeRef.MapTestsStruct:{}.MapOK:{}.FloatVal:float`,
					`TypeRef.MapTestsStruct:{}.MapOK:{}.IntVal:float`,
					`TypeRef.MapTestsStruct:{}.MapOK:{}.ListVal:[]`,
					`TypeRef.MapTestsStruct:{}.MapOK:{}.ListVal:[].float`,
					`TypeRef.MapTestsStruct:{}.MapOK:{}.MapVal:{}`,
					`TypeRef.MapTestsStruct:{}.MapOK:{}.MapVal:{}.Key1:string`,
					`TypeRef.MapTestsStruct:{}.MapOK:{}.MapVal:{}.Key2:{}`,
					`TypeRef.MapTestsStruct:{}.MapOK:{}.MapVal:{}.Key2:{}.DeepKey1:string`,
					`TypeRef.MapTestsStruct:{}.MapOK:{}.MapVal:{}.Key2:{}.DeepKey2:float`,
					`TypeRef.MapTestsStruct:{}.MapOK:{}.StringVal:string`,
					`Root.{}:MapTestsStruct`,
				},
				true: []string{
					`Root.{}`,
					`Root.{}.MapOK:{}`,
					`Root.{}.MapOK:{}.BoolVal:boolean`,
					`Root.{}.MapOK:{}.FloatVal:float`,
					`Root.{}.MapOK:{}.IntVal:float`,
					`Root.{}.MapOK:{}.ListVal:[]`,
					`Root.{}.MapOK:{}.ListVal:[].float`,
					`Root.{}.MapOK:{}.MapVal:{}`,
					`Root.{}.MapOK:{}.MapVal:{}.Key1:string`,
					`Root.{}.MapOK:{}.MapVal:{}.Key2:{}`,
					`Root.{}.MapOK:{}.MapVal:{}.Key2:{}.DeepKey1:string`,
					`Root.{}.MapOK:{}.MapVal:{}.Key2:{}.DeepKey2:float`,
					`Root.{}.MapOK:{}.StringVal:string`,
				},
			},
			"openapi": map[bool][]string{
				false: []string{
					`info:`,
					`  title: golang-map`,
					`  version: v1.0.0`,
					`openapi: 3.0.0`,
					``,
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
					`  /05-compound/golang-map:`,
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
				true: []string{
					`info:`,
					`  title: golang-map`,
					`  version: v1.0.0`,
					`openapi: 3.0.0`,
					``,
					`paths:`,
					`  /05-compound/golang-map:`,
					`    get:`,
					`      summary: Return data.`,
					`      responses:`,
					`        '200':`,
					`          description: Success`,
					`          content:`,
					`            application/json:`,
					`              schema:`,
					`                description: 'From $ref: #/components/schemas/MapTestsStruct'`,
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
		},
	},
	{
		Name:  "json-map",
		Value: fromJSON([]byte(jsonMapTests)),
		Want: map[string]fixtures.WantSet{
			"simple": map[bool][]string{
				false: []string{
					`Root.{}`,
					`Root.{}.MapOK:{}`,
					`Root.{}.MapOK:{}.BoolVal:boolean`,
					`Root.{}.MapOK:{}.FloatVal:float`,
					`Root.{}.MapOK:{}.IntVal:float`,
					`Root.{}.MapOK:{}.ListVal:[]`,
					`Root.{}.MapOK:{}.ListVal:[].float`,
					`Root.{}.MapOK:{}.MapVal:{}`,
					`Root.{}.MapOK:{}.MapVal:{}.Key1:string`,
					`Root.{}.MapOK:{}.MapVal:{}.Key2:{}`,
					`Root.{}.MapOK:{}.MapVal:{}.Key2:{}.DeepKey1:string`,
					`Root.{}.MapOK:{}.MapVal:{}.Key2:{}.DeepKey2:float`,
					`Root.{}.MapOK:{}.StringVal:string`,
				},
				true: []string{
					`Root.{}`,
					`Root.{}.MapOK:{}`,
					`Root.{}.MapOK:{}.BoolVal:boolean`,
					`Root.{}.MapOK:{}.FloatVal:float`,
					`Root.{}.MapOK:{}.IntVal:float`,
					`Root.{}.MapOK:{}.ListVal:[]`,
					`Root.{}.MapOK:{}.ListVal:[].float`,
					`Root.{}.MapOK:{}.MapVal:{}`,
					`Root.{}.MapOK:{}.MapVal:{}.Key1:string`,
					`Root.{}.MapOK:{}.MapVal:{}.Key2:{}`,
					`Root.{}.MapOK:{}.MapVal:{}.Key2:{}.DeepKey1:string`,
					`Root.{}.MapOK:{}.MapVal:{}.Key2:{}.DeepKey2:float`,
					`Root.{}.MapOK:{}.StringVal:string`,
				},
			},
			"openapi": map[bool][]string{
				false: []string{
					`info:`,
					`  title: json-map`,
					`  version: v1.0.0`,
					`openapi: 3.0.0`,
					``,
					`paths:`,
					`  /05-compound/json-map:`,
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
				true: []string{
					`info:`,
					`  title: json-map`,
					`  version: v1.0.0`,
					`openapi: 3.0.0`,
					``,
					`paths:`,
					`  /05-compound/json-map:`,
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
		Want: map[string]fixtures.WantSet{
			"simple": map[bool][]string{
				false: []string{
					`TypeRef.BasicStruct:{}`,
					`TypeRef.BasicStruct:{}.BoolVal:boolean`,
					`TypeRef.BasicStruct:{}.Float64Val:float`,
					`TypeRef.BasicStruct:{}.IntVal:integer`,
					`TypeRef.BasicStruct:{}.StringVal:string`,
					`TypeRef.ReferenceTestsStruct:{}`,
					`TypeRef.ReferenceTestsStruct:{}.!InterfaceVal:invalid! ERROR:interface element is nil`,
					`TypeRef.ReferenceTestsStruct:{}.PtrPtrVal:{}:BasicStruct`,
					`TypeRef.ReferenceTestsStruct:{}.PtrVal:{}:BasicStruct`,
					`Root.{}:ReferenceTestsStruct`,
				},
				true: []string{
					`Root.{}`,
					`Root.{}.!InterfaceVal:invalid! ERROR:interface element is nil`,
					`Root.{}.PtrPtrVal:{}`,
					`Root.{}.PtrPtrVal:{}.BoolVal:boolean`,
					`Root.{}.PtrPtrVal:{}.Float64Val:float`,
					`Root.{}.PtrPtrVal:{}.IntVal:integer`,
					`Root.{}.PtrPtrVal:{}.StringVal:string`,
					`Root.{}.PtrVal:{}`,
					`Root.{}.PtrVal:{}.BoolVal:boolean`,
					`Root.{}.PtrVal:{}.Float64Val:float`,
					`Root.{}.PtrVal:{}.IntVal:integer`,
					`Root.{}.PtrVal:{}.StringVal:string`,
				},
			},
		},
	},
	{
		Name:  "reference-tests-init",
		Value: ReferenceTestsStruct{InterfaceVal: &BasicStruct{}},
		Want: map[string]fixtures.WantSet{
			"simple": map[bool][]string{
				false: []string{
					`TypeRef.BasicStruct:{}`,
					`TypeRef.BasicStruct:{}.BoolVal:boolean`,
					`TypeRef.BasicStruct:{}.Float64Val:float`,
					`TypeRef.BasicStruct:{}.IntVal:integer`,
					`TypeRef.BasicStruct:{}.StringVal:string`,
					`TypeRef.ReferenceTestsStruct:{}`,
					`TypeRef.ReferenceTestsStruct:{}.InterfaceVal:{}:BasicStruct`,
					`TypeRef.ReferenceTestsStruct:{}.PtrPtrVal:{}:BasicStruct`,
					`TypeRef.ReferenceTestsStruct:{}.PtrVal:{}:BasicStruct`,
					`Root.{}:ReferenceTestsStruct`,
				},
				true: []string{
					`Root.{}`,
					`Root.{}.InterfaceVal:{}`,
					`Root.{}.InterfaceVal:{}.BoolVal:boolean`,
					`Root.{}.InterfaceVal:{}.Float64Val:float`,
					`Root.{}.InterfaceVal:{}.IntVal:integer`,
					`Root.{}.InterfaceVal:{}.StringVal:string`,
					`Root.{}.PtrPtrVal:{}`,
					`Root.{}.PtrPtrVal:{}.BoolVal:boolean`,
					`Root.{}.PtrPtrVal:{}.Float64Val:float`,
					`Root.{}.PtrPtrVal:{}.IntVal:integer`,
					`Root.{}.PtrPtrVal:{}.StringVal:string`,
					`Root.{}.PtrVal:{}`,
					`Root.{}.PtrVal:{}.BoolVal:boolean`,
					`Root.{}.PtrVal:{}.Float64Val:float`,
					`Root.{}.PtrVal:{}.IntVal:integer`,
					`Root.{}.PtrVal:{}.StringVal:string`,
				},
			},
			"openapi": map[bool][]string{
				false: []string{
					`info:`,
					`  title: reference-tests-init`,
					`  version: v1.0.0`,
					`openapi: 3.0.0`,
					``,
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
					`  /06-reference/reference-tests-init:`,
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
				true: []string{
					`info:`,
					`  title: reference-tests-init`,
					`  version: v1.0.0`,
					`openapi: 3.0.0`,
					``,
					`paths:`,
					`  /06-reference/reference-tests-init:`,
					`    get:`,
					`      summary: Return data.`,
					`      responses:`,
					`        '200':`,
					`          description: Success`,
					`          content:`,
					`            application/json:`,
					`              schema:`,
					`                description: 'From $ref: #/components/schemas/ReferenceTestsStruct'`,
					`                type: object`,
					`                additionalProperties: false`,
					`                properties:`,
					`                  InterfaceVal:`,
					`                    description: 'From $ref: #/components/schemas/BasicStruct'`,
					`                    type: object`,
					`                    additionalProperties: false`,
					`                    properties:`,
					`                      BoolVal:`,
					`                        type: boolean`,
					`                      Float64Val:`,
					`                        type: number`,
					`                        format: double`,
					`                      IntVal:`,
					`                        type: integer`,
					`                      StringVal:`,
					`                        type: string`,
					`                  PtrPtrVal:`,
					`                    description: 'From $ref: #/components/schemas/BasicStruct'`,
					`                    type: object`,
					`                    additionalProperties: false`,
					`                    properties:`,
					`                      BoolVal:`,
					`                        type: boolean`,
					`                      Float64Val:`,
					`                        type: number`,
					`                        format: double`,
					`                      IntVal:`,
					`                        type: integer`,
					`                      StringVal:`,
					`                        type: string`,
					`                  PtrVal:`,
					`                    description: 'From $ref: #/components/schemas/BasicStruct'`,
					`                    type: object`,
					`                    additionalProperties: false`,
					`                    properties:`,
					`                      BoolVal:`,
					`                        type: boolean`,
					`                      Float64Val:`,
					`                        type: number`,
					`                        format: double`,
					`                      IntVal:`,
					`                        type: integer`,
					`                      StringVal:`,
					`                        type: string`,
				},
			},
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
		Want: map[string]fixtures.WantSet{
			"simple": map[bool][]string{
				false: []string{
					`TypeRef.AStruct:{}`,
					`TypeRef.AStruct:{}.AChild:{}:BStruct`,
					`TypeRef.AStruct:{}.AName:string`,
					`TypeRef.BStruct:{}`,
					`TypeRef.BStruct:{}.BChild:{}:CStruct`,
					`TypeRef.BStruct:{}.BName:string`,
					`TypeRef.CStruct:{}`,
					`TypeRef.CStruct:{}.CChild:{}:AStruct`,
					`TypeRef.CStruct:{}.CName:string`,
					`TypeRef.CycleTest:{}`,
					`TypeRef.CycleTest:{}.CycleA:{}:AStruct`,
					`TypeRef.CycleTest:{}.CycleB:{}:BStruct`,
					`TypeRef.CycleTest:{}.CycleC:{}`,
					`TypeRef.CycleTest:{}.CycleC:{}.C:{}:CStruct`,
					`TypeRef.CycleTest:{}.Level:integer`,
					`Root.{}:CycleTest`,
				},
				true: []string{
					`Root.{}`,
					`Root.{}.CycleA:{}`,
					`Root.{}.CycleA:{}.AChild:{}`,
					`Root.{}.CycleA:{}.AChild:{}.BChild:{}`,
					`Root.{}.CycleA:{}.AChild:{}.BChild:{}.!CChild:{}:AStruct! ERROR:cyclical reference`,
					`Root.{}.CycleA:{}.AChild:{}.BChild:{}.CName:string`,
					`Root.{}.CycleA:{}.AChild:{}.BName:string`,
					`Root.{}.CycleA:{}.AName:string`,
					`Root.{}.CycleB:{}`,
					`Root.{}.CycleB:{}.BChild:{}`,
					`Root.{}.CycleB:{}.BChild:{}.CChild:{}`,
					`Root.{}.CycleB:{}.BChild:{}.CChild:{}.!AChild:{}:BStruct! ERROR:cyclical reference`,
					`Root.{}.CycleB:{}.BChild:{}.CChild:{}.AName:string`,
					`Root.{}.CycleB:{}.BChild:{}.CName:string`,
					`Root.{}.CycleB:{}.BName:string`,
					`Root.{}.CycleC:{}`,
					`Root.{}.CycleC:{}.C:{}`,
					`Root.{}.CycleC:{}.C:{}.CChild:{}`,
					`Root.{}.CycleC:{}.C:{}.CChild:{}.AChild:{}`,
					`Root.{}.CycleC:{}.C:{}.CChild:{}.AChild:{}.!BChild:{}:CStruct! ERROR:cyclical reference`,
					`Root.{}.CycleC:{}.C:{}.CChild:{}.AChild:{}.BName:string`,
					`Root.{}.CycleC:{}.C:{}.CChild:{}.AName:string`,
					`Root.{}.CycleC:{}.C:{}.CName:string`,
					`Root.{}.Level:integer`,
				},
			},
			"openapi": map[bool][]string{
				false: []string{
					`info:`,
					`  title: cycle-test`,
					`  version: v1.0.0`,
					`openapi: 3.0.0`,
					``,
					`components:`,
					`  schemas:`,
					`    AStruct:`,
					`      type: object`,
					`      additionalProperties: false`,
					`      properties:`,
					`        aChild:`,
					`          $ref: '#/components/schemas/BStruct'`,
					`        aName:`,
					`          type: string`,
					`    BStruct:`,
					`      type: object`,
					`      additionalProperties: false`,
					`      properties:`,
					`        bChild:`,
					`          $ref: '#/components/schemas/CStruct'`,
					`        bName:`,
					`          type: string`,
					`    CStruct:`,
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
					`  /07-cycle/cycle-test:`,
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
				true: []string{
					`info:`,
					`  title: cycle-test`,
					`  version: v1.0.0`,
					`openapi: 3.0.0`,
					``,
					`paths:`,
					`  /07-cycle/cycle-test:`,
					`    get:`,
					`      summary: Return data.`,
					`      responses:`,
					`        '200':`,
					`          description: Success`,
					`          content:`,
					`            application/json:`,
					`              schema:`,
					`                description: 'From $ref: #/components/schemas/CycleTest'`,
					`                type: object`,
					`                additionalProperties: false`,
					`                properties:`,
					`                  cycleA:`,
					`                    description: 'From $ref: #/components/schemas/AStruct'`,
					`                    type: object`,
					`                    additionalProperties: false`,
					`                    properties:`,
					`                      aChild:`,
					`                        description: 'From $ref: #/components/schemas/BStruct'`,
					`                        type: object`,
					`                        additionalProperties: false`,
					`                        properties:`,
					`                          bChild:`,
					`                            description: 'From $ref: #/components/schemas/CStruct'`,
					`                            type: object`,
					`                            additionalProperties: false`,
					`                            properties:`,
					`                              cChild:`,
					`                                description: 'From $ref: #/components/schemas/AStruct;ERROR=cyclical reference'`,
					`                                type: object`,
					`                                additionalProperties: false`,
					`                              cName:`,
					`                                type: string`,
					`                          bName:`,
					`                            type: string`,
					`                      aName:`,
					`                        type: string`,
					`                  cycleB:`,
					`                    description: 'From $ref: #/components/schemas/BStruct'`,
					`                    type: object`,
					`                    additionalProperties: false`,
					`                    properties:`,
					`                      bChild:`,
					`                        description: 'From $ref: #/components/schemas/CStruct'`,
					`                        type: object`,
					`                        additionalProperties: false`,
					`                        properties:`,
					`                          cChild:`,
					`                            description: 'From $ref: #/components/schemas/AStruct'`,
					`                            type: object`,
					`                            additionalProperties: false`,
					`                            properties:`,
					`                              aChild:`,
					`                                description: 'From $ref: #/components/schemas/BStruct;ERROR=cyclical reference'`,
					`                                type: object`,
					`                                additionalProperties: false`,
					`                              aName:`,
					`                                type: string`,
					`                          cName:`,
					`                            type: string`,
					`                      bName:`,
					`                        type: string`,
					`                  CycleC:`,
					`                    type: object`,
					`                    additionalProperties: false`,
					`                    properties:`,
					`                      c:`,
					`                        description: 'From $ref: #/components/schemas/CStruct'`,
					`                        type: object`,
					`                        additionalProperties: false`,
					`                        properties:`,
					`                          cChild:`,
					`                            description: 'From $ref: #/components/schemas/AStruct'`,
					`                            type: object`,
					`                            additionalProperties: false`,
					`                            properties:`,
					`                              aChild:`,
					`                                description: 'From $ref: #/components/schemas/BStruct'`,
					`                                type: object`,
					`                                additionalProperties: false`,
					`                                properties:`,
					`                                  bChild:`,
					`                                    description: 'From $ref: #/components/schemas/CStruct;ERROR=cyclical reference'`,
					`                                    type: object`,
					`                                    additionalProperties: false`,
					`                                  bName:`,
					`                                    type: string`,
					`                              aName:`,
					`                                type: string`,
					`                          cName:`,
					`                            type: string`,
				},
			},
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
		Want: map[string]fixtures.WantSet{
			"simple": map[bool][]string{
				false: []string{
					`TypeRef.JSONTagTests:{}`,
					`TypeRef.JSONTagTests:{}.ExcludeTag:string`,
					`TypeRef.JSONTagTests:{}.NoTag:string`,
					`TypeRef.JSONTagTests:{}.RenameOne:string`,
					`TypeRef.JSONTagTests:{}.RenameTwo:string`,
					`Root.{}:JSONTagTests`,
				},
				true: []string{
					`Root.{}`,
					`Root.{}.ExcludeTag:string`,
					`Root.{}.NoTag:string`,
					`Root.{}.RenameOne:string`,
					`Root.{}.RenameTwo:string`,
				},
			},
			"openapi": map[bool][]string{
				false: []string{
					`info:`,
					`  title: json-tags`,
					`  version: v1.0.0`,
					`openapi: 3.0.0`,
					``,
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
					`  /08-json-tag/json-tags:`,
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
				true: []string{
					`info:`,
					`  title: json-tags`,
					`  version: v1.0.0`,
					`openapi: 3.0.0`,
					``,
					`paths:`,
					`  /08-json-tag/json-tags:`,
					`    get:`,
					`      summary: Return data.`,
					`      responses:`,
					`        '200':`,
					`          description: Success`,
					`          content:`,
					`            application/json:`,
					`              schema:`,
					`                description: 'From $ref: #/components/schemas/JSONTagTests'`,
					`                type: object`,
					`                additionalProperties: false`,
					`                properties:`,
					`                  NoTag:`,
					`                    type: string`,
					`                  renameOne:`,
					`                    type: string`,
					`                  something:`,
					`                    type: string`,
				},
			},
		},
	},
}

var nestedTests = []fixtures.TestCase{
	{
		Name:  "nested-struct",
		Value: OuterStruct{},
		Want: map[string]fixtures.WantSet{
			"simple": map[bool][]string{
				false: []string{
					`TypeRef.BasicStruct:{}`,
					`TypeRef.BasicStruct:{}.BoolVal:boolean`,
					`TypeRef.BasicStruct:{}.Float64Val:float`,
					`TypeRef.BasicStruct:{}.IntVal:integer`,
					`TypeRef.BasicStruct:{}.StringVal:string`,
					`TypeRef.InnerStruct:{}`,
					`TypeRef.InnerStruct:{}.ListOfStrings:[]`,
					`TypeRef.InnerStruct:{}.ListOfStrings:[].string`,
					`TypeRef.InnerStruct:{}.ListOfStructs:[]`,
					`TypeRef.InnerStruct:{}.ListOfStructs:[].{}:BasicStruct`,
					`TypeRef.OuterStruct:{}`,
					`TypeRef.OuterStruct:{}.ID:integer`,
					`TypeRef.OuterStruct:{}.Inner:{}:InnerStruct`,
					`Root.{}:OuterStruct`,
				},
				true: []string{
					`Root.{}`,
					`Root.{}.ID:integer`,
					`Root.{}.Inner:{}`,
					`Root.{}.Inner:{}.ListOfStrings:[]`,
					`Root.{}.Inner:{}.ListOfStrings:[].string`,
					`Root.{}.Inner:{}.ListOfStructs:[]`,
					`Root.{}.Inner:{}.ListOfStructs:[].{}`,
					`Root.{}.Inner:{}.ListOfStructs:[].{}.BoolVal:boolean`,
					`Root.{}.Inner:{}.ListOfStructs:[].{}.Float64Val:float`,
					`Root.{}.Inner:{}.ListOfStructs:[].{}.IntVal:integer`,
					`Root.{}.Inner:{}.ListOfStructs:[].{}.StringVal:string`,
				},
			},
			"openapi": map[bool][]string{
				false: []string{
					`info:`,
					`  title: nested-struct`,
					`  version: v1.0.0`,
					`openapi: 3.0.0`,
					``,
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
					`    InnerStruct:`,
					`      type: object`,
					`      additionalProperties: false`,
					`      properties:`,
					`        listOfStrings:`,
					`          type: array`,
					`          items:`,
					`            type: string`,
					`        listOfStructs:`,
					`          type: array`,
					`          items:`,
					`            $ref: '#/components/schemas/BasicStruct'`,
					`    OuterStruct:`,
					`      type: object`,
					`      additionalProperties: false`,
					`      properties:`,
					`        id:`,
					`          type: integer`,
					`        inner:`,
					`          $ref: '#/components/schemas/InnerStruct'`,
					`paths:`,
					`  /09-nested/nested-struct:`,
					`    get:`,
					`      summary: Return data.`,
					`      responses:`,
					`        '200':`,
					`          description: Success`,
					`          content:`,
					`            application/json:`,
					`              schema:`,
					`                $ref: '#/components/schemas/OuterStruct'`,
				},
				true: []string{
					`info:`,
					`  title: nested-struct`,
					`  version: v1.0.0`,
					`openapi: 3.0.0`,
					``,
					`paths:`,
					`  /09-nested/nested-struct:`,
					`    get:`,
					`      summary: Return data.`,
					`      responses:`,
					`        '200':`,
					`          description: Success`,
					`          content:`,
					`            application/json:`,
					`              schema:`,
					`                description: 'From $ref: #/components/schemas/OuterStruct'`,
					`                type: object`,
					`                additionalProperties: false`,
					`                properties:`,
					`                  id:`,
					`                    type: integer`,
					`                  inner:`,
					`                    description: 'From $ref: #/components/schemas/InnerStruct'`,
					`                    type: object`,
					`                    additionalProperties: false`,
					`                    properties:`,
					`                      listOfStrings:`,
					`                        type: array`,
					`                        items:`,
					`                          type: string`,
					`                      listOfStructs:`,
					`                        type: array`,
					`                        items:`,
					`                          description: 'From $ref: #/components/schemas/BasicStruct'`,
					`                          type: object`,
					`                          additionalProperties: false`,
					`                          properties:`,
					`                            BoolVal:`,
					`                              type: boolean`,
					`                            Float64Val:`,
					`                              type: number`,
					`                              format: double`,
					`                            IntVal:`,
					`                              type: integer`,
					`                            StringVal:`,
					`                              type: string`,
				},
			},
		},
	},
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

// Test nested data structures.
type OuterStruct struct {
	ID    int          `json:"id"`
	Inner *InnerStruct `json:"inner"`
}

type InnerStruct struct {
	ListOfStrings []string       `json:"listOfStrings"`
	ListOfStructs []*BasicStruct `json:"listOfStructs"`
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

func TestReflector_AllTests(t *testing.T) {
	// Booleans for deref looping.
	derefFlags := []bool{false, true}

	// want represents strings for deref false/true
	var want map[bool][]string
	var format string
	var gotStrings []string

	r := reflector.NewReflector()
	opt := renderer.NewOptions()

	// testName builds a test name string.
	testName := func(group, name, format string, deref bool) string {
		return fmt.Sprintf("%s/%s/%t/%s", group, name, deref, format)
	}

	// Build sorted list of test keys.
	allKeys := []string{}
	for k := range allTests {
		allKeys = append(allKeys, k)
	}
	sort.Strings(allKeys)

	for _, testGroup := range allKeys {
		testCases := allTests[testGroup]

		for _, test := range testCases {
			r.Reset()
			gotSchema := r.DeriveSchema(test.Value, testGroup+"/"+test.Name)

			format = "reflector"
			want = test.Want[format]
			if want != nil {
				for _, deref := range derefFlags {
					wantStrings := want[deref]
					if len(wantStrings) > 0 {
						name := testName(testGroup, test.Name, format, deref)

						if b, err := yaml.Marshal(gotSchema.CopyWithoutNative()); err != nil {
							t.Errorf("TEST_FAIL %s: yaml err=%s", name, err)
							continue
						} else {
							gotStrings = strings.Split(string(b), "\n")
						}

						util.CompareStrings(t, name, gotStrings, wantStrings)
					}
				}
			}

			format = "simple"
			want = test.Want[format]
			if want != nil {
				for _, deref := range derefFlags {
					wantStrings := want[deref]
					if len(wantStrings) > 0 {
						name := testName(testGroup, test.Name, format, deref)
						opt.DeReference = deref

						r := simple.NewSimpleRenderer(opt)
						gotStrings, err := r.ProcessSchema(gotSchema)
						if err != nil {
							t.Errorf("TEST_FAIL %s: %q err=%s", name, format, err)
							continue
						}

						util.CompareStrings(t, name, gotStrings, wantStrings)
					}
				}
			}

			format = "openapi"
			want = test.Want[format]
			if want != nil {
				for _, deref := range derefFlags {
					wantStrings := want[deref]
					if len(wantStrings) > 0 {
						name := testName(testGroup, test.Name, format, deref)
						opt.DeReference = deref
						opt.Indent = 0

						r := openapi.NewOpenAPIRenderer(openapi.NewMetaData(test.Name, "v1.0.0"), opt)
						gotStrings, err := r.ProcessSchema(gotSchema)

						if err != nil {
							t.Errorf("TEST_FAIL %s: %q err=%s", name, format, err)
							continue
						}

						if !util.CompareStrings(t, name, gotStrings, wantStrings) {
							continue
						}

						// Verify that YAML is valid.
						yamlStr := strings.Join(gotStrings, "\n")
						var yamlOut interface{}
						if err := yaml.Unmarshal([]byte(yamlStr), &yamlOut); err != nil {
							util.OutputErrStrings(t, name, gotStrings, fmt.Errorf("yaml err=%s", err))
							continue
						} else {
							t.Logf("TEST_OK %s: yaml", test.Name)
						}
					}
				}
			}
		}
	}
}
