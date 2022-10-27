package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

const (
	PORT = ":8080"
	TYPE = "tcp"

	CMD_NAME = "/name"
	CMD_MSG  = "/msg"
	CMD_QUIT = "/quit"
	CMD_HELP = "/help"

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
		name:     NAME,
		incoming: make(chan *Message),
		outgoing: make(chan string),
		conn:     conn,
		reader:   reader,
		writer:   writer,
	}
	fmt.Print(client)
	return client
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

// join clients to main room
func (mainRoom *MainRoom) Join(client *Client) {
	mainRoom.clients = append(mainRoom.clients, client)
	client.outgoing <- "You're connected to the server\n"
	go func() {
		for message := range client.incoming {
			mainRoom.incoming <- message
		}
		mainRoom.quit <- client
	}()
}

// handle commands
func (mainRoom *MainRoom) Parse(message *Message) {
	if strings.HasPrefix(message.text, "/") {
		switch {
		case strings.HasPrefix(message.text, CMD_NAME):
			name := strings.TrimSuffix(strings.TrimPrefix(message.text, CMD_NAME+" "), "\n")
			fmt.Print(name)
			fmt.Print("need to create name func")
		case strings.HasPrefix(message.text, CMD_HELP):
			fmt.Print("need to create help func")
		case strings.HasPrefix(message.text, CMD_QUIT):
			message.client.conn.Close()
		case strings.HasPrefix(message.text, CMD_MSG):
			fmt.Print("need to create msg func")
		default:
			message.client.outgoing <- "Unknown command. Type /help for a list of available commands."
		}
	} else {
		fmt.Print("need to create msg func")
	}

}

// Reads in from client's socket and place msg on incoming channel
func (client *Client) Read() {
	for {
		str, err := client.reader.ReadString('\n')
		if err != nil {
			log.Println(err)
			break
		}
		message := NewMessage(client, strings.TrimSuffix(str, "\n"))
		client.incoming <- message
	}
	close(client.incoming)
	log.Println("Closed client's incoming channel.")
}

// Reads from the Client's outgoing channel & write to client's socket
func (client *Client) Write() {
	for str := range client.outgoing {
		_, err := client.writer.WriteString(str)
		if err != nil {
			log.Println(err)
			break
		}
		err = client.writer.Flush()
		if err != nil {
			log.Println(err)
			break
		}
	}
	log.Println("Closed client's write thread")
}

func (client *Client) Listen() {
	go client.Read()
	go client.Write()
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
