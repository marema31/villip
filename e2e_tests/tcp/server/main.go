package main

import (
	"log"
	"net"
	"strconv"
)

// HandleConn read all the data from the connection and resend it back.
func HandleConn(c net.Conn) {
	defer c.Close()

	char := make([]byte, 1)
	header := make([]byte, 0)

	for {
		_, err := c.Read(char)
		if err != nil {
			log.Fatal(err)
		}

		if string(char[0]) == "@" {
			break
		}

		header = append(header, char[0])
	}

	left, err := strconv.Atoi(string(header))
	if err != nil {
		log.Fatal(err)
	}

	data := make([]byte, 0, left)

	for {
		size, err := c.Read(data)
		if err != nil {
			log.Fatal(err)
		}

		_, err = c.Write([]byte(data))
		if err != nil {
			log.Fatal(err)
		}

		left -= size
		if left <= 0 {
			break
		}
	}
}

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:8888")
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	i := 0

	for {
		// accept connection
		log.Println("Waiting for new connection...")

		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}

		i++
		log.Println("received", i)

		// handle connection
		go HandleConn(conn)
	}
}
