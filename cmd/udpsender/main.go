package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	remote, err := net.ResolveUDPAddr("udp4", "localhost:42069")
	if err != nil {
		log.Fatalf("error: resolve local addr : %s", err)
	}

	conn, err := net.DialUDP("udp4", nil, remote)
	if err != nil {
		log.Fatalf("error: dial udp:%s", err)
	}
	defer conn.Close()

	bufreader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print(">")
		line, err := bufreader.ReadString('\n')
		if err != nil {
			log.Printf("error: reading from stdin:%s", err)
		}
		_, err = conn.Write([]byte(line))
		if err != nil {
			log.Printf("error: writing conn:%s", err)
		}

	}
}
