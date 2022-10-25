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



