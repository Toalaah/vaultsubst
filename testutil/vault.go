package testutil

import (
	"net"
	"testing"

	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/http"
	"github.com/hashicorp/vault/vault"
)

// NewTestVault  creates an unsealed in-memory vault and adds a static KV
// secret with a base64-encoded field `username` of value "postgres" to
// `kv/storage/postgres/creds`.
//
// Adapted from https://stackoverflow.com/questions/57771228/mocking-hashicorp-vault-in-go
func NewTestVault(t *testing.T) (*api.Client, net.Listener) {
	t.Helper()

	core, keyShares, rootToken := vault.TestCoreUnsealed(t)
	_ = keyShares

	ln, addr := http.TestServer(t, core)

	conf := api.DefaultConfig()
	conf.Address = addr

	client, err := api.NewClient(conf)
	if err != nil {
		t.Fatal(err)
	}

	client.SetToken(rootToken)

	err = client.Sys().Mount("kv/", &api.MountInput{Type: "kv-v2"})
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.Logical().Write("kv/data/storage/postgres/creds", map[string]interface{}{
		"data": map[string]interface{}{
			"username": "cG9zdGdyZXM=",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	return client, ln
}
