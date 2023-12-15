package controller

import (
	"context"
	"encoding/base64"
	stackv1alpha1 "github.com/zncdata-labs/zncdata-stack-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apitypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
)

// 测试连接是否可用
func (r *DatabaseConnectionReconciler) checkConnection(ctx context.Context, instance *stackv1alpha1.DatabaseConnection) error {
	dsn, err := getDSNFromConnection(ctx, r.Client, instance)
	if err != nil {
		return err
	}
	initializer, err := NewDBInitializer(dsn)
	if err != nil {
		return err
	}
	return initializer.ping()
}

func getDSNFromConnection(ctx context.Context, c client.Client, instance *stackv1alpha1.DatabaseConnection) (*DSN, error) {
	provider := instance.Spec.Provider
	// 查询secret
	dsn := &DSN{
		Driver:  provider.Driver,
		Host:    provider.Host,
		Port:    strconv.Itoa(provider.Port),
		SSLMode: provider.SSL,
	}

	if instance.Spec.Provider.Credential.ExistSecret != "" {
		secret := &corev1.Secret{}
		name := apitypes.NamespacedName{
			Namespace: instance.Namespace,
			Name:      instance.Spec.Provider.Credential.ExistSecret,
		}
		if err := c.Get(ctx, name, secret); err != nil {
			if client.IgnoreNotFound(err) != nil {
				return nil, err
			}
		}
		// data base64 decode
		if username, ok := secret.Data["username"]; ok {
			decodeString, err := base64.StdEncoding.DecodeString(string(username))
			if err != nil {
				return nil, err
			}
			dsn.Username = string(decodeString)
		}
		if password, ok := secret.Data["password"]; ok {
			decodeString, err := base64.StdEncoding.DecodeString(string(password))
			if err != nil {
				return nil, err
			}
			dsn.Password = string(decodeString)
		}
	}
	return dsn, nil
}
