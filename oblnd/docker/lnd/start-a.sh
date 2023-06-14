#!/usr/bin/env bash

# exit from script if error was raised.
set -e

# error function is used within a bash function in order to send the error
# message directly to the stderr output and exit.
error() {
    echo "$1" > /dev/stderr
    exit 0
}

# return is used within bash function in order to return the value.
return() {
    echo "$1"
}

# set_default function gives the ability to move the setting of default
# env variable from docker file to the script thereby giving the ability to the
# user override it during container start.
set_default() {
    # docker initialized env variables with blank string and we can't just
    # use -z flag as usually.
    BLANK_STRING='""'

    VARIABLE="$1"
    DEFAULT="$2"

    if [[ -z "$VARIABLE" || "$VARIABLE" == "$BLANK_STRING" ]]; then

        if [ -z "$DEFAULT" ]; then
            error "You should specify default variable"
        else
            VARIABLE="$DEFAULT"
        fi
    fi

   return "$VARIABLE"
}

# Set default variables if needed.
RPCUSER=$(set_default "$RPCUSER" "omniwallet")
RPCPASS=$(set_default "$RPCPASS" "cB3]iL2@eZ1?cB2?")
DEBUG=$(set_default "$DEBUG" "info")
NETWORK=$(set_default "$NETWORK" "regtest")
CHAIN=$(set_default "$CHAIN" "bitcoin")
#BACKEND="btcd"
#BACKEND="bitcoind"
BACKEND="omnicoreproxy"
HOSTNAME=$(hostname)
if [[ "$CHAIN" == "litecoin" ]]; then
    BACKEND="ltcd"
fi

# CAUTION: DO NOT use the --noseedback for production/mainnet setups, ever!
# Also, setting --rpclisten to $HOSTNAME will cause it to listen on an IP
# address that is reachable on the internal network. If you do this outside of
# docker, this might be a security concern!
set -x
exec ./lnd-debug \
    --autopilot.active \
    --maxpendingchannels=100 \
    --noseedbackup \
    "--lnddir=~/apps/oblnd" \
    --alias=local_a \
    "--$CHAIN.active" \
    "--$CHAIN.$NETWORK" \
    "--$CHAIN.node"="$BACKEND" \
    "--$BACKEND.rpchost"="43.138.107.248:18332" \
    "--rpclisten=$HOSTNAME:10010" \
    "--rpclisten=localhost:10010" \
    --listen=0.0.0.0:9735 \
    "--restlisten=0.0.0.0:18080" \
    --debuglevel="$DEBUG" \
    --$BACKEND.zmqpubrawblock=tcp://43.138.107.248:28332 \
    --$BACKEND.zmqpubrawtx=tcp://43.138.107.248:28333 \
      --nobootstrap \
    "$@"

# ./lncli-debug --rpcserver=localhost:10010 --lnddir=~/apps/oblnd -n testnet

# make build with-rpc=true

#./lnd-debug --noseedbackup '--lnddir=~/apps/oblnd' --bitcoin.active --bitcoin.testnet --bitcoin.node=bitcoind --bitcoind.rpchost=43.138.107.248:18332 --bitcoind.rpcuser=omniwallet '--bitcoind.rpcpass=cB3]iL2@eZ1?cB2?' --rpclisten=wxf-d:10010 --rpclisten=localhost:10010 --debuglevel=debug --bitcoind.zmqpubrawblock=tcp://43.138.107.248:28332 --bitcoind.zmqpubrawtx=tcp://43.138.107.248:28333

#todo  walletbalance add asset
# walletbalance list by address


