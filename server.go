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
