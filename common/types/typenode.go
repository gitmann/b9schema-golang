package types

import (
	"fmt"
	"github.com/gitmann/b9schema-golang/common/enum/generictype"
	"github.com/gitmann/b9schema-golang/common/enum/threeflag"
	"reflect"
	"sort"
	"strings"
	"unicode"
)

// NodePool creates and holds all TypeNodes.
type NodePool struct {
	// Nodes is a map of all TypeNodes by ID.
	Nodes map[string]*TypeNode

	// LastID holds the last assigned ID number. IDs are assigned sequentially.
	LastID int
}

// NewNodePool initializes a new NodePool.
func NewNodePool() *NodePool {
	return &NodePool{
		Nodes:  map[string]*TypeNode{},
		LastID: 0,
	}
}

// nextID returns the next TypeNode ID.
func (np *NodePool) nextID() string {
	np.LastID += 1
	return fmt.Sprintf("n%05d", np.LastID)
}

// NewTypeNode returns a new TypeNode in the current NodePool.
func (np *NodePool) NewTypeNode(name, dialect string) *TypeNode {
	t := &TypeNode{
		ID: np.nextID(),

		Children: []string{},

		Name: name,

		NativeDialect: dialect,
		Native:        map[string]*NativeType{},

		pool: np,
	}
	np.Nodes[t.ID] = t

	if t.NativeDialect != "" {
		t.Native[t.NativeDialect] = NewNativeType(t.NativeDialect)
	}

	return t
}

// NewRootNode creates a new type element that is a root of a tree.
func (np *NodePool) NewRootNode(name, dialect string) *TypeNode {
	r := np.NewTypeNode(name, dialect)

	r.Type = generictype.Root.String()
	r.TypeCategory = generictype.Root.Category().String()

	return r
}

// TypeNode holds type information about an element.
// - TypeNode should be cross-platform and only use basic types.
type TypeNode struct {
	// Unique identifier for an element.
	ID string

	// Optional Name and Description of element.
	// - Name applies to struct/map types with string keys.
	Name        string `json:",omitempty"`
	Description string `json:",omitempty"`

	// Nullable indicates that a field should accept null in addition to values.
	Nullable bool `json:",omitempty"`

	// Generic type of element.
	TypeCategory string
	Type         string

	// TypeRef holds the name of a type (e.g. struct)
	TypeRef string `json:",omitempty"`

	// NativeDialect is the name of the dialect that was the source for the schema.
	NativeDialect string `json:",omitempty"`

	// Native type features by dialect name.
	Native map[string]*NativeType `json:"-"`

	// Capture error if element cannot reflect.
	Error string `json:",omitempty"`

	// Pointers to Parent and Child ID strings.
	Parent   string   `json:",omitempty"`
	Children []string `json:",omitempty"`

	// Keep a pointer to the master NodePool in all Nodes.
	pool *NodePool
}

// Node is a utility wrapper that returns the TypeNode associated with the given ID.
func (t *TypeNode) Node(id string) *TypeNode {
	if id == "" {
		// Default to the current node if no ID.
		id = t.ID
	}
	return t.pool.Nodes[id]
}

// NewChild creates a new type element that is a child of the current one.
func (t *TypeNode) NewChild(name string) *TypeNode {
	childElem := t.pool.NewTypeNode(name, t.NativeDialect)
	t.AddChild(childElem)

	return childElem
}

// AddChild adds a child element to the current element.
// - Sets Parent on the child element.
func (t *TypeNode) AddChild(childElem *TypeNode) {
	// Ignore nil.
	if childElem == nil {
		return
	}

	childElem.SetParent(t)

	t.Children = append(t.Children, childElem.ID)
}

// ChildMap returns a map of Children name --> *TypeNode
// - Output map can be passed to ChildKeys, ContainsChild, ChildByName for reuse.
func (t *TypeNode) ChildMap() map[string]*TypeNode {
	out := map[string]*TypeNode{}
	for _, childID := range t.Children {
		childNode := t.pool.Nodes[childID]
		out[childNode.Name] = childNode
	}
	return out
}

// ChildKeys returns a sorted list of child names.
func (t *TypeNode) ChildKeys(m map[string]*TypeNode) []string {
	if len(m) == 0 {
		m = t.ChildMap()
	}

	out := make([]string, len(m))
	if len(m) > 0 {
		i := 0
		for k := range m {
			out[i] = k
			i++
		}

		sort.Strings(out)
	}

	return out
}

// ContainsChild returns true if a child with the given name exist.
func (t *TypeNode) ContainsChild(name string, m map[string]*TypeNode) bool {
	c := t.ChildByName(name, m)
	return c != nil
}

