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

import "database/sql"

// Synapse specific stats
type ReportStatsSynapse struct {
	Homeserver            string
	LocalTimestamp        int64    // Seconds since epoch, UTC
	RemoteTimestamp       *int64   `json:"timestamp"`                // Seconds since epoch, UTC
	UptimeSeconds         *int64   `json:"uptime_seconds"`           // Seconds since last restart
	TotalUsers            *int64   `json:"total_users"`              // Total users in users table
	TotalNonBridgedUsers  *int64   `json:"total_nonbridged_users"`   // Total native and guest users in users table
	TotalRoomCount        *int64   `json:"total_room_count"`         // Total number of rooms on the server
	DailyActiveUsers      *int64   `json:"daily_active_users"`       // Total number of users in the users ips table seen in the last 24 hours
	DailyMessages         *int64   `json:"daily_messages"`           // Total number of m.room.message in events table in the past 24 hours
	DailySentMessages     *int64   `json:"daily_sent_messages"`      // Total number of m.room.message in events table in the past 24 hours sent from host server
	DailyActiveRooms      *int64   `json:"daily_active_rooms"`       // Total number of rooms with a m.room.message in the event table in the past 24 hours
	DailyE2eeMessages     *int64   `json:"daily_e2ee_messages"`      // Total number of m.room.encrypted in events table in the past 24 hours
	DailySentE2eeMessages *int64   `json:"daily_sent_e2ee_messages"` // Total number of m.room.encrypted in events table in the past 24 hours sent from host server
	DailyActiveE2eeRooms  *int64   `json:"daily_active_e2ee_rooms"`  // Total number of rooms with a m.room.encrypted in the event table in the past 24 hours
	MonthlyActiveUsers    *int64   `json:"monthly_active_users"`     // Total number of users in the users ips table seen in the last 30 days
	R30UsersAll           *int64   `json:"r30_users_all"`            // r30 stat for all users regardless of client
	R30UsersAndroid       *int64   `json:"r30_users_android"`        // r30 stat considering only Riot Android
	R30UsersIOS           *int64   `json:"r30_users_ios"`            // r30 stat considering only Riot iOS
	R30UsersElectron      *int64   `json:"r30_users_electron"`       // r30 stat considering only Riot Electron
	R30UsersWeb           *int64   `json:"r30_users_web"`            // r30 stat considering only web clients (must assume they are Riot)
	R30V2UsersAll         *int64   `json:"r30v2_users_all"`          // r30v2 stat for all users regardless of client
	R30V2UsersAndroid     *int64   `json:"r30v2_users_android"`      // r30v2 stat considering only Riot Android
	R30V2UsersIOS         *int64   `json:"r30v2_users_ios"`          // r30v2 stat considering only Riot iOS
	R30V2UsersElectron    *int64   `json:"r30v2_users_electron"`     // r30v2 stat considering only Riot Electron
	R30V2UsersWeb         *int64   `json:"r30v2_users_web"`          // r30v2 stat considering only web clients (must assume they are Riot)
	MemoryRSS             *int64   `json:"memory_rss"`
	CPUAverage            *int64   `json:"cpu_average"`
	CacheFactor           *float64 `json:"cache_factor"`
	EventCacheSize        *int64   `json:"event_cache_size"`
	DailyUserTypeNative   *int64   `json:"daily_user_type_native"`  // New native users in users table in last 24 hours
	DailyUserTypeGuest    *int64   `json:"daily_user_type_guest"`   // New guest users in users table in the last 24 hours
	DailyUserTypeBridged  *int64   `json:"daily_user_type_bridged"` // New bridged users in the users table in the last 24 hours
	PythonVersion         string   `json:"python_version"`
	DatabaseEngine        string   `json:"database_engine"`
	DatabaseServerVersion string   `json:"database_server_version"`
	ServerContext         string   `json:"server_context"`
	LogLevel              string   `json:"log_level"`
	RemoteAddr            string
	XForwardedFor         string
	UserAgent             string
}

func createTableSynapse(db *sql.DB) error {
	autoincrement := "AUTOINCREMENT"
	if *dbDriver == "mysql" {
		autoincrement = "AUTO_INCREMENT"
	}
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS stats(
		id INTEGER NOT NULL PRIMARY KEY ` + autoincrement + ` ,
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
		cache_factor DOUBLE,
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
