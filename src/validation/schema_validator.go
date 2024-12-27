package validation

import (
	"strconv"
	"strings"

	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting/log"
	"github.com/xeipuuv/gojsonschema"
)

// DynamicStruct Defines a dynamic map to represent the dynamic content of Message
type DynamicStruct map[string]interface{}

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
		if strings.Contains(desc.String(), "\"if\"") {
			continue
		}

		field := sm.cleanString(desc.Field())
		desc := sm.cleanString(desc.Description())
		result[field] = append(result[field], desc)
		sm.logger.Debug(field+": "+desc, sm.pack, "cleanErrors")
	}

	return result
}

// cleanString removes unnecessary information from the field an error fields
//
// Parameters:
//   - value: string to be cleaned
//
// Returns:
//   - string: clean string
func (sm *SchemaValidator) cleanString(value string) string {
	if !strings.Contains(value, "data") {
		return value
	}

	values := strings.Split(value, ".")
	result := ""
	for _, v := range values {
		if !sm.isNumeric(v) {
			if result == "" {
				result = v
			} else {
				result = result + "." + v
			}
		}
	}

	return result
}

// isNumeric indicates if a string contains a numeric value
//
// Parameters:
//   - s: String to validate
//
// Returns:
//   - bool: tru if value is numeric
func (sm *SchemaValidator) isNumeric(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}
