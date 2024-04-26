package substitute_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/toalaah/vaultsubst/internal/substitute"
	"github.com/toalaah/vaultsubst/internal/vault"
)

func TestPatchFileGeneric(t *testing.T) {
	testPatchFileImpl(t, "generic")
}

func TestPatchFileMultipleMatchSingleLine(t *testing.T) {
	testPatchFileImpl(t, "multiple-match-single-line")
}

func testPatchFileImpl(t *testing.T, file string) {
	assert := assert.New(t)

	m := &mockVaultClient{}
	m.On("ReadKVv2", "kv", "storage/postgres/creds").Return(&api.KVSecret{
		Data: map[string]interface{}{
			"username": "cG9zdGdyZXM=",
			"password": "4_5tr0ng_4nd_c0mpl1c4t3d_p455w0rd",
		},
	}, nil)

	client := &vault.Client{Client: m}

	expected, err := os.ReadFile(fmt.Sprintf("./fixtures/%s.expected.txt", file))
	assert.Nil(err)

	f, err := os.Open(fmt.Sprintf("./fixtures/%s.txt", file))
	assert.Nil(err)

	b, err := substitute.PatchSecretsInFile(f, regexp.MustCompile(`@@(.*?)@@`), client)
	assert.Nil(err)

	assert.Equal(expected, b)
}

type mockVaultClient struct{ mock.Mock }

func (m *mockVaultClient) ReadKVv1(mount, path string) (*api.KVSecret, error) {
	args := m.Called(mount, path)
	// nolint:forcetypeassert
	return args.Get(0).(*api.KVSecret), args.Error(1)
}

func (m *mockVaultClient) ReadKVv2(mount, path string) (*api.KVSecret, error) {
	args := m.Called(mount, path)
	// nolint:forcetypeassert
	return args.Get(0).(*api.KVSecret), args.Error(1)
}
