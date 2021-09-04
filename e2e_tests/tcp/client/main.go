package main

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"
)

func SendAndReceive(proto, addr string, out chan<- bool, sum [20]byte, data []byte) {
	c, err := net.Dial(proto, addr)
	if err != nil {
		log.Fatal(err)
	}

	defer c.Close()

	buffer := bytes.NewBuffer(data)
	size := fmt.Sprintf("%d@", len(data))

	_, err = c.Write([]byte(size))
	if err != nil {
		log.Fatal(err)
	}

	_, err = io.Copy(c, buffer)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("sended")
	recdata := make([]byte, 0)
	recbuffer := bytes.NewBuffer(recdata)

	_, err = io.Copy(recbuffer, c)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("received")

	recsum := sha1.Sum(data)

	out <- sum == recsum
}

func main() {
	if len(os.Args) != 3 { //nolint: gomnd
		log.Println("Usage: server filename logfile")

		os.Exit(1)
	}

	resultFile, err := os.Create(os.Args[2])
	if err != nil {
		log.Fatal(err)
	}

	defer resultFile.Close()

	data := make([]byte, 0)

	filedata, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	defer filedata.Close()

	reader := bufio.NewReader(filedata)
	buffer := bytes.NewBuffer(data)

	io.Copy(buffer, reader)

	sum := sha1.Sum(data)

	start := time.Now()

	ch := make(chan bool, 1)

	for i := 0; i < 100; i++ {
		log.Println("sending ", i)

		go SendAndReceive("tcp", "villip:8888", ch, sum, data)
	}

	var m bool

	result := true

	for i := 0; i < 100; i++ {
		m = <-ch
		result = result && m
		log.Println(i+1, m)
	}

	log.Println(time.Since(start))

	fmt.Fprintln(resultFile, time.Since(start))

	if !result {
		log.Println("Test faulted")
		os.Exit(2) //nolint: gomnd
	}

	log.Println("Test succeeded")
	fmt.Fprintln(resultFile, "Test succeeded")

	os.Exit(0)
}
