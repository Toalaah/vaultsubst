package transformations_test

import (
	"encoding/base64"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/toalaah/vaultsubst/internal/transformations"
)

func TestTransformations(t *testing.T) {
	assert := assert.New(t)
	cases := []struct {
		Action        string
		TestValue     string
		ExpectedValue string
		Description   string
		ExpectedErr   error
	}{
		{
			Action:        "base64",
			TestValue:     "postgres",
			ExpectedValue: "cG9zdGdyZXM=",
			ExpectedErr:   nil,
			Description:   "It should decode base64",
		},
		{
			Action:        "trim",
			TestValue:     "  postgres  ",
			ExpectedValue: "postgres",
			ExpectedErr:   nil,
			Description:   "It should trim leading and trailing spaces",
		},
		{
			Action:        "trim",
			TestValue:     "  hello world  ",
			ExpectedValue: "hello world",
			ExpectedErr:   nil,
			Description:   "It should retain internal spacing",
		},
		{
			Action:        "foobarbaz",
			TestValue:     "postgres",
			ExpectedValue: "",
			ExpectedErr:   errors.New("Unknown transformation: foobarbaz"),
			Description:   "It should fail on unknown transformation types",
		},
		{
			Action:        "base64d",
			TestValue:     "InvalidBase64Value",
			ExpectedValue: "",
			ExpectedErr:   base64.CorruptInputError(16),
			Description:   "It should propagate base64 decode errors",
		},
	}

	for _, c := range cases {
		v, err := transformations.Apply(c.Action, c.TestValue)
		assert.Equal(c.ExpectedValue, v)
		assert.Equal(c.ExpectedErr, err)
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
