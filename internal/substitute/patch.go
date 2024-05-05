package substitute

import (
	"io"
	"regexp"
	"strings"

	"github.com/toalaah/vaultsubst/internal/vault"
)

func PatchSecrets(r io.Reader, regexp *regexp.Regexp, client *vault.Client) ([]byte, error) {
	f, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	s := string(f)
	matches := regexp.FindAllStringSubmatch(s, -1)
	for _, match := range matches {
		originalContent := match[0]
		spec, err := vault.NewSecretSpec(match[1])
		if err != nil {
			return nil, err
		}
		res, err := client.ReadKV(spec)
		if err != nil {
			return nil, err
		}
		secret, err := spec.FormatSecret(res)
		if err != nil {
			return nil, err
		}
		s = strings.Replace(s, originalContent, secret, -1)
	}
	return []byte(s), nil
}
