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
	os.Clearenv()

	homeUnsetErr := errors.New("VAULT_TOKEN unset and/or failed to read token from ~/.vault-token: $HOME is not defined")
	if runtime.GOOS == "windows" {
		homeUnsetErr = errors.New("VAULT_TOKEN unset and/or failed to read token from ~/.vault-token: %userprofile% is not defined")
	}

	for _, c := range []struct {
		name     string
		env      map[string]string
		expected error
	}{
		{
			name:     "generic-success",
			expected: nil,
			env: map[string]string{
				"VAULT_TOKEN": "super_secret_token",
			},
		},
		{
			name:     "missing-token-env-var",
			expected: homeUnsetErr,
			env: map[string]string{
				"VAULT_ADDR": "http://localhost:8200",
			},
		},
		{
			name:     "fallback-token-file",
			expected: nil,
			env: map[string]string{
				"VAULT_ADDR": "http://localhost:8200",
				"HOME":       t.TempDir(),
			},
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			// Set environment variable for current test case.
			for k, v := range c.env {
				t.Setenv(k, v)
			}
			// If home is set, write fake ".vault-token" file.
			if home, ok := os.LookupEnv("HOME"); ok {
				// Allow tests on windows to pass as well.
				if runtime.GOOS == "windows" {
					os.Setenv("USERPROFILE", home)
				}
				assert.Nil(os.WriteFile(path.Join(home, ".vault-token"), []byte("super_secret_token"), 0o600), "failed to prepare env")
			}
			client, err := vault.NewClient()
			assert.Equal(c.expected, err)
			if c.expected == nil {
				assert.NotNil(client)
			}
		})
	}
}

func TestClientReadKV(t *testing.T) {
	assert := assert.New(t)
	t.Parallel()

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

	for _, c := range []struct {
		name     string
		spec     *vault.SecretSpec
		expected error
	}{
		{
			name:     "well-defined-secret-spec",
			expected: nil,
			spec: &vault.SecretSpec{
				Path:         "kv/storage/postgres/creds",
				Field:        "username",
				MountVersion: vault.KVv2,
			},
		},
		{
			name:     "kv-v1-mounts",
			expected: nil,
			spec: &vault.SecretSpec{
				Path:         "kv/storage/postgres/creds",
				Field:        "username",
				MountVersion: vault.KVv1,
			},
		},
		{
			name:     "invalid-kv-mount-version",
			expected: errors.New("secret &{Path:kv/storage/postgres/creds Field:username B64:false MountVersion:wrong Transformations:[]}: unknown kv version wrong"),
			spec: &vault.SecretSpec{
				Path:         "kv/storage/postgres/creds",
				Field:        "username",
				MountVersion: "wrong",
			},
		},
		{
			name:     "nonexistent-secret",
			expected: nil,
			spec: &vault.SecretSpec{
				Path:         "kv/does/not/exist",
				Field:        "username",
				MountVersion: vault.KVv2,
			},
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			_, err := client.ReadKV(c.spec)
			assert.Equal(c.expected, err)
		})
	}
}

type mockKVReader struct{ mock.Mock }

func (m *mockKVReader) ReadKVv1(mount, path string) (*api.KVSecret, error) {
	args := m.Called(mount, path)
	s := args.Get(0)
	err := args.Error(1)
	if s == nil {
		return nil, err
	}
	//nolint:forcetypeassert,errcheck // for testing purposes this is fine.
	return s.(*api.KVSecret), err
}

func (m *mockKVReader) ReadKVv2(mount, path string) (*api.KVSecret, error) {
	args := m.Called(mount, path)
	s := args.Get(0)
	err := args.Error(1)
	if s == nil {
		return nil, err
	}
	//nolint:forcetypeassert,errcheck // for testing purposes this is fine.
	return s.(*api.KVSecret), err
}
