#!/usr/bin/env python
# Script to read the stats table and aggregate results down to the sums per day
# The goal of the aggregate datastore is to improve analytics performance.

import pymysql.cursors
import os
import time
from dateutil import tz
from datetime import datetime

from pymysql import Connection

ONE_DAY = 24 * 60 * 60


class Config:
    def __init__(self):
        self.db_name = os.environ["PANOPTICON_DB_NAME"]
        self.db_user = os.environ["PANOPTICON_DB_USER"]
        self.db_password = os.environ["PANOPTICON_DB_PASSWORD"]
        self.db_host = os.environ["PANOPTICON_DB_HOST"]
        self.db_port = int(os.environ["PANOPTICON_DB_PORT"])


def set_up_aggregate_stats_table(db: Connection):
    # Set up aggregate_stats schema
    SCHEMA = """
        CREATE TABLE IF NOT EXISTS `aggregate_stats` (
            `day` bigint(20) NOT NULL,
            `total_users` bigint(20) DEFAULT NULL,
            `total_nonbridged_users` bigint(20) DEFAULT NULL,
            `total_room_count` bigint(20) DEFAULT NULL,
            `daily_active_users` bigint(20) DEFAULT NULL,
            `daily_active_rooms` bigint(20) DEFAULT NULL,
            `daily_messages` bigint(20) DEFAULT NULL,
            `daily_sent_messages` bigint(20) DEFAULT NULL,
            `daily_active_e2ee_rooms` bigint(20) DEFAULT NULL,
            `daily_e2ee_messages` bigint(20) DEFAULT NULL,
            `daily_sent_e2ee_messages` bigint(20) DEFAULT NULL,
            `monthly_active_users` bigint(20) DEFAULT NULL,
            `r30_users_all` bigint(20) DEFAULT NULL,
            `r30_users_android` bigint(20) DEFAULT NULL,
            `r30_users_ios` bigint(20) DEFAULT NULL,
            `r30_users_electron` bigint(20) DEFAULT NULL,
            `r30_users_web` bigint(20) DEFAULT NULL,
            `r30v2_users_all` bigint(20) DEFAULT NULL,
            `r30v2_users_android` bigint(20) DEFAULT NULL,
            `r30v2_users_ios` bigint(20) DEFAULT NULL,
            `r30v2_users_electron` bigint(20) DEFAULT NULL,
            `r30v2_users_web` bigint(20) DEFAULT NULL,
            `daily_user_type_native` bigint(20) DEFAULT NULL,
            `daily_user_type_bridged` bigint(20) DEFAULT NULL,
            `daily_user_type_guest` bigint(20) DEFAULT NULL,
            `daily_active_homeservers` bigint(20) DEFAULT NULL,
            `server_context` text,
            PRIMARY KEY (`day`),
            UNIQUE KEY `day` (`day`)
        ) ENGINE=InnoDB DEFAULT CHARSET=latin1
    """

    create_table(db, SCHEMA)


def main():
    configuration = Config()

    db = pymysql.connect(
        host=configuration.db_host,
        user=configuration.db_user,
        passwd=configuration.db_password,
        db=configuration.db_name,
        port=configuration.db_port,
        ssl={'ssl': {}}
    )

    set_up_aggregate_stats_table(db)

    while True:
        aggregate_once(db)
        time.sleep(ONE_DAY)


