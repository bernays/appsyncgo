package client

import (
	"encoding/base64"
	"encoding/json"
	"os"
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
		t.Errorf(" use /graphql as path")
		t.FailNow()
	}
}

func TestCreateClientUrlHTTPS(t *testing.T) {
	// must use https
	_, err := CreateClient("http://appid.appsync-api.us-east-2.amazonaws.com/graphql", "")
	if err == nil {
		logger.Error(err)
		t.Errorf("Must use HTTPS ")
		t.FailNow()
	}
	_, err = CreateClient("https://appid.appsync-api.us-east-2.amazonaws.com/graphql", "")
	if err != nil {
		logger.Error(err)
		t.Errorf("Must use HTTPS ")
		t.FailNow()
	}
}

func TestCreateClientAuthApiKey(t *testing.T) {
	// if API_KEY then only APIKey can be set
	// if AWS_IAM and profile none of these Token, Key, Secret
	// if AWS_IAM and profile is nil then  Key, Secret are required and Token is optional
}

func TestSubscriptionAuthFieldsAuthType(t *testing.T) {
	// Any value other than AWS_IAM and API_KEY returns an error
	// Test Compute Headers for API_KEY
	// Test Compute Headers for Profile
	// Test Compute Headers for Secret+Key
	// Test Compute Headers for Secret+Key+Token
	AwsProfile := "Testing"
	tempDir := t.TempDir()
	f, err := os.Create(tempDir + "/dat2")
	if err != nil {
		t.FailNow()
	}
	_, err = f.WriteString("[" + AwsProfile + "]\n")

	if err != nil {
		t.FailNow()
	}
	_, err = f.WriteString("aws_access_key_id=KEY\n")
	if err != nil {
		t.FailNow()
	}
	_, err = f.WriteString("aws_secret_access_key=SECRET\n")

	if err != nil {
		t.FailNow()
	}

	f.Close()
	os.Setenv("AWS_SHARED_CREDENTIALS_FILE", tempDir+"/dat2")
	os.Setenv("GO_ENV", "testing")
	ASC := &AppSyncClient{
		URL: "https://appid.appsync-api.us-east-2.amazonaws.com/graphql",
		Auth: APIAuth{
			Profile:  AwsProfile,
			AuthType: "AWS_IAM",
		},
	}
	encoded, err := ASC.subscriptionAuthFields()
	if err != nil {
		logger.Error("decode error:", err)
		return
	}
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	var headers IamHeaders
	if err != nil {
		print(err)
	}
	err = json.Unmarshal([]byte(decoded), &headers)
	if err != nil {
		t.FailNow()
	}
	if headers.Accept != "application/json, text/javascript" {
		t.Errorf("Invalid Headers, Accept")
		t.FailNow()
	}
	if headers.ContentEncoding != "amz-1.0" {
		t.Errorf("Invalid Headers, Accept")
		t.FailNow()
	}
	if headers.ContentType != "application/json; charset=UTF-8" {
		t.Errorf("Invalid Headers, ContentType")
		t.FailNow()
	}
	if headers.Host != "appid.appsync-api.us-east-2.amazonaws.com" {
		t.Errorf("Invalid Headers, Host")
		t.FailNow()
	}

	if headers.XAmzDate != "19700101T000000Z" {
		t.Errorf("Invalid Headers, XAmzDate")
		t.FailNow()
	}

	if headers.Authorization != "AWS4-HMAC-SHA256 Credential=KEY/19700101/us-east-2/appsync/aws4_request, SignedHeaders=host;x-amz-date, Signature=f26d29557ce9c21274b95422e1ee08e606a50f88cfd821ed763c1ebdebba5f54" {
		t.Errorf("Invalid Headers, Authorization")
		t.FailNow()
	}
}
