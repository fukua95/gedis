package main

import (
	"fmt"
	"os"

	"github.com/codecrafters-io/redis-starter-go/server"
)

func main() {
	conf := server.NewConfig(os.Args)

	server := server.NewServer(conf)
	if err := server.ListenAndServe(); err != nil {
		fmt.Println("server error: ", err.Error())
		os.Exit(1)
	}
}
