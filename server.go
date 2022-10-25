package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

const (
	//port     = "8088"
	//connType = "tcp"

	cmdName    = "/name"
	cmdMessage = "/msg"
	cmdQuit    = "/quit"
	cmdHelp    = "/help"

	clientName = "anon"
	serverName = "server"
)

type Room struct {
	clients  []*Client
	incoming chan *Message
	join     chan *Client
	quit     chan *Client
	messages []string
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
