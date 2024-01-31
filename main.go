package main

import (
	"github.com/mbund/nomadic-vpn/cmd"
	"github.com/mbund/nomadic-vpn/db"
)

func main() {
	database, err := db.CreateDb("nomadic-vpn.db")
	if err != nil {
		panic(err)
	}
	db.DB = database

	cmd.Execute()
}
