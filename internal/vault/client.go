package vault

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/hashicorp/vault/api"
)

// Client thinly wraps a vault client. It provides a minimal subset of
// functionality required for interacting with KV stores.
type Client struct {
	Client KVReader
}

const (
	KVv1 = "v1"
	KVv2 = "v2"
)

type KVReader interface {
	ReadKVv1(mount, path string) (*api.KVSecret, error)
	ReadKVv2(mount, path string) (*api.KVSecret, error)
}

type apiClient struct {
	reader *api.Client
}

func (c *apiClient) ReadKVv1(mount, path string) (*api.KVSecret, error) {
	return c.reader.KVv1(mount).Get(context.Background(), path)
}

func (c *apiClient) ReadKVv2(mount, path string) (*api.KVSecret, error) {
	return c.reader.KVv2(mount).Get(context.Background(), path)
}

func (c *Client) ReadKV(spec *SecretSpec) (*api.KVSecret, error) {
	split := strings.Split(spec.Path, "/")
	mnt := split[0]
	// Extra check for second element being empty cause both 'kv/' and 'kv'
	// should be invalid paths.
	if len(split) < 2 || split[1] == "" {
		return nil, fmt.Errorf("no path to query using mountpoint %s", mnt)
	}
	pth := strings.TrimPrefix(spec.Path, mnt+"/")
	switch spec.MountVersion {
	case KVv1:
		return c.Client.ReadKVv1(mnt, pth)
	case KVv2:
		return c.Client.ReadKVv2(mnt, pth)
	}
	return nil, fmt.Errorf("secret %+v: unknown kv version %s", spec, spec.MountVersion)
}

// NewClient returns a new vault client. Address and token initialization are
// handled internally. Any errors encountered during initialization (for
// instance due to lacking environment variables) are returned to the caller.
func NewClient() (*Client, error) {
	c := &Client{}

	api, err := api.NewClient(nil)
	if err != nil {
		return nil, err
	}

	// Try to read from ~/.vault-token if env var is not supplied
	if os.Getenv("VAULT_TOKEN") == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("VAULT_TOKEN unset and/or failed to read token from ~/.vault-token: %s", err)
		}
		token, err := os.ReadFile(path.Join(homeDir, ".vault-token"))
		if err != nil {
			return nil, fmt.Errorf("VAULT_TOKEN unset and/or failed to read token from ~/.vault-token: %s", err)
		}
		api.SetToken(strings.TrimSuffix(string(token), "\n"))
	}

	c.Client = &apiClient{reader: api}
	return c, nil
}
