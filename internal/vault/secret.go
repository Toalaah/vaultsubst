package vault

import (
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/vault/api"
	"github.com/mitchellh/mapstructure"
	"github.com/toalaah/vaultsubst/internal/transformations"
)

// SecretSpec represents a single secret in a file to be patched.
type SecretSpec struct {
	Path            string   `mapstructure:"path"`
	Field           string   `mapstructure:"field"`
	B64             bool     `mapstructure:"b64"`
	MountVersion    string   `mapstructure:"ver"`
	Transformations []string `mapstructure:"transform"`
}

// FormatSecret returns a formatted secret value field from a vault KV secret,
// based on the spec's internally configured transformations.
func (spec *SecretSpec) FormatSecret(secret *api.KVSecret) (string, error) {
	var (
		res string
		err error
	)

	if secret == nil {
		return "", errors.New("secret is nil")
	}

	res, ok := secret.Data[spec.Field].(string)
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

// NewSecretSpec constructs and returns a new SecretSpec from a structured string s.
func NewSecretSpec(s string) (*SecretSpec, error) {
	// "path = ...,field = ..." => "path=...,field=...".
	s = strings.ReplaceAll(s, " ", "")
	// "path=...,field=..." => ["path=...", "field=..."].
	attrs := strings.Split(s, ",")

	m := make(map[string]string)
	for _, v := range attrs {
		// "path=..." => ["path", "..."].
		kv := strings.Split(v, "=")
		if len(kv) != 2 {
			return nil, fmt.Errorf("Unable to parse option: %s (value %s)", s, v)
		}
		m[kv[0]] = kv[1]
	}

	spec := new(SecretSpec)

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		WeaklyTypedInput: true,
		Result:           &spec,
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
	if spec.Path == "" {
		return nil, fmt.Errorf("Path may not be empty")
	}
	if spec.Field == "" {
		return nil, fmt.Errorf("Field may not be empty")
	}
	// Default to KVv2 unless specified otherwise.
	if spec.MountVersion == "" {
		spec.MountVersion = KVv2
	}

	return spec, nil
}
