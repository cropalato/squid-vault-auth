https://gist.github.com/teknoraver/5ffacb8757330715bcbcc90e6d46ac74


running in docker vault-dev

apk add --no-cache gcompat

export VAULT_ADDR=http://127.0.0.1:8200
export VAULT_TOKEN=hvs.KBJ4mqVesjO2XKEukITIfPGC

# Add that in policy
path "sys/plugins/catalog/database" {
  capabilities = ["read", "update", "create", "list", "delete", "sudo"]
}

chown -R vault /vault_plugin

vault secrets enable database

vault write sys/plugins/catalog/database/squiddb sha256="$(sha256sum /vault_plugin/squid-database-plugin|cut -d" " -f1)" command=squid-database-plugin

vault write database/config/squiddb plugin_name=squiddb allowed_roles="readonly" connection_url="http://10.0.0.81:8080" username="admin" password="secret"

vault write database/roles/readonly db_name=squiddb default_ttl="5m" max_ttl="24h"

vault read database/creds/readonly
