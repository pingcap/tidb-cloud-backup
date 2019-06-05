package pkg

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"gocloud.dev/blob"
	"gocloud.dev/blob/gcsblob"
	"gocloud.dev/blob/s3blob"
	"gocloud.dev/gcp"
)

// SetupBucket creates a connection to a particular cloud provider's blob storage.
func SetupBucket(ctx context.Context, cloud, region, bucket, endpoint string) (*blob.Bucket, error) {
	switch cloud {
	case "aws":
		return SetupAWS(ctx, region, bucket)
	case "gcp":
		return SetupGCP(ctx, bucket)
	case "ceph":
		return SetupCeph(ctx, region, bucket, endpoint)
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
func SetupAWS(ctx context.Context, region, bucket string) (*blob.Bucket, error) {
	if len(region) == 0 {
		// Backward compatible
		region = "us-east-2"
	}
	c := &aws.Config{
		// Either hard-code the region or use AWS_REGION.
		Region: aws.String(region),
		// credentials.NewEnvCredentials assumes two environment variables are
		// present:
		// 1. AWS_ACCESS_KEY_ID, and
		// 2. AWS_SECRET_ACCESS_KEY.
		Credentials: credentials.NewEnvCredentials(),
	}
	s := session.Must(session.NewSession(c))
	return s3blob.OpenBucket(ctx, s, bucket, nil)
}

// S3Helper contains pointer to s3 client and wrappers for basic object store operations
type S3Helper struct {
	s3client *s3.S3
}

// IsBucketPresent returns true if a bucket is present and false if it's not present
func (h *S3Helper) IsBucketPresent(bucket string) (bool, error) {
	_, err := h.s3client.HeadBucket(&s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil && strings.Contains(err.Error(), "NotFound") {
		return false, nil
	} else if err == nil {
		return true, nil
	}
	return false, err
}

// CreateBucket creates bucket using s3 client
func (h *S3Helper) CreateBucket(name string) error {
	_, err := h.s3client.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(name),
	})
	return err
}

// checkBucket creates bucket if it's not present
func checkBucket(bucket string, config *aws.Config) error {
	s := session.Must(session.NewSession(config))
	c := s3.New(s, config)
	s3helper := &S3Helper{c}
	result, err := s3helper.IsBucketPresent(bucket)
	if result == false && err == nil {
		return s3helper.CreateBucket(bucket)
	}
	return err
}

// SetupCeph creates a connection to ROOK Ceph object storage with the S3 API.
// See here for more information:
// https://rook.io/docs/rook/v0.9/ceph-object.html
func SetupCeph(ctx context.Context, region, bucket, endpoint string) (*blob.Bucket, error) {
	// credentials.NewEnvCredentials assumes two environment variables are
	// present:
	// 1. AWS_ACCESS_KEY_ID, and
	// 2. AWS_SECRET_ACCESS_KEY.
	creds := credentials.NewEnvCredentials()

	if len(region) == 0 {
		region = ":default-placement"
	}
	awsConfig := aws.NewConfig().
		WithRegion(region).
		WithCredentials(creds).
		WithEndpoint(endpoint).
		WithS3ForcePathStyle(true).
		WithDisableSSL(true).
		WithMaxRetries(20)

	err := checkBucket(bucket, awsConfig)
	if err != nil {
		return nil, err
	}

	s := session.Must(session.NewSession(awsConfig))
	return s3blob.OpenBucket(ctx, s, bucket, nil)
}
