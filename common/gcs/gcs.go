package common

import (
	"bytes"
	"cloud.google.com/go/storage"
	"fmt"

	"context"
	"encoding/json"

	log "github.com/sirupsen/logrus"

	"google.golang.org/api/option"

	"io"
	"time"
)

type GCSPackage struct {
	ServiceAccountKeyJSON ServiceAccountKeyJSON
	BucketName            string
	TimeoutInSeconds      uint
}

type ServiceAccountKeyJSON struct {
	Type                    string `json:"type"`
	ProjectID               string `json:"project_id"`
	PrivateKeyID            string `json:"private_key_id"`
	PrivateKey              string `json:"private_key"`
	ClientEmail             string `json:"client_email"`
	ClientID                string `json:"client_id"`
	AuthURI                 string `json:"auth_uri"`
	TokenURI                string `json:"token_uri"`
	AuthProviderX509CertURL string `json:"auth_provider_x509_cert_url"`
	ClientX509CertURL       string `json:"client_x509_cert_url"`
	UniverseDomain          string `json:"universe_domain"`
}

type IGCSClient interface {
	UploadFile(ctx context.Context, fileName string, data []byte) (string, error)
	DeleteFile(ctx context.Context, fileName string) error
}

func NewGCSClient(
	serviceAccountKey ServiceAccountKeyJSON,
	bucketName string) IGCSClient {
	return &GCSPackage{
		ServiceAccountKeyJSON: serviceAccountKey,
		BucketName:            bucketName,
	}
}

func (c *GCSPackage) createClient(ctx context.Context) (*storage.Client, error) {
	reqBodyBytes := new(bytes.Buffer)
	err := json.NewEncoder(reqBodyBytes).Encode(c.ServiceAccountKeyJSON)
	if err != nil {
		log.Errorf("an error occurred when encode invoice account key json : %v", err)
		return nil, err
	}

	jsonByte := reqBodyBytes.Bytes()
	client, err := storage.NewClient(ctx, option.WithCredentialsJSON(jsonByte))
	if err != nil {
		log.Errorf("an error occurred when create gcs client : %v", err)
		return nil, err
	}

	return client, nil
}

func (c *GCSPackage) UploadFile(ctx context.Context, fileName string, data []byte) (string, error) {
	var (
		contentType      = "application/octet-stream"
		timeoutInSeconds uint
	)
	client, err := c.createClient(ctx)
	if err != nil {
		log.Errorf("an error occurred when create client : %v", err)
		return "", err
	}

	defer func(Client *storage.Client) {
		err := Client.Close()
		if err != nil {
			log.Errorf("an error occurred when close client : %v", err)
			return
		}
	}(client)

	if c.TimeoutInSeconds > 0 {
		timeoutInSeconds = c.TimeoutInSeconds
	} else {
		timeoutInSeconds = 30
	}

	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeoutInSeconds)*time.Second)
	defer cancel()

	buc := client.Bucket(c.BucketName)
	obj := buc.Object(fileName)
	buf := bytes.NewBuffer(data)

	writer := obj.NewWriter(ctx)
	writer.ChunkSize = 0

	if _, err = io.Copy(writer, buf); err != nil {
		log.Errorf("an error occurred when copy file : %v", err)
		return "", err
	}
	if err = writer.Close(); err != nil {
		log.Errorf("an error occurred when close writer : %v", err)
		return "", err
	}

	if _, err = obj.Update(ctx, storage.ObjectAttrsToUpdate{ContentType: contentType}); err != nil {
		log.Errorf("an error occurred when update object : %v", err)
		return "", err
	}

	url := fmt.Sprintf("https://storage.googleapis.com/%s/%s", c.BucketName, fileName)
	return url, nil
}

func (c *GCSPackage) DeleteFile(ctx context.Context, fileName string) error {
	client, err := c.createClient(ctx)
	if err != nil {
		log.Errorf("an error occurred when creating client: %v", err)
		return err
	}

	defer func(Client *storage.Client) {
		if err = Client.Close(); err != nil {
			log.Errorf("an error occurred when closing client: %v", err)
		}
	}(client)

	ctx, cancel := context.WithTimeout(ctx, time.Duration(40)*time.Second)
	defer cancel()

	buc := client.Bucket(c.BucketName)
	obj := buc.Object(fileName)

	if err = obj.Delete(ctx); err != nil {
		log.Errorf("an error occurred when deleting object: %v", err)
		return err
	}

	log.Infof("Object %s successfully deleted from bucket %s", fileName, c.BucketName)
	return nil
}
