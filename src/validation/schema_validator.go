package validation

import (
	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting/log"
	"github.com/xeipuuv/gojsonschema"
)

// Validator of JSON Schema
type SchemaValidator struct {
	pack   string     // Package name
	schema string     // JSON Schema
	logger log.Logger // Logger
}

// GetSchemaValidator is for creating a SchemaValidator
// @author AB
// @params
// logger: Logger to be used
// schema: JSON Schema to be used for validation
// @return
// SchemaValidator instance
func GetSchemaValidator(logger log.Logger, schema string) *SchemaValidator {
	return &SchemaValidator{
		pack:   "SchemaValidator",
		schema: schema,
		logger: logger,
	}
}

// ValidateWithSchema is for Validating a dynamic structure using a JSON Schema
// @author AB
// @params
// data: DynamicStruct to be validated
// schemaPath: Path for the Schema file to be loaded
// @return
// Error if validation fails.
func (v *SchemaValidator) Validate(data DynamicStruct) (*ValidationResult, error) {
	v.logger.Info("Starting Validation With Schema", v.pack, "Validate")

	validationResult := ValidationResult{Valid: true}

	loader := gojsonschema.NewStringLoader(v.schema)
	documentLoader := gojsonschema.NewGoLoader(data)
	result, err := gojsonschema.Validate(loader, documentLoader)

	if err != nil {
		v.logger.Error(err, "error validating message", v.pack, "ValidateWithSchema")
		return nil, err
	}

	if !result.Valid() {
		validationResult.Errors = v.cleanErrors(result.Errors())
		validationResult.Valid = false
		return &validationResult, nil
	}

	return &validationResult, nil
}

// cleanErrors Creates an array or clean error based on the validations
// @author AB
// @params
// error: List of errors generated during the validation
// @return
// ErrorDetail: List of errors found
func (v *SchemaValidator) cleanErrors(errors []gojsonschema.ResultError) map[string][]string {
	result := make(map[string][]string)

	for _, desc := range errors {
		result[desc.Field()] = append(result[desc.Field()], desc.Description())
		v.logger.Debug(desc.Field()+": "+desc.Description(), v.pack, "cleanErrors")
	}

	return result
}
