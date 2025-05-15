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
	t.Parallel()

	for _, c := range []struct {
		name          string
		parseStr      string
		expectedValue *vault.SecretSpec
		expectedErr   error
	}{
		{
			parseStr: "path=kv/storage/postgres/creds,field=password",
			expectedValue: &vault.SecretSpec{
				Path:            "kv/storage/postgres/creds",
				Field:           "password",
				B64:             false,
				Transformations: nil,
				MountVersion:    vault.KVv2,
			},
			expectedErr: nil,
			name:        "parse-generic",
		},
		{
			parseStr: "path=kv/storage/postgres/creds,field=username,b64=true,transform=trim|upper",
			expectedValue: &vault.SecretSpec{
				Path:            "kv/storage/postgres/creds",
				Field:           "username",
				B64:             true,
				Transformations: []string{"trim", "upper"},
				MountVersion:    vault.KVv2,
			},
			expectedErr: nil,
			name:        "parse-transforms",
		},
		{
			parseStr: "path=kv/storage/postgres/creds,field=password,ver=v1",
			expectedValue: &vault.SecretSpec{
				Path:            "kv/storage/postgres/creds",
				Field:           "password",
				B64:             false,
				Transformations: nil,
				MountVersion:    vault.KVv1,
			},
			expectedErr: nil,
			name:        "parse-kv-version",
		},
		{
			parseStr: "path=kv/storage/postgres/creds,field=password,ver=foobarbaz",
			expectedValue: &vault.SecretSpec{
				Path:            "kv/storage/postgres/creds",
				Field:           "password",
				B64:             false,
				Transformations: nil,
				MountVersion:    "foobarbaz",
			},
			expectedErr: nil,
			// This is the job of the vault client, see ReadKV() in client.go.
			name: "parse-invalid-kv-version",
		},
		{
			parseStr: "path =       kv/storage/postgres/creds ,    field= username,b64=true",
			expectedValue: &vault.SecretSpec{
				Path:         "kv/storage/postgres/creds",
				Field:        "username",
				B64:          true,
				MountVersion: vault.KVv2,
			},
			expectedErr: nil,
			name:        "parse-trim-spaces",
		},
		{
			parseStr:      "path=kv/storage/postgres/creds,,fieldpassword",
			expectedValue: nil,
			expectedErr:   errors.New("unable to parse option: path=kv/storage/postgres/creds,,fieldpassword (value )"),
			name:          "parse-invalid-str",
		},
		{
			parseStr:      "field=username,b64=true,transform=trim|upper",
			expectedValue: nil,
			expectedErr:   errors.New("path may not be empty"),
			name:          "missing-required-fields-1",
		},
		{
			parseStr:      "path=kv/storage/postgres/creds,b64=true,transform=trim|upper",
			expectedValue: nil,
			expectedErr:   errors.New("field may not be empty"),
			name:          "missing-required-fields-2",
		},
		{
			parseStr:      "path=kv/storage/postgres/creds,transform=trim,upper,b64d",
			expectedValue: nil,
			expectedErr:   errors.New("unable to parse option: path=kv/storage/postgres/creds,transform=trim,upper,b64d (value upper)"),
			name:          "transform-delimiters",
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			s, err := vault.NewSecretSpec(c.parseStr)
			assert.Equal(c.expectedValue, s, c.name)
			assert.Equal(c.expectedErr, err, c.name)
		})
	}
}

func TestSecretFormatting(t *testing.T) {
	assert := assert.New(t)
	t.Parallel()

	dummySecret := &api.KVSecret{
		Data: map[string]any{
			"username": "cG9zdGdyZXM=",
			"password": "4_5tr0ng_4nd_c0mpl1c4t3d_p455w0rd",
		},
	}

	for _, c := range []struct {
		name          string
		spec          *vault.SecretSpec
		expectedValue string
		expectedErr   error
		secret        *api.KVSecret
	}{
		{
			spec: &vault.SecretSpec{
				Path:  "kv/storage/postgres/creds",
				Field: "username",
				B64:   true,
			},
			expectedValue: "postgres",
			expectedErr:   nil,
			secret:        dummySecret,
			name:          "generic-1",
		},
		{
			spec: &vault.SecretSpec{
				Path:  "kv/storage/doesnotexist",
				Field: "username",
			},
			expectedValue: "",
			expectedErr:   errors.New("secret is nil"),
			secret:        nil,
			name:          "nil-secret",
		},
		{
			spec: &vault.SecretSpec{
				Path:  "kv/storage/postgres/creds",
				Field: "doesnotexist",
			},
			expectedValue: "",
			expectedErr:   errors.New("could not cast data at field doesnotexist to string"),
			secret:        dummySecret,
			name:          "nonexistent-field",
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			s, err := c.spec.FormatSecret(c.secret)
			assert.Equal(c.expectedValue, s, c.name)
			assert.Equal(c.expectedErr, err, c.name)
		})
	}
}