// ChildByName gets the child with the given element name.
// - Returns nil if child does not exist.
func (t *TypeNode) ChildByName(name string, m map[string]*TypeNode) *TypeNode {
	if len(m) == 0 {
		m = t.ChildMap()
	}
	return m[name]
}

// RemoveAllChildren removes all children from the current element.
func (t *TypeNode) RemoveAllChildren() {
	for _, childID := range t.Children {
		childNode := t.pool.Nodes[childID]
		childNode.RemoveParent()
	}

	t.Children = []string{}
}

// RemoveChild removes the given child from the Children list.
// - Uses ID for matching.
// - Sets Parent on child to nil.
func (t *TypeNode) RemoveChild(childElem *TypeNode) {
	if childElem == nil {
		return
	}

	// Copy all children except the given one.
	newChildren := []string{}
	for _, childID := range t.Children {
		if childID != childElem.ID {
			newChildren = append(newChildren, childID)
		} else {
			// Remove parent from child element.
			childElem.Parent = ""
		}
	}

	t.Children = newChildren
}

// Copy makes a copy of a TypeNode and its Children.
// - The copied element has no Parent.
func (t *TypeNode) Copy() *TypeNode {
	n := t.pool.NewTypeNode(t.Name, t.NativeDialect)

	// Copy simple fields.
	n.Parent = ""
	n.Description = t.Description
	n.Type = t.Type
	n.TypeCategory = t.TypeCategory
	n.TypeRef = t.TypeRef
	n.Error = t.Error

	// Copy Children with new element as parent.
	for _, childID := range t.Children {
		childElem := t.pool.Nodes[childID]
		newChild := childElem.Copy()
		n.AddChild(newChild)
	}

	// Copy Dialects.
	for dialect, native := range t.Native {
		n.Native[dialect] = native.Copy()
	}

	return n
}

// GetNativeType returns a new NativeType with Name,Type,TypeRef,Include set.
func (t *TypeNode) GetNativeType(dialect string) *NativeType {
	// Start with a new native type that is a clone of the current type element.
	newType := NewNativeType(dialect)
	newType.Name = t.Name
	newType.Type = t.Type
	newType.TypeRef = t.NativeDefault().TypeRef
	newType.Include = threeflag.Undefined

	// Check if a native type exists for the dialect.
	oldType := t.Native[dialect]
	if oldType != nil {
		// Replace with values from oldType if set.
		if oldType.Name != "" {
			newType.Name = oldType.Name
		}
		if oldType.Type != "" {
			newType.Type = oldType.Type
		}
		if oldType.TypeRef != "" {
			newType.TypeRef = oldType.TypeRef
		}
		if oldType.Include != threeflag.Undefined {
			newType.Include = oldType.Include
		}
	}

	return newType
}

// RemoveParent removes the Parent.
func (t *TypeNode) RemoveParent() {
	if t.Parent != "" {
		parentElem := t.pool.Nodes[t.Parent]
		parentElem.RemoveChild(t)
	}
}

// SetParent set the Parent and ParentID.
func (t *TypeNode) SetParent(p *TypeNode) {
	if p == nil {
		panic("parent cannot be nil")
	}

	if t.ID == p.ID {
		panic("element cannot be its own parent")
	}

	t.RemoveParent()

	t.Parent = p.ID
}

// GetName returns the alias for the given lang or Name.
func (t *TypeNode) GetName(lang string) string {
	if t.Native != nil {
		if t.Native[lang] != nil {
			if a := t.Native[lang].Name; a != "" {
				return a
			}
		}
	}
	return t.Name
}

// SetName sets the GetName for the native dialect.
func (t *TypeNode) SetName(dialect, alias string) {
	if t.Native == nil {
		t.Native = make(map[string]*NativeType)
	}
	if a := t.Native[dialect]; a == nil {
		t.Native[dialect] = NewNativeType(dialect)
	}
	t.Native[dialect].Name = alias
}

// NativeDefault returns the native element for the NativeDialect.
func (t *TypeNode) NativeDefault() *NativeType {
	return t.Native[t.NativeDialect]
}

// IsBasicType returns true if the element is a basic type.
func (t *TypeNode) IsBasicType() bool {
	switch t.Type {
	case "string", "integer", "float", "boolean":
		return true
	}
	return false
}

// IsExported returns true if the element Name starts with an uppercase letter.
func (t *TypeNode) IsExported() bool {
	if t.Name == "" {
		return false
	}

	r := []rune(t.Name)
	return unicode.IsUpper(r[0])
}

// Ancestors returns a slice of all ancestors of the given TypeNode.
func (t *TypeNode) Ancestors() []*TypeNode {
	if t.Parent == "" {
		// RootID element. Start a new path.
		return []*TypeNode{t}
	}

	parentNode := t.pool.Nodes[t.Parent]
	return append(parentNode.Ancestors(), t)
}

