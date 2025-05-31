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

	// Start server listening on port 3000 for HTTPS connections
	server.Router.RunTLS(":3000", "./ssl/public_certificate.pem", "./ssl/private_key.pem")
	if err != nil {
		log.Fatal("Could not start the server")
	}
}
