package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/nanobox-io/golang-scribble"
)

var ServerConfig *scribble.Driver
var ServerDB *sql.DB

func InitServerConfig(dir string) {
	var err error
	ServerConfig, err = scribble.New(dir, nil)

	if err != nil {
		panic(err)
	}
}

func InitServerDB(filepath string) {
	var err error
	ServerDB, err = sql.Open("sqlite3", filepath)

	if err != nil {
		panic(err)
	}

	if ServerDB == nil {
		panic("db nil")
	}
}