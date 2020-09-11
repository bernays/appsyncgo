package client

import "encoding/json"

// IamHeaders object contains headers required to sign websocket requests
type IamHeaders struct {
	Accept            string `json:"accept"`
	ContentEncoding   string `json:"content-encoding"`
	ContentType       string `json:"content-type"`
	Host              string `json:"host"`
	XAmzDate          string `json:"x-amz-date"`
	XAmzSecurityToken string `json:"X-Amz-Security-Token"`
	Authorization     string `json:"Authorization"`
}

// SubscriptionRequest a request submitted to the client object to subscribe and handle data via a subscription
type SubscriptionRequest struct {
	ID      string                     `json:"id"`
	Payload subscriptionRequestPayload `json:"payload"`
	Type    string                     `json:"type"`
	Handler CallbackFn                 `json:"-"`
}

type subscriptionRequestPayload struct {
	Data       string                        `json:"data"`
	Extensions subscriptionRequestExtensions `json:"extensions"`
}

type subscriptionRequestExtensions struct {
	Authorization IamHeaders `json:"authorization"`
}

type appSyncMessage struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Payload struct {
		Data json.RawMessage
	} `json:"payload"`
}

// AppSyncResponse response from appsync, Data can be null if their is an error or no data was returned
// Both fields are strings as the data is dynamic
type AppSyncResponse struct {
	Data   string `json:"data"`
	Errors string `json:"errors"`
}

type appSyncResponseInternal struct {
	Data   json.RawMessage `json:"data"`
	Errors json.RawMessage `json:"errors"`
}

// AppSyncRequest Struct that contains all of the information for a Synchronous request
type AppSyncRequest struct {
	Query     string `json:"query"`
	Variables string `json:"variables"`
}
