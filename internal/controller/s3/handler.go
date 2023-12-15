package s3

import (
	"context"
	"encoding/base64"
	"github.com/minio/madmin-go/v3"
	stackv1alpha1 "github.com/zncdata-labs/zncdata-stack-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateS3Config(s3Bucket *stackv1alpha1.S3Bucket) *Config {
	config := &Config{}
	return config
}

func CreateS3AdminClient(ctx context.Context, c client.Client, s3Connection *stackv1alpha1.S3Connection) (*madmin.AdminClient, error) {
	config, err := GetAdminConfig(ctx, c, s3Connection)
	if err != nil {
		return nil, err
	}
	adminClient, err := NewAdminClient(config)
	if err != nil {
		return nil, err
	}
	return adminClient, nil
}

func GetAdminConfig(ctx context.Context, c client.Client, s3Connection *stackv1alpha1.S3Connection) (*Config, error) {
	config := &Config{}
	if s3Connection.Spec.S3Credential.ExistingSecret != "" {
		secret := &corev1.Secret{}
		err := c.Get(ctx, client.ObjectKey{
			Namespace: s3Connection.Namespace,
			Name:      s3Connection.Spec.S3Credential.ExistingSecret,
		}, secret)
		if err != nil {
			return nil, err
		}
		if endpoint, ok := secret.Data["endpoint"]; ok {
			decodeString, err := base64.StdEncoding.DecodeString(string(endpoint))
			if err != nil {
				return nil, err
			}
			config.Endpoint = string(decodeString)
		}
		if accessKey, ok := secret.Data["accessKey"]; ok {
			decodeString, err := base64.StdEncoding.DecodeString(string(accessKey))
			if err != nil {
				return nil, err
			}
			config.AccessKey = string(decodeString)
		}
		if secretKey, ok := secret.Data["secretKey"]; ok {
			decodeString, err := base64.StdEncoding.DecodeString(string(secretKey))
			if err != nil {
				return nil, err
			}
			config.SecretKey = string(decodeString)
		}
		if region, ok := secret.Data["region"]; ok {
			decodeString, err := base64.StdEncoding.DecodeString(string(region))
			if err != nil {
				return nil, err
			}
			config.Region = string(decodeString)
		}
		if ssl, ok := secret.Data["ssl"]; ok {
			decodeString, err := base64.StdEncoding.DecodeString(string(ssl))
			if err != nil {
				return nil, err
			}
			config.SSL = string(decodeString) == "true"
		}
	} else {
		config.Endpoint = s3Connection.Spec.S3Credential.Endpoint
		config.AccessKey = s3Connection.Spec.S3Credential.AccessKey
		config.SecretKey = s3Connection.Spec.S3Credential.SecretKey
		config.Region = s3Connection.Spec.S3Credential.Region
		config.SSL = s3Connection.Spec.S3Credential.SSL
	}
	return config, nil
}
