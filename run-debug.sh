#!/usr/bin/env bash

rm ordx-debug
go build -o ordx-debug

if [ $# -eq 0 ]; then
  nohup ./ordx-debug &
  disown
else
  if [ "$1" = "off" ]; then
    ./ordx-debug
  else
    echo "unknown parameter"
  fi
fi
