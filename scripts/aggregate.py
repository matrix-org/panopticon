# Script to read the stats table and aggregate results down to the sums per day
# The goal of the aggregate datastore is to improve analytics performance.

import pymysql.cursors
import yaml
from os.path import expanduser
from dateutil import tz
from datetime import datetime


class Config:
    def __init__(self):
        with open(expanduser("~") + "/.panopticon", "r") as config_file:
            config = yaml.safe_load(config_file)
            self.DB_NAME = config["db_name"]
            self.DB_USER = config["db_user"]
            self.DB_PASSWORD = config["db_password"]
            self.DB_HOST = config["db_host"]
            self.DB_PORT = config["db_port"]


CONFIG = Config()

db = pymysql.connect(
    host=CONFIG.DB_HOST,
    user=CONFIG.DB_USER,
    passwd=CONFIG.DB_PASSWORD,
    db=CONFIG.DB_NAME,
    port=CONFIG.DB_PORT
)

ONE_DAY = 24 * 60 * 60

try:
    with db.cursor() as cursor:
        start_date_query = """
            SELECT day from aggregate_stats
            ORDER BY day DESC
            LIMIT 1
            """
        cursor.execute(start_date_query)
        overal_start_day = cursor.fetchone()[0] + ONE_DAY
        # overal_start_day = 1443657600

    today = datetime.utcnow().date()
    today_start = datetime(today.year, today.month, today.day, tzinfo=tz.tzutc())
    final_day = int(today_start.strftime('%s'))
    start_day = overal_start_day

    while start_day < final_day:

        with db.cursor() as cursor:
            query = """
                SELECT
                    SUM(total_users) as 'total_users',
                    SUM(total_nonbridged_users) as 'total_nonbridged_users',
                    SUM(total_room_count) as 'total_room_count',
                    SUM(daily_active_users) as 'daily_active_users',
                    SUM(daily_active_rooms) as 'daily_active_rooms',
                    SUM(daily_messages) as 'daily_messages',
                    SUM(daily_sent_messages) as 'daily_sent_messages',
                    SUM(r30_users_all) as 'r30_users_all',
                    SUM(r30_users_android) as 'r30_users_android',
                    SUM(r30_users_ios) as 'r30_users_ios',
                    SUM(r30_users_electron) as 'r30_users_electron',
                    SUM(r30_users_web) as 'r30_users_web',
                    SUM(daily_user_type_native) as 'daily_user_type_native',
                    SUM(daily_user_type_bridged) as 'daily_user_type_bridged',
                    SUM(daily_user_type_guest) as 'daily_user_type_guest',
                    COUNT(homeserver) as 'homeserver'
                FROM (
                    SELECT *, MAX(local_timestamp)
                    FROM stats
                    WHERE local_timestamp >= %s and local_timestamp < %s
                    GROUP BY homeserver
                ) as s;
                """

            date_range = (start_day, start_day + ONE_DAY)
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
                        r30_users_all,
                        r30_users_android,
                        r30_users_ios,
                        r30_users_electron,
                        r30_users_web,
                        daily_user_type_native,
                        daily_user_type_bridged,
                        daily_user_type_guest,
                        daily_active_homservers,
                        server_context
                ) VALUES (%s, %s,%s, %s,%s, %s,%s, %s,%s, %s,%s, %s,%s, %s,%s, %s, %s, %s)
            """
            insert_data = [x if x is None else int(x) for x in result]
            # insert day at the front
            insert_data.insert(0, start_day)
            # append context at the end
            insert_data.append(None)
            cursor.execute(insert_query, insert_data)
            db.commit()
            start_day = start_day + ONE_DAY

finally:
    db.close()
