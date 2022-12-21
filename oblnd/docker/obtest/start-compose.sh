mkdir -p ./volumes/lnd/alice  ./volumes/lnd/bob ./volumes/lnd/carl ./volumes/omnicored
docker compose up -d
docker compose  ps
./mine.sh