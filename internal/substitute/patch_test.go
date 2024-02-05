package substitute

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/toalaah/vaultsubst/testutil"
)

func TestPatchFileGeneric(t *testing.T) {
  testPatchFileImpl(t, "generic")
}

func TestPatchFileMultipleMatchSingleLine(t *testing.T) {
  testPatchFileImpl(t, "multiple-match-single-line")
}

func testPatchFileImpl(t *testing.T, file string) {
	assert := assert.New(t)

	client := testutil.NewTestVault(t)

	expected, err := os.ReadFile(fmt.Sprintf("./fixtures/%s.expected.txt", file))
	assert.Nil(err)

	f, err := os.Open(fmt.Sprintf("./fixtures/%s.txt", file))
	assert.Nil(err)

	b, err := PatchSecretsInFile(f, regexp.MustCompile(`@@(.*?)@@`), client)
	assert.Nil(err)

	assert.Equal(expected, b)
}
