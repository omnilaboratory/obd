set -x
docker compose exec -u 1000 carl lncli-debug -n regtest  $@