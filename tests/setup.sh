#!/bin/bash -eu

function assert_eq {
  if [[ "$1" != "$2" ]]; then
    echo >&2 "$(caller): Expected \"$1\" to equal \"$2\""
    exit 1
  fi
}

function log {
  echo >&2 "$@"
}

function log_verbose {
  if [[ "${VERBOSE:-0}" == "1" ]]; then
    log "$@"
  fi
}

dir=$(mktemp -d)
trap "rm -rf ${dir}" EXIT

port=9002

cd $(dirname $(dirname $(realpath $0)))
./panopticon --port=${port} --db=${dir}/stats.db 2>$1 &
PID=$! 
function kill_server {
  kill $PID
}

trap kill_server EXIT

log_verbose "Waiting for server to come up"
until curl -k http://localhost:${port}/healthz >/dev/null 2>/dev/null; do
  sleep 0.1
done
log_verbose "Server came up"

