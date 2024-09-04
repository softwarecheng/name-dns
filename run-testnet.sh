#!/usr/bin/env bash

rm ordx-testnet
go build -o ordx-testnet

if [ $# -eq 0 ]; then
  nohup ./ordx-testnet &
  disown
else
  if [ "$1" = "off" ]; then
    ./ordx-testnet
  else
    echo "unknown parameter"
  fi
fi

