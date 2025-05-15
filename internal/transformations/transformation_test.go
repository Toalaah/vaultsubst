package transformations_test

import (
	"encoding/base64"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/toalaah/vaultsubst/internal/transformations"
)

func TestTransformations(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	for _, c := range []struct {
		action        string
		testValue     string
		expectedValue string
		name          string
		expectedErr   error
	}{
		{
			name:          "decode-base64",
			action:        "base64",
			testValue:     "postgres",
			expectedValue: "cG9zdGdyZXM=",
			expectedErr:   nil,
		},
		{
			name:          "trim",
			action:        "trim",
			testValue:     "  postgres  ",
			expectedValue: "postgres",
			expectedErr:   nil,
		},
		{
			name:          "trim-retain-internal-spacing",
			action:        "trim",
			testValue:     "  hello world  ",
			expectedValue: "hello world",
			expectedErr:   nil,
		},
		{
			name:          "fail-on-invalid-transform",
			action:        "foobarbaz",
			testValue:     "postgres",
			expectedValue: "",
			expectedErr:   errors.New("unknown transformation: foobarbaz"),
		},
		{
			name:          "fail-illegal-base64",
			action:        "base64d",
			testValue:     "InvalidBase64Value",
			expectedValue: "",
			expectedErr:   base64.CorruptInputError(16),
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			v, err := transformations.Apply(c.action, c.testValue)
			assert.Equal(c.expectedValue, v)
			assert.Equal(c.expectedErr, err)
		})
	}
}

func TestTransformationsChained(t *testing.T) {
	assert := assert.New(t)
	cases := []struct {
		Actions                  []string
		TestValue, ExpectedValue string
	}{
		{
			Actions:       []string{"upper", "base64", "base64d"},
			TestValue:     "postgres",
			ExpectedValue: "POSTGRES",
		},
		{
			Actions:       []string{"trim", "lower"},
			TestValue:     "  postgres  ",
			ExpectedValue: "postgres",
		},
	}

	for _, c := range cases {
		var err error
		s := c.ExpectedValue
		for _, a := range c.Actions {
			s, err = transformations.Apply(a, s)
			assert.Nil(err)
		}
		assert.Equal(c.ExpectedValue, s)
	}
}
