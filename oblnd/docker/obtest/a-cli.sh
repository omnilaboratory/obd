set -x
docker compose exec -u 1000 alice lncli-debug -n regtest  $@