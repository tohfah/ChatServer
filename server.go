package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

const (
	PORT     = ":8080"
	TYPE 	 = "tcp"

	CMD_NAME    = "/name"
	CMD_MSG     = "/msg"
	CMD_QUIT    = "/quit"
	CMD_HELP    = "/help"

	NAME = "anon"
)

type MainRoom struct {
	clients  []*Client
	incoming chan *Message
	join     chan *Client
	quit     chan *Client
}

type Client struct {
	name     string
	incoming chan *Message
	outgoing chan string
	conn     net.Conn
	reader   *bufio.Reader
	writer   *bufio.Writer
}

type Message struct {
	client *Client
	text   string
}

func NewMainRoom() *MainRoom {
	mainRoom := &MainRoom{
		clients:  make([]*Client, 0),
		incoming: make(chan *Message),
		join:     make(chan *Client),
		quit:     make(chan *Client),
	}
	mainRoom.Listen()
	return mainRoom
}

func NewClient(conn net.Conn) *Client {
	writer := bufio.NewWriter(conn)
	reader := bufio.NewReader(conn)

	client := &Client{
		name:     CLIENT_NAME,
		incoming: make(chan *Message),
		outgoing: make(chan string),
		conn:     conn,
		reader:   reader,
		writer:   writer,
	}

}

func NewMessage(client *Client, text string) *Message {
	return &Message{
		client: client,
		text:   text,
	}
}

// constantly listen to all channels in main room
func (mainRoom *MainRoom) Listen() {
	go func() {
		for {
			select {
				case message := <-mainRoom.incoming:
					mainRoom.Parse(message)
				case client := <-mainRoom.join:
					mainRoom.Join(client)
				case client := <-mainRoom.quit:
					client.conn.Close()
			}
		}
	}()
}

func main() {

	mainRoom := NewMainRoom()

	listener, err := net.Listen(TYPE, PORT)
	if err != nil {
		log.Println("Error: ", err)
		os.Exit(1)
	}
	defer listener.Close()
	log.Println("Listening on " + PORT)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error: ", err)
			continue
		}
		mainRoom.Join(NewClient(conn))
	}
}
