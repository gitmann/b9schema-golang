package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"
	"testing"
	"time"
	"unsafe"

	"github.com/ghodss/yaml"
	"github.com/gitmann/b9schema-golang/common/util"
	"github.com/gitmann/b9schema-golang/fixtures"
	"github.com/gitmann/b9schema-golang/reflector"
	"github.com/gitmann/b9schema-golang/renderer"
	"github.com/gitmann/b9schema-golang/renderer/openapi"
	"github.com/gitmann/b9schema-golang/renderer/simple"
)

const (
	OPENAPI_CLI      = "swagger-cli"
	OPENAPI_CLI_FILE = "swagger-validate.yaml"
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

// Redefine types.
type MyBool bool
type MyInt int
type MyInt8 int8
type MyInt16 int16
type MyInt32 int32
type MyInt64 int64
type MyUint uint
type MyUint8 uint8
type MyUint16 uint16
type MyUint32 uint32
type MyUint64 uint64
type MyUintptr uintptr
type MyFloat32 float32
type MyFloat64 float64
type MyString string
type MyComplex64 complex64
type MyComplex128 complex128
type MyArray0 [0]string
type MyArray3 [3]string
type MyInterface interface{}
type MyMap map[int]int
type MyPtr *StringStruct
type MyPrivatePtr *PrivateStruct
type MySlice []interface{}
type MyStruct struct{}
type MyDateTime time.Time

type RedefineStruct struct {
	Bool       MyBool
	Int        MyInt
	Int8       MyInt8
	Int16      MyInt16
	Int32      MyInt32
	Int64      MyInt64
	Uint       MyUint
	Uint8      MyUint8
	Uint16     MyUint16
	Uint32     MyUint32
	Uint64     MyUint64
	Uintptr    MyUintptr
	Float32    MyFloat32
	Float64    MyFloat64
	String     MyString
	Complex64  MyComplex64
	Complex128 MyComplex128
	Array0     MyArray0
	Array3     MyArray3
	Interface  MyInterface
	Map        MyMap
	Ptr        MyPtr
	PrivatePtr MyPrivatePtr
	Slice      MySlice
	Struct     MyStruct
	DateTime   MyDateTime
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
					`  /03-type/boolean:`,
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
					`paths:`,
					`  /03-type/boolean:`,
					`    get:`,
					`      summary: Return data.`,
					`      responses:`,
					`        '200':`,
					`          description: Success`,
					`          content:`,
					`            application/json:`,
					`              schema:`,
					`                description: 'From $ref: #/components/schemas/BoolTypes'`,
					`                type: object`,
					`                additionalProperties: false`,
					`                properties:`,
					`                  Bool:`,
					`                    type: boolean`,
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
					`  /03-type/integer:`,
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
					`paths:`,
					`  /03-type/integer:`,
					`    get:`,
					`      summary: Return data.`,
					`      responses:`,
					`        '200':`,
					`          description: Success`,
					`          content:`,
					`            application/json:`,
					`              schema:`,
					`                description: 'From $ref: #/components/schemas/IntegerTypes'`,
					`                type: object`,
					`                additionalProperties: false`,
					`                properties:`,
					`                  Int:`,
					`                    type: integer`,
					`                  Int16:`,
					`                    type: integer`,
					`                  Int32:`,
					`                    type: integer`,
					`                  Int64:`,
					`                    type: integer`,
					`                    format: int64`,
					`                  Int8:`,
					`                    type: integer`,
					`                  Uint:`,
					`                    type: integer`,
					`                  Uint16:`,
					`                    type: integer`,
					`                  Uint32:`,
					`                    type: integer`,
					`                  Uint64:`,
					`                    type: integer`,
					`                    format: int64`,
					`                  Uint8:`,
					`                    type: integer`,
					`                  Uintptr:`,
					`                    type: integer`,
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
					`  /03-type/float:`,
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
					`paths:`,
					`  /03-type/float:`,
					`    get:`,
					`      summary: Return data.`,
					`      responses:`,
					`        '200':`,
					`          description: Success`,
					`          content:`,
					`            application/json:`,
					`              schema:`,
					`                description: 'From $ref: #/components/schemas/FloatTypes'`,
					`                type: object`,
					`                additionalProperties: false`,
					`                properties:`,
					`                  Float32:`,
					`                    type: number`,
					`                  Float64:`,
					`                    type: number`,
					`                    format: double`,
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
					`  /03-type/string:`,
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
					`paths:`,
					`  /03-type/string:`,
					`    get:`,
					`      summary: Return data.`,
					`      responses:`,
					`        '200':`,
					`          description: Success`,
					`          content:`,
					`            application/json:`,
					`              schema:`,
					`                description: 'From $ref: #/components/schemas/StringTypes'`,
					`                type: object`,
					`                additionalProperties: false`,
					`                properties:`,
					`                  String:`,
					`                    type: string`,
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
					`          description: 'ERROR=kind not supported;Kind=invalid:chan'`,
					`          type: string`,
					`        Complex128:`,
					`          description: 'ERROR=kind not supported;Kind=invalid:complex128'`,
					`          type: string`,
					`        Complex64:`,
					`          description: 'ERROR=kind not supported;Kind=invalid:complex64'`,
					`          type: string`,
					`        Func:`,
					`          description: 'ERROR=kind not supported;Kind=invalid:func'`,
					`          type: string`,
					`        UnsafePointer:`,
					`          description: 'ERROR=kind not supported;Kind=invalid:unsafe.Pointer'`,
					`          type: string`,
					`paths:`,
					`  /03-type/invalid:`,
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
					`paths:`,
					`  /03-type/invalid:`,
					`    get:`,
					`      summary: Return data.`,
					`      responses:`,
					`        '200':`,
					`          description: Success`,
					`          content:`,
					`            application/json:`,
					`              schema:`,
					`                description: 'From $ref: #/components/schemas/InvalidTypes'`,
					`                type: object`,
					`                additionalProperties: false`,
					`                properties:`,
					`                  Chan:`,
					`                    description: 'ERROR=kind not supported;Kind=invalid:chan'`,
					`                    type: string`,
					`                  Complex128:`,
					`                    description: 'ERROR=kind not supported;Kind=invalid:complex128'`,
					`                    type: string`,
					`                  Complex64:`,
					`                    description: 'ERROR=kind not supported;Kind=invalid:complex64'`,
					`                    type: string`,
					`                  Func:`,
					`                    description: 'ERROR=kind not supported;Kind=invalid:func'`,
					`                    type: string`,
					`                  UnsafePointer:`,
					`                    description: 'ERROR=kind not supported;Kind=invalid:unsafe.Pointer'`,
					`                    type: string`,
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
					`          description: 'ERROR=interface element is nil'`,
					`          type: string`,
					`        Map:`,
					`          description: 'ERROR=map key type must be string'`,
					`          type: object`,
					`          additionalProperties: false`,
					`        PrivatePtr:`,
					`          $ref: '#/components/schemas/PrivateStruct'`,
					`        Ptr:`,
					`          $ref: '#/components/schemas/StringStruct'`,
					`        Slice:`,
					`          type: array`,
					`          items:`,
					`            description: 'ERROR=interface element is nil'`,
					`            type: string`,
					`        Struct:`,
					`          description: 'ERROR=empty struct not supported'`,
					`          type: object`,
					`          additionalProperties: false`,
					`    PrivateStruct:`,
					`      description: 'ERROR=struct has no exported fields'`,
					`      type: object`,
					`      additionalProperties: false`,
					`    StringStruct:`,
					`      type: object`,
					`      additionalProperties: false`,
					`      properties:`,
					`        Value:`,
					`          type: string`,
					`paths:`,
					`  /03-type/compound:`,
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
					`paths:`,
					`  /03-type/compound:`,
					`    get:`,
					`      summary: Return data.`,
					`      responses:`,
					`        '200':`,
					`          description: Success`,
					`          content:`,
					`            application/json:`,
					`              schema:`,
					`                description: 'From $ref: #/components/schemas/CompoundTypes'`,
					`                type: object`,
					`                additionalProperties: false`,
					`                properties:`,
					`                  Array0:`,
					`                    type: array`,
					`                    items:`,
					`                      type: string`,
					`                  Array3:`,
					`                    type: array`,
					`                    items:`,
					`                      type: string`,
					`                  Interface:`,
					`                    description: 'ERROR=interface element is nil'`,
					`                    type: string`,
					`                  Map:`,
					`                    description: 'ERROR=map key type must be string'`,
					`                    type: object`,
					`                    additionalProperties: false`,
					`                  PrivatePtr:`,
					`                    description: 'From $ref: #/components/schemas/PrivateStruct;ERROR=struct has no exported fields'`,
					`                    type: object`,
					`                    additionalProperties: false`,
					`                  Ptr:`,
					`                    description: 'From $ref: #/components/schemas/StringStruct'`,
					`                    type: object`,
					`                    additionalProperties: false`,
					`                    properties:`,
					`                      Value:`,
					`                        type: string`,
					`                  Slice:`,
					`                    type: array`,
					`                    items:`,
					`                      description: 'ERROR=interface element is nil'`,
					`                      type: string`,
					`                  Struct:`,
					`                    description: 'ERROR=empty struct not supported'`,
					`                    type: object`,
					`                    additionalProperties: false`,
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
					`  /03-type/special:`,
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
					`paths:`,
					`  /03-type/special:`,
					`    get:`,
					`      summary: Return data.`,
					`      responses:`,
					`        '200':`,
					`          description: Success`,
					`          content:`,
					`            application/json:`,
					`              schema:`,
					`                description: 'From $ref: #/components/schemas/SpecialTypes'`,
					`                type: object`,
					`                additionalProperties: false`,
					`                properties:`,
					`                  DateTime:`,
					`                    type: string`,
					`                    format: date-time`,
				},
			},
		},
	},
	{
		Name:  "redefined",
		Value: RedefineStruct{},
		Want: map[string]fixtures.WantSet{
			"simple": map[bool][]string{
				false: []string{
					`TypeRef.MyArray0:[]`,
					`TypeRef.MyArray0:[].string`,
					`TypeRef.MyArray3:[]`,
					`TypeRef.MyArray3:[].string`,
					`TypeRef.MyBool:boolean`,
					`TypeRef.MyDateTime:datetime`,
					`TypeRef.MyFloat32:float`,
					`TypeRef.MyFloat64:float`,
					`TypeRef.MyInt:integer`,
					`TypeRef.MyInt16:integer`,
					`TypeRef.MyInt32:integer`,
					`TypeRef.MyInt64:integer`,
					`TypeRef.MyInt8:integer`,
					`TypeRef.!MyInterface:invalid! ERROR:interface element is nil`,
					`TypeRef.!MyMap:map{}! ERROR:map key type must be string`,
					`TypeRef.MySlice:[]`,
					`TypeRef.MySlice:[].!invalid! ERROR:interface element is nil`,
					`TypeRef.MyString:string`,
					`TypeRef.!MyStruct:{}! ERROR:empty struct not supported`,
					`TypeRef.MyUint:integer`,
					`TypeRef.MyUint16:integer`,
					`TypeRef.MyUint32:integer`,
					`TypeRef.MyUint64:integer`,
					`TypeRef.MyUint8:integer`,
					`TypeRef.MyUintptr:integer`,
					`TypeRef.!PrivateStruct:{}! ERROR:struct has no exported fields`,
					`TypeRef.RedefineStruct:{}`,
					`TypeRef.RedefineStruct:{}.Array0:[]:MyArray0`,
					`TypeRef.RedefineStruct:{}.Array3:[]:MyArray3`,
					`TypeRef.RedefineStruct:{}.Bool:boolean:MyBool`,
					`TypeRef.RedefineStruct:{}.!Complex128:invalid:complex128! ERROR:kind not supported`,
					`TypeRef.RedefineStruct:{}.!Complex64:invalid:complex64! ERROR:kind not supported`,
					`TypeRef.RedefineStruct:{}.DateTime:datetime:MyDateTime`,
					`TypeRef.RedefineStruct:{}.Float32:float:MyFloat32`,
					`TypeRef.RedefineStruct:{}.Float64:float:MyFloat64`,
					`TypeRef.RedefineStruct:{}.Int:integer:MyInt`,
					`TypeRef.RedefineStruct:{}.Int16:integer:MyInt16`,
					`TypeRef.RedefineStruct:{}.Int32:integer:MyInt32`,
					`TypeRef.RedefineStruct:{}.Int64:integer:MyInt64`,
					`TypeRef.RedefineStruct:{}.Int8:integer:MyInt8`,
					`TypeRef.RedefineStruct:{}.Interface:invalid:MyInterface`,
					`TypeRef.RedefineStruct:{}.Map:map{}:MyMap`,
					`TypeRef.RedefineStruct:{}.PrivatePtr:{}:PrivateStruct`,
					`TypeRef.RedefineStruct:{}.Ptr:{}:StringStruct`,
					`TypeRef.RedefineStruct:{}.Slice:[]:MySlice`,
					`TypeRef.RedefineStruct:{}.String:string:MyString`,
					`TypeRef.RedefineStruct:{}.Struct:{}:MyStruct`,
					`TypeRef.RedefineStruct:{}.Uint:integer:MyUint`,
					`TypeRef.RedefineStruct:{}.Uint16:integer:MyUint16`,
					`TypeRef.RedefineStruct:{}.Uint32:integer:MyUint32`,
					`TypeRef.RedefineStruct:{}.Uint64:integer:MyUint64`,
					`TypeRef.RedefineStruct:{}.Uint8:integer:MyUint8`,
					`TypeRef.RedefineStruct:{}.Uintptr:integer:MyUintptr`,
					`TypeRef.StringStruct:{}`,
					`TypeRef.StringStruct:{}.Value:string`,
					`Root.{}:RedefineStruct`,
				},
				true: []string{
					`Root.{}`,
					`Root.{}.Array0:[]`,
					`Root.{}.Array0:[].string`,
					`Root.{}.Array3:[]`,
					`Root.{}.Array3:[].string`,
					`Root.{}.Bool:boolean`,
					`Root.{}.!Complex128:invalid:complex128! ERROR:kind not supported`,
					`Root.{}.!Complex64:invalid:complex64! ERROR:kind not supported`,
					`Root.{}.DateTime:datetime`,
					`Root.{}.Float32:float`,
					`Root.{}.Float64:float`,
					`Root.{}.Int:integer`,
					`Root.{}.Int16:integer`,
					`Root.{}.Int32:integer`,
					`Root.{}.Int64:integer`,
					`Root.{}.Int8:integer`,
					`Root.{}.!Interface:invalid! ERROR:interface element is nil`,
					`Root.{}.!Map:map{}! ERROR:map key type must be string`,
					`Root.{}.!PrivatePtr:{}! ERROR:struct has no exported fields`,
					`Root.{}.Ptr:{}`,
					`Root.{}.Ptr:{}.Value:string`,
					`Root.{}.Slice:[]`,
					`Root.{}.Slice:[].!invalid! ERROR:interface element is nil`,
					`Root.{}.String:string`,
					`Root.{}.!Struct:{}! ERROR:empty struct not supported`,
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
					`  title: redefined`,
					`  version: v1.0.0`,
					`openapi: 3.0.0`,
					``,
					`components:`,
					`  schemas:`,
					`    MyArray0:`,
					`      type: array`,
					`      items:`,
					`        type: string`,
					`    MyArray3:`,
					`      type: array`,
					`      items:`,
					`        type: string`,
					`    MyBool:`,
					`      type: boolean`,
					`    MyDateTime:`,
					`      type: string`,
					`      format: date-time`,
					`    MyFloat32:`,
					`      type: number`,
					`    MyFloat64:`,
					`      type: number`,
					`      format: double`,
					`    MyInt:`,
					`      type: integer`,
					`    MyInt16:`,
					`      type: integer`,
					`    MyInt32:`,
					`      type: integer`,
					`    MyInt64:`,
					`      type: integer`,
					`      format: int64`,
					`    MyInt8:`,
					`      type: integer`,
					`    MyInterface:`,
					`      description: 'ERROR=interface element is nil'`,
					`      type: string`,
					`    MyMap:`,
					`      description: 'ERROR=map key type must be string'`,
					`      type: object`,
					`      additionalProperties: false`,
					`    MySlice:`,
					`      type: array`,
					`      items:`,
					`        description: 'ERROR=interface element is nil'`,
					`        type: string`,
					`    MyString:`,
					`      type: string`,
					`    MyStruct:`,
					`      description: 'ERROR=empty struct not supported'`,
					`      type: object`,
					`      additionalProperties: false`,
					`    MyUint:`,
					`      type: integer`,
					`    MyUint16:`,
					`      type: integer`,
					`    MyUint32:`,
					`      type: integer`,
					`    MyUint64:`,
					`      type: integer`,
					`      format: int64`,
					`    MyUint8:`,
					`      type: integer`,
					`    MyUintptr:`,
					`      type: integer`,
					`    PrivateStruct:`,
					`      description: 'ERROR=struct has no exported fields'`,
					`      type: object`,
					`      additionalProperties: false`,
					`    RedefineStruct:`,
					`      type: object`,
					`      additionalProperties: false`,
					`      properties:`,
					`        Array0:`,
					`          $ref: '#/components/schemas/MyArray0'`,
					`        Array3:`,
					`          $ref: '#/components/schemas/MyArray3'`,
					`        Bool:`,
					`          $ref: '#/components/schemas/MyBool'`,
					`        Complex128:`,
					`          description: 'ERROR=kind not supported;Kind=invalid:complex128'`,
					`          type: string`,
					`        Complex64:`,
					`          description: 'ERROR=kind not supported;Kind=invalid:complex64'`,
					`          type: string`,
					`        DateTime:`,
					`          $ref: '#/components/schemas/MyDateTime'`,
					`        Float32:`,
					`          $ref: '#/components/schemas/MyFloat32'`,
					`        Float64:`,
					`          $ref: '#/components/schemas/MyFloat64'`,
					`        Int:`,
					`          $ref: '#/components/schemas/MyInt'`,
					`        Int16:`,
					`          $ref: '#/components/schemas/MyInt16'`,
					`        Int32:`,
					`          $ref: '#/components/schemas/MyInt32'`,
					`        Int64:`,
					`          $ref: '#/components/schemas/MyInt64'`,
					`        Int8:`,
					`          $ref: '#/components/schemas/MyInt8'`,
					`        Interface:`,
					`          $ref: '#/components/schemas/MyInterface'`,
					`        Map:`,
					`          $ref: '#/components/schemas/MyMap'`,
					`        PrivatePtr:`,
					`          $ref: '#/components/schemas/PrivateStruct'`,
					`        Ptr:`,
					`          $ref: '#/components/schemas/StringStruct'`,
					`        Slice:`,
					`          $ref: '#/components/schemas/MySlice'`,
					`        String:`,
					`          $ref: '#/components/schemas/MyString'`,
					`        Struct:`,
					`          $ref: '#/components/schemas/MyStruct'`,
					`        Uint:`,
					`          $ref: '#/components/schemas/MyUint'`,
					`        Uint16:`,
					`          $ref: '#/components/schemas/MyUint16'`,
					`        Uint32:`,
					`          $ref: '#/components/schemas/MyUint32'`,
					`        Uint64:`,
					`          $ref: '#/components/schemas/MyUint64'`,
					`        Uint8:`,
					`          $ref: '#/components/schemas/MyUint8'`,
					`        Uintptr:`,
					`          $ref: '#/components/schemas/MyUintptr'`,
					`    StringStruct:`,
					`      type: object`,
					`      additionalProperties: false`,
					`      properties:`,
					`        Value:`,
					`          type: string`,
					`paths:`,
					`  /03-type/redefined:`,
					`    get:`,
					`      summary: Return data.`,
					`      responses:`,
					`        '200':`,
					`          description: Success`,
					`          content:`,
					`            application/json:`,
					`              schema:`,
					`                $ref: '#/components/schemas/RedefineStruct'`,
				},
				true: []string{
					`info:`,
					`  title: redefined`,
					`  version: v1.0.0`,
					`openapi: 3.0.0`,
					``,
					`paths:`,
					`  /03-type/redefined:`,
					`    get:`,
					`      summary: Return data.`,
					`      responses:`,
					`        '200':`,
					`          description: Success`,
					`          content:`,
					`            application/json:`,
					`              schema:`,
					`                description: 'From $ref: #/components/schemas/RedefineStruct'`,
					`                type: object`,
					`                additionalProperties: false`,
					`                properties:`,
					`                  Array0:`,
					`                    description: 'From $ref: #/components/schemas/MyArray0'`,
					`                    type: array`,
					`                    items:`,
					`                      type: string`,
					`                  Array3:`,
					`                    description: 'From $ref: #/components/schemas/MyArray3'`,
					`                    type: array`,
					`                    items:`,
					`                      type: string`,
					`                  Bool:`,
					`                    description: 'From $ref: #/components/schemas/MyBool'`,
					`                    type: boolean`,
					`                  Complex128:`,
					`                    description: 'ERROR=kind not supported;Kind=invalid:complex128'`,
					`                    type: string`,
					`                  Complex64:`,
					`                    description: 'ERROR=kind not supported;Kind=invalid:complex64'`,
					`                    type: string`,
					`                  DateTime:`,
					`                    description: 'From $ref: #/components/schemas/MyDateTime'`,
					`                    type: string`,
					`                    format: date-time`,
					`                  Float32:`,
					`                    description: 'From $ref: #/components/schemas/MyFloat32'`,
					`                    type: number`,
					`                  Float64:`,
					`                    description: 'From $ref: #/components/schemas/MyFloat64'`,
					`                    type: number`,
					`                    format: double`,
					`                  Int:`,
					`                    description: 'From $ref: #/components/schemas/MyInt'`,
					`                    type: integer`,
					`                  Int16:`,
					`                    description: 'From $ref: #/components/schemas/MyInt16'`,
					`                    type: integer`,
					`                  Int32:`,
					`                    description: 'From $ref: #/components/schemas/MyInt32'`,
					`                    type: integer`,
					`                  Int64:`,
					`                    description: 'From $ref: #/components/schemas/MyInt64'`,
					`                    type: integer`,
					`                    format: int64`,
					`                  Int8:`,
					`                    description: 'From $ref: #/components/schemas/MyInt8'`,
					`                    type: integer`,
					`                  Interface:`,
					`                    description: 'From $ref: #/components/schemas/MyInterface;ERROR=interface element is nil'`,
					`                    type: string`,
					`                  Map:`,
					`                    description: 'From $ref: #/components/schemas/MyMap;ERROR=map key type must be string'`,
					`                    type: object`,
					`                    additionalProperties: false`,
					`                  PrivatePtr:`,
					`                    description: 'From $ref: #/components/schemas/PrivateStruct;ERROR=struct has no exported fields'`,
					`                    type: object`,
					`                    additionalProperties: false`,
					`                  Ptr:`,
					`                    description: 'From $ref: #/components/schemas/StringStruct'`,
					`                    type: object`,
					`                    additionalProperties: false`,
					`                    properties:`,
					`                      Value:`,
					`                        type: string`,
					`                  Slice:`,
					`                    description: 'From $ref: #/components/schemas/MySlice'`,
					`                    type: array`,
					`                    items:`,
					`                      description: 'ERROR=interface element is nil'`,
					`                      type: string`,
					`                  String:`,
					`                    description: 'From $ref: #/components/schemas/MyString'`,
					`                    type: string`,
					`                  Struct:`,
					`                    description: 'From $ref: #/components/schemas/MyStruct;ERROR=empty struct not supported'`,
					`                    type: object`,
					`                    additionalProperties: false`,
					`                  Uint:`,
					`                    description: 'From $ref: #/components/schemas/MyUint'`,
					`                    type: integer`,
					`                  Uint16:`,
					`                    description: 'From $ref: #/components/schemas/MyUint16'`,
					`                    type: integer`,
					`                  Uint32:`,
					`                    description: 'From $ref: #/components/schemas/MyUint32'`,
					`                    type: integer`,
					`                  Uint64:`,
					`                    description: 'From $ref: #/components/schemas/MyUint64'`,
					`                    type: integer`,
					`                    format: int64`,
					`                  Uint8:`,
					`                    description: 'From $ref: #/components/schemas/MyUint8'`,
					`                    type: integer`,
					`                  Uintptr:`,
					`                    description: 'From $ref: #/components/schemas/MyUintptr'`,
					`                    type: integer`,
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
					`  /04-list/arrays:`,
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
					`paths:`,
					`  /04-list/arrays:`,
					`    get:`,
					`      summary: Return data.`,
					`      responses:`,
					`        '200':`,
					`          description: Success`,
					`          content:`,
					`            application/json:`,
					`              schema:`,
					`                description: 'From $ref: #/components/schemas/ArrayStruct'`,
					`                type: object`,
					`                additionalProperties: false`,
					`                properties:`,
					`                  Array0:`,
					`                    type: array`,
					`                    items:`,
					`                      type: string`,
					`                  Array2_3:`,
					`                    type: array`,
					`                    items:`,
					`                      type: array`,
					`                      items:`,
					`                        type: string`,
					`                  Array3:`,
					`                    type: array`,
					`                    items:`,
					`                      type: string`,
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
					`  /04-list/json-array:`,
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
					`                      description: 'ERROR=interface element is nil'`,
					`                      type: string`,
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
					`  /04-list/json-array:`,
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
					`                      description: 'ERROR=interface element is nil'`,
					`                      type: string`,
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
					`  /04-list/slices:`,
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
					`paths:`,
					`  /04-list/slices:`,
					`    get:`,
					`      summary: Return data.`,
					`      responses:`,
					`        '200':`,
					`          description: Success`,
					`          content:`,
					`            application/json:`,
					`              schema:`,
					`                description: 'From $ref: #/components/schemas/SliceStruct'`,
					`                type: object`,
					`                additionalProperties: false`,
					`                properties:`,
					`                  Array2:`,
					`                    type: array`,
					`                    items:`,
					`                      type: array`,
					`                      items:`,
					`                        type: string`,
					`                  Slice:`,
					`                    type: array`,
					`                    items:`,
					`                      type: string`,
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
						} else if b, err := json.Marshal(yamlOut); err != nil {
							t.Errorf("TEST_FAIL %s: json err=%s", name, err)
						} else {
							_ = b
							//t.Logf("TEST_OK %s: yaml/json=%s", test.Name, string(b))
							//t.Logf("TEST_OK %s: yaml/json len=%d", test.Name, len(b))
						}

						// Verify that OpenAPI is valid.
						if !validateOpenAPI(t, name, yamlStr) {
							continue
						}

						t.Logf("TEST_OK %s", name)
					}
				}
			}
		}
	}
}

