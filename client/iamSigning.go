package client

import (
	// "github.com/google/uuid"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func iamAuth(canonicalURI, profile, payload string) (*IamHeaders, string, error) {

	cfg, err := external.LoadDefaultAWSConfig(
		external.WithSharedConfigProfile(profile),
		//aws.WithLogLevel(aws.LogDebugWithSigning),
		//external.WithDefaultRegion("us-east-2"),
	)
	if err != nil {
		log.Printf("%+v", err)
		panic("unable to load SDK config, " + err.Error())

	}
	signer := v4.NewSigner(cfg.Credentials, func(s *v4.Signer) {
		// s.Logger = cfg.Logger
		s.Debug = aws.LogDebugWithSigning
	})

	hashBytes, err := makeSha256Reader(strings.NewReader(payload))
	if err != nil {
		logger.Errorf("Error: %+v", err)
	}
	sha1Hash := hex.EncodeToString(hashBytes)

	req, err := http.NewRequest("POST", canonicalURI, nil)
	if err != nil {
		log.Printf("Error constructing request object")
		log.Printf("Error: %v", err)
		return &IamHeaders{}, "", err
	}
	var signingTime time.Time

	if os.Getenv("GO_ENV") == "testing" {
		signingTime = time.Unix(0, 0)
	} else {
		signingTime = time.Now()
	}
	err = signer.SignHTTP(context.Background(), req, sha1Hash, "appsync", "us-east-2", signingTime)

	if err != nil {
		log.Printf("%+v", err)
		panic("unable to load SDK config, " + err.Error())

	}
	host := strings.Split(canonicalURI, "/")
	iamHeaders := &IamHeaders{
		Accept:            "application/json, text/javascript",
		ContentEncoding:   "amz-1.0",
		ContentType:       "application/json; charset=UTF-8",
		Host:              host[2],
		XAmzDate:          req.Header.Get("X-Amz-Date"),
		XAmzSecurityToken: req.Header.Get("X-Amz-Security-Token"),
		Authorization:     req.Header.Get("Authorization"),
	}
	return iamHeaders, req.Header.Get("X-Amz-Security-Token"), nil
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
