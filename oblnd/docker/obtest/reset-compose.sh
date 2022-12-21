./clean-compose.sh
mkdir -p ./volumes/lnd/alice  ./volumes/lnd/bob ./volumes/lnd/carl ./volumes/omnicored
#docker compose up omnicored -d
#sleep 5
docker compose up -d
docker compose  ps
./mine.sh