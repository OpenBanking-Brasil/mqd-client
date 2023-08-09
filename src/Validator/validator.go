package validator

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// Define a dynamic structure map to represent the dynamic content of Message
type DynamicStruct map[string]interface{}

// Structure to store validation results
type ValidationResult struct {
	Valid   bool   // Indicates the result of the validation
	ErrType string // Indicates the typ of error based on the validation executed
}

// Structure to store the validation rules
type ValidationRule struct {
	FieldName      string // Name of the field that needs validation
	ValidationRule string // Name of the validation rule to be applied
}

// Struture to represent a validator with ValidationRules
type Validator struct {
	Rules []ValidationRule // List of validation rules to execute
}

/**
 * Func: NewValidator is for creating a new instance of a validator
 *
 * @author AB
 *
 * @params
 * rules: List of rules that will apply during the validation
 * @return
 * Validator created
 */
func NewValidator(rules []ValidationRule) *Validator {
	return &Validator{Rules: rules}
}

/**
 * Func: Validate is for Validating a dynamic structure based on a set of validation rules
 *
 * @author AB
 *
 * @params
 * data: DynamicStruct to be validated
 * @return
 * Error if validation fails.
 */
func (v *Validator) Validate(data DynamicStruct) error {
	for _, rule := range v.Rules {
		fieldName := rule.FieldName
		validationRule := rule.ValidationRule

		fieldValue, exists := data[fieldName]
		if !exists {
			return fmt.Errorf("Field '%s' not found in data", fieldName)
		}

		if err := ApplyValidationRule(fieldName, validationRule, fieldValue); err != nil {
			return err
		}
	}

	return nil
}

/**
 * Func: ApplyValidationRule Applies the specified validation rule to the field indicated
 *
 * @author AB
 *
 * @params
 * fieldName: Name of the field to be validated
 * validationRule: Name of the validation to execute
 * value: Value of the field to be validated
 * @return
 * Returns error if validation fails
 */
func ApplyValidationRule(fieldName, validationRule string, value interface{}) error {
	// Implement validation logic based on the validation rule
	// might need to use a validation library or custom logic here
	//fmt.Printf("Validating field '%s' with rule '%s' and value '%v'\n", fieldName, validationRule, value)
	return nil
}

/**
 * Func: LoadValidationRules is for loading the validation rules from the configuration file
 *
 * @author AB
 *
 * @params
 * filename: Name of the configuration file that contains the validation rules
 * @return
 * ValidationRule: List of validation rules loaded
 * error in case of read error.
 */
func LoadValidationRules(filename string) ([]ValidationRule, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var rules []ValidationRule
	err = json.Unmarshal(data, &rules)
	if err != nil {
		return nil, err
	}

	return rules, nil
}
