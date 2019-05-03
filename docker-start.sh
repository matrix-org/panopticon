#!/bin/sh
#
# Converts environment variables into flags for panopticon

exec /root/panopticon --db-driver=$PANOPTICON_DB_DRIVER --db=$PANOPTICON_DB --port=$PANOPTICON_PORT
