mkdir -p ./volumes/lnd/alice  ./volumes/lnd/bob ./volumes/lnd/carl ./volumes/omnicored
docker compose up -d
docker compose  ps
./mine.sh
set -x
sleep 5
./a-cli.sh newaddress
./b-cli.sh newaddress
./c-cli.sh newaddress
export asset_id=2147483651