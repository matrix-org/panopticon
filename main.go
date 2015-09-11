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
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var (
	dbPath   = flag.String("db", "stats.db", "Path to sqlite3 database")
	port     = flag.Int("port", 9001, "Port on which to serve HTTPS")
	certFile = flag.String("cert_file", "tls.crt", "Path to SSL cert")
	keyFile  = flag.String("key_file", "tls.key", "Path to SSL key")
)

type StatsReport struct {
	Homeserver       string
	LocalTimestamp   int64 // Seconds since epoch, UTC
	RemoteTimestamp  int64 `json:"timestamp"` // Seconds since epoch, UTC
	TotalUsers       int64 `json:"total_users"`
	TotalRoomCount   int64 `json:"total_room_count"`
	DailyActiveUsers int64 `json:"daily_active_users"`
	DailyMessages    int64 `json:"daily_messages"`
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
	http.HandleFunc("/healthz", serveText("ok"))
	log.Fatal(http.ListenAndServeTLS(fmt.Sprintf(":%d", *port), *certFile, *keyFile, nil))
}

type Recorder struct {
	DB *sql.DB
}

func (r *Recorder) Handle(w http.ResponseWriter, req *http.Request) {
	dec := json.NewDecoder(req.Body)
	defer req.Body.Close()
	var sr StatsReport
	if err := dec.Decode(&sr); err != nil {
		logAndReplyError(w, err)
		return
	}
	sr.LocalTimestamp = time.Now().UTC().Unix()
	if err := r.Save(sr); err != nil {
		logAndReplyError(w, err)
		return
	}
	io.WriteString(w, "{}")
}

func (r *Recorder) Save(sr StatsReport) error {
	_, err := r.DB.Exec(`INSERT INTO stats (
			homeserver,
			local_timestamp,
			remote_timestamp,
			total_users,
			total_room_count,
			daily_active_users,
			daily_messages
		) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		sr.Homeserver,
		sr.LocalTimestamp,
		sr.RemoteTimestamp,
		sr.TotalUsers,
		sr.TotalRoomCount,
		sr.DailyActiveUsers,
		sr.DailyMessages,
	)
	return err
}

func logAndReplyError(w http.ResponseWriter, err error) {
	log.Printf("Error reading JSON: %v", err)
	w.WriteHeader(500)
	io.WriteString(w, `{"errcode": "unable to process request"}`)
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
		total_users BIGINT,
		total_room_count BIGINT,
		daily_active_users BIGINT,
		daily_messages BIGINT
		)`)
	return err
}
