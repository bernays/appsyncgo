package client

import (
	// "github.com/google/uuid"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
)

type AppSyncClient struct {
	Connection    *(websocket.Conn)
	URL           string
	Auth          APIAuth
	Subscriptions []Subscription
}

// APIAuth parameters used to authenticate with AppSync
type APIAuth struct {
	AuthType         string // AWS_IAM, API_KEY
	APIKey           string // Only required if Type is API_KEY
	Profile          string // Only required if Type is API_KEY
	AwsAccessKey     string // Only required if Type is AWS_IAM and not using profile
	AwsAccessSecret  string // Only required if Type is AWS_IAM and not using profile
	AwsSecurityToken string // Only required if Type is AWS_IAM and not using profile

}

type Subscription struct {
	ID string
	//OnData  function
	//OnError function
}

func CreateClient(url, profile string) *AppSyncClient {
	return &AppSyncClient{
		URL: url,
		Auth: APIAuth{
			Profile:  profile,
			AuthType: "AWS_IAM",
		},
	}
}

func (client *AppSyncClient) GenerateAuthFields() (string, error) {
	apiURL := client.URL
	if client.Auth.AuthType == "API_KEY" {
		host := strings.ReplaceAll(strings.ReplaceAll(apiURL, "https://", ""), "/graphql", "")
		encodedBytes := []byte("{\"host\":\"" + host + "\",\"x-api-key\": \"" + client.Auth.APIKey + "\"}")
		return base64.StdEncoding.EncodeToString(encodedBytes), nil
	} else if client.Auth.AuthType == "AWS_IAM" {
		canonicalURI := strings.ReplaceAll(apiURL, "https://", "wss://") + "/connect"
		iamHeaders, _, err := iamAuth(canonicalURI, client.Auth.Profile, "{}")
		if err != nil {
			return "", err
		}
		encodedBytes, err := json.Marshal(iamHeaders)
		if err != nil {
			return "", err
		}
		return base64.StdEncoding.EncodeToString([]byte(encodedBytes)), err
	}
	return "", nil
}
func (client *AppSyncClient) StartConnection() error {

	encoded, err := client.GenerateAuthFields()
	h := http.Header{}
	h.Add("Sec-WebSocket-Protocol", "graphql-ws")
	u := url.URL{
		Scheme:   "wss",
		Host:     strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(client.URL, "https://", ""), "/graphql", ""), "appsync-api", "appsync-realtime-api"),
		Path:     "/graphql",
		RawQuery: fmt.Sprintf("%s%s%s", "header=", encoded, "&payload=e30="),
	}
	if err != nil {

	}
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	c, _, err := websocket.DefaultDialer.Dial(u.String(), h)

	if err != nil {
		log.Printf("%+v", err)
	}

	// // send message
	err = c.WriteMessage(websocket.TextMessage, []byte("{\"type\": \"connection_init\"}"))
	if err != nil {
		log.Printf("%+v", err)
	}

	// receive message
	_, message, err := c.ReadMessage()
	if err != nil {
		// handle error
		log.Printf("%+v", err)
	}
	log.Print(string(message))
	defer c.Close()
	return err
}

func (client *AppSyncClient) CloseConnection() error {
	return nil
}
func (client *AppSyncClient) Query(method, variables, query string) (string, error) {
	return "", nil
}
