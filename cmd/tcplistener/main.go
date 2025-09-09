package main

import (
	"fmt"
	"go-http/internal/request"
	"log"
	"net"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func handleConnection(conn net.Conn) {
	req, err := request.RequestFromReader(conn)

	check(err)

	fmt.Printf("Request line:\n- Method: %s\n- Target: %s\n- Version: %s\n",
		req.RequestLine.Method,
		req.RequestLine.RequestTarget,
		req.RequestLine.HttpVersion)

	fmt.Println("Headers:")

	for key, value := range req.Headers.GetHeaders() {
		fmt.Printf("- %s: %s \n", key, value)
	}

	// fmt.Printf("Body: %s", string(req.Body))

}

func main() {
	ln, err := net.Listen("tcp", ":42069")
	defer ln.Close()

	check(err)

	for {
		conn, err := ln.Accept()

		if err != nil {
			log.Fatal("TCP connection error:", err)
		}

		go handleConnection(conn)
	}
}
