package client

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	guuid "github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type AppSyncClient struct {
	Connection    *(websocket.Conn)
	URL           string
	Auth          APIAuth
	Subscriptions []Subscription
	Data          chan []byte
	Done          chan struct{}
	Interrupt     chan os.Signal
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
type DataHandler func(string) error
type Subscription struct {
	ID      string
	Query   string
	Handler DataHandler

	//OnData  function
	//OnError function
}

var logger = logrus.New()

func init() {
	logger.SetLevel(logrus.WarnLevel)
}
func CreateClient(url, profile string) *AppSyncClient {

	//TODO: Validate inputs and write tests
	// 1. Ensure clean URL
	// 2. Ensure that correct auth fields are present

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
	return "", errors.New("Unknown AuthType")
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
	// client.Interrupt = make(chan os.Signal, 1)
	// signal.Notify(client.Interrupt, os.Interrupt)
	// go client.internalCloseConnection(client.Interrupt)
	client.Connection, _, err = websocket.DefaultDialer.Dial(u.String(), h)

	if err != nil {
		logger.Printf("%+v", err)
	}

	// // send message
	err = client.Connection.WriteMessage(websocket.TextMessage, []byte("{\"type\": \"connection_init\"}"))
	if err != nil {
		logger.Printf("%+v", err)
	}

	// receive message
	_, message, err := client.Connection.ReadMessage()
	if err != nil {
		// handle error
		logger.Printf("%+v", err)
	}
	logger.Debug(string(message))
	client.Data = make(chan []byte)
	go client.readData()
	go client.processData()
	return err
}

func (client *AppSyncClient) processData() {
	for {
		message := <-client.Data
		logger.Printf("process: %s", message)
		var wsMessage AppSyncMessage
		err := json.Unmarshal([]byte(message), &wsMessage)
		if err != nil {
			fmt.Println("error:", err)
			continue
		}
		logger.Debug(wsMessage)
		for _, s := range client.Subscriptions {
			messageString := string(wsMessage.Payload.Data)
			if s.ID == wsMessage.ID && len(messageString) > 0 {
				s.Handler(messageString)
			}
		}

	}
}
func (client *AppSyncClient) readData() {
	for {
		logger.Printf("Waiting for message")
		_, message, err := client.Connection.ReadMessage()
		if err != nil {
			logger.Println("read:", err)
			return
		}
		logger.Printf("recv: %s", message)
		client.Data <- message
	}

}

func (client *AppSyncClient) Query(method, variables, query string) (string, error) {
	return "", nil
}

func (client *AppSyncClient) CloseConnection() error {

	if client.Connection != nil {
		logger.Printf("Closing Connection")
		client.Connection.Close()
	} else {
		logger.Printf("No Connection to close")
	}
	return nil
}
func (client *AppSyncClient) Subscribe(Query string, Handler DataHandler) (string, error) {
	iamHeaders, _, err := iamAuth(client.URL, client.Auth.Profile, Query)
	uuid := guuid.New()
	subRequest := &SubscriptionRequest{
		ID: uuid.String(),
		Payload: SubscriptionRequestPayload{
			Data: Query,
			Extensions: SubscriptionRequestExtensions{
				Authorization: *iamHeaders,
			},
		},
		Type:    "start",
		Handler: Handler,
	}

	client.Subscriptions = append(client.Subscriptions, Subscription{
		Handler: Handler,
		ID:      uuid.String(),
	})

	encodedBytes, err := json.Marshal(subRequest)
	if err != nil {
		logger.Println("marshalling:", err)
	}
	logger.Debug(string(encodedBytes))
	err = client.Connection.WriteMessage(websocket.TextMessage, encodedBytes)
	return "", err
}
