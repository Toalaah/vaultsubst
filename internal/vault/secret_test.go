package vault_test

import (
	"errors"
	"testing"

	"github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
	"github.com/toalaah/vaultsubst/internal/vault"
)

func TestSecretSpecParsing(t *testing.T) {
	assert := assert.New(t)
	cases := []struct {
		SpecStr       string
		ExpectedValue *vault.SecretSpec
		ExpectedErr   error
		Description   string
	}{
		{
			SpecStr: "path=kv/storage/postgres/creds,field=username,b64=true,transform=trim|upper",
			ExpectedValue: &vault.SecretSpec{
				Path:            "kv/storage/postgres/creds",
				Field:           "username",
				B64:             true,
				Transformations: []string{"trim", "upper"},
				MountVersion:    vault.KVv2,
			},
			ExpectedErr: nil,
			Description: "It should correctly construct with transform opts",
		},
		{
			SpecStr: "path=kv/storage/postgres/creds,field=password",
			ExpectedValue: &vault.SecretSpec{
				Path:            "kv/storage/postgres/creds",
				Field:           "password",
				B64:             false,
				Transformations: nil,
				MountVersion:    vault.KVv2,
			},
			ExpectedErr: nil,
			Description: "It should correctly construct from input with minimum required fields",
		},
		{
			SpecStr: "path=kv/storage/postgres/creds,field=password,ver=v1",
			ExpectedValue: &vault.SecretSpec{
				Path:            "kv/storage/postgres/creds",
				Field:           "password",
				B64:             false,
				Transformations: nil,
				MountVersion:    vault.KVv1,
			},
			ExpectedErr: nil,
			Description: "It should correctly set the kv version",
		},
		{
			SpecStr: "path=kv/storage/postgres/creds,field=password,ver=foobarbaz",
			ExpectedValue: &vault.SecretSpec{
				Path:            "kv/storage/postgres/creds",
				Field:           "password",
				B64:             false,
				Transformations: nil,
				MountVersion:    "foobarbaz",
			},
			ExpectedErr: nil,
			// This is the job of the vault client, see Read() in client.go
			Description: "It should not verify that KV version is valid",
		},
		{
			SpecStr: "path =       kv/storage/postgres/creds ,    field= username,b64=true",
			ExpectedValue: &vault.SecretSpec{
				Path:         "kv/storage/postgres/creds",
				Field:        "username",
				B64:          true,
				MountVersion: vault.KVv2,
			},
			ExpectedErr: nil,
			Description: "It should trim spaces from input string",
		},
		{
			SpecStr:       "path=kv/storage/postgres/creds,,fieldpassword",
			ExpectedValue: nil,
			ExpectedErr:   errors.New("Unable to parse option: path=kv/storage/postgres/creds,,fieldpassword (value )"),
			Description:   "It should fail to construct on poorly delimited input",
		},
		{
			SpecStr:       "field=username,b64=true,transform=trim|upper",
			ExpectedValue: nil,
			ExpectedErr:   errors.New("Path may not be empty"),
			Description:   "It should fail to construct if path is empty",
		},
		{
			SpecStr:       "path=kv/storage/postgres/creds,b64=true,transform=trim|upper",
			ExpectedValue: nil,
			ExpectedErr:   errors.New("Field may not be empty"),
			Description:   "It should fail to construct if field is empty",
		},
		{
			SpecStr:       "path=kv/storage/postgres/creds,transform=trim,upper,b64d",
			ExpectedValue: nil,
			ExpectedErr:   errors.New("Unable to parse option: path=kv/storage/postgres/creds,transform=trim,upper,b64d (value upper)"),
			Description:   "Transform opts should be delimited by '|', not ','",
		},
	}

	for _, c := range cases {
		s, err := vault.NewSecretSpec(c.SpecStr)
		assert.Equal(c.ExpectedValue, s, c.Description)
		assert.Equal(c.ExpectedErr, err, c.Description)
	}
}

func TestSecretFormatting(t *testing.T) {
	assert := assert.New(t)

	dummySecret := &api.KVSecret{
		Data: map[string]interface{}{
			"username": "cG9zdGdyZXM=",
			"password": "4_5tr0ng_4nd_c0mpl1c4t3d_p455w0rd",
		},
	}

	cases := []struct {
		Spec          *vault.SecretSpec
		ExpectedValue string
		ExpectedErr   error
		Description   string
		Secret        *api.KVSecret
	}{
		{
			Spec: &vault.SecretSpec{
				Path:  "kv/storage/postgres/creds",
				Field: "username",
				B64:   true,
			},
			ExpectedValue: "postgres",
			ExpectedErr:   nil,
			Secret:        dummySecret,
			Description:   "It should read and format an existing field from a secret",
		},
		{
			Spec: &vault.SecretSpec{
				Path:  "kv/storage/doesnotexist",
				Field: "username",
			},
			ExpectedValue: "",
			ExpectedErr:   errors.New("secret is nil"),
			Secret:        nil,
			Description:   "It should fail to format a nil secret",
		},
		{
			Spec: &vault.SecretSpec{
				Path:  "kv/storage/postgres/creds",
				Field: "doesnotexist",
			},
			ExpectedValue: "",
			ExpectedErr:   errors.New("could not cast data at field doesnotexist to string"),
			Secret:        dummySecret,
			Description:   "It should fail to format a non-existent field",
		},
	}

	for _, c := range cases {
		s, err := c.Spec.FormatSecret(c.Secret)
		assert.Equal(c.ExpectedValue, s, c.Description)
		assert.Equal(c.ExpectedErr, err, c.Description)
	}
}
