package minio

import (
	"context"
	"fmt"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/okysetiawan/retainr/config"
	"github.com/okysetiawan/retainr/internal/ports"
)

type BucketHandle interface {
}

type Client struct {
	client *minio.Client
	config config.MinioConfig
}

func (cli *Client) Close() error {
	return nil
}

func (cli *Client) PutReader(ctx context.Context, key string, reader ports.Reader) error {
	option := minio.PutObjectOptions{
		ContentType: reader.ContentType(),
	}

	if _, err := cli.client.PutObject(ctx, cli.config.BucketName, key, reader, reader.Size(), option); err != nil {
		return fmt.Errorf("failed minio PutObject: %w", err)
	}

	return nil
}

func NewClient(conf config.MinioConfig) (*Client, error) {
	option := &minio.Options{
		Creds:  credentials.NewStaticV4(conf.AccessKey, conf.SecretKey, ""),
		Secure: conf.Secure,
	}

	minioClient, err := minio.New(conf.Endpoint, option)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	bucketExists, err := minioClient.BucketExists(ctx, conf.BucketName)
	if err != nil {
		return nil, err
	}
	if !bucketExists {
		return nil, fmt.Errorf("bucket %s does not exist", conf.BucketName)
	}

	return &Client{
		client: minioClient,
		config: conf,
	}, nil
}
