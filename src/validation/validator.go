package validation

// Structure to store validation results
type ValidationResult struct {
	Valid  bool                // Indicates the result of the validation
	Errors map[string][]string // Stores the error details for the validation
}

type Validator interface {
	Validate(data DynamicStruct) (*ValidationResult, error)
}
