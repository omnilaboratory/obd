#!/bin/bash

set -e

# generate compiles the *.pb.go stubs from the *.proto files.
function generate() {
  echo "Generating root gRPC server protos"

  PROTOS="*.proto "

  # For each of the sub-servers, we then generate their protos, but a restricted
  # set as they don't yet require REST proxies, or swagger docs.
  for file in $PROTOS; do
    DIRECTORY=$(dirname "${file}")
    echo "Generating protos from ${file}, into ${DIRECTORY}"

    # Generate the protos. --go-grpc_opt=require_unimplemented_servers=false
    protoc -I/usr/local/include -I. \
      --go_out . --go_opt paths=source_relative \
      --go-grpc_out . --go-grpc_opt paths=source_relative \
      "${file}"

    # Generate the REST reverse proxy.
    annotationsFile=${file//proto/yaml}
    protoc -I/usr/local/include -I. \
      --grpc-gateway_out . \
      --grpc-gateway_opt logtostderr=true \
      --grpc-gateway_opt paths=source_relative \
      --grpc-gateway_opt grpc_api_configuration=${annotationsFile} \
      "${file}"

    # Generate the swagger file which describes the REST API in detail.
    protoc -I/usr/local/include -I. \
      --openapiv2_out ./swagger/ \
      --openapiv2_opt logtostderr=true \
      --openapiv2_opt grpc_api_configuration=${annotationsFile} \
      --openapiv2_opt json_names_for_fields=false \
      --openapiv2_opt enums_as_ints=true \
      "${file}"
  done
}
generate
