package main

import (
	"github.com/bernays/appsync-go-client/cmd"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)
var (
	authType = "AWS_IAM"
	profile  = "default"
)

func main() {
	v := cmd.VersionInfo{
		Version: version,
		Commit:  commit,
		Date:    cmd.ParseDate(date),
	}
	i := cmd.InputArguments{
		URL: "https://whom3blq6vhxhd6rkt3offziva.appsync-api.us-east-2.amazonaws.com/graphql",
		APIAuth: cmd.APIAuth{
			AuthType: authType,
			Profile:  profile,
		},
	}
	cmd.Execute(i, v)
}
