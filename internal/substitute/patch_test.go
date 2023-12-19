package substitute

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/toalaah/vaultsubst/testutil"
)

var defaultRegex = regexp.MustCompile(fmt.Sprintf(`%s(?P<Data>.*)%s`, "@@", "@@"))

func TestPatchFile(t *testing.T) {
	assert := assert.New(t)

	client := testutil.NewTestVault(t)

	expected, err := os.ReadFile("./fixtures/test.expected.txt")
	assert.Nil(err)

	f, err := os.Open("./fixtures/test.txt")
	assert.Nil(err)

	b, err := PatchSecretsInFile(f, defaultRegex, client)
	assert.Nil(err)

	assert.Equal(expected, b)
}
