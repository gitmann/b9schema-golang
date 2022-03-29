package renderer

type Options struct {
	// DeReference converts TypeRef to their included types.
	// - If TyepRefs have a cyclical relationship, the last TypeRef is kept as a TypeRef.
	DeReference bool

	// Dialects uses dialect resolution to override defaults.
	// - May be overridden or ignored by renderers.
	Dialects []string

	// IncludeNative includes details on native types if set.
	// - May be overridden or ignored by renderers.
	IncludeNative bool

	// Prefix is a string used as a prefix for indented lines.
	Prefix string

	// Indent is used for rendering where indent matters.
	Indent int
}

func NewOptions() *Options {
	opt := &Options{
		Dialects: []string{},
	}
	return opt
}
