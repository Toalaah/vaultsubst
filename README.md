# vaultsubst

`vaultsubst` is a tool for injecting and optionally formatting vault KV secrets
into files. It acts similarly to `envsubst`, but instead of environment
variables it uses vault secrets. This is primarily useful if you have a large
number of arbitrary KV-paths that you want to query and don't necessarily know
ahead of time and/or are unable to set environment variables. It works by
extracting structured strings out of an input file which are then used to query
and finally inject the correct secret into the file (more precisely, `stdout`
by default).

`vaultsubst` also supports basic formatting options which are applied
sequentially to each secret prior to injection. The following transformations
are currently supported:

- `upper`: convert secret to uppercase
- `lower`: convert secret to lowercase
- `base64`: encode in base64
- `base64d`: decode from base64
- `trim`: trim leading and trailing white-spaces

As it is quite common that secrets are stored in base64, an additional option
`b64` can be supplied separately from `transform` to indicate that the fetched
secret should be *decoded* as such once fetched (this is equivalent to
specifying `transform=base64d`).

## Install

```bash
go install github.com/toalaah/vaultsubst@latest
```

## Example

```bash
$ cat test.yml
apiVersion: v1
data:
  username: "@@path=kv/storage/postgres/creds,field=username,b64=true,transform=trim|upper@@"
  password: "@@path=kv/storage/postgres/creds,field=password@@"
  test: "static dont change"
kind: Secret
metadata:
  name: supersecretdata
type: Opaque

$ vaultsubst --delim=@@ test.yml
apiVersion: v1
data:
  username: "POSTGRES"
  password: "4_5tr0ng_4nd_c0mpl1c4t3d_p455w0rd"
  test: "static dont change"
kind: Secret
metadata:
  name: supersecretdata
type: Opaque

```


## Roadmap

- [x] read from stdin if no file passed (or add check for '-' arg)
- [x] tests
- [x] basic CI
