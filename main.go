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

// panopticon collects statistics posted to it, and records them in a sqlite3 or mysql database.
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
	tableName := "stats"
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
	if !isDendrite {
		cols, vals = appendIfNonNilFloat(cols, vals, "cache_factor", sr.CacheFactor)
		cols, vals = appendIfNonNil(cols, vals, "event_cache_size", sr.EventCacheSize)
	}
	cols, vals = appendIfNonNil(cols, vals, "daily_user_type_native", sr.DailyUserTypeNative)
	cols, vals = appendIfNonNil(cols, vals, "daily_user_type_guest", sr.DailyUserTypeGuest)
	cols, vals = appendIfNonNil(cols, vals, "daily_user_type_bridged", sr.DailyUserTypeBridged)
	if !isDendrite {
		cols, vals = appendIfNonEmpty(cols, vals, "python_version", sr.PythonVersion)
	}
	cols, vals = appendIfNonEmpty(cols, vals, "database_engine", sr.DatabaseEngine)
	cols, vals = appendIfNonEmpty(cols, vals, "database_server_version", sr.DatabaseServerVersion)

	if !isDendrite {
		cols, vals = appendIfNonEmpty(cols, vals, "server_context", sr.ServerContext)
	}
	cols, vals = appendIfNonEmpty(cols, vals, "log_level", sr.LogLevel)

	if isDendrite {
		tableName = "dendrite_stats"
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
	}

	var valuePlaceholders []string
	for i := range vals {
		if *dbDriver == "mysql" {
			valuePlaceholders = append(valuePlaceholders, "?")
		} else {
			valuePlaceholders = append(valuePlaceholders, fmt.Sprintf("$%d", i+1))
		}
	}
	qry := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", tableName, strings.Join(cols, ", "), strings.Join(valuePlaceholders, ", "))
	_, err := r.DB.Exec(qry, vals...)
	return err
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

