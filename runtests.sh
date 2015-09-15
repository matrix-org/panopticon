#!/bin/bash -u

pass=0
fail=0
green="\e[32m"
red="\e[31m"
prefix="${green}"

log_dir=$(mktemp -d)

function reset_terminal_color {
  echo -ne "\033[0m"
}

cd $(dirname $(realpath $0))
set -e
go build || ( echo -e >&2 "${red}Build failed" ; reset_terminal_color ; exit 1 )
set +e

for t in $(dirname $0)/tests/test_*.sh; do
  log="${log_dir}/$(basename ${t})"
  if ${t} ${log} ; then
    echo -e >&2 "${green}Passed"
    reset_terminal_color
    pass=$((pass + 1))
  else
    fail=$((fail + 1))
    prefix="${red}"
    echo -e >&2 "${red}Failed"
    reset_terminal_color
    echo >&2 "Server log:"
    cat >&2 ${log}
  fi
done

echo -e "${prefix}${pass} test(s) passed, ${fail} test(s) failed"
reset_terminal_color
