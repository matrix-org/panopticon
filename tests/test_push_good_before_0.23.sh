#!/bin/bash -eu

. $(dirname $0)/setup.sh
log "Testing /push with 0.15.x - 0.23.2 pushes"
assert_eq "{}" "$(curl -k -d '{"daily_active_users": 10, "timestamp": 20, "total_users": 123, "total_room_count": 17, "daily_messages": 9, "uptime_seconds": 19, "homeserver": "many.turtles"}' http://localhost:${port}/push 2>/dev/null)"
assert_eq "10|123|17|9|20|19" "$(sqlite3 ${dir}/stats.db 'SELECT daily_active_users, total_users, total_room_count, daily_messages, remote_timestamp, uptime_seconds FROM stats WHERE homeserver == "many.turtles"')"
