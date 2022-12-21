#!/bin/bash

Target="$1"
if [ -z "$Target" ]; then
  echo "miss address"
    exit 1
fi

set -x
omnicore-cli sendtoaddress $1  1
omnicore-cli omni_send mtowceAw2yeftR1pPg15QcsDqsnSik7Spz "$1"  2147483651  100
omnicore-cli generatetoaddress 3 mtowceAw2yeftR1pPg15QcsDqsnSik7Spz
omnicore-cli omni_getbalance  $1 2147483651