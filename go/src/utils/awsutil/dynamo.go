package awsutil

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

// NewDynamoClient builds a DynamoDB client. If AWS_ENDPOINT_URL_DYNAMODB is
// set (i.e. tests running against DynamoDB Local via with-dynamodb-local.sh),
// it points there with dummy credentials. Otherwise it uses the standard
// AWS SDK credential chain, suitable for real deployments.
func NewDynamoClient() *dynamodb.DynamoDB {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	cfg := aws.NewConfig()
	if endpoint := os.Getenv("AWS_ENDPOINT_URL_DYNAMODB"); endpoint != "" {
		cfg = cfg.
			WithEndpoint(endpoint).
			WithRegion("local").
			WithCredentials(credentials.NewStaticCredentials("dummy", "dummy", ""))
	}
	return dynamodb.New(sess, cfg)
}
