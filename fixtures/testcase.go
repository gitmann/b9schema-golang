package fixtures

type TestCase struct {
	Name  string
	Value interface{}

	// Expected strings for reference and de-reference.
	RefStrings     []string
	DerefStrings   []string
	JSONStrings    []string
	OpenAPIStrings []string
}
