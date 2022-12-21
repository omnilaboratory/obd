set -x
docker compose exec -u 0 btc  bitcoin-cli --regtest $@