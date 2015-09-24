// panopticon collects statistics posted to it, and records them in a sqlite3 database.
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

	_ "github.com/mattn/go-sqlite3"
)

var (
	dbPath = flag.String("db", "stats.db", "Path to sqlite3 database")
	port   = flag.Int("port", 9001, "Port on which to serve HTTP")
)

type StatsReport struct {
	Homeserver       string
	LocalTimestamp   int64  // Seconds since epoch, UTC
	RemoteTimestamp  *int64 `json:"timestamp"` // Seconds since epoch, UTC
	UptimeSeconds    *int64 `json:"uptime_seconds"`
	TotalUsers       *int64 `json:"total_users"`
	TotalRoomCount   *int64 `json:"total_room_count"`
	DailyActiveUsers *int64 `json:"daily_active_users"`
	DailyMessages    *int64 `json:"daily_messages"`
	RemoteAddr       string
	XForwardedFor    string
	UserAgent        string
}

func main() {
	flag.Parse()

	db, err := sql.Open("sqlite3", *dbPath)
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
	cols, vals = appendIfNonNil(cols, vals, "total_room_count", sr.TotalRoomCount)
	cols, vals = appendIfNonNil(cols, vals, "daily_active_users", sr.DailyActiveUsers)
	cols, vals = appendIfNonNil(cols, vals, "daily_messages", sr.DailyMessages)
	cols, vals = appendIfNonEmpty(cols, vals, "forwarded_for", sr.XForwardedFor)
	cols, vals = appendIfNonEmpty(cols, vals, "user_agent", sr.UserAgent)

	var valuePlaceholders []string
	for i := range vals {
		valuePlaceholders = append(valuePlaceholders, fmt.Sprintf("$%d", i+1))
	}
	_, err := r.DB.Exec(`INSERT INTO stats (
			`+strings.Join(cols, ", ")+`
		) VALUES (`+strings.Join(valuePlaceholders, ", ")+`)`,
		vals...,
	)
	return err
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
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS stats(
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		homeserver VARCHAR(256),
		local_timestamp BIGINT,
		remote_timestamp BIGINT,
		remote_addr TEXT,
		forwarded_for TEXT,
		uptime_seconds BIGINT,
		total_users BIGINT,
		total_room_count BIGINT,
		daily_active_users BIGINT,
		daily_messages BIGINT,
		user_agent STRING
		)`)
	return err
}
