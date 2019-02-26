#!/bin/bash -eu
# You almost certainly do not want to add more tests to this file
. $(dirname $0)/setup.sh
log "Testing /push with 0.27.2 - 0.33.5 pushes"
assert_eq "{}" "$(curl -k -d '{"daily_active_users": 10, "timestamp": 20, "total_users": 123, "total_room_count": 17, "daily_messages": 9, "uptime_seconds": 19, "r30_users_all": 5, "r30_users_android": 4, "r30_users_ios": 3, "r30_users_electron": 2, "r30_users_web": 1, "daily_user_type_native": 21,  "daily_user_type_guest": 22, "daily_user_type_bridged": 23, "homeserver": "many.turtles", "memory_rss": 12, "cpu_average": 125, "cache_factor": 5.501, "event_cache_size": 10000}' http://localhost:${port}/push 2>/dev/null)"
assert_eq "10|123|17|9|20|19|5|4|3|2|1|21|22|23|125|12|5.501|10000" "$(sqlite3 ${dir}/stats.db 'SELECT daily_active_users, total_users, total_room_count, daily_messages, remote_timestamp, uptime_seconds, r30_users_all, r30_users_android, r30_users_ios, r30_users_electron, r30_users_web, daily_user_type_native, daily_user_type_guest, daily_user_type_bridged, cpu_average, memory_rss, cache_factor, event_cache_size FROM stats WHERE homeserver == "many.turtles"')"


