package main

import (
	"time"
	"net/http"
	"database/sql"
	"github.com/ventu-io/go-shortid"
)

type User struct {
	ID string `json:"id"`
	Name string `json:"name"`
	PublicKey string `json:"publickey"`
	SharedSecret []byte `json:"sharedsecret"`
	Joined int64 `json:"joined"`
}

func InitUsersTable() {
	sql_query := `
	CREATE TABLE IF NOT EXISTS users(
		ID TEXT NOT NULL PRIMARY KEY,
		Name TEXT,
		PublicKey TEXT,
		SharedSecret BINARY,
		Joined INT
	);
	`

	_, err := ServerDB.Exec(sql_query)

	if err != nil {
		panic(err)
	}
}

func FindUserIDByPublicKey(publicKey string) string {
	var userid string

	sql_query := `SELECT ID FROM users WHERE users.PublicKey = ?`
	err := ServerDB.QueryRow(sql_query, publicKey).Scan(&userid)

	if err != nil {
		if err == sql.ErrNoRows {
			return ""
		} else {
			panic(err)
		}
	}

	return userid
}

func CreateUser(publicKey string) *User {
	sharedSecret := GenerateSharedSecret(publicKey, serverkeypair.Private);
	uid, _ := shortid.Generate()

	user := &User{
		ID: uid,
		Name: "Defaulto",
		PublicKey: publicKey,
		SharedSecret: sharedSecret,
		Joined: time.Now().Unix(),
	}

	sql_query := `
	INSERT INTO users(
		ID,
		Name,
		PublicKey,
		SharedSecret,
		Joined
	) values(?, ?, ?, ?, ?)
	`

	stmt, err := ServerDB.Prepare(sql_query)

	if err != nil {
		panic(err)
	}

	result, err := stmt.Exec(user.ID, user.Name, user.PublicKey, user.SharedSecret, user.Joined)
	result = result

	if err != nil {
		panic(err)
	}

	stmt.Close()
	return user
}

// Debug Functions
func ListUsersFrontEnd(w http.ResponseWriter, r *http.Request) {
	sql_query := `
	SELECT ID, Name, PublicKey, SharedSecret, Joined FROM users
	`

	rows, err := ServerDB.Query(sql_query)

	if err != nil {
		panic(err)
	}

	var users []User

	for rows.Next() {
		user := User{}
		err2 := rows.Scan(&user.ID, &user.Name, &user.PublicKey, &user.SharedSecret, &user.Joined)

		if err2 != nil {
			panic(err2)
		}

		users = append(users, user)
	}

	rows.Close()
	WriteJSON(w, users)
}

func UpdateUsername(w http.ResponseWriter, r *http.Request) {
	sql_query := `
	SELECT ID, Name, PublicKey, SharedSecret, Joined FROM users
	`

	rows, err := ServerDB.Query(sql_query)

	if err != nil {
		panic(err)
	}

	var users []User

	for rows.Next() {
		user := User{}
		err2 := rows.Scan(&user.ID, &user.Name, &user.PublicKey, &user.SharedSecret, &user.Joined)

		if err2 != nil {
			panic(err2)
		}

		users = append(users, user)
	}

	rows.Close()
	WriteJSON(w, users)
}

func CreateUserFrontEnd(w http.ResponseWriter, r *http.Request) {
	user := CreateUser(GenerateKeyPair().Public)
	WriteJSON(w, user)
}