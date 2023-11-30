package vault

import (
	"encoding/base64"
	"fmt"
	"os"
	"regexp"
	"strings"

	vault "github.com/hashicorp/vault/api"
	"github.com/mitchellh/mapstructure"
	"github.com/toalaah/vaultsubst/internal/transformations"
)

// SecretSpec represents a single secret in a file to be patched.
type SecretSpec struct {
	Path            string   `mapstructure:"path"`
	Field           string   `mapstructure:"field"`
	B64             bool     `mapstructure:"b64"`
	Transformations []string `mapstructure:"transformations"`
}

func (v *SecretSpec) FormatSecret(data VaultData) (string, error) {
	var (
		res string
		err error
	)

	res, ok := data[v.Field].(string)
	if !ok {
		return "", fmt.Errorf("could not cast data at field %s to string\n", v.Field)
	}

	if v.B64 {
		b, err := base64.StdEncoding.DecodeString(res)
		if err != nil {
			return "", err
		}
		res = string(b)
	}

	for _, t := range v.Transformations {
		res, err = transformations.Apply(t, res)
		if err != nil {
			return "", err
		}
	}

	return res, nil
}

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
		// instead. Semantially speaking, this may even be desireable as multiple
		// transformations will be piped in order anyways.
		DecodeHook: mapstructure.StringToSliceHookFunc("|"),
	})

	if err != nil {
		return nil, nil
	}

	err = decoder.Decode(m)
	return result, err
}

func GetSecretFromSpec(spec *SecretSpec, client *vault.Client) (string, error) {
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

func PatchSecretsInFile(file string, regexp *regexp.Regexp, client *vault.Client, inPlace bool) error {
	f, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	s := string(f)
	matches := regexp.FindAllStringSubmatch(s, -1)
	for _, match := range matches {
		originalContent := match[0]
		spec, err := NewSecretSpec(match[1])
		if err != nil {
			return err
		}
		secret, err := GetSecretFromSpec(spec, client)
		if err != nil {
			return err
		}
		s = strings.Replace(s, originalContent, secret, -1)
	}

	if inPlace {
		return os.WriteFile(file, []byte(s), 0644)
	} else {
		fmt.Fprint(os.Stdout, s)
	}

	return nil
}