// PathList keeps a list of path string elements that form a unique identifier for a TypeNode.
// - PathList behaves like a stack with Push/Pop operators.
type PathList struct {
	paths []string
}

func NewPathList() *PathList {
	return &PathList{paths: make([]string, 0)}
}

func (p *PathList) Len() int {
	return len(p.paths)
}

func (p *PathList) Push(elem string) {
	// Ignore empty elements.
	if elem == "" {
		return
	}
	p.paths = append(p.paths, elem)
}

func (p *PathList) Pop() string {
	if len(p.paths) == 0 {
		return ""
	}

	elem := p.paths[len(p.paths)-1]
	p.paths = p.paths[:len(p.paths)-1]
	return elem
}

func (p *PathList) Copy() *PathList {
	n := &PathList{paths: make([]string, len(p.paths))}
	copy(n.paths, p.paths)
	return n
}

func (p *PathList) String() string {
	out := make([]string, len(p.paths))
	for i, s := range p.paths {
		if strings.Contains(s, ".") {
			s = fmt.Sprintf("%q", s)
		}
		out[i] = s
	}
	return strings.Join(out, ".")
}

// NativeOption stores options as key-value pairs but returns a list of strings.
// - Value-only entries are unique by value.
// - Values with keys are unique by key.
type NativeOption struct {
	OptionMap map[string]string
}

func NewNativeOption() *NativeOption {
	return &NativeOption{
		OptionMap: make(map[string]string),
	}
}

// Equals returns true if both NativeOption struct have the same values.
func (n *NativeOption) Equals(other *NativeOption) bool {
	if n == nil && other == nil {
		// Both are nil so equal.
		return true
	}

	// Treat nil option map as zero-length.
	var thisLen, otherLen int
	if n != nil && n.OptionMap != nil {
		thisLen = len(n.OptionMap)
	}
	if other != nil && other.OptionMap != nil {
		otherLen = len(other.OptionMap)
	}

	if thisLen != otherLen {
		return false
	} else if thisLen == 0 {
		return true
	}

	return reflect.DeepEqual(n.OptionMap, n.OptionMap)
}

// AsList returns options as a slice of strings.
func (n *NativeOption) AsList() []string {
	// Return empty slice if no options are set.
	if len(n.OptionMap) == 0 {
		return make([]string, 0)
	}

	s := make([]string, len(n.OptionMap))
	i := 0
	for k, v := range n.OptionMap {
		if v == "" {
			//	Value only
			s[i] = k
		} else {
			// Key-Value pair
			s[i] = fmt.Sprintf("%s=%s", k, v)
		}
		i++
	}

	// Sort slice for output.
	sort.Strings(s)
	return s
}

// AddVal adds an option value string.
func (n *NativeOption) AddVal(val string) {
	// Ignore if value is empty.
	if val == "" {
		return
	}
	n.OptionMap[val] = ""
}

// Delete removes an entry from the option map.
// - key will match either key-value pairs or value-only settings.
func (n *NativeOption) Delete(key string) {
	// Ignore empty key.
	if key == "" {
		return
	}
	delete(n.OptionMap, key)
}

// AddKeyVal adds an option string key=val
func (n *NativeOption) AddKeyVal(key, val string) {
	// Ignore if key is empty.
	if key == "" {
		return
	}

	// If value is empty, delete key.
	if val == "" {
		n.Delete(key)
		return
	}

	// Set value.
	n.OptionMap[key] = val
}

// AddBool adds a boolean as an option string.
// - key is required
//   - if key is empty, nothing is added
// - val is boolean value
func (n *NativeOption) AddBool(key string, val bool) {
	// Ignore if key is missing.
	if key == "" {
		return
	}

	n.OptionMap[key] = fmt.Sprintf("%t", val)
}

// AddThreeFlag adds a ThreeFlag value as a string.
// - key is required
//   - if key is empty, nothing is added
// - val is ThreeFlag value
func (n *NativeOption) AddThreeFlag(key string, val threeflag.ThreeFlag) {
	// Ignore if key is missing.
	if key == "" {
		return
	}

	n.OptionMap[key] = val.String()
}

// UpdateFrom updates with values from another NativeOption.
func (n *NativeOption) UpdateFrom(other *NativeOption) {
	for k, v := range other.OptionMap {
		n.OptionMap[k] = v
	}
}

// Copy makes a copy of the NativeOption.
func (n *NativeOption) Copy() *NativeOption {
	c := NewNativeOption()

	for k, v := range n.OptionMap {
		c.OptionMap[k] = v
	}

	return c
}

