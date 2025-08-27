package main

import (
	"fmt"
	"httpfromtcp/internal/request"
	"log"
	"net"
)

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatalf("error: listen : %s", err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("error: accept : %s", err)
		}

		req, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatalf("Error: RequestFromReader: %s", err)
		}
		fmt.Println("Request line:")
		fmt.Printf("- Method: %s\n", req.RequestLine.Method)
		fmt.Printf("- Target: %s\n", req.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", req.RequestLine.HttpVersion)

		fmt.Println("Headers:")
		req.Headers.ForEach(func(n, v string) {
			fmt.Printf("- %s: %s\n", n, v)
		})

		fmt.Println("Body:")
		fmt.Println(req.Body)

	}
}
