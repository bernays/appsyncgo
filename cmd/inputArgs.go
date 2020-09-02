package cmd

// InputArguments holds all of the arguments that are used
type InputArguments struct {
	URL     string
	APIAuth APIAuth
}

// APIAuth parameters used to authenticate with AppSync
type APIAuth struct {
	AuthType         string // AWS_IAM, API_KEY
	APIKey           string // Only required if Type is API_KEY
	Profile          string // Only required if Type is API_KEY
	AwsAccessKey     string // Only required if Type is AWS_IAM
	AwsAccessSecret  string // Only required if Type is AWS_IAM
	AwsSecurityToken string // Only required if Type is AWS_IAM

}
