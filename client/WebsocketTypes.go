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
	Handler DataHandler                `json:"-"`
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
