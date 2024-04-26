package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/server"
)

func main() {
	port := flag.Int("port", 6379, "redis port number")
	flag.Parse()

	server := server.NewServer("tcp", strconv.Itoa(*port))
	if err := server.ListenAndServe(); err != nil {
		fmt.Println("server error: ", err.Error())
		os.Exit(1)
	}
}
