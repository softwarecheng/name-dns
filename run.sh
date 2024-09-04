#!/usr/bin/env bash

rm ordx-mainnet
go build -o ordx-mainnet

if [ $# -eq 0 ]; then
  nohup ./ordx-mainnet &
  disown
else
  if [ "$1" = "off" ]; then
    ./ordx-mainnet
  else
    echo "unknown parameter"
  fi
fi

