#!/bin/sh

container=$(docker run --rm -d -p 8200:8200 vault:1.13.3)
proc=$$

echo "Container ID: $container"

token=""
while [ -z "$token" ]; do
  sleep 1
  token=$(docker logs $container 2>/dev/null | grep 'Root Token:' | cut -d' ' -f3)
done

export VAULT_ADDR=http://127.0.0.1:8200
export VAULT_TOKEN=$token

echo -e "\nEnabling KV engine and populating demo secret at kv/storage/postgres/creds"

vault secrets enable -version=2 kv >/dev/null
vault kv put kv/storage/postgres/creds \
  username="$(echo -n postgres | base64)" \
  password="4_5tr0ng_4nd_c0mpl1c4t3d_p455w0rd" >/dev/null

echo "Done"

cat << EOF

You can now interact with the vault server by exporting the following
variables:

export VAULT_ADDR=http://127.0.0.1:8200
export VAULT_TOKEN=$token

To stop the server gracefully send a SIGINT to $proc or '^C' this script
EOF

cleanup() {
  echo -e "\nCaught interrupt, performing cleanup..."
  docker stop $container 2>/dev/null 1>&2
  echo "Done"
  exit 0
}

trap cleanup INT

sleep infinity
