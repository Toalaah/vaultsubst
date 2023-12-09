package secret

import (
	"fmt"
	"strings"

	vault "github.com/hashicorp/vault/api"
	"github.com/mitchellh/mapstructure"
	"github.com/toalaah/vaultsubst/internal/transformations"
)

type VaultData map[string]interface{}

// SecretSpec represents a single secret in a file to be patched.
type SecretSpec struct {
	Path            string   `mapstructure:"path"`
	Field           string   `mapstructure:"field"`
	B64             bool     `mapstructure:"b64"`
	Transformations []string `mapstructure:"transform"`
}

// FormatSecret returns a formatted secret from "raw" Vault data, based on the
// Spec's configured transformations
func (spec *SecretSpec) FormatSecret(data VaultData) (string, error) {
	var (
		res string
		err error
	)

	res, ok := data[spec.Field].(string)
	if !ok {
		return "", fmt.Errorf("could not cast data at field %s to string", spec.Field)
	}

	if spec.B64 {
		res, err = transformations.Apply("base64d", res)
		if err != nil {
			return "", err
		}
	}

	for _, t := range spec.Transformations {
		res, err = transformations.Apply(t, res)
		if err != nil {
			return "", err
		}
	}

	return res, nil
}

// NewSecretSpec constructs and returns a new SecretSpec from a structured
// string.
func NewSecretSpec(s string) (*SecretSpec, error) {
	// ["path=...", "field=..."]
	specs := strings.Split(s, ",")

	m := make(map[string]interface{})
	for _, v := range specs {
		// "path=..." => ["path", "..."]
		kv := strings.Split(v, "=")
		if len(kv) != 2 {
			return nil, fmt.Errorf("Unable to parse option: %s (value %s)", s, v)
		}
		m[kv[0]] = kv[1]
	}

	result := &SecretSpec{}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		WeaklyTypedInput: true,
		Result:           &result,
		// Since we use commas as a field separator, arrays are assigned pipes
		// instead. Semantically speaking, this may even be desirable as multiple
		// transformations will be piped in order anyways.
		DecodeHook: mapstructure.StringToSliceHookFunc("|"),
	})

	if err != nil {
		return nil, err
	}

	if err := decoder.Decode(m); err != nil {
		return nil, err
	}

	// Some light validation on the decoded spec string. Without a path/field to
	// query, we are kind of useless.
	if result.Path == "" {
		return nil, fmt.Errorf("Path may not be empty")
	}
	if result.Field == "" {
		return nil, fmt.Errorf("Field may not be empty")
	}

	return result, nil
}

// Fetch fetches and returns a formatted vault secret string from a SecretSpec
func (spec *SecretSpec) Fetch(client *vault.Client) (string, error) {
	path := strings.TrimPrefix(spec.Path, "kv/")
	secret, err := client.Logical().Read("kv/data/" + path)
	if err != nil {
		return "", err
	}
	if secret == nil {
		return "", fmt.Errorf("secret is nil")
	}
	data := secret.Data["data"].(map[string]interface{})
	return spec.FormatSecret(data)
}
