package client

import (
	"testing"
)

func TestCreateClientUrl(t *testing.T) {
	// path needs to be /graphql
	_, err := CreateClient("https://appid.appsync-api.us-east-2.amazonaws.com", "")
	if err == nil {
		t.Errorf("Bad URL, need a valid url")
		t.FailNow()
	}
	// must use https
	_, err = CreateClient("http://appid.appsync-api.us-east-2.amazonaws.com/graphql", "")
	if err == nil {
		t.Errorf("Bad URL, need to ")
		t.FailNow()
	}
}
