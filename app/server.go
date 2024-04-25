package main

import (
	"fmt"
	"os"

	"github.com/codecrafters-io/redis-starter-go/server"
)

func main() {
	server := server.NewServer()
	if err := server.ListenAndServe("tcp", "0.0.0.0:6379"); err != nil {
		fmt.Println("server error: ", err.Error())
		os.Exit(1)
	}
}
