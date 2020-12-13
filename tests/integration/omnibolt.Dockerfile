FROM golang:1.13-alpine as builder
RUN echo "ABBC"
ARG checkout="working-local-test"

WORKDIR /go

# Install dependencies and build the binaries.
RUN apk add --no-cache --update alpine-sdk \
    git \
    make \
    gcc \
    bash \
    && git clone https://github.com/johng/obd.git \
    && cd obd \
    && git checkout ${checkout}


RUN cd obd && go build obdserver.go && go build tracker_server.go
COPY start.sh .
COPY conf.tracker.ini /go/conf.tracker.ini
ENTRYPOINT [ "./start.sh" ]