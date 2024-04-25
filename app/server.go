package main

import (
	"fmt"
	"io"
	"net"
	"os"
)

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	defer l.Close()

	fmt.Println("Server start to accept requests")
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 1024)
	for {
		_, err := conn.Read(buf)
		if err == io.EOF {
			return
		}
		if err != nil {
			fmt.Println("Error reading from conn: ", err.Error())
			return
		}
		res := []byte("+PONG\r\n")
		n, err := conn.Write(res)
		if err != nil || n != len(res) {
			fmt.Println("Error writing to conn: ", err.Error())
		}
	}
}
