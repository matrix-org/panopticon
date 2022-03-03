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
	if isDendrite {
		s := sr.ReportStatsDendrite
		s.Common = sr.ReportStatsSynapse
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

