/*
Copyright 2015 - 2017 OpenMarket Ltd
Copyright 2017 Vector Creations Ltd
Copyright 2017, 2018 New Vector Ltd

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// panopticon collects statistics posted to it, and records them in a sqlite3, mysql or postgres database.
package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	_ "github.com/lib/pq"
)

var (
	dbDriver = flag.String("db-driver", "sqlite3", "the database driver to use")
	dbPath   = flag.String("db", "stats.db", "the data source to use, for sqlite this is the path to the file")
	port     = flag.Int("port", 9001, "Port on which to serve HTTP")
)

type StatsReport struct {
	ReportStatsSynapse
	ReportStatsDendrite
}

// CommonStats defines statistics every server should report to be comparable.
// Uncommon statistics should be added to the specific homeserver struct.
type CommonStats struct {
	Homeserver            string
	LocalTimestamp        int64  // Seconds since epoch, UTC
	RemoteTimestamp       *int64 `json:"timestamp"`                // Seconds since epoch, UTC
	UptimeSeconds         *int64 `json:"uptime_seconds"`           // Seconds since last restart
	TotalUsers            *int64 `json:"total_users"`              // Total users in users table
	TotalNonBridgedUsers  *int64 `json:"total_nonbridged_users"`   // Total native and guest users in users table
	TotalRoomCount        *int64 `json:"total_room_count"`         // Total number of rooms on the server
	DailyActiveUsers      *int64 `json:"daily_active_users"`       // Total number of users in the users ips table seen in the last 24 hours
	DailyMessages         *int64 `json:"daily_messages"`           // Total number of m.room.message in events table in the past 24 hours
	DailySentMessages     *int64 `json:"daily_sent_messages"`      // Total number of m.room.message in events table in the past 24 hours sent from host server
	DailyActiveRooms      *int64 `json:"daily_active_rooms"`       // Total number of rooms with a m.room.message in the event table in the past 24 hours
	DailyE2eeMessages     *int64 `json:"daily_e2ee_messages"`      // Total number of m.room.encrypted in events table in the past 24 hours
	DailySentE2eeMessages *int64 `json:"daily_sent_e2ee_messages"` // Total number of m.room.encrypted in events table in the past 24 hours sent from host server
	DailyActiveE2eeRooms  *int64 `json:"daily_active_e2ee_rooms"`  // Total number of rooms with a m.room.encrypted in the event table in the past 24 hours
	MonthlyActiveUsers    *int64 `json:"monthly_active_users"`     // Total number of users in the users ips table seen in the last 30 days
	R30UsersAll           *int64 `json:"r30_users_all"`            // r30 stat for all users regardless of client
	R30UsersAndroid       *int64 `json:"r30_users_android"`        // r30 stat considering only Riot Android
	R30UsersIOS           *int64 `json:"r30_users_ios"`            // r30 stat considering only Riot iOS
	R30UsersElectron      *int64 `json:"r30_users_electron"`       // r30 stat considering only Riot Electron
	R30UsersWeb           *int64 `json:"r30_users_web"`            // r30 stat considering only web clients (must assume they are Riot)
	R30V2UsersAll         *int64 `json:"r30v2_users_all"`          // r30v2 stat for all users regardless of client
	R30V2UsersAndroid     *int64 `json:"r30v2_users_android"`      // r30v2 stat considering only Riot Android
	R30V2UsersIOS         *int64 `json:"r30v2_users_ios"`          // r30v2 stat considering only Riot iOS
	R30V2UsersElectron    *int64 `json:"r30v2_users_electron"`     // r30v2 stat considering only Riot Electron
	R30V2UsersWeb         *int64 `json:"r30v2_users_web"`          // r30v2 stat considering only web clients (must assume they are Riot)
	MemoryRSS             *int64 `json:"memory_rss"`
	CPUAverage            *int64 `json:"cpu_average"`
	DailyUserTypeNative   *int64 `json:"daily_user_type_native"`  // New native users in users table in last 24 hours
	DailyUserTypeGuest    *int64 `json:"daily_user_type_guest"`   // New guest users in users table in the last 24 hours
	DailyUserTypeBridged  *int64 `json:"daily_user_type_bridged"` // New bridged users in the users table in the last 24 hours
	DatabaseEngine        string `json:"database_engine"`
	DatabaseServerVersion string `json:"database_server_version"`
	LogLevel              string `json:"log_level"`
	RemoteAddr            string
	XForwardedFor         string
	UserAgent             string
}

func main() {
	flag.Parse()

	db, err := sql.Open(*dbDriver, *dbPath)
	if err != nil {
		log.Fatalf("Could not open database: %v", err)
	}
	defer db.Close()

	if err := createTableSynapse(db); err != nil {
		log.Fatalf("Error creating database: %v", err)
	}
	if err := createTableDendrite(db); err != nil {
		log.Fatalf("Error creating database: %v", err)
	}

	r := &Recorder{db}

	http.HandleFunc("/push", r.Handle)
	http.HandleFunc("/test", serveText("ok"))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}

type Recorder struct {
	DB *sql.DB
}

func (r *Recorder) Handle(w http.ResponseWriter, req *http.Request) {
	dec := json.NewDecoder(req.Body)
	defer req.Body.Close()
	var sr StatsReport
	if err := dec.Decode(&sr); err != nil {
		logAndReplyError(w, err, 400, "Error decoding JSON")
		return
	}
	sr.LocalTimestamp = time.Now().UTC().Unix()
	sr.RemoteAddr = req.RemoteAddr
	sr.XForwardedFor = req.Header.Get("X-Forwarded-For")
	sr.UserAgent = req.Header.Get("User-Agent")
	if err := r.Save(sr, strings.HasPrefix(sr.UserAgent, "Dendrite")); err != nil {
		logAndReplyError(w, err, 500, "Error saving to DB")
		return
	}
	io.WriteString(w, "{}")
}

func (r *Recorder) Save(sr StatsReport, isDendrite bool) error {
	if isDendrite {
		s := sr.ReportStatsDendrite
		s.Common = sr.ReportStatsSynapse.CommonStats
		return s.Save(r.DB)
	}
	return sr.ReportStatsSynapse.Save(r.DB)
}

func appendIfNonNilBool(cols []string, vals []interface{}, name string, value *bool) ([]string, []interface{}) {
	if value != nil {
		cols = append(cols, name)
		vals = append(vals, value)
	}
	return cols, vals
}

func appendIfNonNilFloat(cols []string, vals []interface{}, name string, value *float64) ([]string, []interface{}) {
	if value != nil {
		cols = append(cols, name)
		vals = append(vals, value)
	}
	return cols, vals
}
func appendIfNonNil(cols []string, vals []interface{}, name string, value *int64) ([]string, []interface{}) {
	if value != nil {
		cols = append(cols, name)
		vals = append(vals, value)
	}
	return cols, vals
}

func appendIfNonEmpty(cols []string, vals []interface{}, name string, value string) ([]string, []interface{}) {
	if value != "" {
		cols = append(cols, name)
		vals = append(vals, value)
	}
	return cols, vals
}

func logAndReplyError(w http.ResponseWriter, err error, code int, description string) {
	log.Printf("%s: %v", description, err)
	w.WriteHeader(code)
	io.WriteString(w, `{"error_message": "unable to process request"}`)
}

func serveText(s string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		io.WriteString(w, s)
	}
}
