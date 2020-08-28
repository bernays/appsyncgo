package main

import (
	appsync "github.com/bernays/appsync-go-client/client"
)

func main() {
	client := appsync.CreateClient("https://whom3blq6vhxhd6rkt3offziva.appsync-api.us-east-2.amazonaws.com/graphql", "default")
	client.StartConnection()
}
