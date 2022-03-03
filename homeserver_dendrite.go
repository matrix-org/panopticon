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

import "database/sql"

// Dendrite specific stats
type ReportStatsDendrite struct {
	GoOS               string  `json:"go_os,omitempty"`
	GoArch             string  `json:"go_arch,omitempty"`
	GoVersion          string  `json:"go_version,omitempty"`
	FederationDisabled *bool   `json:"federation_disabled,omitempty"`
	Monolith           *bool   `json:"monolith,omitempty"`
	NATSEmbedded       *bool   `json:"nats_embedded,omitempty"`
	NATSInMemory       *bool   `json:"nats_in_memory,omitempty"`
	NumCPU             *int64  `json:"num_cpu,omitempty"`
	NumGoRoutine       *int64  `json:"num_go_routine,omitempty"`
	Version            string `json:"version,omitempty"`
}

func createTableDendrite(db *sql.DB) error {
	autoincrement := "AUTOINCREMENT"
	if *dbDriver == "mysql" {
		autoincrement = "AUTO_INCREMENT"
	}
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS dendrite_stats(
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
