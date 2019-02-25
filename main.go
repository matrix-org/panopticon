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
	Homeserver           string
	LocalTimestamp       int64    // Seconds since epoch, UTC
	RemoteTimestamp      *int64   `json:"timestamp"` // Seconds since epoch, UTC
	UptimeSeconds        *int64   `json:"uptime_seconds"`
	TotalUsers           *int64   `json:"total_users"`
	TotalNonBridgedUsers *int64   `json:"total_nonbridged_users"`
	TotalRoomCount       *int64   `json:"total_room_count"`
	DailyActiveUsers     *int64   `json:"daily_active_users"`
	DailyMessages        *int64   `json:"daily_messages"`
	DailySentMessages    *int64   `json:"daily_sent_messages"`
	DailyActiveRooms     *int64   `json:"daily_active_rooms"`
	R30UsersAll          *int64   `json:"r30_users_all"`
	R30UsersAndroid      *int64   `json:"r30_users_android"`
	R30UsersIOS          *int64   `json:"r30_users_ios"`
	R30UsersElectron     *int64   `json:"r30_users_electron"`
	R30UsersWeb          *int64   `json:"r30_users_web"`
	MemoryRSS            *int64   `json:"memory_rss"`
	CPUAverage           *int64   `json:"cpu_average"`
	CacheFactor          *float64 `json:"cache_factor"`
	EventCacheSize       *int64   `json:"event_cache_size"`
	DailyUserTypeNative  *int64   `json:"daily_user_type_native"`
	DailyUserTypeGuest   *int64   `json:"daily_user_type_guest"`
	DailyUserTypeBridged *int64   `json:"daily_user_type_bridged"`
	PythonVersion        string   `json:"python_version"`
	RemoteAddr           string
	XForwardedFor        string
	UserAgent            string
}

func main() {
	flag.Parse()

	db, err := sql.Open(*dbDriver, *dbPath)
	if err != nil {
		log.Fatalf("Could not open database: %v", err)
	}
	defer db.Close()

	if err := createTable(db); err != nil {
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
	if err := r.Save(sr); err != nil {
		logAndReplyError(w, err, 500, "Error saving to DB")
		return
	}
	io.WriteString(w, "{}")
}

func (r *Recorder) Save(sr StatsReport) error {
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

	cols, vals = appendIfNonNil(cols, vals, "r30_users_all", sr.R30UsersAll)
	cols, vals = appendIfNonNil(cols, vals, "r30_users_android", sr.R30UsersAndroid)
	cols, vals = appendIfNonNil(cols, vals, "r30_users_ios", sr.R30UsersIOS)
	cols, vals = appendIfNonNil(cols, vals, "r30_users_electron", sr.R30UsersElectron)
	cols, vals = appendIfNonNil(cols, vals, "r30_users_web", sr.R30UsersWeb)

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

	var valuePlaceholders []string
	for i := range vals {
		if *dbDriver == "mysql" {
			valuePlaceholders = append(valuePlaceholders, "?")
		} else {
			valuePlaceholders = append(valuePlaceholders, fmt.Sprintf("$%d", i+1))
		}
	}
	_, err := r.DB.Exec(`INSERT INTO stats (
			`+strings.Join(cols, ", ")+`
		) VALUES (`+strings.Join(valuePlaceholders, ", ")+`)`,
		vals...,
	)
	return err
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

func createTable(db *sql.DB) error {
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
		r30_users_all BIGINT,
		r30_users_android BIGINT,
		r30_users_ios BIGINT,
		r30_users_electron BIGINT,
		r30_users_web BIGINT,
		cpu_average BIGINT,
		memory_rss BIGINT,
		cache_factor DOUBLE,
		event_cache_size BIGINT,
		user_agent TEXT,
		daily_user_type_native BIGINT,
		daily_user_type_bridged BIGINT,
		daily_user_type_guest BIGINT,
		python_version TEXT,
		)`)
	return err
}
