package validator

import (
	"fmt"
	"strings"

	"github.com/OpenBanking-Brasil/MQD_Client/configuration"
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
func (v *Validator) ValidateWithSchema(data DynamicStruct, settings *configuration.EndPointSettings) error {
	// println("validating with schema")
	loader := gojsonschema.NewStringLoader(settings.JSONSchema)
	documentLoader := gojsonschema.NewGoLoader(data)
	result, err := gojsonschema.Validate(loader, documentLoader)
	if err != nil {
		println("error validating: " + err.Error())
		return err
	}

	if !result.Valid() {
		errMsgs := make([]string, len(result.Errors()))
		for i, desc := range result.Errors() {
			errMsgs[i] = desc.Field() + ": " + desc.Description()
		}
		return fmt.Errorf(strings.Join(errMsgs, ", "))
	}

	return nil
}
