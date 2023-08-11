package validator

import (
	"fmt"
	"os"
	"strings"

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
func (v *Validator) ValidateWithSchema(data DynamicStruct, schemaPath string) error {

	_, err := os.ReadFile(schemaPath)
	if err != nil {
		return err
	}

	loader := gojsonschema.NewReferenceLoader("file://" + schemaPath)
	documentLoader := gojsonschema.NewGoLoader(data)

	result, err := gojsonschema.Validate(loader, documentLoader)
	if err != nil {
		return err
	}

	if !result.Valid() {
		errMsgs := make([]string, len(result.Errors()))
		for i, desc := range result.Errors() {
			errMsgs[i] = desc.Description()
		}
		return fmt.Errorf(strings.Join(errMsgs, ", "))
	}

	return nil
}
