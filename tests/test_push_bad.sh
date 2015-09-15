#!/bin/bash -eu

. $(dirname $0)/setup.sh

log "Testing /push with bad input"
assert_eq '{"error_message": "unable to process request"}' "$(curl -k -d "not an object" http://localhost:${port}/push 2>/dev/null)"
assert_eq '{"error_message": "unable to process request"}' "$(curl -k -d "123" http://localhost:${port}/push 2>/dev/null)"
