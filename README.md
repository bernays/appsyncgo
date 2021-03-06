# appsync-go-client


![Go](https://github.com/bernays/appsyncgo/workflows/Go/badge.svg)  [![PkgGoDev](https://pkg.go.dev/badge/mod/github.com/bernays/appsyncgo)](https://pkg.go.dev/mod/github.com/bernays/appsyncgo)



This client library is designed to provide a stable interface for programs to interact with AppSync.

It uses native websockets so is able to take advantage of the realtime ability of subscriptions.

Connection recycling and reconnection is built in to handle network issues


## Example:

```
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

// Can run in main scope or in parallel go routine
func main() {
	// default profile will be used to grab credentials and sign reqs
	client, err := appsync.CreateClient("https://whom3blq6vhxhd6rkt3offziva.appsync-api.us-east-2.amazonaws.com/graphql", "default")
	if err != nil {
		logger.Error(err)
	}
	
	// Close connection and subscriptions on function end
	defer client.CloseConnection(false, false)
	client.StartConnection()

	data := "{\"query\":\"subscription { addedPost{ id title } }\",\"variables\":{}}"
	client.Subscribe(data, HandleData)
	for {
		time.Sleep(2 * time.Second)
	}
}
```
