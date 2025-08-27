package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	raddr, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		log.Fatal("UPD resolver exception: ", err)
	}

	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		return
	}

	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print(">")

		data, err := reader.ReadString('\n')

		if err != nil {
			fmt.Printf("Input reader exception: %s\n", err)
		}

		_, err = conn.Write([]byte(data))

		if err != nil {
			fmt.Printf("UPD write error: %s\n", err)
		}
	}
}
