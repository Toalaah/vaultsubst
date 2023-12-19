package vault

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/hashicorp/vault/api"
)

// Client simply wraps a vault client. It satisfies the SecretReader interface.
type Client struct {
	*api.Client
}

// SecretReader is an interface describing anything able to read vault data.
type SecretReader interface {
	Read(path string) (*api.Secret, error)
}

func (c *Client) Read(path string) (*api.Secret, error) {
	return c.Logical().Read(path)
}

// NewClient returns a new vault client. Address and token initialization are
// handled internally. Any errors encountered during initialization (for
// instance due to lacking environment variables) are returned to the caller.
func NewClient() (*Client, error) {
	c := &Client{}
	vaultClient, err := api.NewClient(nil)
	if err != nil {
		return nil, err
	}
	c.Client = vaultClient

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
		c.SetToken(strings.TrimSuffix(string(token), "\n"))
	}
	return c, nil
}
