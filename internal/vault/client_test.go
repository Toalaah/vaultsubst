package vault_test

import (
	"errors"
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/toalaah/vaultsubst/internal/vault"
)

func TestClientConstruction(t *testing.T) {
	assert := assert.New(t)

	homeUnsetErr := errors.New("VAULT_TOKEN unset and/or failed to read token from ~/.vault-token: $HOME is not defined")
	if runtime.GOOS == "windows" {
		homeUnsetErr = errors.New("VAULT_TOKEN unset and/or failed to read token from ~/.vault-token: %userprofile% is not defined")
	}

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
			ExpectedErr: homeUnsetErr,
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

	secretStub := &api.KVSecret{
		Data: map[string]interface{}{
			"username": "cG9zdGdyZXM=",
			"password": "4_5tr0ng_4nd_c0mpl1c4t3d_p455w0rd",
		},
	}
	m := &mockKVReader{}
	m.
		On("ReadKVv1", "kv", "storage/postgres/creds").Return(secretStub, nil).
		On("ReadKVv2", "kv", "storage/postgres/creds").Return(secretStub, nil).
		On("ReadKVv1", mock.Anything, mock.Anything).Return(nil, nil).
		On("ReadKVv2", mock.Anything, mock.Anything).Return(nil, nil)

	client := &vault.Client{KVReader: m}

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
			Description:   "It should not error with a well-defined secret spec",
		},
		{
			SecretSpec: &vault.SecretSpec{
				Path:         "kv/storage/postgres/creds",
				Field:        "username",
				MountVersion: vault.KVv1,
			},
			ExpectedError: nil,
			Description:   "It should work with KVv1 mounts",
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
		{
			SecretSpec: &vault.SecretSpec{
				Path:         "kv/does/not/exist",
				Field:        "username",
				MountVersion: vault.KVv2,
			},
			ExpectedError: nil,
			Description:   "It should not error when accessing non-existent secret",
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
		// Allow tests on windows to pass as well
		if runtime.GOOS == "windows" {
			os.Setenv("USERPROFILE", home)
		}
		return os.WriteFile(path.Join(home, ".vault-token"), []byte("super_secret_token"), 0600)
	}
	return nil
}

type mockKVReader struct{ mock.Mock }

func (m *mockKVReader) ReadKVv1(mount, path string) (*api.KVSecret, error) {
	args := m.Called(mount, path)
	s := args.Get(0)
	err := args.Error(1)
	if s == nil {
		return nil, err
	}
	// nolint:forcetypeassert
	return s.(*api.KVSecret), err
}

func (m *mockKVReader) ReadKVv2(mount, path string) (*api.KVSecret, error) {
	args := m.Called(mount, path)
	s := args.Get(0)
	err := args.Error(1)
	if s == nil {
		return nil, err
	}
	// nolint:forcetypeassert
	return s.(*api.KVSecret), err
}
