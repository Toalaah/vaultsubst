package substitute

import (
	"io"
	"regexp"
	"strings"

	"github.com/toalaah/vaultsubst/internal/secret"
	"github.com/toalaah/vaultsubst/internal/vault"
)

func PatchSecretsInFile(r io.Reader, regexp *regexp.Regexp, client vault.SecretReader) ([]byte, error) {
	f, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	s := string(f)
	matches := regexp.FindAllStringSubmatch(s, -1)
	for _, match := range matches {
		originalContent := match[0]
		spec, err := secret.NewSecretSpec(match[1])
		if err != nil {
			return nil, err
		}
		secret, err := spec.Fetch(client)
		if err != nil {
			return nil, err
		}
		s = strings.Replace(s, originalContent, secret, -1)
	}

	return []byte(s), nil
}
