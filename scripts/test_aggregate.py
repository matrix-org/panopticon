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
    *,
    remote_addr: str = "192.42.42.42",
    forwarded_for: Optional[str] = None,
    table: str = "stats",
):
    """
    Insert a row that emulates a row that Panopticon would update after a server
    phones home.
    """
    metric_set_lines = ",\n".join(f"`{metric}` = %s" for metric in METRIC_COLUMNS)
    cursor.execute(
        f"""
        INSERT INTO {table}
        SET
            homeserver = %s,
            local_timestamp = %s,
            remote_timestamp = %s,
            remote_addr = %s,
            forwarded_for = %s,
            user_agent = %s,
            {metric_set_lines};
        """,
        (
            homeserver,
            timestamp,
            timestamp,
            remote_addr,
            forwarded_for,
            "FakeStats/42.x.y",
        )
        + tuple(metrics.values()),
    )


def select_aggregate(cursor: Cursor, day: int) -> Optional[Dict[str, int]]:
    """
    Select the aggregated statistics for a given day.
    """

    extra_columns = ("daily_active_homeservers",)
    all_columns = METRIC_COLUMNS + extra_columns

    metric_select_lines = "\n,".join(f"`{metric}`" for metric in all_columns)
    cursor.execute(
        f"""
        SELECT
            {metric_select_lines},
            daily_active_homeservers
        FROM aggregate_stats
        WHERE day = %s
        """,
        (day,),
    )
    row = cursor.fetchone()
    if row is None:
        return None
    else:
        return dict(zip(all_columns, row))


class AggregateTestCase(TestCase):
    def setUp(self) -> None:
        self.config = Config()
        db = self.config.connect_db()
        with db.cursor() as cursor:
            cursor.execute("DROP TABLE IF EXISTS aggregate_stats;")

            for stats_table in ("stats", "dendrite_stats"):
                cursor.execute(f"DROP TABLE IF EXISTS {stats_table};")
                metric_lines = ",\n".join(
                    f"`{metric}` BIGINT" for metric in METRIC_COLUMNS
                )
                cursor.execute(
                    f"""
                    CREATE TABLE {stats_table} (
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
            insert_recording(
                cursor,
                "hs3",
                INITIAL_DAY + ONE_DAY + 300,
                {metric: 2 for metric in METRIC_COLUMNS},
                table="dendrite_stats",
            )

        aggregate_until_today(db, today=INITIAL_DAY + 2 * ONE_DAY)

        with db.cursor() as cursor:
            row = select_aggregate(cursor, INITIAL_DAY + ONE_DAY)
            self.assertIsNot(row, None)
            self.assertEqual(row["total_users"], 6)

    def test_empty_homeservers_not_counted(self):
        """
        Tests that empty servers are not counted (because they are likely to be
        standby backup servers).
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
                "hs2-standby-backup",
                INITIAL_DAY + ONE_DAY + 300,
                dict(
                    {metric: 3 for metric in METRIC_COLUMNS},
                    total_users=0,
                ),
            )

        aggregate_until_today(db, today=INITIAL_DAY + 2 * ONE_DAY)

        with db.cursor() as cursor:
            row = select_aggregate(cursor, INITIAL_DAY + ONE_DAY)
            self.assertIsNot(row, None)
            self.assertEqual(row["total_users"], 1)
            self.assertEqual(row["daily_active_homeservers"], 1)

    def test_same_homeserver_from_different_addresses_counted_separately(self):
        """
        Tests that reports claiming to be the same `homeserver` name *but* coming from
        different `remote_addr` are counted separately.
        """
        db = self.config.connect_db()
        with db.cursor() as cursor:
            insert_recording(
                cursor,
                # Both recordings have the same `homeserver`
                "hs1",
                INITIAL_DAY + ONE_DAY + 300,
                {metric: 1 for metric in METRIC_COLUMNS},
                # But come from different people
                remote_addr="10.0.0.1",
                forwarded_for=None,
            )
            insert_recording(
                cursor,
                "hs1",
                INITIAL_DAY + ONE_DAY + 600,
                {metric: 3 for metric in METRIC_COLUMNS},
                remote_addr="10.0.0.2",
                forwarded_for=None,
            )
        aggregate_until_today(db, today=INITIAL_DAY + 2 * ONE_DAY)
        with db.cursor() as cursor:
            row = select_aggregate(cursor, INITIAL_DAY + ONE_DAY)
            self.assertIsNot(row, None)
            self.assertEqual(row["total_users"], 4)
            self.assertEqual(row["daily_active_homeservers"], 2)

    def test_same_homeserver_from_different_addresses_behind_proxy_counted_separately(
        self,
    ):
        """
        Tests that reports claiming to be the same `homeserver` name *but* coming from
        different addresses (according to the `X-Forwarded-For` header) are counted
        separately.

        This mimicks what happens when you run behind a reverse-proxy.
        """
        db = self.config.connect_db()
        with db.cursor() as cursor:
            insert_recording(
                cursor,
                # Both recordings have the same `homeserver`
                "hs1",
                INITIAL_DAY + ONE_DAY + 300,
                {metric: 1 for metric in METRIC_COLUMNS},
                # Same `remote_addr` as both come from the reverse-proxy (ephemeral port
                # is different)
                remote_addr="192.42.42.42:40000",
                # But come from different people according to the `X-Forwarded-For`
                # header
                forwarded_for="10.0.0.1",
            )
            insert_recording(
                cursor,
                "hs1",
                INITIAL_DAY + ONE_DAY + 600,
                {metric: 3 for metric in METRIC_COLUMNS},
                remote_addr="192.42.42.42:50000",
                forwarded_for="10.0.0.2",
            )
        aggregate_until_today(db, today=INITIAL_DAY + 2 * ONE_DAY)
        with db.cursor() as cursor:
            row = select_aggregate(cursor, INITIAL_DAY + ONE_DAY)
            self.assertIsNot(row, None)
            # Expect to see each
            self.assertEqual(row["total_users"], 4)
            self.assertEqual(row["daily_active_homeservers"], 2)

    def test_same_address_on_different_ports_counted_once(self):
        """
        Tests that reports from the same `remote_addr` are counted once, even though
        each request comes from a different (ephemeral) source port.

        This is mimicking what you would when using the Docker container and requests
        come over the Docker bridge gateway. Each phone home request gets a new randomly
        assigned ephemeral source port.
        """
        db = self.config.connect_db()
        with db.cursor() as cursor:
            for port in (40000, 40001, 40002):
                insert_recording(
                    cursor,
                    "hs1",
                    INITIAL_DAY + ONE_DAY + port - 39000,
                    {metric: 1 for metric in METRIC_COLUMNS},
                    remote_addr=f"172.17.0.1:{port}",
                    forwarded_for=None,
                )
        aggregate_until_today(db, today=INITIAL_DAY + 2 * ONE_DAY)
        with db.cursor() as cursor:
            row = select_aggregate(cursor, INITIAL_DAY + ONE_DAY)
            self.assertIsNot(row, None)
            # We expect to see the stats from *one* of the recordings
            self.assertEqual(row["total_users"], 1)
            self.assertEqual(row["daily_active_homeservers"], 1)
