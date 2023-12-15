package s3

import (
	"context"
	"fmt"
	"github.com/zncdata-labs/zncdata-stack-operator/internal/util"
	"os"
	"strings"
	"testing"
)

// NewDefaultConfig minio 官网提供的配置
func NewDefaultConfig() *Config {
	getenv := os.Getenv("ENV")
	if getenv == "local" {
		return NewLocalConfig()
	}
	return NewConfigFromEnv()
}

func NewLocalConfig() *Config {
	return &Config{
		Endpoint:  "127.0.0.1:9000",
		AccessKey: "admin",
		SecretKey: "admin123456",
		Region:    "us-east-1",
		Token:     "",
		SSL:       false,
	}
}

func NewConfigFromEnv() *Config {
	endpoint := os.Getenv("MINIO_ENDPOINT")
	accessKey := os.Getenv("MINIO_ACCESS_KEY")
	secretKey := os.Getenv("MINIO_SECRET_KEY")
	region := os.Getenv("MINIO_REGION")
	Token := os.Getenv("MINIO_TOKEN")
	ssl := os.Getenv("MINIO_SSL")
	return &Config{
		Endpoint:  endpoint,
		AccessKey: accessKey,
		SecretKey: secretKey,
		Region:    region,
		Token:     Token,
		SSL:       ssl == "true",
	}
}

func TestNewClient(t *testing.T) {
	config := NewDefaultConfig()
	client, err := NewClient(config)
	ctx := context.Background()
	if err != nil {
		t.Error(err)
	}
	buckets, err := client.ListBuckets(ctx)
	if err != nil {
		t.Error(err)
	}
	t.Log("list buckets", "buckets: ", buckets)
}

func TestListUser(t *testing.T) {
	config := NewDefaultConfig()
	client, err := NewAdminClient(config)
	if err != nil {
		t.Error("new admin", "err: ", err)
	}
	t.Log("new admin", "admin: ", client)

	users, err := client.ListUsers(context.Background())
	if err != nil {
		t.Error("list users", "err: ", err)
	}
	t.Log("list users", "users: ", users)
}

func TestS3Suit(t *testing.T) {
	config := NewDefaultConfig()
	ctx := context.Background()
	initializer, err := NewMinioInitializer(config)
	if err != nil {
		t.Error("new minio initializer", "err: ", err)
	}

	str := fmt.Sprintf("zncdata_minio_%s", util.GenerateRandomStr(5))
	accessKey := fmt.Sprintf("%s_user", str)
	secretKey, err := util.GenerateSecretAccessKey()
	if err != nil {
		secretKey = util.GenerateRandomStr(10)
	}
	policyName := fmt.Sprintf("%s_policy", str)
	bucketName := strings.ToLower(util.RemoveSpecialCharacter(fmt.Sprintf("%s_bucket", str)))

	newUserConfig := &Config{
		Endpoint:  config.Endpoint,
		AccessKey: accessKey,
		SecretKey: secretKey,
		Region:    config.Region,
		Token:     "",
		SSL:       config.SSL,
	}
	err = initializer.createUser(ctx, newUserConfig)
	if err != nil {
		t.Error("create user", "err: ", err)
	}
	t.Log("create user", "accessKey: ", accessKey, "secretKey: ", secretKey)

	err = initializer.createBucket(ctx, bucketName, newUserConfig.Region)
	if err != nil {
		t.Error("create bucket", "err: ", err)
	}
	t.Log("create bucket", "bucketName: ", bucketName)

	err = initializer.createUserPolicy(ctx, policyName, accessKey, bucketName)
	if err != nil {
		t.Error("create user policy", "err: ", err)
	}
	t.Log("create user policy", "policyName: ", policyName, "accessKey: ", accessKey, "bucketName: ", bucketName)

	err = initializer.bindPolicy(ctx, accessKey, policyName)
	if err != nil {
		t.Error("bind policy", "err: ", err)
	}
	t.Log("bind policy", "policyName: ", policyName, "accessKey: ", accessKey)

	client, err := NewClient(newUserConfig)
	if err != nil {
		t.Error("new client", "err: ", err)
	}
	t.Log("new client", "client: ", client)

	exists, err := client.BucketExists(ctx, bucketName)
	if err != nil {
		t.Error("bucket exists", "err: ", err)
	}
	t.Log("bucket exists", "exists: ", exists)

	err = client.RemoveBucket(ctx, bucketName)
	if err != nil {
		t.Error("remove bucket", "err: ", err)
	}
	t.Log("remove bucket", "bucketName: ", bucketName)

	err = initializer.removeUser(ctx, accessKey)
	if err != nil {
		t.Error("remove user", "err: ", err)
	}
	t.Log("remove user", "accessKey: ", accessKey)
}

func Test_generatePolicy(t *testing.T) {
	type args struct {
		accessKey string
		bucket    string
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "test generate policy 1",
			args: args{
				accessKey: "test_000_zncdata_1",
				bucket:    "test_000_zncdata_1",
			},
			want: []byte("{\"Version\":\"2012-10-17\",\"Statement\":[{\"Sid\":\"ObjectLevel\",\"Effect\":\"Allow\",\"Principal\":{\"AWS\":[\"arn:aws:iam::test_000_zncdata_1:root\"]},\"Resource\":\"arn:aws:s3:::test_000_zncdata_1/*\",\"Action\":[\"s3:*\"]},{\"Sid\":\"BucketLevel\",\"Effect\":\"Allow\",\"Principal\":{\"AWS\":[\"arn:aws:iam::test_000_zncdata_1:root\"]},\"Resource\":\"arn:aws:s3:::test_000_zncdata_1/*\",\"Action\":[\"s3:*\"]}]}"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generatePolicy(tt.args.accessKey, tt.args.bucket)
			t.Log("generate policy", "got: ", string(got), "want: ", string(tt.want))
		})
	}
}
