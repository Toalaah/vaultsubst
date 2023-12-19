package secret

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/toalaah/vaultsubst/testutil"
)

func TestSecretSpecParsing(t *testing.T) {
	assert := assert.New(t)
	cases := []struct {
		SpecStr       string
		ExpectedValue *SecretSpec
		ExpectedErr   error
	}{
		{
			SpecStr: "path=kv/storage/postgres/creds,field=username,b64=true,transform=trim|upper",
			ExpectedValue: &SecretSpec{
				Path:            "kv/storage/postgres/creds",
				Field:           "username",
				B64:             true,
				Transformations: []string{"trim", "upper"},
			},
			ExpectedErr: nil,
		},
		{
			SpecStr: "path=kv/storage/postgres/creds,field=password",
			ExpectedValue: &SecretSpec{
				Path:            "kv/storage/postgres/creds",
				Field:           "password",
				B64:             false,
				Transformations: nil,
			},
			ExpectedErr: nil,
		},
		{
			SpecStr:       "path = kv/storage/postgres/creds,,fieldpassword",
			ExpectedValue: nil,
			ExpectedErr:   errors.New("Unable to parse option: path = kv/storage/postgres/creds,,fieldpassword (value )"),
		},
		{
			SpecStr:       "field=username,b64=true,transform=trim|upper",
			ExpectedValue: nil,
			ExpectedErr:   errors.New("Path may not be empty"),
		},
		{
			SpecStr:       "path=kv/storage/postgres/creds,b64=true,transform=trim|upper",
			ExpectedValue: nil,
			ExpectedErr:   errors.New("Field may not be empty"),
		},

		{
			SpecStr:       "path=kv/storage/postgres/creds,transform=trim,upper,b64d",
			ExpectedValue: nil,
			ExpectedErr:   errors.New("Unable to parse option: path=kv/storage/postgres/creds,transform=trim,upper,b64d (value upper)"),
		},
	}

	for _, c := range cases {
		ss, err := NewSecretSpec(c.SpecStr)
		assert.Equal(c.ExpectedValue, ss)
		assert.Equal(c.ExpectedErr, err)
	}
}

func TestSecretFetching(t *testing.T) {
	assert := assert.New(t)

	client := testutil.NewTestVault(t)

	cases := []struct {
		Spec          *SecretSpec
		ExpectedValue string
		ExpectedErr   error
	}{
		{
			Spec: &SecretSpec{
				Path:  "kv/storage/postgres/creds",
				Field: "username",
				B64:   true,
			},
			ExpectedValue: "postgres",
			ExpectedErr:   nil,
		},
		{
			Spec: &SecretSpec{
				Path:  "kv/storage/doesnotexist",
				Field: "username",
			},
			ExpectedValue: "",
			ExpectedErr:   errors.New("secret is nil"),
		},
		{
			Spec: &SecretSpec{
				Path:  "kv/storage/postgres/creds",
				Field: "doesnotexist",
			},
			ExpectedValue: "",
			ExpectedErr:   errors.New("could not cast data at field doesnotexist to string"),
		},
	}

	for _, c := range cases {
		secret, err := c.Spec.Fetch(client)
		assert.Equal(c.ExpectedValue, secret)
		assert.Equal(c.ExpectedErr, err)
	}
}
