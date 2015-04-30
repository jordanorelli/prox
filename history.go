package main

import (
	"bytes"
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"io/ioutil"
	"log"
	"net/http"
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
	var dbpath string
	var err error

	err = conf.Get("dbpath", &dbpath)
	if err != nil {
		return err
	}
	log.Printf("opening sqlite file at %s", dbpath)

	db, err = sql.Open("sqlite3", dbpath)
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
        blocked integer not null default 0
    );

    create table if not exists requests (
        id text primary key,
        host text not null,
        path text not null
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
	_, err := db.Exec(`insert or ignore into requests (id, host, path)
    values (?, ?, ?)`, id.String(), r.URL.Host, r.URL.Path)
	if err != nil {
		log.Printf("unable to save request: %v", err)
		return
	}
}
