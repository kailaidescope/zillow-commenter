package main

import (
	//"fmt"
	//"log"

	//server_app "server/router"

	"log"

	"github.com/aws/aws-lambda-go/lambda"
	"zillow-commenter.com/m/api"
)

var server *api.Server

func init() {
	// Server //
	var err error
	server, err = api.GetNewServer(api.Production)
	if err != nil {
		log.Fatal("Could not start the server")
	}
}

func main() {
	/* // Start server listening on port 3000 for HTTPS connections
	server.Router.RunTLS(":3000", "./ssl/public_certificate.pem", "./ssl/private_key.pem")
	if err != nil {
		log.Fatal("Could not start the server")
	} */

	// Proxy the server to AWS Lambda
	if server.LambdaAdapter == nil {
		log.Fatal("LambdaAdapter is not initialized")
	}
	lambda.Start(server.LambdaAdapter.ProxyWithContext)
}
