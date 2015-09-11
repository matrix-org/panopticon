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

openssl req -x509 -newkey rsa:2048 -keyout ${dir}/tls.key -out ${dir}/tls.crt -nodes -days 10 -subj "/C=/ST=/L=/O=/CN=localhost" 2>/dev/null

cd $(dirname $(dirname $(realpath $0)))
./panopticon --port=${port} --cert_file=${dir}/tls.crt --key_file=${dir}/tls.key --db=${dir}/stats.db 2>$1 &
function kill_server {
  kill $(lsof -i | grep ":${port} " | awk '{print $2}')
}

trap kill_server EXIT

log_verbose "Waiting for server to come up"
until curl -k https://localhost:${port}/healthz >/dev/null 2>/dev/null; do
  sleep 0.1
done
log_verbose "Server came up"

