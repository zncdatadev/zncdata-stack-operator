package s3

import (
	"context"

	"fmt"
	"github.com/minio/madmin-go/v3"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"k8s.io/apimachinery/pkg/util/json"
)

type Policy struct {
	Version    string       `json:"Version,omitempty"`
	Statements []*Statement `json:"Statement,omitempty"`
}
type Statement struct {
	Sid       string    `json:"Sid,omitempty"`
	Effect    string    `json:"Effect,omitempty"`
	Principal Principal `json:"Principal,omitempty"`
	Resource  string    `json:"Resource,omitempty"`
	Action    []string  `json:"Action,omitempty"`
}
type Principal struct {
	AWS []string `json:"AWS,omitempty"`
}

type Config struct {
	// 访问地址,ip:port
	Endpoint  string `json:"endpoint,omitempty"`
	AccessKey string `json:"access_key,omitempty"`
	SecretKey string `json:"secret_key,omitempty"`
	Region    string `json:"region,omitempty"`
	Token     string `json:"token,omitempty"`
	SSL       bool   `json:"ssl,omitempty"`
}

// NewAdminClient 用来管理权限的客户端
func NewAdminClient(s3Config *Config) (*madmin.AdminClient, error) {
	// Initialize minio client object.
	mdmClnt, err := madmin.New(s3Config.Endpoint, s3Config.AccessKey, s3Config.SecretKey, s3Config.SSL)
	if err != nil {
		return nil, err
	}
	return mdmClnt, nil
}

// NewClient 用来操作s3的客户端(上传下载,处理通相关逻辑)
func NewClient(s3Config *Config) (*minio.Client, error) {
	// Initialize minio client object.
	minioClient, err := minio.New(s3Config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(s3Config.AccessKey, s3Config.SecretKey, s3Config.Token),
		Secure: s3Config.SSL,
	})
	if err != nil {
		return nil, err
	}
	return minioClient, nil
}

type Initializer struct {
	IS3Initializer
}

func (i Initializer) createBucket(ctx context.Context, bucketName, region string) error {
	panic("implement me")
}

func (i Initializer) createUser(ctx context.Context, s3Config *Config) error {
	panic("implement me")
}

func (i Initializer) createUserPolicy(ctx context.Context, policyName, accessKey, bucketName string) error {
	panic("implement me")
}

func (i Initializer) bindPolicy(ctx context.Context, accessKey, policyName string) error {
	panic("implement me")
}

func (i Initializer) removeUser(ctx context.Context, accessKey string) error {
	panic("implement me")
}

func (i Initializer) removeBucket(ctx context.Context, bucketName string) error {
	panic("implement me")
}

type MinioInitializer struct {
	Initializer
	adminClient *madmin.AdminClient
	client      *minio.Client
}

func (mi *MinioInitializer) createUser(ctx context.Context, s3Config *Config) error {
	return mi.adminClient.AddUser(ctx, s3Config.AccessKey, s3Config.SecretKey)
}

func (mi *MinioInitializer) createBucket(ctx context.Context, bucketName, region string) error {
	exists, err := mi.client.BucketExists(ctx, bucketName)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	return mi.client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: region})
}

// generatePolicy - generate policy for user
func generatePolicy(accessKey, bucket string) []byte {
	principalAws := fmt.Sprintf("arn:aws:iam::%s:root", accessKey)
	resource := fmt.Sprintf("arn:aws:s3:::%s/*", bucket)
	policy := &Policy{
		Version: "2012-10-17",
		Statements: []*Statement{
			{
				Sid:    "ObjectLevel",
				Effect: "Allow",
				Principal: Principal{
					AWS: []string{principalAws},
				},
				Resource: resource,
				Action:   []string{"s3:*"},
			},
			{
				Sid:    "BucketLevel",
				Effect: "Allow",
				Principal: Principal{
					AWS: []string{principalAws},
				},
				Resource: resource,
				Action:   []string{"s3:*"},
			},
		},
	}
	marshal, err := json.Marshal(policy)
	if err != nil {
		return nil
	}
	return marshal
}

func (mi *MinioInitializer) createUserPolicy(ctx context.Context, policyName, accessKey, bucketName string) error {
	policy := generatePolicy(accessKey, bucketName)
	return mi.adminClient.AddCannedPolicy(ctx, policyName, policy)

}

func (mi *MinioInitializer) bindPolicy(ctx context.Context, accessKeyID, policyName string) error {
	return mi.adminClient.SetPolicy(ctx, policyName, accessKeyID, false)
}

func (mi *MinioInitializer) removeUser(ctx context.Context, accessKey string) error {
	return mi.adminClient.RemoveUser(ctx, accessKey)
}

func (mi *MinioInitializer) removeBucket(ctx context.Context, bucketName string) error {
	return mi.client.RemoveBucket(ctx, bucketName)
}

func NewMinioInitializer(s3Config *Config) (*MinioInitializer, error) {
	admin, err := NewAdminClient(s3Config)
	if err != nil {
		return nil, err
	}

	s3Client, err := NewClient(s3Config)
	if err != nil {
		return nil, err
	}

	return &MinioInitializer{
		adminClient: admin,
		client:      s3Client,
	}, nil
}

type IS3Initializer interface {
	createBucket(ctx context.Context, bucketName, region string) error
	createUser(ctx context.Context, s3Config *Config) error
	removeUser(ctx context.Context, accessKey string) error
	removeBucket(ctx context.Context, bucketName string) error
	createUserPolicy(ctx context.Context, policyName, accessKey, bucketName string) error
	bindPolicy(ctx context.Context, accessKey, policyName string) error
}
