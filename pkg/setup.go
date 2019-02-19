package pkg

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"gocloud.dev/blob"
	"gocloud.dev/blob/gcsblob"
	"gocloud.dev/blob/s3blob"
	"gocloud.dev/gcp"
)

// SetupBucket creates a connection to a particular cloud provider's blob storage.
func SetupBucket(ctx context.Context, cloud, bucket, endpoint string) (*blob.Bucket, error) {
	switch cloud {
	case "aws":
		return SetupAWS(ctx, bucket)
	case "gcp":
		return SetupGCP(ctx, bucket)
	case "ceph":
		return SetupCeph(ctx, bucket, endpoint)
	default:
		return nil, fmt.Errorf("invalid cloud provider: %s", cloud)
	}
}

// SetupGCP creates a connection to Google Cloud Storage (GCS).
func SetupGCP(ctx context.Context, bucket string) (*blob.Bucket, error) {
	// DefaultCredentials assumes a user has logged in with gcloud.
	// See here for more information:
	// https://cloud.google.com/docs/authentication/getting-started
	creds, err := gcp.DefaultCredentials(ctx)
	if err != nil {
		return nil, err
	}
	c, err := gcp.NewHTTPClient(gcp.DefaultTransport(), gcp.CredentialsTokenSource(creds))
	if err != nil {
		return nil, err
	}
	return gcsblob.OpenBucket(ctx, c, bucket, nil)
}

// SetupAWS creates a connection to Simple Cloud Storage Service (S3).
func SetupAWS(ctx context.Context, bucket string) (*blob.Bucket, error) {
	c := &aws.Config{
		// Either hard-code the region or use AWS_REGION.
		Region: aws.String("us-east-2"),
		// credentials.NewEnvCredentials assumes two environment variables are
		// present:
		// 1. AWS_ACCESS_KEY_ID, and
		// 2. AWS_SECRET_ACCESS_KEY.
		Credentials: credentials.NewEnvCredentials(),
	}
	s := session.Must(session.NewSession(c))
	return s3blob.OpenBucket(ctx, s, bucket, nil)
}

// SetupCeph creates a connection to ROOK Ceph object storage with the S3 API.
// See here for more information:
// https://rook.io/docs/rook/v0.9/ceph-object.html
func SetupCeph(ctx context.Context, bucket, endpoint string) (*blob.Bucket, error) {
	// credentials.NewEnvCredentials assumes two environment variables are
	// present:
	// 1. AWS_ACCESS_KEY_ID, and
	// 2. AWS_SECRET_ACCESS_KEY.
	creds := credentials.NewEnvCredentials()

	awsConfig := aws.NewConfig().
		WithRegion("us-east-1").
		WithCredentials(creds).
		WithEndpoint(endpoint).
		WithS3ForcePathStyle(true).
		WithDisableSSL(true).
		WithMaxRetries(20)

	s := session.Must(session.NewSession(awsConfig))
	return s3blob.OpenBucket(ctx, s, bucket, nil)
}
