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
	"time"
)

// AppSyncClient main object containing all of the information about connections to the API
type AppSyncClient struct {
	Connection    *(websocket.Conn)
	URL           string
	Auth          APIAuth
	Subscriptions []subscription
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

// CallbackFn is the function that the client invokes upon reciept
type CallbackFn func(string) error
type subscription struct {
	ID      string
	Query   string
	Handler CallbackFn

	//OnData  function
	//OnError function
}

var logger = logrus.New()

func init() {
	logger.SetLevel(logrus.DebugLevel)
}

// CreateClient initialization function that enables users to create a Client object
func CreateClient(urlAppSync, profile string) (*AppSyncClient, error) {

	//TODO: Validate inputs and write tests
	// 1. Ensure clean URL
	// 2. Ensure that correct auth fields are present
	val, err := url.ParseRequestURI(urlAppSync)
	if err != nil {
		logger.Error("unknown url")
		logger.Error(err)
		return &AppSyncClient{}, errors.New("INVALID_URI")
	}

	if val.Path != "/graphql" {
		logger.Errorf("unknown path: %s", val.Path)
		return &AppSyncClient{}, errors.New("INVALID_URI_PATH")
	}
	if val.Scheme != "https" {
		logger.Errorf("must use https: %s", val.Scheme)
		return &AppSyncClient{}, errors.New("INVALID_URI_SCHEME")
	}
	return &AppSyncClient{
		URL: urlAppSync,
		Auth: APIAuth{
			Profile:  profile,
			AuthType: "AWS_IAM",
		},
	}, nil
}

func (client *AppSyncClient) generateAuthFields() (string, error) {
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

// StartConnection Starts long running websocket connection to AppSync,
func (client *AppSyncClient) StartConnection() error {

	encoded, err := client.generateAuthFields()
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
		logger.Errorf("%+v", err)
		// TODO: Handle retryable errors
		// Ensure backoff and jitter
		return err
	}

	// // send message
	err = client.Connection.WriteMessage(websocket.TextMessage, []byte("{\"type\": \"connection_init\"}"))
	if err != nil {
		logger.Errorf("%+v", err)
	}

	// receive message
	_, message, err := client.Connection.ReadMessage()
	if err != nil {
		// handle error
		logger.Errorf("%+v", err)
	}

	var newSubscriptions []subscription
	for _, s := range client.Subscriptions {
		s.ID = guuid.New().String()
		err := client.internalSubscribe(s)
		if err != nil {
			logger.Error("Unable to resubscribe to subscription")
		} else {
			newSubscriptions = append(newSubscriptions, s)
		}

	}
	client.Subscriptions = newSubscriptions
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
		var wsMessage appSyncMessage
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
		logger.Debug("Waiting for message")
		client.Connection.SetReadDeadline(time.Now().Add(10 * time.Second))
		_, message, err := client.Connection.ReadMessage()
		if os.IsTimeout(err) {
			logger.Error("Timeout!:")
			logger.Error(err)
			client.CloseConnection(true, true)
			break
		} else if err != nil {
			logger.Error("read:", err)
			// TODO: understand if error is because connection was closed
			// Close connection, and retry
			client.CloseConnection(true, false)
			break
		}
		logger.Printf("recv: %s", message)
		client.Data <- message
	}
}

// Query allows user to synchronously interact with API
func (client *AppSyncClient) Query(method, variables, query string) (str string, err error) {
	return str, err
}

// CloseConnection closes connections and unsubscribes from subscriptions (if necessary)
func (client *AppSyncClient) CloseConnection(restart, timeout bool) error {
	// TODO: Close subscriptions with AppSync
	if client.Connection != nil {
		logger.Printf("Closing Connection")
		client.Connection.Close()
		if !timeout {
			for _, s := range client.Subscriptions {
				closingString := "{ \"type\":\"stop\",\"id\":\"" + s.ID + "\"}"
				err := client.Connection.WriteMessage(websocket.TextMessage, []byte(closingString))
				if err != nil {
					logger.Errorf("%+v", err)
				}
			}
		}
		if restart {
			client.StartConnection()
		} else {

		}
	} else {
		logger.Printf("No Connection to close")
	}
	return nil
}

func (client *AppSyncClient) internalSubscribe(subscription subscription) error {
	iamHeaders, _, err := iamAuth(client.URL, client.Auth.Profile, subscription.Query)
	subRequest := &SubscriptionRequest{
		ID: subscription.ID,
		Payload: subscriptionRequestPayload{
			Data: subscription.Query,
			Extensions: subscriptionRequestExtensions{
				Authorization: *iamHeaders,
			},
		},
		Type:    "start",
		Handler: subscription.Handler,
	}
	encodedBytes, err := json.Marshal(subRequest)
	if err != nil {
		logger.Println("marshalling:", err)
	}
	logger.Debug(string(encodedBytes))
	err = client.Connection.WriteMessage(websocket.TextMessage, encodedBytes)
	return err
}

//Subscribe subscribes to a specific websocket. This uses the generic websocket subscriptions
func (client *AppSyncClient) Subscribe(Query string, Handler CallbackFn) (string, error) {
	uuid := guuid.New()

	subscription := subscription{
		Handler: Handler,
		ID:      uuid.String(),
		Query:   Query,
	}
	err := client.internalSubscribe(subscription)
	if err != nil {
		return "", err
	}
	client.Subscriptions = append(client.Subscriptions, subscription)
	return "", err
}
