package validation

import (
	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting/log"
	"github.com/xeipuuv/gojsonschema"
)

// SchemaValidator Validator that uses JSON Schemas
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

// Validate is for Validating a dynamic structure using a JSON Schema
// @author AB
// @params
// data: DynamicStruct to be validated
// schemaPath: Path for the Schema file to be loaded
// @return
// Error if validation fails.
func (sm *SchemaValidator) Validate(data DynamicStruct) (*Result, error) {
	sm.logger.Info("Starting Validation With Schema", sm.pack, "Validate")

	validationResult := Result{Valid: true}
	if sm.schema == "" {
		return &validationResult, nil
	}

	loader := gojsonschema.NewStringLoader(sm.schema)
	documentLoader := gojsonschema.NewGoLoader(data)
	result, err := gojsonschema.Validate(loader, documentLoader)
	if err != nil {
		sm.logger.Error(err, "error validating message", sm.pack, "Validate")
		return nil, err
	}

	if !result.Valid() {
		validationResult.Errors = sm.cleanErrors(result.Errors())
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
func (sm *SchemaValidator) cleanErrors(errors []gojsonschema.ResultError) map[string][]string {
	result := make(map[string][]string)
	for _, desc := range errors {
		result[desc.Field()] = append(result[desc.Field()], desc.Description())
		sm.logger.Debug(desc.Field()+": "+desc.Description(), sm.pack, "cleanErrors")
	}

	return result
}
