# squid-vault-auth

Tools to use vault to manage temporary accounts for squid proxy.

## Tools

### squid-database

Rest API service responsible for maintaining the user database.
Vault should be able to call the server to create/delete users.

It can be configured using the following environment variables:

| Variable | Default | Description |
|--- | --- | --- |
| SQUIDDB_LISTEN | :8080 | IP and port used by squid db service. format: '[\<ip>]:\<port>'. default: ':8080' |
| SQUIDDB_USER | admin | admin account used to call squid db service API |
| SQUIDDB_PASS | hash(admin) | admin password used to call squid db service API. It is a bcrypt hash |
| SQUIDDB_PATH | /etc/squid-vault.json | squid db file path |
| SQUIDDB_CORS | * | configure Access-Control-Allow-Origin header |
| SQUIDDB_DEBUG | false | activate debug mode |

You can use the following command to generate a new password hash
```
python -c 'import bcrypt; print(bcrypt.hashpw(b"admin", bcrypt.gensalt(rounds=15)).decode("ascii"))'
```

### squid-database-auth

Tool used by squid to validate user http basic authentication.

| Variable | Default | Description |
|--- | --- | --- |
| SQUIDDB_URL | http://127.0.0.1:8080 | squid db service URL. format: 'http[s]://(\<fqdn>|\<ip>)[:\<port>]' |
| SQUIDDB_USER | admin | admin account used to call squid db service API |
| SQUIDDB_PASS | admin | admin password used to call squid db service API |

### squid-database-validator

Tool used by squid to check is a user is member of a specific group.

| Variable | Default | Description |
|--- | --- | --- |
| SQUIDDB_URL | http://127.0.0.1:8080 | squid db service URL. format: 'http[s]://(\<fqdn>|\<ip>)[:\<port>]' |
| SQUIDDB_USER | admin | admin account used to call squid db service API |
| SQUIDDB_PASS | admin | admin password used to call squid db service API |


### squid-database-plugin

Vault plugin used to integrate vault with squid-database.


## Building

On repository root directory, just run the command

```bash
go build -o bin ./...
```

## Testing it

This test will run two docker containers (squid proxy, vault). To make thing easy we will use four terminals and we will call them T1, T2, T3, T4.

### T1
You should run a squid proxy server.

- export IP where you are running squid-database
```
export SERVER_IP=<IP>
```
- run docker
```bash
docker run -it --rm --name squid-container -e TZ=UTC -e "SQUIDDB_URL=http://192.168.102.62:8080" -p 3128:3128 -v $(pwd)/docker/squid/squid.conf:/etc/squid/squid.conf -v $(pwd)/bin/squid-database-auth:/app/squid-database-auth -v $(pwd)/bin/squid-database-validator:/app/squid-database-validator ubuntu/squid
```

In squid.conf, you will find the following snippet:
```
auth_param basic program /squid-database-auth
auth_param basic children 5
auth_param basic realm Please enter your proxy server username and password
auth_param basic credentialsttl 2 hours
auth_param basic casesensitive on
```
to enable authentication, and

```
external_acl_type custom_acl ttl=10 %LOGIN /squid-database-validator
acl proxy_pac_authentication external custom_acl mygroup
```
to handle the group validation.


### T2
You should start squid-database
```
export SQUIDDB_PATH=/tmp/squid-vault.json
./bin/squid-database
```

### T3
You will start vault
```
docker run -it --rm --cap-add=IPC_LOCK -p 8200:8200 -v $(pwd)/docker/vault/policy/default.hcl/:/vault/policy/default.hcl -v $(pwd)/bin/squid-database-plugin:/vault/plugin/squid-database-plugin -v $(pwd)/docker/vault/config/:/vault/config -e SKIP_CHOWN=true --name=dev-vault hashicorp/vault
```
Please not, it will export a "Root Token" we will use in T4.

### T4
you will ...

- export token to be easy to copy and past the other commands
```
export MYTOKEN=<root token from T3>
```

- export IP where you are running squid-database
```
export SERVER_IP=<IP>
```

- full open vault, we don't care about that to our test.
```
docker exec -it -e "VAULT_ADDR=http://0.0.0.0:8200" -e "VAULT_TOKEN=${MYTOKEN}" dev-vault vault policy write default /vault/policy/default.hcl
```
- enable the docker container to execute /vault/plugin/squid-database-plugin
```
docker exec -it -e "VAULT_ADDR=http://0.0.0.0:8200" -e "VAULT_TOKEN=${MYTOKEN}" dev-vault apk add gcompat
```

- enable secret database
```
docker exec -it -e "VAULT_ADDR=http://0.0.0.0:8200" -e "VAULT_TOKEN=${MYTOKEN}" dev-vault vault secrets enable database
```

- load our plugin
```
docker exec -it -e "VAULT_ADDR=http://0.0.0.0:8200" -e "VAULT_TOKEN=${MYTOKEN}" dev-vault sh -c 'vault write sys/plugins/catalog/database/squiddb sha256="$(sha256sum /vault/plugin/squid-database-plugin|cut -d" " -f1)" command=squid-database-plugin'
```

- configure the plugin
```
docker exec -it -e "VAULT_ADDR=http://0.0.0.0:8200" -e "VAULT_TOKEN=${MYTOKEN}" dev-vault sh -c 'vault write database/config/squiddb plugin_name=squiddb allowed_roles="myrole" connection_url="http://'${SERVER_IP}':8080" username="admin" password="admin"'
```

- create a new role vault
```
docker exec -it -e "VAULT_ADDR=http://0.0.0.0:8200" -e "VAULT_TOKEN=${MYTOKEN}" dev-vault vault write database/roles/myrole db_name=squiddb default_ttl="5m" max_ttl="24h"
```
After 5 minutes the account will be removed.

- test the proxy without new account. Yes, it should fail.
```
https_proxy=http://127.0.0.1:3128 curl -si https://ifconfig.info
```

- generating a user
```
docker exec -it -e "VAULT_ADDR=http://0.0.0.0:8200" -e "VAULT_TOKEN=${MYTOKEN}" dev-vault vault read database/creds/myrole
```

- do the same with the account created by vault
```
https_proxy=http://<user>:<password>@127.0.0.1:3128 curl -si https://ifconfig.info
```

