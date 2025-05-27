package main

import (
	//"fmt"
	//"log"

	//server_app "server/router"

	"log"

	"zillow-commenter.com/m/api"
)

func main() {

	// Server //
	server, err := api.GetNewServer()
	if err != nil {
		log.Fatal("Could not start the server")
	}

	server.Router.Run(":3000")
}
