// Copyright 2020 The Matrix.org Foundation C.I.C.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"database/sql"
	"fmt"
	"strings"
)

// Synapse specific stats
type ReportStatsSynapse struct {
	CommonStats
	CacheFactor    *float64 `json:"cache_factor"`
	EventCacheSize *int64   `json:"event_cache_size"`
	PythonVersion  string   `json:"python_version"`
	ServerContext  string   `json:"server_context"`
}

func createTableSynapse(db *sql.DB) error {
	autoincrement := "AUTOINCREMENT"
	primaryKeyType := "INTEGER"
	doubleType := "DOUBLE"

	if *dbDriver == "mysql" {
		autoincrement = "AUTO_INCREMENT"
	} else if *dbDriver == "postgres" {
		autoincrement = ""
		primaryKeyType = "SERIAL"
		doubleType = "DOUBLE PRECISION"
	}
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS stats(
		id ` + primaryKeyType + ` NOT NULL PRIMARY KEY ` + autoincrement + ` ,
		homeserver VARCHAR(256),
		local_timestamp BIGINT,
		remote_timestamp BIGINT,
		remote_addr TEXT,
		forwarded_for TEXT,
		uptime_seconds BIGINT,
		total_users BIGINT,
		total_nonbridged_users BIGINT,
		total_room_count BIGINT,
		daily_active_users BIGINT,
		daily_active_rooms BIGINT,
		daily_messages BIGINT,
		daily_sent_messages BIGINT,
		daily_active_e2ee_rooms BIGINT,
		daily_e2ee_messages BIGINT,
		daily_sent_e2ee_messages BIGINT,
		monthly_active_users BIGINT,
		r30_users_all BIGINT,
		r30_users_android BIGINT,
		r30_users_ios BIGINT,
		r30_users_electron BIGINT,
		r30_users_web BIGINT,
		r30v2_users_all BIGINT,
		r30v2_users_android BIGINT,
		r30v2_users_ios BIGINT,
		r30v2_users_electron BIGINT,
		r30v2_users_web BIGINT,
		cpu_average BIGINT,
		memory_rss BIGINT,
		cache_factor ` + doubleType + `,
		event_cache_size BIGINT,
		user_agent TEXT,
		daily_user_type_native BIGINT,
		daily_user_type_bridged BIGINT,
		daily_user_type_guest BIGINT,
		python_version TEXT,
		database_engine TEXT,
		database_server_version TEXT,
		server_context TEXT,
		log_level TEXT
		)`)

	return err
}

func (sr *ReportStatsSynapse) Save(db *sql.DB) error {
	cols := []string{"homeserver", "local_timestamp", "remote_addr"}
	vals := []interface{}{sr.Homeserver, sr.LocalTimestamp, sr.RemoteAddr}

	cols, vals = appendIfNonNil(cols, vals, "remote_timestamp", sr.RemoteTimestamp)
	cols, vals = appendIfNonNil(cols, vals, "uptime_seconds", sr.UptimeSeconds)
	cols, vals = appendIfNonNil(cols, vals, "total_users", sr.TotalUsers)
	cols, vals = appendIfNonNil(cols, vals, "total_nonbridged_users", sr.TotalNonBridgedUsers)
	cols, vals = appendIfNonNil(cols, vals, "total_room_count", sr.TotalRoomCount)
	cols, vals = appendIfNonNil(cols, vals, "daily_active_users", sr.DailyActiveUsers)
	cols, vals = appendIfNonNil(cols, vals, "daily_active_rooms", sr.DailyActiveRooms)
	cols, vals = appendIfNonNil(cols, vals, "daily_messages", sr.DailyMessages)
	cols, vals = appendIfNonNil(cols, vals, "daily_sent_messages", sr.DailySentMessages)
	cols, vals = appendIfNonNil(cols, vals, "daily_active_e2ee_rooms", sr.DailyActiveE2eeRooms)
	cols, vals = appendIfNonNil(cols, vals, "daily_e2ee_messages", sr.DailyE2eeMessages)
	cols, vals = appendIfNonNil(cols, vals, "daily_sent_e2ee_messages", sr.DailySentE2eeMessages)
	cols, vals = appendIfNonNil(cols, vals, "monthly_active_users", sr.MonthlyActiveUsers)

	cols, vals = appendIfNonNil(cols, vals, "r30_users_all", sr.R30UsersAll)
	cols, vals = appendIfNonNil(cols, vals, "r30_users_android", sr.R30UsersAndroid)
	cols, vals = appendIfNonNil(cols, vals, "r30_users_ios", sr.R30UsersIOS)
	cols, vals = appendIfNonNil(cols, vals, "r30_users_electron", sr.R30UsersElectron)
	cols, vals = appendIfNonNil(cols, vals, "r30_users_web", sr.R30UsersWeb)

	cols, vals = appendIfNonNil(cols, vals, "r30v2_users_all", sr.R30V2UsersAll)
	cols, vals = appendIfNonNil(cols, vals, "r30v2_users_android", sr.R30V2UsersAndroid)
	cols, vals = appendIfNonNil(cols, vals, "r30v2_users_ios", sr.R30V2UsersIOS)
	cols, vals = appendIfNonNil(cols, vals, "r30v2_users_electron", sr.R30V2UsersElectron)
	cols, vals = appendIfNonNil(cols, vals, "r30v2_users_web", sr.R30V2UsersWeb)

	cols, vals = appendIfNonEmpty(cols, vals, "forwarded_for", sr.XForwardedFor)
	cols, vals = appendIfNonEmpty(cols, vals, "user_agent", sr.UserAgent)

	cols, vals = appendIfNonNil(cols, vals, "cpu_average", sr.CPUAverage)
	cols, vals = appendIfNonNil(cols, vals, "memory_rss", sr.MemoryRSS)
	cols, vals = appendIfNonNilFloat(cols, vals, "cache_factor", sr.CacheFactor)

	cols, vals = appendIfNonNil(cols, vals, "event_cache_size", sr.EventCacheSize)
	cols, vals = appendIfNonNil(cols, vals, "daily_user_type_native", sr.DailyUserTypeNative)
	cols, vals = appendIfNonNil(cols, vals, "daily_user_type_guest", sr.DailyUserTypeGuest)
	cols, vals = appendIfNonNil(cols, vals, "daily_user_type_bridged", sr.DailyUserTypeBridged)

	cols, vals = appendIfNonEmpty(cols, vals, "python_version", sr.PythonVersion)
	cols, vals = appendIfNonEmpty(cols, vals, "database_engine", sr.DatabaseEngine)
	cols, vals = appendIfNonEmpty(cols, vals, "database_server_version", sr.DatabaseServerVersion)
	cols, vals = appendIfNonEmpty(cols, vals, "server_context", sr.ServerContext)
	cols, vals = appendIfNonEmpty(cols, vals, "log_level", sr.LogLevel)

	var valuePlaceholders []string
	for i := range vals {
		if *dbDriver == "mysql" {
			valuePlaceholders = append(valuePlaceholders, "?")
		} else {
			valuePlaceholders = append(valuePlaceholders, fmt.Sprintf("$%d", i+1))
		}
	}
	qry := fmt.Sprintf("INSERT INTO stats (%s) VALUES (%s)", strings.Join(cols, ", "), strings.Join(valuePlaceholders, ", "))
	_, err := db.Exec(qry, vals...)
	return err
}