// NativeType holds key-value attributes specific to one dialect.
// - A dialect is the name of a language (e.g. golang) or implementation (e.g. json-schema)
type NativeType struct {
	// Name of language of dialect represented by NativeType.
	Dialect string

	// Name of element if different from generic Name.
	Name string

	// Native type of element if different from the generic Type.
	Type string

	// TypeRef holds the native name of a type if different from the generic TypeRef.
	TypeRef string

	// Include indicates whether an element should be included in output for a dialect.
	// Include has three value values:
	// - "" (empty string) means value is not set
	// - "yes" = include element in output
	// - "no" = exclude element from output
	Include threeflag.ThreeFlag

	// Options contains a list of strings representing dialect-specific options.
	// - Format is one of:
	//   - "value"
	//   - "key=value"
	Options *NativeOption

	// Capture error if element cannot reflect.
	Error string
}

// NewNativeType initializes a new NativeType with default settings.
func NewNativeType(dialect string) *NativeType {
	n := &NativeType{
		// Default to the native dialect.
		Dialect: dialect,

		// Include fields by default.
		Include: threeflag.True,

		// Empty options list.
		Options: NewNativeOption(),
	}

	return n
}

// UpdateFromTag sets NativeType fields from a StructFieldTag.
func (n *NativeType) UpdateFromTag(t *StructFieldTag) {
	if t == nil {
		return
	}

	if t.Ignore {
		n.Include = threeflag.False
	}

	if t.Alias != "" {
		n.Name = t.Alias
	}

	n.Options.UpdateFrom(t.Options)
}

// AsMap returns a map[string]string representation of the NativeType struct.
func (n *NativeType) AsMap() map[string]string {
	m := map[string]string{}

	if n.Include != threeflag.Undefined {
		m["Include"] = n.Include.String()
	}
	if n.Name != "" {
		m["Name"] = n.Name
	}
	if n.Type != "" {
		m["Type"] = n.Type
	}
	if n.TypeRef != "" {
		m["TypeRef"] = n.TypeRef
	}
	if n.Error != "" {
		m["Error"] = n.Error
	}

	for i, s := range n.Options.AsList() {
		k := fmt.Sprintf("Options[%03d]", i)
		m[k] = s
	}

	return m
}

// Copy makes a copy of a NativeType.
func (n *NativeType) Copy() *NativeType {
	c := &NativeType{
		Dialect: n.Dialect,
		Name:    n.Name,
		Type:    n.Type,
		TypeRef: n.TypeRef,
		Include: n.Include,
		Options: n.Options.Copy(),
		Error:   n.Error,
	}

	return c
}

// TypeList holds a slice of TypeElements.
// - Behavior is similar to a stack with Push/Pop methods to add/remove elements from the end
type TypeList struct {
	types []*TypeNode
}

func NewTypeList() *TypeList {
	// Initialize an empty TypeList.
	return &TypeList{
		types: make([]*TypeNode, 0),
	}
}

// Len returns the number of elements in the TypeList.
func (typeList *TypeList) Len() int {
	return len(typeList.types)
}

// Push adds an element to the list.
func (typeList *TypeList) Push(elem *TypeNode) {
	typeList.types = append(typeList.types, elem)
}

// Pop removes the last element from the list an returns it.
// - Returns nil is list is empty.
func (typeList *TypeList) Pop() *TypeNode {
	if len(typeList.types) > 0 {
		lastElem := typeList.types[len(typeList.types)-1]
		typeList.types = typeList.types[:len(typeList.types)-1]

		return lastElem
	}

	// Empty list.
	return nil
}

// Copy makes a copy of the current TypeList.
// - Parent is set if parentElem is not nil.
func (typeList *TypeList) Copy(parentElem *TypeNode) *TypeList {
	c := NewTypeList()

	// Copy all elements to new list.
	for _, elem := range typeList.types {
		newElem := elem.Copy()
		c.Push(newElem)

		if parentElem != nil {
			parentElem.AddChild(newElem)
		}
	}

	return c
}

// Elements returns the internal slice of TypeElements.
func (typeList *TypeList) Elements() []*TypeNode {
	return typeList.types
}

// AncestorTypeRef keeps track of type references that are ancestors of the current element.
// - Stores a count of references found.
// - If count > 1, a cyclical reference exists.
type AncestorTypeRef map[string]int

// NewAncestorTypeRef initializes a new ancestor list.
func NewAncestorTypeRef() AncestorTypeRef {
	return make(AncestorTypeRef)
}

// Copy makes a copy of the ancestor list.
func (a AncestorTypeRef) Copy() AncestorTypeRef {
	n := make(AncestorTypeRef)
	for k, v := range a {
		n[k] = v
	}
	return n
}

// Contains returns true if the key exists in ancestor list.
func (a AncestorTypeRef) Contains(key string) bool {
	return a[key] > 0
}

// Add adds a reference count to the ancestor list.
func (a AncestorTypeRef) Add(key string) int {
	if key == "" {
		return 0
	}

	a[key]++
	return a[key]
}
