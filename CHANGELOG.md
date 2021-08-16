Panopticon 0.2.0 (2021-08-16)
=============================

Schema changes
--------------

**Beware**: upgrading to this release requires manual schema changes.

The following SQL can be applied before applying the upgrade:
```sql
ALTER TABLE stats
  ADD COLUMN daily_active_e2ee_rooms BIGINT AFTER daily_sent_messages,
  ADD COLUMN daily_e2ee_messages BIGINT AFTER daily_active_e2ee_rooms,
  ADD COLUMN daily_sent_e2ee_messages BIGINT AFTER daily_e2ee_messages,
  ADD COLUMN r30v2_users_all BIGINT AFTER r30_users_web,
  ADD COLUMN r30v2_users_android BIGINT AFTER r30v2_users_all,
  ADD COLUMN r30v2_users_ios BIGINT AFTER r30v2_users_android,
  ADD COLUMN r30v2_users_electron BIGINT AFTER r30v2_users_ios,
  ADD COLUMN r30v2_users_web BIGINT AFTER r30v2_users_electron;

ALTER TABLE aggregate_stats
  ADD COLUMN daily_active_e2ee_rooms BIGINT AFTER daily_sent_messages,
  ADD COLUMN daily_e2ee_messages BIGINT AFTER daily_active_e2ee_rooms,
  ADD COLUMN daily_sent_e2ee_messages BIGINT AFTER daily_e2ee_messages,
  ADD COLUMN r30v2_users_all BIGINT AFTER r30_users_web,
  ADD COLUMN r30v2_users_android BIGINT AFTER r30v2_users_all,
  ADD COLUMN r30v2_users_ios BIGINT AFTER r30v2_users_android,
  ADD COLUMN r30v2_users_electron BIGINT AFTER r30v2_users_ios,
  ADD COLUMN r30v2_users_web BIGINT AFTER r30v2_users_electron;
```


New features
------------

- New encrypted message metrics have been added. ([\#20](https://github.com/matrix-org/panopticon/pull/20))
- A new 30-day retention metric has been introduced (R30v2). ([\#22](https://github.com/matrix-org/panopticon/pull/22), [\#23](https://github.com/matrix-org/panopticon/pull/23))


Internal changes
----------------

- *GitHub Actions* is now being used to run CI checks and to build Docker images. ([\#24](https://github.com/matrix-org/panopticon/pull/24), [\#25](https://github.com/matrix-org/panopticon/pull/25))


Prior versions
==============

Prior versions did not have changelog entries.
