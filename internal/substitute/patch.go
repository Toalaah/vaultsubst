package substitute

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	vault "github.com/hashicorp/vault/api"

	"github.com/toalaah/vaultsubst/internal/secret"
)

func PatchSecretsInFile(file string, regexp *regexp.Regexp, client *vault.Client, inPlace bool) error {
	f, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	s := string(f)
	matches := regexp.FindAllStringSubmatch(s, -1)
	for _, match := range matches {
		originalContent := match[0]
		spec, err := secret.NewSecretSpec(match[1])
		if err != nil {
			return err
		}
		secret, err := spec.Fetch(client)
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
