package cmd

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws/external"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/gorilla/websocket"
)

type iamHeaders struct {
	Accept            string `json:"accept"`
	ContentEncoding   string `json:"content-encoding"`
	ContentType       string `json:"content-type"`
	Host              string `json:"host"`
	XAmzDate          string `json:"x-amz-date"`
	XAmzSecurityToken string `json:"X-Amz-Security-Token"`
	Authorization     string `json:"Authorization"`
}

type nilBody struct{}

func (nilBody) Read(p []byte) (int, error) {
	return 0, io.EOF
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(i InputArguments, v VersionInfo) {
	version := v

	if err := main(i.URL, i.APIAuth); err != nil {
		fmt.Println(err)
		log.Print(version)
		os.Exit(1)
	}
}

func makeSha256Reader(reader io.ReadSeeker) (hashBytes []byte, err error) {
	hash := sha256.New()
	start, err := reader.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, err
	}
	defer func() {
		// ensure error is return if unable to seek back to start if payload
		_, err = reader.Seek(start, io.SeekStart)
	}()

	io.Copy(hash, reader)
	return hash.Sum(nil), nil
}
func iamAuth(url, profile, payload string) (string, string, error) {
	canonicalURI := strings.ReplaceAll(url, "https://", "wss://") + "/connect"

	cfg, err := external.LoadDefaultAWSConfig(
		// external.WithSharedConfigProfile(profile),
		// aws.WithLogLevel(aws.LogDebug),
		external.WithDefaultRegion("us-east-2"),
	)
	if err != nil {
		log.Printf("%+v", err)
		panic("unable to load SDK config, " + err.Error())

	}
	signer := v4.NewSigner(cfg.Credentials, func(s *v4.Signer) {
		// s.Logger = cfg.Logger
		// s.Debug = aws.LogDebugWithSigning
	})

	hashBytes, err := makeSha256Reader(strings.NewReader(payload))
	if err != nil {
		log.Printf("Error: %v", err)
	}
	sha1Hash := hex.EncodeToString(hashBytes)

	req, err := http.NewRequest("POST", canonicalURI, nil)
	if err != nil {
		log.Printf("Error constructing request object")
		log.Printf("Error: %v", err)
		return "", "", err
	}

	err = signer.SignHTTP(context.Background(), req, sha1Hash, "appsync", "us-east-2", time.Now())

	if err != nil {
		log.Printf("%+v", err)
		panic("unable to load SDK config, " + err.Error())

	}

	iamHeaders := &iamHeaders{
		Accept:            "application/json, text/javascript",
		ContentEncoding:   "amz-1.0",
		ContentType:       "application/json; charset=UTF-8",
		Host:              strings.ReplaceAll(strings.ReplaceAll(url, "https://", ""), "/graphql", ""),
		XAmzDate:          req.Header.Get("X-Amz-Date"),
		XAmzSecurityToken: req.Header.Get("X-Amz-Security-Token"),
		Authorization:     req.Header.Get("Authorization"),
	}
	encodedBytes, err := json.Marshal(iamHeaders)
	return string(encodedBytes), req.Header.Get("X-Amz-Security-Token"), nil
}
func main(apiURL string, apiAuth APIAuth) error {
	var encoded string
	h := http.Header{}
	h.Add("Sec-WebSocket-Protocol", "graphql-ws")

	if apiAuth.AuthType == "API_KEY" {
		encoded = base64.StdEncoding.EncodeToString([]byte("{\"host\":\"" + strings.ReplaceAll(strings.ReplaceAll(apiURL, "https://", ""), "/graphql", "") + "\",\"x-api-key\": \"" + apiAuth.APIKey + "\"}"))
	} else if apiAuth.AuthType == "AWS_IAM" {
		encodedBytes, _, err := iamAuth(apiURL, apiAuth.Profile, "{}")
		if err != nil {
			return err
		}
		encoded = base64.StdEncoding.EncodeToString([]byte(encodedBytes))
	}

	u := url.URL{
		Scheme:   "wss",
		Host:     strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(apiURL, "https://", ""), "/graphql", ""), "appsync-api", "appsync-realtime-api"),
		Path:     "/graphql",
		RawQuery: fmt.Sprintf("%s%s%s", "header=", encoded, "&payload=e30="),
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

	// // send message
	err = c.WriteMessage(websocket.TextMessage, []byte("{\"type\": \"connection_init\"}"))
	if err != nil {
		log.Printf("%+v", err)
	}

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s", message)
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return nil
		case t := <-ticker.C:
			log.Println("write:", t.String())
			err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
			if err != nil {
				log.Println("write:", err)
				return nil
			}
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return nil
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return nil
		}
	}
}
