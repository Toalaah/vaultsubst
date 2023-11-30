package vault

import (
	"os"

	vault "github.com/hashicorp/vault/api"
)

type VaultData map[string]interface{}

func NewClient() (*vault.Client, error) {
	client, err := vault.NewClient(&vault.Config{
		Address: os.Getenv("VAULT_ADDR"),
	})
	if err != nil {
		return nil, err
	}
	client.SetToken(os.Getenv("VAULT_TOKEN"))
	return client, nil
}
