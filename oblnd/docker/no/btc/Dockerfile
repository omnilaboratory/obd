FROM ubuntu:22.04

# RUN apt-get update
#
# RUN apt-get -y install curl

# RUN curl -OL https://bitcoin.org/bin/bitcoin-core-0.16.0/bitcoin-0.16.0-x86_64-linux-gnu.tar.gz
#
# RUN tar zxvf bitcoin-0.16.0-x86_64-linux-gnu.tar.gz

# RUN ln -s /bitcoin-0.16.0/bin/bitcoin-cli /bitcoin-cli

COPY bin/bitcoind /usr/bin/
COPY bin/bitcoin-cli /usr/bin/
COPY bitcoin.conf /root/.bitcoin/bitcoin.conf

# p2p
EXPOSE 18444/tcp
# rpc
EXPOSE 18443/tcp

ENTRYPOINT ["bitcoind", "-printtoconsole"]