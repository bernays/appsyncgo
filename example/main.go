package main

import (
	appsync "github.com/bernays/appsyncgo/client"
	"github.com/sirupsen/logrus"
	"time"
)

func HandleData(data string) error {
	logger.Printf("Client Side data: %s", data)
	return nil
}

var logger = logrus.New()

func init() {
	logger.SetLevel(logrus.DebugLevel)
}
func main() {
	logger.Error("started")
	client, err := appsync.CreateClient("https://whom3blq6vhxhd6rkt3offziva.appsync-api.us-east-2.amazonaws.com/graphql", "default")
	if err != nil {
		logger.Error(err)
	}
	defer func() {
		err = client.CloseConnection(false, false)
		logger.Error(err)

	}()
	err = client.StartConnection()
	if err != nil {
		logger.Error(err)
	}
	reqc := appsync.AppSyncRequest{Query: `query { singlePost(id: "22") {id title } }`}
	client.Query(reqc)
	data := "{\"query\":\"subscription { addedPost{ id title } }\",\"variables\":{}}"
	_, err = client.Subscribe(data, HandleData)
	if err != nil {
		logger.Error(err)
	}
	for {
		time.Sleep(2 * time.Second)
	}
}
