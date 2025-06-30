//go:build unit

package lang

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testStruct struct {
	Localizable
}

func TestLocalize(t *testing.T) {
	SetBasePath("../../../resources/locales")
	SetSupportedLanguages("en,es")

	testCases := []struct {
		name        string
		lang        string
		subject     testStruct
		expectedMsg string
	}{
		{
			name: "MessageID only - Spanish",
			lang: "es",
			subject: testStruct{
				Localizable{
					MessageID: "errors.invalid_type",
				},
			},
			expectedMsg: "Tipo inválido encontrado!",
		},
		{
			name: "MessageID only - English",
			lang: "en",
			subject: testStruct{
				Localizable{
					MessageID: "errors.invalid_type",
				},
			},
			expectedMsg: "Invalid type encountered!",
		},
		{
			name: "MessageID only - Spanish with dialect",
			lang: "es_MX",
			subject: testStruct{
				Localizable{
					MessageID: "errors.invalid_type",
				},
			},
			expectedMsg: "Tipo inválido encontrado!",
		},
		{
			name: "With message variables - English",
			lang: "en_GB",
			subject: testStruct{
				Localizable{
					MessageID:   "errors.record_not_found",
					MessageVars: map[string]interface{}{"name": "User"},
				},
			},
			expectedMsg: "User record not found!",
		},
		{
			name: "Missing message variables - Spanish",
			lang: "es",
			subject: testStruct{
				Localizable{
					MessageID: "errors.record_not_found",
				},
			},
			expectedMsg: "Registro de <no value> no encontrado!",
		},
		{
			name: "Unsupported language - use default language",
			lang: "ar",
			subject: testStruct{
				Localizable{
					MessageID: "errors.invalid_type",
				},
			},
			expectedMsg: "Invalid type encountered!",
		},
		{
			name: "MessageID not found - use default message",
			lang: "en",
			subject: testStruct{
				Localizable{
					MessageID: "errors.foo_bar",
					Message:   "Foo Bar",
				},
			},
			expectedMsg: "Foo Bar",
		},
		{
			name: "Language is not set - use default language",
			lang: "",
			subject: testStruct{
				Localizable{
					MessageID: "errors.invalid_type",
				},
			},
			expectedMsg: "Invalid type encountered!",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			localized := testCase.subject.Localize(testCase.lang)
			assert.Equal(t, testCase.expectedMsg, localized)
		})
	}
}
