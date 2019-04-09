#!/bin/bash -eu

. $(dirname $0)/setup.sh
log "Testing /push with beyond 0.99.2 pushes"

assert_eq "{}" "$(curl -k -d '{"daily_active_users": 10, "timestamp": 20, "total_users": 123, "total_room_count": 17, "daily_messages": 9, "uptime_seconds": 19, "r30_users_all": 5, "r30_users_android": 4, "r30_users_ios": 3, "r30_users_electron": 2, "r30_users_web": 1, "daily_user_type_native": 21,  "daily_user_type_guest": 22, "daily_user_type_bridged": 23, "homeserver": "many.turtles", "memory_rss": 12, "cpu_average": 125, "cache_factor": 5.501, "event_cache_size": 10000, "python_version":"3.6.1", "database_engine":"PostgreSql", "database_server_version":"9.5.0", "server_context":"my_context"}' http://localhost:${port}/push 2>/dev/null)"

assert_eq "10|123|17|9|20|19|5|4|3|2|1|21|22|23|125|12|5.501|10000|3.6.1|PostgreSql|9.5.0|my_context" "$(sqlite3 ${dir}/stats.db 'SELECT daily_active_users, total_users, total_room_count, daily_messages, remote_timestamp, uptime_seconds, r30_users_all, r30_users_android, r30_users_ios, r30_users_electron, r30_users_web, daily_user_type_native, daily_user_type_guest, daily_user_type_bridged, cpu_average, memory_rss, cache_factor, event_cache_size, python_version, database_engine, database_server_version, server_context FROM stats WHERE homeserver == "many.turtles"')"


assert_eq "1" "$(sqlite3 ${dir}/stats.db 'SELECT COUNT(*) AS count FROM stats WHERE homeserver == "many.turtles" AND (remote_addr LIKE "127.0.0.1%" OR remote_addr LIKE "[::1]%")')"

sleep 2
assert_eq "{}" "$(curl -k -d '{"daily_active_users": 456, "timestamp": 19, "homeserver": "many.turtles"}' http://localhost:${port}/push 2>/dev/null)"
assert_eq "10
456" "$(sqlite3 ${dir}/stats.db 'SELECT daily_active_users FROM stats WHERE homeserver == "many.turtles" ORDER BY daily_active_users ASC')"

assert_eq "456
10" "$(sqlite3 ${dir}/stats.db 'SELECT daily_active_users FROM stats WHERE homeserver == "many.turtles" ORDER BY remote_timestamp ASC')"
assert_eq "10
456" "$(sqlite3 ${dir}/stats.db 'SELECT daily_active_users FROM stats WHERE homeserver == "many.turtles" ORDER BY local_timestamp ASC')"

assert_eq "{}" "$(curl -k -d '{"homeserver": "few.turtles"}' http://localhost:${port}/push 2>/dev/null)"
assert_eq "|||||||||||||||||||||" "$(sqlite3 ${dir}/stats.db 'SELECT daily_active_users, total_users, total_room_count, daily_messages, remote_timestamp, uptime_seconds, r30_users_all, r30_users_android, r30_users_ios, r30_users_electron, r30_users_web, daily_user_type_native, daily_user_type_guest, daily_user_type_bridged, cpu_average, memory_rss, cache_factor, event_cache_size, python_version, database_engine, database_server_version, server_context FROM stats WHERE homeserver == "few.turtles"')"

assert_eq "{}" "$(curl -k -H "X-Forwarded-For: faraway.turtles" -d '{"homeserver": "proxied.turtles"}' http://localhost:${port}/push 2>/dev/null)"
assert_eq "faraway.turtles" "$(sqlite3 ${dir}/stats.db 'SELECT forwarded_for FROM stats WHERE homeserver == "proxied.turtles"')"

assert_eq "{}" "$(curl -k -H "x-forwarded-for: lower.faraway.turtles" -d '{"homeserver": "lower.proxied.turtles"}' http://localhost:${port}/push 2>/dev/null)"
assert_eq "lower.faraway.turtles" "$(sqlite3 ${dir}/stats.db 'SELECT forwarded_for FROM stats WHERE homeserver == "lower.proxied.turtles"')"

assert_eq "{}" "$(curl -k -H "User-Agent: turtle/agent/0.0.7" -d '{"homeserver": "agent.turtles"}' http://localhost:${port}/push 2>/dev/null)"
assert_eq "turtle/agent/0.0.7" "$(sqlite3 ${dir}/stats.db 'SELECT user_agent FROM stats WHERE homeserver == "agent.turtles"')"
