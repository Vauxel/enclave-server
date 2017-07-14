package main

import (
	"log"
	"net/http"
)

const serverversion string = "0.1"

func main() {
	InitServerDB("./data.db")
	InitUsersTable()
	log.Println("Database successfully initialized")

	InitServerKeyPair("./keypair.json")
	log.Println("Server keypair loaded with the public key:", serverkeypair.Public)

	apirouter := ApiRouter()
	go http.ListenAndServe(":8242", apirouter)
	log.Println("API initialized and waiting for requests")

	go socketmanager.Listen()
	log.Println("Socket opened and started listening")

	select{}
}