func validateOpenAPI(t *testing.T, name, yamlStr string) bool {
	if err := os.WriteFile(OPENAPI_CLI_FILE, []byte(yamlStr), 0644); err != nil {
		t.Errorf("TEST_FAIL %s: writing yaml file err=%s", name, err)
		return false
	}

	cmd := exec.Command(OPENAPI_CLI, "validate", OPENAPI_CLI_FILE)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Errorf("TEST_FAIL %s: stdout pipe err=%s", name, err)
		return false
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		t.Errorf("TEST_FAIL %s: stderr pipe err=%s", name, err)
		return false
	}

	if err := cmd.Start(); err != nil {
		t.Errorf("TEST_FAIL %s: cmd start err=%s\nIs swagger-cli installed? 'npm install -g @apidevtools/swagger-cli'", name, err)
		return false
	}

	cmdOutput, err := io.ReadAll(stdout)
	if err != nil {
		t.Errorf("TEST_FAIL %s: stdout read err=%s", name, err)
	}

	cmdErr, err := io.ReadAll(stderr)
	if err != nil {
		t.Errorf("TEST_FAIL %s: stderr read err=%s", name, err)
	}

	if err := cmd.Wait(); err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			t.Errorf("TEST_FAIL %s: cmd wait err=%s", name, err)
			return false
		} else {
			outLines := []string{}
			if len(cmdOutput) > 0 {
				outLines = append(outLines,
					"STDOUT:",
					string(cmdOutput),
				)
			}
			if len(cmdErr) > 0 {
				outLines = append(outLines,
					"STDERR:",
					string(cmdErr),
				)
			}

			util.OutputErrStrings(t, name, []string{yamlStr}, fmt.Errorf("openapi validation\n%s", strings.Join(outLines, "\n")))
			return false
		}
	}
	//t.Errorf("TEST_OK %s: openapi validation", name)
	return true
}
