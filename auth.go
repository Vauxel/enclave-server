package main

import (
	"os"
	"time"
	"fmt"
	"log"
	"strings"
	"strconv"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"database/sql"
	"github.com/GoKillers/libsodium-go/cryptobox"
	"github.com/GoKillers/libsodium-go/cryptoauth"
)

const AUTH_REQUEST_BUFFER = 60 // In seconds

var serverkeypair *KeyPair

type KeyPair struct {
	Public string `json:"public"`
	Private string `json:"private"`
}

type AuthHandshakeRequest struct {
	PublicKey string `json:"publickey"`
}

type AuthHandshakeResponse struct {
	PublicKey string `json:"publickey"`
	UserID string `json:"userid"`
}

func InitServerKeyPair(filepath string) {
	if _, err := os.Stat(filepath); err == nil {
		file, err2 := os.Open(filepath)
		defer file.Close()

		if err2 != nil {
			panic(err2)
		}

		serverkeypair = &KeyPair{}
		err3 := json.NewDecoder(file).Decode(&serverkeypair)

		if err3 != nil {
			panic(err3)
		}
	} else {
		file, err2 := os.Create(filepath)
		defer file.Close()

		if err2 != nil {
			panic(err2)
		}

		serverkeypair = GenerateKeyPair()
		err3 := json.NewEncoder(file).Encode(serverkeypair)

		if err3 != nil {
			panic(err3)
		}
	}
}

func EncodeKey(rawKey []byte) (encodedKey string) {
	encodedKey = hex.EncodeToString(rawKey)
	return
}

func DecodeKey(encodedKey string) (rawKey []byte) {
	rawKey, err := hex.DecodeString(encodedKey)

	if err != nil {
		panic(err)
	}

	return
}

func GenerateKeyPair() *KeyPair {
	privateKey, publicKey, exit := cryptobox.CryptoBoxKeyPair()

	if exit != 0 {
		panic(exit)
	}

	return &KeyPair{
		Public: EncodeKey(publicKey),
		Private: EncodeKey(privateKey),
	}
}

func GenerateSharedSecret(publicKey, privateKey string) (sharedEncryptKey []byte) {
	publicKeyRaw := DecodeKey(publicKey)
	privateKeyRaw := DecodeKey(privateKey)

	sharedEncryptKey, exit := cryptobox.CryptoBoxBeforeNm(publicKeyRaw, privateKeyRaw)

	if exit != 0 {
		panic(exit)
	}

	return
}

func ValidateAuthentication(authparcel string, uniquedata string) (bool, *User) {
	authparcelparts := strings.SplitN(authparcel, " ", 2)

	if len(authparcelparts) != 2 {
		return false, nil
	}

	if authparcelparts[0] != "HMAC" {
		return false, nil
	}

	authstring, err := hex.DecodeString(authparcelparts[1])

	if err != nil {
		return false, nil
	}

	authparts := strings.SplitN(string(authstring), ":", 3)

	if len(authparts) != 3 {
		return false, nil
	}

	userid := authparts[0]
	authtime, err2 := strconv.ParseInt(authparts[1], 0, 64)

	if err2 != nil {
		return false, nil
	}

	authtoken := authparts[2]

	if ((authtime - time.Now().Unix()) > AUTH_REQUEST_BUFFER) || ((authtime - time.Now().Unix()) < -AUTH_REQUEST_BUFFER) {
		return false, nil
	}

	user := &User{}
	sql_query := `SELECT ID, Name, PublicKey, SharedSecret, Joined FROM users WHERE users.ID = ?`
	err3 := ServerDB.QueryRow(sql_query, userid).Scan(&user.ID, &user.Name, &user.PublicKey, &user.SharedSecret, &user.Joined)

	if err3 != nil {
		if err3 == sql.ErrNoRows {
			return false, nil
		} else {
			panic(err3)
		}
	}

	payload := []byte(fmt.Sprintf("%s|%d", uniquedata, authtime))
	trueauthtoken, exit := cryptoauth.CryptoAuth(payload, user.SharedSecret)
	trueauthtoken = trueauthtoken[0:cryptoauth.CryptoAuthBytes()]

	if exit != 0 {
		return false, nil
	}

	return authtoken == hex.EncodeToString(trueauthtoken), user
}

func GetUserFromAuthToken(authparcel string) *User {
	authparcelparts := strings.SplitN(authparcel, " ", 2)
	authstring, err := hex.DecodeString(authparcelparts[1])

	if err != nil {
		panic(err)
	}

	user := &User{}
	sql_query := `SELECT ID, Name, PublicKey, SharedSecret, Joined FROM users WHERE users.ID = ?`
	err3 := ServerDB.QueryRow(sql_query, strings.SplitN(string(authstring), ":", 3)[0]).Scan(&user.ID, &user.Name, &user.PublicKey, &user.SharedSecret, &user.Joined)

	if err3 != nil {
		if err3 == sql.ErrNoRows {
			return nil
		} else {
			panic(err3)
		}
	}

	return user
}

func AuthHandshake(w http.ResponseWriter, r *http.Request) {
	var jsonReq AuthHandshakeRequest
	json.NewDecoder(r.Body).Decode(&jsonReq)

	senderPublicKeyRaw, err := hex.DecodeString(jsonReq.PublicKey)

	if err != nil {
		panic(err)
	}

	if len(senderPublicKeyRaw) != cryptobox.CryptoBoxPublicKeyBytes() {
		log.Println("Client tried to auth with invalid public key")
	}

	var userid string
	userid = FindUserIDByPublicKey(jsonReq.PublicKey)

	if userid == "" {
		userid = CreateUser(jsonReq.PublicKey).ID
	}

	WriteJSON(w, AuthHandshakeResponse{
		PublicKey: serverkeypair.Public,
		UserID: userid,
	})
}

func ValidateRequest(inner http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authed, _ := ValidateAuthentication(r.Header.Get("Authorization"), r.URL.Path)

		if !authed {
			log.Println("Client tried to connect, but not authed")
			http.Error(w, "Invalid authorization token", 401)
			return
		}

		inner.ServeHTTP(w, r)
	})
}