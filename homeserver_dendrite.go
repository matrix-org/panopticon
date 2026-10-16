// Copyright 2022 The Matrix.org Foundation C.I.C.
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

// Dendrite specific stats
type ReportStatsDendrite struct {
	// We're using mostly Synapse defined fields
	Common             CommonStats
	GoOS               string `json:"go_os,omitempty"`
	GoArch             string `json:"go_arch,omitempty"`
	GoVersion          string `json:"go_version,omitempty"`
	FederationDisabled *bool  `json:"federation_disabled,omitempty"`
	Monolith           *bool  `json:"monolith,omitempty"`
	NATSEmbedded       *bool  `json:"nats_embedded,omitempty"`
	NATSInMemory       *bool  `json:"nats_in_memory,omitempty"`
	NumCPU             *int64 `json:"num_cpu,omitempty"`
	NumGoRoutine       *int64 `json:"num_go_routine,omitempty"`
	Version            string `json:"version,omitempty"`
}

func createTableDendrite(db *sql.DB) error {
	autoincrement := "AUTOINCREMENT"
	primaryKeyType := "INTEGER"

	if *dbDriver == "mysql" {
		autoincrement = "AUTO_INCREMENT"
	} else if *dbDriver == "postgres" {
		autoincrement = ""
		primaryKeyType = "SERIAL"
	}
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS dendrite_stats(
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
		user_agent TEXT,
		daily_user_type_native BIGINT,
		daily_user_type_bridged BIGINT,
		daily_user_type_guest BIGINT,
		database_engine TEXT,
		database_server_version TEXT,
		log_level TEXT,
		goos TEXT,
		goarch TEXT,
		goversion TEXT,
		federation_disabled INT,
		monolith INT,
		nats_embedded INT,
		nats_in_memory INT,
		num_cpu INT,
		num_go_routine INT,
		version TEXT
		)`)

	return err
}

func (sr *ReportStatsDendrite) Save(db *sql.DB) error {
	cols := []string{"homeserver", "local_timestamp", "remote_addr"}
	vals := []interface{}{sr.Common.Homeserver, sr.Common.LocalTimestamp, sr.Common.RemoteAddr}

	cols, vals = appendIfNonNil(cols, vals, "remote_timestamp", sr.Common.RemoteTimestamp)
	cols, vals = appendIfNonNil(cols, vals, "uptime_seconds", sr.Common.UptimeSeconds)
	cols, vals = appendIfNonNil(cols, vals, "total_users", sr.Common.TotalUsers)
	cols, vals = appendIfNonNil(cols, vals, "total_nonbridged_users", sr.Common.TotalNonBridgedUsers)
	cols, vals = appendIfNonNil(cols, vals, "total_room_count", sr.Common.TotalRoomCount)
	cols, vals = appendIfNonNil(cols, vals, "daily_active_users", sr.Common.DailyActiveUsers)
	cols, vals = appendIfNonNil(cols, vals, "daily_active_rooms", sr.Common.DailyActiveRooms)
	cols, vals = appendIfNonNil(cols, vals, "daily_messages", sr.Common.DailyMessages)
	cols, vals = appendIfNonNil(cols, vals, "daily_sent_messages", sr.Common.DailySentMessages)
	cols, vals = appendIfNonNil(cols, vals, "daily_active_e2ee_rooms", sr.Common.DailyActiveE2eeRooms)
	cols, vals = appendIfNonNil(cols, vals, "daily_e2ee_messages", sr.Common.DailyE2eeMessages)
	cols, vals = appendIfNonNil(cols, vals, "daily_sent_e2ee_messages", sr.Common.DailySentE2eeMessages)
	cols, vals = appendIfNonNil(cols, vals, "monthly_active_users", sr.Common.MonthlyActiveUsers)

	cols, vals = appendIfNonNil(cols, vals, "r30_users_all", sr.Common.R30UsersAll)
	cols, vals = appendIfNonNil(cols, vals, "r30_users_android", sr.Common.R30UsersAndroid)
	cols, vals = appendIfNonNil(cols, vals, "r30_users_ios", sr.Common.R30UsersIOS)
	cols, vals = appendIfNonNil(cols, vals, "r30_users_electron", sr.Common.R30UsersElectron)
	cols, vals = appendIfNonNil(cols, vals, "r30_users_web", sr.Common.R30UsersWeb)

	cols, vals = appendIfNonNil(cols, vals, "r30v2_users_all", sr.Common.R30V2UsersAll)
	cols, vals = appendIfNonNil(cols, vals, "r30v2_users_android", sr.Common.R30V2UsersAndroid)
	cols, vals = appendIfNonNil(cols, vals, "r30v2_users_ios", sr.Common.R30V2UsersIOS)
	cols, vals = appendIfNonNil(cols, vals, "r30v2_users_electron", sr.Common.R30V2UsersElectron)
	cols, vals = appendIfNonNil(cols, vals, "r30v2_users_web", sr.Common.R30V2UsersWeb)

	cols, vals = appendIfNonEmpty(cols, vals, "forwarded_for", sr.Common.XForwardedFor)
	cols, vals = appendIfNonEmpty(cols, vals, "user_agent", sr.Common.UserAgent)

	cols, vals = appendIfNonNil(cols, vals, "cpu_average", sr.Common.CPUAverage)
	cols, vals = appendIfNonNil(cols, vals, "memory_rss", sr.Common.MemoryRSS)

	cols, vals = appendIfNonNil(cols, vals, "daily_user_type_native", sr.Common.DailyUserTypeNative)
	cols, vals = appendIfNonNil(cols, vals, "daily_user_type_guest", sr.Common.DailyUserTypeGuest)
	cols, vals = appendIfNonNil(cols, vals, "daily_user_type_bridged", sr.Common.DailyUserTypeBridged)

	cols, vals = appendIfNonEmpty(cols, vals, "database_engine", sr.Common.DatabaseEngine)
	cols, vals = appendIfNonEmpty(cols, vals, "database_server_version", sr.Common.DatabaseServerVersion)

	cols, vals = appendIfNonEmpty(cols, vals, "log_level", sr.Common.LogLevel)

	cols, vals = appendIfNonEmpty(cols, vals, "goos", sr.GoOS)
	cols, vals = appendIfNonEmpty(cols, vals, "goarch", sr.GoArch)
	cols, vals = appendIfNonEmpty(cols, vals, "goversion", sr.GoVersion)
	cols, vals = appendIfNonNilBool(cols, vals, "federation_disabled", sr.FederationDisabled)
	cols, vals = appendIfNonNilBool(cols, vals, "monolith", sr.Monolith)
	cols, vals = appendIfNonNilBool(cols, vals, "nats_embedded", sr.NATSEmbedded)
	cols, vals = appendIfNonNilBool(cols, vals, "nats_in_memory", sr.NATSInMemory)
	cols, vals = appendIfNonNil(cols, vals, "num_cpu", sr.NumCPU)
	cols, vals = appendIfNonNil(cols, vals, "num_go_routine", sr.NumGoRoutine)
	cols, vals = appendIfNonEmpty(cols, vals, "version", sr.Version)

	var valuePlaceholders []string
	for i := range vals {
		if *dbDriver == "mysql" {
			valuePlaceholders = append(valuePlaceholders, "?")
		} else {
			valuePlaceholders = append(valuePlaceholders, fmt.Sprintf("$%d", i+1))
		}
	}
	qry := fmt.Sprintf("INSERT INTO dendrite_stats (%s) VALUES (%s)", strings.Join(cols, ", "), strings.Join(valuePlaceholders, ", "))
	_, err := db.Exec(qry, vals...)
	return err
}
