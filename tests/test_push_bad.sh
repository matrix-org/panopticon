#!/bin/bash -eu

. $(dirname $0)/setup.sh

log "Testing /push with bad input"
assert_eq '{"errcode": "unable to process request"}' "$(curl -k -d "not an object" https://localhost:${port}/push 2>/dev/null)"
assert_eq '{"errcode": "unable to process request"}' "$(curl -k -d "123" https://localhost:${port}/push 2>/dev/null)"
