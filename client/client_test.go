package client

import (
	"testing"
)

func TestCreateClientUrlMalformed(t *testing.T) {
	// path needs to be /graphql
	_, err := CreateClient("https//appid.appsync-api.us-east-2.amazonaws.com/graphql", "")
	if err == nil {
		t.Errorf("Bad URL, need a valid url")
		t.FailNow()
	}
	_, err = CreateClient("https://appid.appsync-api.us-east-2.amazonaws.com/graphql", "")
	if err != nil {
		t.Errorf("Bad URL, need a valid url")
		t.FailNow()
	}
}

func TestCreateClientUrlPath(t *testing.T) {
	// path needs to be /graphql
	_, err := CreateClient("https://appid.appsync-api.us-east-2.amazonaws.com/graphfql", "")
	if err == nil {
		t.Errorf("Must use /graphql as path")
		t.FailNow()
	}

	_, err = CreateClient("https://appid.appsync-api.us-east-2.amazonaws.com/graphql", "")
	if err != nil {
		t.Errorf("Must use /graphql as path")
		t.FailNow()
	}
}

func TestCreateClientUrlHTTPS(t *testing.T) {
	// must use https
	_, err := CreateClient("http://appid.appsync-api.us-east-2.amazonaws.com/graphql", "")
	if err == nil {
		print(err)
		t.Errorf("Must use HTTPS ")
		t.FailNow()
	}
	_, err = CreateClient("https://appid.appsync-api.us-east-2.amazonaws.com/graphql", "")
	if err != nil {
		print(err)
		t.Errorf("Must use HTTPS ")
		t.FailNow()
	}
}

func TestCreateClientAuth(t *testing.T) {
	// if API_KEY then only APIKey can be set
	// if AWS_IAM and profile none of these Token, Key, Secret
	// if AWS_IAM and profile is nil then  Key, Secret are required and Token is optional
}

func TestGenerateAuthFieldsAuthType(t *testing.T) {
	// Any value other than AWS_IAM and API_KEY returns an error
	// Test Compute Headers for API_KEY
	// Test Compute Headers for Profile
	// Test Compute Headers for Secret+Key
	// Test Compute Headers for Secret+Key+Token

}
