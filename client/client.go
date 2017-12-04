package client

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Client struct {
	Conn *websocket.Conn

	ProcessChannel chan []byte
	SendChannel    chan []byte
}

func (c *Client) ReadPump() {
	defer c.Conn.Close()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			fmt.Println("Closing a connection")
			break
		}

		c.ProcessChannel <- message
		fmt.Println("Received msg:", string(message[:40])+"...")
	}
}

func (c *Client) ProcessPump() {
	defer c.Conn.Close()

	outFile, _ := os.Create("out.ogg")

	var i int = 1
	for {
		select {
		case message := <-c.ProcessChannel:
			decoded, err := base64.StdEncoding.DecodeString(strings.Split(string(message), ",")[1])
			if err != nil {
				log.Fatal("Bad b64: ", err)
			}
			outFile.Write(decoded)

			i += 1
			processedMessage := []byte(fmt.Sprintf("Received %d bytes", len(message)))
			c.SendChannel <- processedMessage
		}
	}
}

func (c *Client) WritePump() {
	defer c.Conn.Close()

	for {
		select {
		case message := <-c.SendChannel:
			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Fatal("No writer")
			}

			w.Write(message)
			w.Close()
		}
	}
}

func ServeClient(w http.ResponseWriter, r *http.Request) {
	fmt.Println("New client")

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal("upgrader.Upgrade: ", err)
	}

	client := Client{
		Conn:           conn,
		ProcessChannel: make(chan []byte, 256),
		SendChannel:    make(chan []byte, 256),
	}

	go client.ReadPump()
	go client.ProcessPump()
	go client.WritePump()
}
