package vault

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/hashicorp/vault/api"
)

// Client simply wraps a vault client. If fullfills the SecretReader
// interface.
type Client struct {
	*api.Client
}

// SecretReader is an interface which groups anything able to read and write
// vault data.
type SecretReader interface {
	Read(path string) (*api.Secret, error)
}

func (c *Client) Read(path string) (*api.Secret, error) {
	return c.Logical().Read(path)
}

// NewClient returns a new vault client. It handles address and token
// initialization. It returns any errors encountered during construction of
// it's nested client or if it is unable to properly configure the client
// token.
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
