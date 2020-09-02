package client

import "encoding/json"

// IamHeaders- all headers required to sign websocket requests
type IamHeaders struct {
	Accept            string `json:"accept"`
	ContentEncoding   string `json:"content-encoding"`
	ContentType       string `json:"content-type"`
	Host              string `json:"host"`
	XAmzDate          string `json:"x-amz-date"`
	XAmzSecurityToken string `json:"X-Amz-Security-Token"`
	Authorization     string `json:"Authorization"`
}

// SubscriptionRequest- APpSync template
type SubscriptionRequest struct {
	ID      string                     `json:"id"`
	Payload SubscriptionRequestPayload `json:"payload"`
	Type    string                     `json:"type"`
	Handler DataHandler                `json:"-"`
}

type SubscriptionRequestPayload struct {
	Data       string                        `json:"data"`
	Extensions SubscriptionRequestExtensions `json:"extensions"`
}

type SubscriptionRequestExtensions struct {
	Authorization IamHeaders `json:"authorization"`
}

type AppSyncMessage struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Payload struct {
		Data json.RawMessage
	} `json:"payload"`
}
