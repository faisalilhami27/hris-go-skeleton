package common

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	configAws "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/sirupsen/logrus"
	"mime/multipart"
)

type S3Client struct {
	AccessKeyID     string
	SecretAccessKey string
	Region          string
	BucketName      string
}

type IS3Client interface {
	UploadFile(ctx context.Context, fileHeader multipart.FileHeader, filename string) (string, error)
}

func NewS3Client(accessKeyID, secretAccessKey, region, bucketName string) *S3Client {
	return &S3Client{
		AccessKeyID:     accessKeyID,
		SecretAccessKey: secretAccessKey,
		Region:          region,
		BucketName:      bucketName,
	}
}

func (s *S3Client) createClient(ctx context.Context) (*s3.Client, error) {
	cfg, err := configAws.LoadDefaultConfig(ctx,
		configAws.WithRegion(s.Region),
		configAws.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				s.AccessKeyID,
				s.SecretAccessKey,
				"",
			),
		),
	)
	if err != nil {
		logrus.Errorf("unable to load SDK config, %v", err)
		return nil, err
	}

	return s3.NewFromConfig(cfg), nil
}

func (s *S3Client) UploadFile(ctx context.Context, fileHeader multipart.FileHeader, filename string) (string, error) {
	client, err := s.createClient(ctx)
	if err != nil {
		logrus.Errorf("an error occurred when create client : %v", err)
		return "", err
	}

	file, err := fileHeader.Open()
	if err != nil {
		logrus.Errorf("failed to open uploaded file: %v", err)
		return "", err
	}
	defer file.Close()

	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(filename),
		Body:   file,
	})
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.BucketName, s.Region, filename), nil
}
