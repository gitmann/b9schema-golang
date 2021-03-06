package generictype

import (
	"fmt"
	"github.com/gitmann/b9schema-golang/common/enum/typecategory"
	"reflect"
	"time"
)

// GenericType defines generic types for shiny schemas.
// Uses slugs from: https://threedots.tech/post/safer-enums-in-go/
type GenericType struct {
	slug        string
	pathDefault string
	cat         typecategory.TypeCategory
	kinds       []string
}

// String returns GenericType as a string.
func (t *GenericType) String() string {
	return t.slug
}

// Category returns the TypeCategory for the GenericType.
func (t *GenericType) Category() typecategory.TypeCategory {
	return t.cat
}

// PathDefault returns the default path string for the GenericType.
func (t *GenericType) PathDefault() string {
	if t.pathDefault != "" {
		return t.pathDefault
	}
	return t.slug
}

// ContainsKind returns true if the generic type contains the kind string.
func (t *GenericType) ContainsKind(kind string) bool {
	for _, k := range t.kinds {
		if kind == k {
			return true
		}
	}
	return false
}

// Invalid types are not allowed in shiny schemas.
var Invalid = &GenericType{
	slug: "invalid",
	cat:  typecategory.Invalid,
	kinds: []string{
		reflect.Invalid.String(),
		reflect.Complex64.String(),
		reflect.Complex128.String(),
		reflect.Chan.String(),
		reflect.Func.String(),
		reflect.UnsafePointer.String(),
	},
}

// Basic types.
var Boolean = &GenericType{
	slug: "boolean",
	cat:  typecategory.Basic,
	kinds: []string{
		reflect.Bool.String(),
	},
}

var Integer = &GenericType{
	slug: "integer",
	cat:  typecategory.Basic,
	kinds: []string{
		reflect.Int.String(),
		reflect.Int8.String(),
		reflect.Int16.String(),
		reflect.Int32.String(),
		reflect.Int64.String(),
		reflect.Uint.String(),
		reflect.Uint8.String(),
		reflect.Uint16.String(),
		reflect.Uint32.String(),
		reflect.Uint64.String(),
		reflect.Uintptr.String(),
	},
}

var Float = &GenericType{
	slug: "float",
	cat:  typecategory.Basic,
	kinds: []string{
		reflect.Float32.String(),
		reflect.Float64.String(),
	},
}

var String = &GenericType{
	slug: "string",
	cat:  typecategory.Basic,
	kinds: []string{
		reflect.String.String(),
	},
}

// Compound types.
var List = &GenericType{
	slug:        "list",
	pathDefault: "[]",
	cat:         typecategory.Compound,
	kinds: []string{
		reflect.Array.String(),
		reflect.Slice.String(),
	},
}

var Struct = &GenericType{
	slug:        "struct",
	pathDefault: "{}",
	cat:         typecategory.Compound,
	kinds: []string{
		reflect.Struct.String(),
	},
}

var Map = &GenericType{
	slug:        "map",
	pathDefault: "map{}",
	cat:         typecategory.Compound,
	kinds: []string{
		reflect.Map.String(),
	},
}

// Known types map Go standard types to b9schema types.
// - kinds is a list of "PkgPath.Type"
// These are a subset of protobuf well-known types:
// https://developers.google.com/protocol-buffers/docs/reference/google.protobuf

var DateTime = &GenericType{
	slug: "datetime",
	cat:  typecategory.Known,
	kinds: []string{
		"time.Time",
	},
}

// Reference types.
var Interface = &GenericType{
	slug:        "interface",
	pathDefault: "{?}",
	cat:         typecategory.Reference,
	kinds: []string{
		reflect.Interface.String(),
	},
}

var Pointer = &GenericType{
	slug:        "pointer",
	pathDefault: "*",
	cat:         typecategory.Reference,
	kinds: []string{
		reflect.Ptr.String(),
	},
}

// Internal types.
// These have no meaning outside of a b9schema.

// Root is at the top of any type tree.
var Root = &GenericType{
	slug:        "root",
	pathDefault: "$",
	cat:         typecategory.Internal,
	kinds:       []string{},
}

// lookupByKind provides lookups from reflect.Kind.String to GenericType.
var lookupByKind map[string]*GenericType

// lookupByType provides lookups from generic type string to GenericType.
var lookupByType map[string]*GenericType

// init() initializes the lookupByKind map.
func init() {
	lookupByKind = map[string]*GenericType{}
	lookupByType = map[string]*GenericType{}

	// mapTypes is a utility function to create map entries for the given GenericType.
	mapTypes := func(t *GenericType) {
		for _, k := range t.kinds {
			// Panic if duplicate type mappings exist.
			if lookupByKind[k] != nil {
				panic(fmt.Sprintf("duplicate lookupByKind mapping for %q", k))
			}
			lookupByKind[k] = t
		}

		if lookupByType[t.String()] != nil {
			panic(fmt.Sprintf("duplicate lookupByType mapping for %q", t.String()))
		}
		lookupByType[t.String()] = t
	}

	mapTypes(Invalid)

	mapTypes(Boolean)
	mapTypes(Integer)
	mapTypes(Float)
	mapTypes(String)

	mapTypes(List)
	mapTypes(Struct)
	mapTypes(Map)

	mapTypes(DateTime)

	mapTypes(Interface)
	mapTypes(Pointer)

	mapTypes(Root)
}

// GenericTypeOf returns the GenericType of the given reflect.Value.
func GenericTypeOf(v reflect.Value) *GenericType {
	if t := lookupByKind[v.Kind().String()]; t != nil {
		if t == Invalid {
			// Return invalid types immediately.
			return t
		}

		// Look for special types.
		if v.Type().PkgPath() != "" {
			fullPath := FullPathOf(v)
			if specialType := lookupByKind[fullPath]; specialType != nil {
				return specialType
			}
		}

		// Check for type definitions for special types.
		if t == Struct {
			if _, ok := tryConversion(v, reflect.TypeOf(time.Time{})); ok {
				return DateTime
			}
		}

		// Not a special type.
		return t
	}

	return Invalid
}

// FromType returns the GenericType associated with a given string or nil if not found.
func FromType(typeString string) *GenericType {
	return lookupByType[typeString]
}

// PathDefaultOfType returns the path default for a given generic type string.
func PathDefaultOfType(typeString string) string {
	if gt := FromType(typeString); gt != nil {
		return gt.PathDefault()
	}
	return typeString
}

// FullPathOf returns the full package path for a Value.
func FullPathOf(v reflect.Value) string {
	return fmt.Sprintf("%s.%s", v.Type().PkgPath(), v.Type().Name())
}

// tryConversion attempts to convert a Value to the given Type.
// - 2nd return value is true if conversion succeeds, false otherwise.
func tryConversion(v reflect.Value, t reflect.Type) (newValue reflect.Value, ok bool) {
	defer func() {
		if r := recover(); r != nil {
			// recover from panic
			ok = false
		}
	}()

	return v.Convert(t), true
}
