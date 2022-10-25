package main

import (
	"fmt"
	"log"
	"net"
	"os"
)

const (
	port     = "8080"
	connType = "tcp"
)

func main() {
	listener, err := net.Listen(connType, "localhost:"+port)
	if err != nil {
		log.Fatal(err)
	}

	go broadcast()

	for {
		conn, err := listener.Accept()
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
		}
		fmt.Println(conn)
	}
}

type client chan string

var (
	enter = make(chan client)
	leave  = make(chan client)
	msg = make(chan string)
)

//broadcast message to all clients
func broadcast() {
	clients := make(map[client]bool)

	for {
		select {
			case msg := <-messages:
				for cli := range clients {
					cli <- msg
				}
			case newCli := <-entering:
				clients[newCli] = true
			case leaveCli := <-leaving:
				delete(clients, leaveCli)
				close(leaveCli)
		}
	}
}


