package fixtures

type TestCase struct {
	Name  string
	Value interface{}

	// Expected strings for reference and de-reference.
	// - map[render engine][de-reference flag][]string
	Want map[string]WantSet
}

type WantSet map[bool][]string
