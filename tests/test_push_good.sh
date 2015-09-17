#!/bin/bash -eu

. $(dirname $0)/setup.sh

log "Testing /push"
assert_eq "{}" "$(curl -k -d '{"daily_active_users": 10, "timestamp": 20, "total_users": 123, "total_room_count": 17, "daily_messages": 9, "uptime_seconds": 19, "homeserver": "many.turtles"}' http://localhost:${port}/push 2>/dev/null)"
assert_eq "10|123|17|9|20|19" "$(sqlite3 ${dir}/stats.db 'SELECT daily_active_users, total_users, total_room_count, daily_messages, remote_timestamp, uptime_seconds FROM stats WHERE homeserver == "many.turtles"')"
assert_eq "1" "$(sqlite3 ${dir}/stats.db 'SELECT COUNT(*) AS count FROM stats WHERE homeserver == "many.turtles" AND (remote_addr LIKE "127.0.0.1%" OR remote_addr LIKE "::1%")')"

sleep 2
assert_eq "{}" "$(curl -k -d '{"daily_active_users": 456, "timestamp": 19, "homeserver": "many.turtles"}' http://localhost:${port}/push 2>/dev/null)"
assert_eq "10
456" "$(sqlite3 ${dir}/stats.db 'SELECT daily_active_users FROM stats WHERE homeserver == "many.turtles" ORDER BY daily_active_users ASC')"

assert_eq "456
10" "$(sqlite3 ${dir}/stats.db 'SELECT daily_active_users FROM stats WHERE homeserver == "many.turtles" ORDER BY remote_timestamp ASC')"
assert_eq "10
456" "$(sqlite3 ${dir}/stats.db 'SELECT daily_active_users FROM stats WHERE homeserver == "many.turtles" ORDER BY local_timestamp ASC')"

assert_eq "{}" "$(curl -k -d '{"homeserver": "few.turtles"}' http://localhost:${port}/push 2>/dev/null)"
assert_eq "|||||" "$(sqlite3 ${dir}/stats.db 'SELECT daily_active_users, total_users, total_room_count, daily_messages, remote_timestamp, uptime_seconds FROM stats WHERE homeserver == "few.turtles"')"
