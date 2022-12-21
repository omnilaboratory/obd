FROM ubuntu:22.04 as baseImage
COPY omnicore/sources.list /etc/apt/sources.list
COPY omnicore/apt.conf /etc/apt/apt.conf
RUN apt-get update -y
#RUN apt-get install iputils-ping
RUN apt-get install gosu -y
RUN apt-get install  libevent-dev libboost-system-dev libboost-filesystem-dev libboost-test-dev libboost-thread-dev libdb5.3++-dev libdb-dev libdb++-dev libzmq5 -y
#libdb5.3++-dev libdb-dev libdb++-dev

FROM baseImage
ADD omnicore/initdata.tar.gz /initdata/

COPY omnicore/send_coin.sh /bin/
RUN  chmod a+x  /bin/send_coin.sh

COPY omnicore/polar/docker-entrypoint.sh /entrypoint.sh
RUN chmod a+x /entrypoint.sh

# Copy the compiled binaries from the builder image.
ADD  tmp/omnicored.tar.gz /bin/
RUN ln -s /bin/omnicored /bin/bitcoind
RUN ln -s /bin/omnicore-cli /bin/bitcoin-cli

#VOLUME ["/home/bitcoin/.bitcoin"]
EXPOSE 18332 28332 28333

ENTRYPOINT ["/entrypoint.sh"]
CMD ["omnicored"]

