#!/bin/bash -eu

. $(dirname $0)/setup.sh
log "Testing /push with 0.15.x - 0.23.2 pushes"
assert_eq "{}" "$(curl -k -d '{"daily_active_users": 10, "timestamp": 20, "total_users": 123, "total_room_count": 17, "daily_messages": 9, "uptime_seconds": 19, "r30_users_all": 5, "r30_users_android": 4, "r30_users_ios": 3, "r30_users_electron": 2, "r30_users_web": 1,"homeserver": "many.turtles"}' http://localhost:${port}/push 2>/dev/null)"
assert_eq "10|123|17|9|20|19|5|4|3|2|1" "$(sqlite3 ${dir}/stats.db 'SELECT daily_active_users, total_users, total_room_count, daily_messages, remote_timestamp, uptime_seconds, r30_users_all, r30_users_android, r30_users_ios, r30_users_electron, r30_users_web FROM stats WHERE homeserver == "many.turtles"')"
