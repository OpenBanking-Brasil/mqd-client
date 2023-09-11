package validator

import (
	"github.com/xeipuuv/gojsonschema"
)

/**
 * Func: ValidateWithSchema is for Validating a dynamic structure using a JSON Schema
 *
 * @author AB
 *
 * @params
 * data: DynamicStruct to be validated
 * schemaPath: Path for the Schema file to be loaded
 * @return
 * Error if validation fails.
 */
func (v *Validator) ValidateWithSchema(data DynamicStruct, schema string) (*ValidationResult, error) {
	validationResult := ValidationResult{Valid: true}

	// println("validating with schema")
	loader := gojsonschema.NewStringLoader(schema)
	documentLoader := gojsonschema.NewGoLoader(data)
	result, err := gojsonschema.Validate(loader, documentLoader)

	if err != nil {
		println("error validating: " + err.Error())
		return nil, err
	}

	if !result.Valid() {
		validationResult.Errors = cleanErrors(result.Errors())
		validationResult.Valid = false
		return &validationResult, nil
	}

	return &validationResult, nil
}

/**
 * Func: cleanErrors Creates an array or clean error based on the validations
 *
 * @author AB
 *
 * @params
 * error: List of errors generated during the validation
 * @return
 * ErrorDetail: List of errors found
 */
func cleanErrors(errors []gojsonschema.ResultError) map[string][]string {
	result := make(map[string][]string)

	for _, desc := range errors {
		result[desc.Field()] = append(result[desc.Field()], desc.Description())
		println(desc.Field() + ": " + desc.Description())
	}

	return result
}
