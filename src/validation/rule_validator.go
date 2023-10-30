package validation

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting/log"
)

// Define a dynamic structure map to represent the dynamic content of Message
type DynamicStruct map[string]interface{}

// Structure to store the validation rules
type ValidationRule struct {
	FieldName      string // Name of the field that needs validation
	ValidationRule string // Name of the validation rule to be applied
}

// Struture to represent a validator with ValidationRules
type RuleValidator struct {
	pack   string
	Rules  []ValidationRule // List of validation rules to execute
	logger log.Logger
}

// Func: NewRuleValidator is for creating a new instance of a validator
// @author AB
// @params
// rules: List of rules that will apply during the validation
// @return
// Validator created
func NewRuleValidator(logger log.Logger, rules []ValidationRule) *RuleValidator {
	return &RuleValidator{
		Rules:  rules,
		pack:   "RuleValidator",
		logger: logger,
	}
}

// Func: NewRuleValidatorFromFile is for creating a new instance of a validator
// @author AB
// @params
// fileName: Name of the configuration file that contains the validation rules
// @return
// RuleValidator instance
func NewRuleValidatorFromFile(fileName string) *RuleValidator {
	result := &RuleValidator{pack: "RuleValidator"}
	err := result.loadValidationRules(fileName)
	if err != nil {
		return nil
	}

	return result
}

// Func: Validate is for Validating a dynamic structure based on a set of validation rules
// @author AB
// @params
// data: DynamicStruct to be validated
// @return
// Error if validation fails.
func (v *RuleValidator) Validate(data DynamicStruct) (*ValidationResult, error) {
	//// TODO Update function to return Validation results for each field
	for _, rule := range v.Rules {
		fieldName := rule.FieldName
		validationRule := rule.ValidationRule

		fieldValue, exists := data[fieldName]
		if !exists {
			return nil, fmt.Errorf("Field '%s' not found in data", fieldName)
		}

		if err := v.applyValidationRule(fieldName, validationRule, fieldValue); err != nil {
			return nil, err
		}
	}

	return nil, nil
}

// Func: ApplyValidationRule Applies the specified validation rule to the field indicated
// @author AB
// @params
// fieldName: Name of the field to be validated
// validationRule: Name of the validation to execute
// value: Value of the field to be validated
// @return
// Returns error if validation fails
func (v *RuleValidator) applyValidationRule(fieldName, validationRule string, value interface{}) error {
	//// Implement validation logic based on the validation rule
	//// might need to use a validation library or custom logic here
	//// fmt.Printf("Validating field '%s' with rule '%s' and value '%v'\n", fieldName, validationRule, value)
	return nil
}

// Func: LoadValidationRules is for loading the validation rules from the configuration file
// @author AB
// @params
// filename: Name of the configuration file that contains the validation rules
// @return
// ValidationRule: List of validation rules loaded
// error in case of read error.
func (v *RuleValidator) loadValidationRules(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		v.logger.Error(err, "error reading file", "Validator", "LoadValidationRules")
		return err
	}

	var rules []ValidationRule
	err = json.Unmarshal(data, &rules)
	if err != nil {
		v.logger.Error(err, "error unmarshal file", "Validator", "LoadValidationRules")
		return err
	}

	v.Rules = rules
	return nil
}
