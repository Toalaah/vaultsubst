package substitute_test

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/toalaah/vaultsubst/internal/substitute"
	"github.com/toalaah/vaultsubst/internal/vault"
)

func TestSecretPatching(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	client := newMockClient()

	for _, c := range []struct {
		name        string
		expectedErr error
		body        string
		expectedRes string
	}{
		{
			name:        "generic",
			expectedErr: nil,
			body: `
The username is "@@path=kv/storage/postgres/creds,field=username,b64=true,transform=trim|upper@@"

This value should not be substituted due to a different delimiter:  "$$path=kv/storage/postgres/creds,field=username,b64=true,transform=trim|upper$$"
`,
			expectedRes: `
The username is "POSTGRES"

This value should not be substituted due to a different delimiter:  "$$path=kv/storage/postgres/creds,field=username,b64=true,transform=trim|upper$$"
`,
		},
		{
			name:        "multple-match-per-line",
			expectedErr: nil,
			body:        "username=@@path=kv/storage/postgres/creds,field=username,b64=true,transform=trim|upper@@,password=@@path=kv/storage/postgres/creds,field=password@@",
			expectedRes: "username=POSTGRES,password=4_5tr0ng_4nd_c0mpl1c4t3d_p455w0rd",
		},
		{
			name:        "invalid-spec-unknown-field",
			expectedErr: errors.New("Unable to parse option: incorrect-spec (value incorrect-spec)"),
			body:        "Some text here @@incorrect-spec@@",
		},
		{
			name:        "invalid-spec-invalid-path",
			expectedErr: errors.New("no path to query using mountpoint kv"),
			body:        "Some text here @@path=kv/,field=something@@",
		},
		{
			name:        "invalid-spec-format-errors",
			expectedErr: errors.New("Unknown transformation: wrong"),
			body:        "Some text here @@path=kv/storage/postgres/creds,field=username,transform=wrong@@",
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			b, err := substitute.PatchSecrets(strings.NewReader(c.body), regexp.MustCompile(fmt.Sprintf(`%s(.*?)%s`, "@@", "@@")), client)
			assert.Equal(c.expectedRes, string(b))
			assert.Equal(c.expectedErr, err)
		})
	}
}

func TestSecretPatchingWithReaderError(t *testing.T) {
	assert := assert.New(t)
	client := newMockClient()
	b, err := substitute.PatchSecrets(&errReader{}, regexp.MustCompile(""), client)
	assert.Equal(errors.New("read error"), err)
	assert.Nil(b)
}

type errReader struct{}

func (r *errReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error")
}

type mockKVReader struct{ mock.Mock }

func (m *mockKVReader) ReadKVv1(mount, path string) (*api.KVSecret, error) {
	args := m.Called(mount, path)
	//nolint:forcetypeassert,errcheck // for testing purposes this is fine.
	return args.Get(0).(*api.KVSecret), args.Error(1)
}

func (m *mockKVReader) ReadKVv2(mount, path string) (*api.KVSecret, error) {
	args := m.Called(mount, path)
	//nolint:forcetypeassert,errcheck // for testing purposes this is fine.
	return args.Get(0).(*api.KVSecret), args.Error(1)
}

func newMockClient() *vault.Client {
	m := &mockKVReader{}
	m.On("ReadKVv2", "kv", "storage/postgres/creds").Return(&api.KVSecret{
		Data: map[string]interface{}{
			"username": "cG9zdGdyZXM=",
			"password": "4_5tr0ng_4nd_c0mpl1c4t3d_p455w0rd",
		},
	}, nil)
	return &vault.Client{KVReader: m}
}
