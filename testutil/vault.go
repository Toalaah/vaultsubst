package testutil

import (
	"testing"

	"github.com/hashicorp/vault/api"
)

type FakeVaultClient struct {
	kvStore map[string](map[string]interface{})
}

func (c *FakeVaultClient) Read(path string) (*api.KVSecret, error) {
	data := c.kvStore[path]
	if data == nil {
		// Seems to be in line with vault api when making calls to non-existent paths
		return nil, nil
	}
	return &api.KVSecret{Data: data}, nil
}

// NewTestVault creates an unsealed in-memory vault and adds a static KV
// secret with a base64-encoded field `username` of value "postgres" to
// `kv/storage/postgres/creds`.
func NewTestVault(t *testing.T) *FakeVaultClient {
	t.Helper()
	c := &FakeVaultClient{}
	c.kvStore = make(map[string](map[string]interface{}))

	// populate with fake secret
	c.kvStore["storage/postgres/creds"] = map[string]interface{}{
		"username": "cG9zdGdyZXM=",
		"password": "4_5tr0ng_4nd_c0mpl1c4t3d_p455w0rd",
	}

	return c
}
