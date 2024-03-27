package vault_test

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/toalaah/vaultsubst/internal/vault"
)

func TestClientConstruction(t *testing.T) {
	assert := assert.New(t)
	cases := []struct {
		Env         map[string]string
		ExpectedErr error
		Description string
	}{
		{
			Env: map[string]string{
				"VAULT_TOKEN": "super_secret_token",
			},
			ExpectedErr: nil,
			Description: "It should return a valid client if minimum required environment variables are set",
		},
		{
			Env: map[string]string{
				"VAULT_ADDR": "http://localhost:8200",
			},
			ExpectedErr: errors.New("VAULT_TOKEN unset and/or failed to read token from ~/.vault-token: $HOME is not defined"),
			Description: "It should fail if VAULT_TOKEN is not defined and ~/.vault-token does not exist",
		},
		{
			Env: map[string]string{
				"VAULT_ADDR": "http://localhost:8200",
				"HOME":       t.TempDir(),
			},
			ExpectedErr: nil,
			Description: "It should fallback to ~/.vault-token if VAULT_TOKEN is not defined",
		},
	}

	for _, c := range cases {
		assert.Nil(prepareTestCaseEnv(c.Env), "failed to prepare test case environment")
		client, err := vault.NewClient()
		assert.Equal(c.ExpectedErr, err, c.Description)
		if c.ExpectedErr == nil {
			assert.NotNil(client)
		}
	}
}

func TestClientReadKV(t *testing.T) {
	assert := assert.New(t)
	// Ensure vault token is defined so that NewClient() does not fail
	os.Setenv("VAULT_TOKEN", "dummy-value")
	client, err := vault.NewClient()
	assert.Nil(err)

	cases := []struct {
		SecretSpec    *vault.SecretSpec
		ExpectedError error
		Description   string
	}{
		{
			SecretSpec: &vault.SecretSpec{
				Path:         "kv/storage/postgres/creds",
				Field:        "username",
				MountVersion: vault.KVv2,
			},
			ExpectedError: nil,
			Description:   "It should not error with well-defined secret spec",
		},
		{
			SecretSpec: &vault.SecretSpec{
				Path:         "kv/storage/postgres/creds",
				Field:        "username",
				MountVersion: "wrong",
			},
			ExpectedError: errors.New("secret &{Path:kv/storage/postgres/creds Field:username B64:false MountVersion:wrong Transformations:[]}: unknown kv version wrong"),
			Description:   "It should error with an invalid KV mount version",
		},
	}

	for _, c := range cases {
		_, err := client.ReadKV(c.SecretSpec)
		assert.Equal(c.ExpectedError, err)
	}
}

func prepareTestCaseEnv(env map[string]string) error {
	os.Clearenv()
	for k, v := range env {
		os.Setenv(k, v)
	}
	if home, ok := os.LookupEnv("HOME"); ok {
		return os.WriteFile(home+"/.vault-token", []byte("super_secret_token"), 0600)
	}
	return nil
}
