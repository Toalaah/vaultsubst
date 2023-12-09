package transformations

import (
	"encoding/base64"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransformations(t *testing.T) {
	assert := assert.New(t)
	cases := []struct {
		Action, TestValue, ExpectedValue string
		ExpectedErr                      error
	}{
		{
			Action:        "base64",
			TestValue:     "postgres",
			ExpectedValue: "cG9zdGdyZXM=",
			ExpectedErr:   nil,
		},
		{
			Action:        "trim",
			TestValue:     "  postgres  ",
			ExpectedValue: "postgres",
			ExpectedErr:   nil,
		},
		{
			Action:        "foobarbaz",
			TestValue:     "postgres",
			ExpectedValue: "",
			ExpectedErr:   errors.New("Unknown transformation: foobarbaz"),
		},
		{
			Action:        "base64d",
			TestValue:     "InvalidBase64Value",
			ExpectedValue: "",
			ExpectedErr:   base64.CorruptInputError(16),
		},
	}

	for _, c := range cases {
		v, err := Apply(c.Action, c.TestValue)
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
			s, err = Apply(a, s)
			assert.Nil(err)
		}
		assert.Equal(c.ExpectedValue, s)
	}
}
