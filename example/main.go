package main

import (
	appsync "github.com/bernays/appsyncgo/client"
	"github.com/sirupsen/logrus"
	//"time"
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
	defer client.CloseConnection(false, false)
	client.StartConnection()
	reqc := appsync.AppSyncRequest{Query: `query { singlePost(id: "22") {id title } }`}
	client.Query(reqc)
	// data := "{\"query\":\"subscription { addedPost{ id title } }\",\"variables\":{}}"
	// client.Subscribe(data, HandleData)
	// for {
	// 	time.Sleep(2 * time.Second)
	// }
}
