package main

import (
	"bytes"
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

var lastId int
var db *sql.DB

func freezeRequest(r *http.Request) error {
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r.Body); err != nil {
		return fmt.Errorf("unable to clone request: error reading original request body: %s", err)
	}

	if err := r.Body.Close(); err != nil {
		return fmt.Errorf("unable to clone request: cannot close original request body: %s", err)
	}
	r.Body = ioutil.NopCloser(&buf)
	return nil
}

func openDB() error {
	var err error

	log.Printf("opening sqlite file at %s", config.DbPath)
	db, err = sql.Open("sqlite3", config.DbPath)
	if err != nil {
		return err
	}

	setupDB()
	return nil
}

func setupDB() {
	sql := `
    create table if not exists domains (
        hostname text primary key,
        blocked  integer not null default 0
    );

    create table if not exists requests (
        id       text primary key,
        host     text not null,
        path     text not null
    );

    create table if not exists responses (
        id           text primary key,
        status       integer,
        length       integer,          -- content-length header
        content_type text,             -- content-type header
        duration     integer           -- time taken in milliseconds
    );
    `
	res, err := db.Exec(sql)
	if err != nil {
		log.Printf("unable to setup db: %v", err)
	} else {
		log.Printf("db was set up: %v", res)
	}
}

func saveHostname(hostname string) {
	res, err := db.Exec(`insert or ignore into domains (hostname) values (?)`, hostname)
	if err != nil {
		log.Printf("unable to save hostname: %v", err)
		return
	}
	if n, _ := res.RowsAffected(); n > 0 {
		log.Printf("saved new hostname: %s", hostname)
	}
}

func saveRequest(id RequestId, r *http.Request) {
	_, err := db.Exec(`insert into requests (id, host, path)
    values (?, ?, ?)`, id.String(), r.URL.Host, r.URL.Path)
	if err != nil {
		log.Printf("unable to save request: %v", err)
		return
	}
}

func saveResponse(id RequestId, res *http.Response, elapsed time.Duration) {
	_, err := db.Exec(`insert into responses (id, status, length, content_type, duration)
    values (?, ?, ?, ?, ?)`, id.String(), res.StatusCode, res.ContentLength, res.Header.Get("Content-Type"), elapsed.Nanoseconds()/1000000)
	if err != nil {
		log.Printf("unable to save response: %v", err)
		return
	}
}
