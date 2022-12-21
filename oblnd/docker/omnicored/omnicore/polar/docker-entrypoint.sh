#!/bin/sh
set -e

# containers on linux share file permissions with hosts.
# assigning the same uid/gid from the host user
# ensures that the files can be read/write from both sides
if ! id bitcoin > /dev/null 2>&1; then
  USERID=${USERID:-1000}
  GROUPID=${GROUPID:-1000}

  echo "adding user bitcoin ($USERID:$GROUPID)"
  groupadd -f -g $GROUPID bitcoin
  useradd -r -u $USERID -g $GROUPID bitcoin

  #init data
  set -e
  SOURCE_DIR=/initdata/
  TARGET_DIR="/home/bitcoin/.bitcoin"
  mkdir -p $TARGET_DIR
  chown bitcoin  $TARGET_DIR
  if [ $(find $TARGET_DIR -maxdepth 0 -type d -empty 2>/dev/null) ]; then
     cp -r --preserve=all $SOURCE_DIR/* $TARGET_DIR/
  fi
fi


if [ $(echo "$1" | cut -c1) = "-" ]; then
  echo "$0: assuming arguments for omnicored"

  set -- omnicored "$@"
fi

if [ "$1" = "omnicored" ] || [ "$1" = "omnicore-cli" ] || [ "$1" = "bitcoin-tx" ]; then
  echo "Running as omnicored user: $@"
  exec gosu bitcoin "$@"
fi

echo "$@"
exec "$@"
