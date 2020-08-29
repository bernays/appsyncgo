package main

import (
	appsync "github.com/bernays/appsync-go-client/client"
	"github.com/sirupsen/logrus"
)

func WeGotData(data string) error {
	log.Printf("Client Side data: %s", data)
	return nil
}

var log = logrus.New()

func main() {
	client := appsync.CreateClient("https://whom3blq6vhxhd6rkt3offziva.appsync-api.us-east-2.amazonaws.com/graphql", "default")
	defer client.CloseConnection()
	client.StartConnection()
	data := "{\"query\":\"subscription { addedPost{ id title } }\",\"variables\":{}}"
	client.Subscribe(data, WeGotData)
	for {

	}
}
