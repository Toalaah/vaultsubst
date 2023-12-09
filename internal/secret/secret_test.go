package secret

import (
	"errors"
	"net"
	"testing"

	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/http"
	"github.com/hashicorp/vault/vault"
	"github.com/stretchr/testify/assert"
)

func TestSecretSpecParsing(t *testing.T) {
	assert := assert.New(t)
	cases := []struct {
		SpecStr       string
		ExpectedValue *SecretSpec
		ExpectedErr   error
	}{
		{
			SpecStr: "path=kv/storage/postgres/creds,field=username,b64=true,transform=trim|upper",
			ExpectedValue: &SecretSpec{
				Path:            "kv/storage/postgres/creds",
				Field:           "username",
				B64:             true,
				Transformations: []string{"trim", "upper"},
			},
			ExpectedErr: nil,
		},
		{
			SpecStr: "path=kv/storage/postgres/creds,field=password",
			ExpectedValue: &SecretSpec{
				Path:            "kv/storage/postgres/creds",
				Field:           "password",
				B64:             false,
				Transformations: nil,
			},
			ExpectedErr: nil,
		},
		{
			SpecStr:       "path = kv/storage/postgres/creds,,fieldpassword",
			ExpectedValue: nil,
			ExpectedErr:   errors.New("Unable to parse option: path = kv/storage/postgres/creds,,fieldpassword (value )"),
		},
		{
			SpecStr:       "field=username,b64=true,transform=trim|upper",
			ExpectedValue: nil,
			ExpectedErr:   errors.New("Path may not be empty"),
		},
		{
			SpecStr:       "path=kv/storage/postgres/creds,b64=true,transform=trim|upper",
			ExpectedValue: nil,
			ExpectedErr:   errors.New("Field may not be empty"),
		},

		{
			SpecStr:       "path=kv/storage/postgres/creds,transform=trim,upper,b64d",
			ExpectedValue: nil,
			ExpectedErr:   errors.New("Unable to parse option: path=kv/storage/postgres/creds,transform=trim,upper,b64d (value upper)"),
		},
	}

	for _, c := range cases {
		ss, err := NewSecretSpec(c.SpecStr)
		assert.Equal(c.ExpectedValue, ss)
		assert.Equal(c.ExpectedErr, err)
	}
}

func TestSecretFetching(t *testing.T) {
	assert := assert.New(t)

	client, ln := newTestVault(t)
	defer ln.Close()

	cases := []struct {
		Spec          *SecretSpec
		ExpectedValue string
		ExpectedErr   error
	}{
		{
			Spec: &SecretSpec{
				Path:  "kv/storage/postgres/creds",
				Field: "username",
				B64:   true,
			},
			ExpectedValue: "postgres",
			ExpectedErr:   nil,
		},
		{
			Spec: &SecretSpec{
				Path:  "kv/storage/doesnotexist",
				Field: "username",
			},
			ExpectedValue: "",
			ExpectedErr:   errors.New("secret is nil"),
		},
		{
			Spec: &SecretSpec{
				Path:  "kv/storage/postgres/creds",
				Field: "doesnotexist",
			},
			ExpectedValue: "",
			ExpectedErr:   errors.New("could not cast data at field doesnotexist to string"),
		},
	}

	for _, c := range cases {
		secret, err := c.Spec.Fetch(client)
		assert.Equal(c.ExpectedValue, secret)
		assert.Equal(c.ExpectedErr, err)
	}
}

// Adapted from https://stackoverflow.com/questions/57771228/mocking-hashicorp-vault-in-go
// Creates an unsealed in-memory vault and adds a static KV secret to `kv/storage/postgres/creds`
func newTestVault(t *testing.T) (*api.Client, net.Listener) {
	t.Helper()

	core, keyShares, rootToken := vault.TestCoreUnsealed(t)
	_ = keyShares

	ln, addr := http.TestServer(t, core)

	conf := api.DefaultConfig()
	conf.Address = addr

	client, err := api.NewClient(conf)
	if err != nil {
		t.Fatal(err)
	}

	client.SetToken(rootToken)

	err = client.Sys().Mount("kv/", &api.MountInput{Type: "kv-v2"})
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.Logical().Write("kv/data/storage/postgres/creds", map[string]interface{}{
		"data": map[string]interface{}{
			"username": "cG9zdGdyZXM=",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	return client, ln
}
