package ypss3

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type YPSS3 struct {
	client *s3.Client

	bucket          string
	uploadKeyPrefix string
	uploadURLPrefix string
}

func Open(bucket, uploadKeyPrefix, uploadURLPrefix string) (*YPSS3, error) {
	// Load the Shared AWS Configuration (~/.aws/config)
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}

	theS3 := YPSS3{
		bucket:          bucket,
		uploadKeyPrefix: uploadKeyPrefix,
		uploadURLPrefix: uploadURLPrefix,
	}

	// Create an Amazon S3 service client
	theS3.client = s3.NewFromConfig(cfg)

	return &theS3, nil
}
