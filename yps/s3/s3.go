package ypss3

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
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

func (ys3 *YPSS3) Upload(key string, body io.Reader) (*S3Upload, error) {
	name := fmt.Sprintf("%s%s", ys3.uploadKeyPrefix, key)
	_, err := ys3.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(ys3.bucket),
		Key:    aws.String(name),
		Body:   body,
	})
	if err != nil {
		return nil, err
	}
	return &S3Upload{
		Filename: name,
		URL:      fmt.Sprintf("%s%s", ys3.uploadURLPrefix, key),
	}, nil
}

func (ys3 *YPSS3) Delete(key string) error {
	name := fmt.Sprintf("%s%s", ys3.uploadKeyPrefix, key)
	_, err := ys3.client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(ys3.bucket),
		Key:    aws.String(name),
	})
	return err
}

func (ys3 *YPSS3) FileExists(key string) (bool, error) {
	name := fmt.Sprintf("%s%s", ys3.uploadKeyPrefix, key)
	_, err := ys3.client.HeadObject(context.TODO(), &s3.HeadObjectInput{
		Bucket: aws.String(ys3.bucket),
		Key:    aws.String(name),
	})
	var apiError smithy.APIError
	if err == nil {
		return true, nil
	} else if errors.As(err, &apiError) && apiError.ErrorCode() == "NotFound" {
		return false, nil
	}
	fmt.Println("Other type of error encountered from S3 HeadObject:", apiError.ErrorCode(), err)
	return true, err
}

func (ys3 *YPSS3) EntryFileKey(entryID, filename string) string {
	return fmt.Sprintf("entries/%s/%s", entryID, filename)
}

func (ys3 *YPSS3) EntryFileURL(entryID, filename string) string {
	return fmt.Sprintf("%sentries/%s/%s", ys3.uploadURLPrefix, entryID, filename)
}
