 FROM ubuntu:20.04 as baseImage
COPY omnicore/sources.list /etc/apt/sources.list
RUN apt-get update
#RUN apt-get install iputils-ping
RUN apt-get install  libevent-dev libboost-system-dev libboost-filesystem-dev libboost-test-dev libboost-thread-dev libdb5.3++-dev libdb-dev libdb++-dev libzmq5 -y

FROM baseImage
RUN mkdir /root/.bitcoin/
ADD omnicore/initdata.tar.gz /root/.bitcoin/
RUN echo "hello world." >> /root/.bitcoin/greeting

COPY omnicore/send_coin.sh /bin/
RUN  chmod a+x  /bin/send_coin.sh

# Copy the compiled binaries from the builder image.
ADD  tmp/omnicored.tar.gz /bin/

EXPOSE 18332 28332 28333

ENTRYPOINT ["omnicored"]

