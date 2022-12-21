set -x
docker compose exec -u 1000 bob lncli-debug -n regtest  $@