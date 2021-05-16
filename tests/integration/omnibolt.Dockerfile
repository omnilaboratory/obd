FROM golang:1.13-alpine as builder

WORKDIR /obd

# Install dependencies and build the binaries.
RUN apk add --no-cache --update alpine-sdk \
    git \
    make \
    gcc \
    bash


COPY bean /obd/bean
COPY config /obd/config
COPY conn /obd/conn
COPY dao /obd/dao
COPY lightclient /obd/lightclient
COPY omnicore /obd/omnicore
COPY service /obd/service
COPY tracker /obd/tracker
COPY tool /obd/tool
COPY proxy /obd/proxy
COPY admin /obd/admin
COPY obdserver.go /obd/obdserver.go
COPY go.mod /obd/go.mod

RUN go build /obd/obdserver.go && go build /obd/tracker/tracker_server.go
COPY tests/integration/start.sh /obd/start.sh
COPY tests/integration/conf.tracker.ini /obd/conf.tracker.ini
ENTRYPOINT [ "/obd/start.sh" ]