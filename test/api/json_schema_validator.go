//go:build integration

package api

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/xeipuuv/gojsonschema"
)

func validateJSONSchema(jsonSchemaPath, document string) error {
	schemaPath, err := filepath.Abs(jsonSchemaPath)
	if err != nil {
		return fmt.Errorf("unable to get JSON schema absolute path: %w", err)
	}

	schemaLoader := gojsonschema.NewReferenceLoader(fmt.Sprintf("file://%s", schemaPath))
	documentLoader := gojsonschema.NewStringLoader(document)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if !result.Valid() {
		return fmt.Errorf("validation errors: %s", joinResultErrors(result.Errors(), ", "))
	}

	return nil
}

func joinResultErrors(stringers []gojsonschema.ResultError, separator string) string {
	var strs []string
	for _, stringer := range stringers {
		strs = append(strs, stringer.String())
	}

	return strings.Join(strs, separator)
}