def aggregate_once(db: Connection):
    with db.cursor() as cursor:
        start_date_query = """
            SELECT day from aggregate_stats
            ORDER BY day DESC
            LIMIT 1
        """
        cursor.execute(start_date_query)
        try:
            last_day_in_db = cursor.fetchone()[0]
        except IndexError:
            # If no data to read assume is empty revert to 2015-10-01
            # which is when the stats table is populated from.
            last_day_in_db = 1443657600
    now = datetime.utcnow().date()
    today = int(datetime(now.year, now.month, now.day, tzinfo=tz.tzutc()).strftime('%s'))
    processing_day = last_day_in_db + ONE_DAY
    while processing_day < today:
        with db.cursor() as cursor:
            # Need to filter on "AND total_users > 0" since some installs
            # run with a standby unused server with an empty db. This means
            # that picking a recent entry for a given server is likely to
            # under report. Filtering on total_users removes the standbys.
            # It also filters out genuinely unused servers, but the value of
            # aggregating these servers is limited.
            query = """
                SELECT
                    SUM(total_users) as 'total_users',
                    SUM(total_nonbridged_users) as 'total_nonbridged_users',
                    SUM(total_room_count) as 'total_room_count',
                    SUM(daily_active_users) as 'daily_active_users',
                    SUM(daily_active_rooms) as 'daily_active_rooms',
                    SUM(daily_messages) as 'daily_messages',
                    SUM(daily_sent_messages) as 'daily_sent_messages',
                    SUM(daily_active_e2ee_rooms) as 'daily_active_e2ee_rooms',
                    SUM(daily_e2ee_messages) as 'daily_e2ee_messages',
                    SUM(daily_sent_e2ee_messages) as 'daily_sent_e2ee_messages',
                    SUM(monthly_active_users) as 'monthly_active_users',
                    SUM(r30_users_all) as 'r30_users_all',
                    SUM(r30_users_android) as 'r30_users_android',
                    SUM(r30_users_ios) as 'r30_users_ios',
                    SUM(r30_users_electron) as 'r30_users_electron',
                    SUM(r30_users_web) as 'r30_users_web',
                    SUM(r30v2_users_all) as 'r30v2_users_all',
                    SUM(r30v2_users_android) as 'r30v2_users_android',
                    SUM(r30v2_users_ios) as 'r30v2_users_ios',
                    SUM(r30v2_users_electron) as 'r30v2_users_electron',
                    SUM(r30v2_users_web) as 'r30v2_users_web',
                    SUM(daily_user_type_native) as 'daily_user_type_native',
                    SUM(daily_user_type_bridged) as 'daily_user_type_bridged',
                    SUM(daily_user_type_guest) as 'daily_user_type_guest',
                    COUNT(homeserver) as 'homeserver'
                FROM (
                    SELECT *, MAX(local_timestamp)
                    FROM stats
                    WHERE local_timestamp >= %s and local_timestamp < %s
                    AND total_users > 0
                    GROUP BY homeserver
                ) as s;
            """

            date_range = (processing_day, processing_day + ONE_DAY)
            cursor.execute(query, date_range)
            result = cursor.fetchone()

            insert_query = """
                INSERT into aggregate_stats
                (
                        day,
                        total_users,
                        total_nonbridged_users,
                        total_room_count,
                        daily_active_users,
                        daily_active_rooms,
                        daily_messages,
                        daily_sent_messages,
                        daily_active_e2ee_rooms,
                        daily_e2ee_messages,
                        daily_sent_e2ee_messages,
                        monthly_active_users,
                        r30_users_all,
                        r30_users_android,
                        r30_users_ios,
                        r30_users_electron,
                        r30_users_web,
                        r30v2_users_all,
                        r30v2_users_android,
                        r30v2_users_ios,
                        r30v2_users_electron,
                        r30v2_users_web,
                        daily_user_type_native,
                        daily_user_type_bridged,
                        daily_user_type_guest,
                        daily_active_homeservers,
                        server_context
                ) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s,
                          %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
            """
            insert_data = [x if x is None else int(x) for x in result]
            # insert day at the front
            insert_data.insert(0, processing_day)
            # append context at the end
            insert_data.append(None)
            cursor.execute(insert_query, insert_data)
            db.commit()
            processing_day = processing_day + ONE_DAY


def create_table(db, schema):
    """This method executes a CREATE TABLE IF NOT EXISTS command
    _without_ generating a mysql warning if the table already exists."""
    cursor = db.cursor()
    cursor.execute('SET sql_notes = 0;')
    cursor.execute(schema)
    cursor.execute('SET sql_notes = 1;')
    db.commit()


if __name__ == "__main__":
    main()
