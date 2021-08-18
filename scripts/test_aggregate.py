from typing import Dict, Optional
from unittest import TestCase

from pymysql.cursors import Cursor

from aggregate import Config
from aggregate import set_up_aggregate_stats_table
from aggregate import METRIC_COLUMNS
from aggregate import INITIAL_DAY, aggregate_until_today
from aggregate import ONE_DAY


def insert_recording(
    cursor: Cursor,
    homeserver: str,
    timestamp: int,
    metrics: Dict[str, int],
    remote_addr: str = "192.42.42.42",
):
    """
    Insert a row that emulates a row that Panopticon would update after a server
    phones home.
    """
    metric_set_lines = ",\n".join(f"`{metric}` = %s" for metric in METRIC_COLUMNS)
    cursor.execute(
        f"""
        INSERT INTO stats
        SET
            homeserver = %s,
            local_timestamp = %s,
            remote_timestamp = %s,
            remote_addr = %s,
            forwarded_for = %s,
            user_agent = %s,
            {metric_set_lines};
        """,
        (homeserver, timestamp, timestamp, remote_addr, remote_addr, "FakeStats/42.x.y")
        + tuple(metrics.values()),
    )


def select_aggregate(cursor: Cursor, day: int) -> Optional[Dict[str, int]]:
    """
    Select the aggregated statistics for a given day.
    """
    metric_select_lines = "\n,".join(f"`{metric}`" for metric in METRIC_COLUMNS)
    cursor.execute(
        f"""
        SELECT
            {metric_select_lines}
        FROM aggregate_stats
        WHERE day = %s
        """,
        (day,),
    )
    row = cursor.fetchone()
    if row is None:
        return None
    else:
        return dict(zip(METRIC_COLUMNS, row))


class AggregateTestCase(TestCase):
    def setUp(self) -> None:
        self.config = Config()
        db = self.config.connect_db()
        with db.cursor() as cursor:
            cursor.execute("DROP TABLE IF EXISTS aggregate_stats;")
            cursor.execute("DROP TABLE IF EXISTS stats;")
            metric_lines = ",\n".join(f"`{metric}` BIGINT" for metric in METRIC_COLUMNS)
            cursor.execute(
                f"""
                CREATE TABLE stats (
                    id INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT,
                    homeserver VARCHAR(256),
                    local_timestamp BIGINT,
                    remote_timestamp BIGINT,
                    remote_addr TEXT,
                    forwarded_for TEXT,
                    user_agent TEXT,
                    {metric_lines}
                );
                """
            )
        set_up_aggregate_stats_table(db)

    def test_sum_of_metrics(self):
        """
        Tests that the aggregator reports the sum of metrics.
        """

        db = self.config.connect_db()
        with db.cursor() as cursor:
            insert_recording(
                cursor,
                "hs1",
                INITIAL_DAY + ONE_DAY + 300,
                {metric: 1 for metric in METRIC_COLUMNS},
            )
            insert_recording(
                cursor,
                "hs2",
                INITIAL_DAY + ONE_DAY + 300,
                {metric: 3 for metric in METRIC_COLUMNS},
            )

        aggregate_until_today(db, today=INITIAL_DAY + 2 * ONE_DAY)

        with db.cursor() as cursor:
            row = select_aggregate(cursor, INITIAL_DAY + ONE_DAY)
            self.assertIsNot(row, None)
            self.assertEqual(row["total_users"], 4)
