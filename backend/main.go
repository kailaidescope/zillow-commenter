package main

import (
	//"fmt"
	//"log"

	//server_app "server/router"

	"flag"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
	"zillow-commenter.com/m/api"
)

var server *api.Server

func init() {
	/* // Server //
	var err error
	server, err = api.GetNewServer(api.Production)
	if err != nil {
		log.Fatal("Could not start the server", err)
	} */
}

func main() {
	/* // Start server listening on port 3000 for HTTPS connections
	server.Router.RunTLS(":3000", "./ssl/public_certificate.pem", "./ssl/private_key.pem")
	if err != nil {
		log.Fatal("Could not start the server")
	} */

	// Define server command line flags
	runLocalServerFlag := flag.Bool("l", false, "Runs the server locally rather than through a Lambda proxy.")
	flag.Parse()

	if *runLocalServerFlag {
		runLocalServer()
	} else {
		runLambdaServer()
	}
}

func runLambdaServer() {
	// Server //
	var err error
	server, err = api.GetNewServer(api.Production)
	if err != nil {
		log.Fatal("Could not start the server", err)
	}

	// Proxy the server to AWS Lambda
	if server.LambdaAdapter == nil {
		log.Fatal("LambdaAdapter is not initialized")
	}
	lambda.Start(server.LambdaAdapter.ProxyWithContext)
}

func runLocalServer() {
	log.Println("STARTING LOCAL SERVER INSTANCE...")
	// Server //
	server, err := api.GetNewServer(api.Test)
	if err != nil {
		log.Fatal("Could not start the local server", err)
	}

	server.Router.Run(":3000")
}
