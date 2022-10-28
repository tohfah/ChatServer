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
	PORT = ":3030"
	TYPE = "tcp"

	CMD_MSG   = "/msg"
	CMD_SHOUT = "/shout"
	CMD_NAME  = "/name"
	CMD_HELP  = "/help"
	CMD_QUIT  = "/quit"

	CLIENT_NAME = "Anon"
	PASSWORD    = "password"

	NAME_CHANGE = ">> Your name has been changed to \"%s\".\n"
)

// A MainRoom receives messages on its channels, and keeps track of the currently
// connected clients, and currently created chat rooms.
type MainRoom struct {
	clients  []*Client
	rooms    map[string]*Room
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

type Room struct {
	name     string
	clients  []*Client
	messages []string
}

// Creates a mainRoom which beings listening over its channels.
func NewMainRoom() *MainRoom {
	mainRoom := &MainRoom{
		clients:  make([]*Client, 0),
		rooms:    make(map[string]*Room),
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
	client.Listen()
	return client
}

// Creates a new message with the given time, client and text.
func NewMessage(client *Client, text string) *Message {
	return &Message{
		client: client,
		text:   text,
	}
}

// Create a new room with name, and stores messages and clients in the room
func NewRoom(name string) *Room {
	return &Room{
		name:     name,
		clients:  make([]*Client, 0),
		messages: make([]string, 0),
	}
}

// Starts a new thread which listens over the MainRoom's various channels.
func (mainRoom *MainRoom) Listen() {
	go func() {
		for {
			select {
			case client := <-mainRoom.join:
				mainRoom.Join(client)
			case message := <-mainRoom.incoming:
				mainRoom.Parse(message)
			case client := <-mainRoom.quit:
				client.conn.Close()
			}
		}
	}()
}

func (mainRoom *MainRoom) CheckPassword(client *Client) bool {
	password := <-client.incoming
	args := strings.Split(strings.Trim(password.text, "\r\n"), " ")
	pass := strings.TrimSpace(args[0])
	log.Print(pass)
	if pass != "password" {
		client.outgoing <- "Incorrect Password."
		log.Print("incorrect")
		client.conn.Close()
	}
	return true
}

func (mainRoom *MainRoom) Join(client *Client) {
	mainRoom.clients = append(mainRoom.clients, client)
	if mainRoom.CheckPassword(client) {
		go func() {
			for message := range client.incoming {
				mainRoom.incoming <- message
			}
			mainRoom.quit <- client
		}()
		client.outgoing <- "You're connected to the server\n"
	}
}

// Handles messages sent to the mainRoom
func (mainRoom *MainRoom) Parse(message *Message) {
	if strings.HasPrefix(message.text, "/") {
		switch {
		case strings.HasPrefix(message.text, CMD_MSG):
			mainRoom.SendMessage(message)
		case strings.HasPrefix(message.text, CMD_SHOUT):
			mainRoom.Shout(message)
		case strings.HasPrefix(message.text, CMD_NAME):
			name := strings.TrimSuffix(strings.TrimPrefix(message.text, CMD_NAME+" "), "\n")
			mainRoom.Name(message.client, name)
		case strings.HasPrefix(message.text, CMD_HELP):
			mainRoom.Help(message.client)
		case strings.HasPrefix(message.text, CMD_QUIT):
			message.client.conn.Close()
		default:
			message.client.outgoing <- "Unknown command. Type /help for a list of available commands."
		}
	} else {
		mainRoom.SendMessage(message)
	}

}

// Send the given message to the client's current room.
func (mainRoom *MainRoom) SendMessage(message *Message) {
	msg := message.text
	msgString := strings.Join([]string{message.text}, " ")
	if strings.HasPrefix(message.text, CMD_MSG) {
		msg = msgString[len(CMD_MSG)+1:]
	}
	fullMsg := ">> " + message.client.name + ": " + msg + "\n"
	for _, client := range mainRoom.clients {
		client.outgoing <- fullMsg
	}
	log.Println("client sent message")
}

// Send the upper case message to the client's current room.
func (mainRoom *MainRoom) Shout(message *Message) {
	var msg string
	msgString := strings.Join([]string{message.text}, " ")
	msg = strings.ToUpper(msgString[len(CMD_SHOUT)+1:])
	fullMsg := ">> " + message.client.name + ": " + msg + "\n"
	for _, client := range mainRoom.clients {
		client.outgoing <- fullMsg
	}
	log.Println("client sent shout message")
}

// Changes the client's name to the given name.
func (mainRoom *MainRoom) Name(client *Client, name string) {
	client.outgoing <- fmt.Sprintf(NAME_CHANGE, name)
	client.name = name
	log.Println("client changed their name")
}

// Sends to the client the list of possible commands to the client.
func (mainRoom *MainRoom) Help(client *Client) {
	client.outgoing <- "\n"
	client.outgoing <- "Commands:\n"
	client.outgoing <- "/help - lists all commands\n"
	client.outgoing <- "/name Tohfah - changes your name to Tohfah\n"
	client.outgoing <- "/msg Hello - sends \"Hello\" to all members in the chat\n"
	client.outgoing <- "/shout Hello - sends an upper case \"HELLO\" to all members in the chat\n"
	client.outgoing <- "/quit - removes the client from the chat\n"
	client.outgoing <- "\n"
	log.Println("client requested help")
}

func (client *Client) Listen() {
	go client.Read()
	go client.Write()
}

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
	log.Println("Closed client's incoming channel read thread")
}

// Reads in messages from the Client's outgoing channel, and writes them to the
// Client's socket.
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

// Creates a mainRoom, listens for client connections, and connects them to the
// mainRoom.
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
		newCli := NewClient(conn)
		mainRoom.Join(newCli)
		log.Println("New client has connected")
	}

}
