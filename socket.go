package main

import (
	"log"
	"time"
	"bytes"
	"encoding/json"
	"net/http"
	"github.com/gorilla/websocket"
	"github.com/GoKillers/libsodium-go/randombytes"
	"github.com/GoKillers/libsodium-go/cryptosecretbox"
)

type SocketManager struct {
	Clients map[string]*Client
	Register chan *Client
	Unregister chan *Client
}

type Client struct {
	ID string
	User *User
	Conn *websocket.Conn
	Send chan Message
}

type Message interface {
	serialize() []byte
}

type ServerMessage struct {
	Type int `json:"type"`
	Command string `json:"command"`
	Data map[string]interface{} `json:"data"`
}

type TextMessage struct {
	Type int `json:"type"`
	Sender string `json:"sender"`
	Time int64 `json:"time"`
	Message string `json:"message"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize: 1024,
	WriteBufferSize: 1024,
	/*CheckOrigin: func(r *http.Request) bool {
		return true
	},*/
}

var socketmanager = &SocketManager{
	Clients: make(map[string]*Client),
	Register: make(chan *Client),
	Unregister: make(chan *Client),
}

func (sm *ServerMessage) serialize() []byte {
	parsed, err := json.Marshal(sm)

	if err != nil {
		panic(err)
	}

	return parsed
}

func (tm *TextMessage) serialize() []byte {
	parsed, err := json.Marshal(tm)

	if err != nil {
		panic(err)
	}

	return parsed
}

func (sm *SocketManager) AddClient(c *Client) {
	sm.Register <- c
}

func (sm *SocketManager) DeleteClient(c *Client) {
	sm.Unregister <- c
}

func (sm *SocketManager) Listen() {
	for {
		select {
			// Add (Register) a new client
			case client := <-sm.Register:
				log.Println("Registered a new client:", client.ID, client.User.PublicKey)
				sm.Clients[client.ID] = client
				client.Listen()
				sm.BroadcastTextMessage(client.ID + " has joined the server!", "SERVER")
				log.Println("Currently connected clients:", len(sm.Clients))
			// Drop/Delete (Unregister) an existing client
			case client := <-sm.Unregister:
				log.Println("Unregistered client")
				delete(sm.Clients, client.ID)
		}
	}
}

func (c *Client) Listen() {
	go c.ListenWrite()
	go c.ListenRead()
}

func (c *Client) ListenWrite() {
	for {
		msg := <-c.Send
		serialized := msg.serialize()
		encrypted := EncryptMessage(serialized, c.User.SharedSecret)

		log.Println("Sent:", c.ID, string(serialized))
		err := c.Conn.WriteMessage(websocket.BinaryMessage, encrypted)

		if err != nil {
			break
			panic(err)
		}
	}
}

func (c *Client) ListenRead() {
	for {
		_, message, err := c.Conn.ReadMessage()

		if err != nil {
			break
			panic(err)
		}

		decrypted := DecryptMessage(message, c.User.SharedSecret)
		parts := bytes.SplitN(decrypted, []byte("|"), 2)

		cmdName := parts[0]
		data := parts[1]

		log.Println("Received:", c.ID, string(cmdName), string(data))
		ParseCommand(cmdName, data).Handle(c)
	}
}

func (sm *SocketManager) BroadcastMessage(msg Message) {
	for _, c := range sm.Clients {
		c.SendMessage(msg)
	}
}

func (sm *SocketManager) BroadcastTextMessage(message, id string) {
	sm.BroadcastMessage(&TextMessage{
		Type: 2,
		Sender: id,
		Time: time.Now().Unix(),
		Message: message,
	})
}

func (c *Client) SendMessage(msg Message) {
	c.Send <- msg
}

func (c *Client) SendTextMessage(message, id string) {
	c.SendMessage(&TextMessage{
		Type: 2,
		Sender: id,
		Time: time.Now().Unix(),
		Message: message,
	})
}

func EncryptMessage(message, secretkey []byte) []byte {
	nonce := randombytes.RandomBytes(secretbox.CryptoSecretBoxNonceBytes())
	cipher, exit := secretbox.CryptoSecretBoxEasy(message, nonce, secretkey)

	if exit != 0 {
		panic(exit)
	}

	return append(nonce, cipher...)
}

func DecryptMessage(cipher, secretkey []byte) []byte {
	nonce := cipher[0:secretbox.CryptoSecretBoxNonceBytes()]
	cipher = cipher[len(nonce):len(cipher)]
	message, exit := secretbox.CryptoSecretBoxOpenEasy(cipher, nonce, secretkey)

	if exit != 0 {
		panic(exit)
	}

	return message
}

func SocketLogin(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromAuthToken(r.Header.Get("Authorization"))
	ws, err := upgrader.Upgrade(w, r, nil)

	log.Println("Client", user.ID, "connected and connection upgraded")

	if err != nil {
		panic(err)
	}

	socketmanager.AddClient(&Client{
		ID: user.ID,
		User: user,
		Conn: ws,
		Send: make(chan Message),
	})
}