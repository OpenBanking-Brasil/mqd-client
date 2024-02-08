package validation

// Result stores the results for the validations
type Result struct {
	Valid  bool                // Indicates the result of the validation
	Errors map[string][]string // Stores the error details for the validation
}

// Validator is the Interface that exposes the methods to validate structures
type Validator interface {
	Validate(data DynamicStruct) (*Result, error)
}
