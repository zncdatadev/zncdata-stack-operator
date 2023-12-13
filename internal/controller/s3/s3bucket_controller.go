/*
Copyright 2023 zncdata-labs.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package s3

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/zncdata-labs/zncdata-stack-operator/internal/util"
	corev1 "k8s.io/api/core/v1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	stackv1alpha1 "github.com/zncdata-labs/zncdata-stack-operator/api/v1alpha1"
)

// S3BucketReconciler reconciles a S3Bucket object
type S3BucketReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

//+kubebuilder:rbac:groups=stack.zncdata.net,resources=s3buckets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=stack.zncdata.net,resources=s3buckets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=stack.zncdata.net,resources=s3buckets/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the S3Bucket object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *S3BucketReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	s3Bucket := &stackv1alpha1.S3Bucket{}
	if err := r.Get(ctx, req.NamespacedName, s3Bucket); err != nil {
		if client.IgnoreNotFound(err) != nil {
			r.Log.Error(err, "unable to fetch instance")
			return ctrl.Result{}, err
		}
		r.Log.Info("s3Bucket resource not found. Ignoring since object must be deleted")
		return ctrl.Result{}, nil
	}
	//// Get the status condition, if it exists and its generation is not the
	////same as the s3Bucket's generation, reset the status conditions
	readCondition := apimeta.FindStatusCondition(s3Bucket.Status.Conditions, stackv1alpha1.ConditionTypeProgressing)
	if readCondition == nil || readCondition.ObservedGeneration != s3Bucket.GetGeneration() {
		s3Bucket.InitStatusConditions()
		if err := r.UpdateStatus(ctx, s3Bucket); err != nil {
			return ctrl.Result{}, err
		}
	}

	if err := r.CreateBucket(ctx, s3Bucket); err != nil {
		r.Log.Error(err, "create bucket")
		return ctrl.Result{}, err
	}
	r.Log.Info("create bucket success")

	isDatabaseMarkedToBeDeleted := s3Bucket.GetDeletionTimestamp() != nil
	if isDatabaseMarkedToBeDeleted {
		if controllerutil.ContainsFinalizer(s3Bucket, stackv1alpha1.S3BucketFinalizer) {
			if err := r.finalizeDatabase(ctx, s3Bucket); err != nil {
				return ctrl.Result{}, err
			}
			controllerutil.RemoveFinalizer(s3Bucket, stackv1alpha1.S3BucketFinalizer)
			if err := r.Update(ctx, s3Bucket); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// Add finalizer for this CR
	if !controllerutil.ContainsFinalizer(s3Bucket, stackv1alpha1.S3BucketFinalizer) {
		controllerutil.AddFinalizer(s3Bucket, stackv1alpha1.S3BucketFinalizer)
		err := r.Update(ctx, s3Bucket)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	if s3Bucket.IsAvailable() {
		return ctrl.Result{}, nil
	}

	s3Bucket.SetStatusCondition(metav1.Condition{
		Type:               stackv1alpha1.ConditionTypeAvailable,
		Status:             metav1.ConditionTrue,
		Reason:             stackv1alpha1.ConditionReasonRunning,
		Message:            "DatabaseConnection is running",
		ObservedGeneration: s3Bucket.GetGeneration(),
	})

	if err := r.UpdateStatus(ctx, s3Bucket); err != nil {
		r.Log.Error(err, "Failed to update status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *S3BucketReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&stackv1alpha1.S3Bucket{}).
		Complete(r)
}

func (r *S3BucketReconciler) UpdateStatus(ctx context.Context, instance *stackv1alpha1.S3Bucket) error {
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		return r.Status().Update(ctx, instance)
	})
	if retryErr != nil {
		r.Log.Error(retryErr, "Failed to update vfm status after retries")
		return retryErr
	}
	r.Log.V(1).Info("Successfully patched object status")
	return nil
}

func (r *S3BucketReconciler) CreateBucket(ctx context.Context, s3bucket *stackv1alpha1.S3Bucket) error {
	// 如果可用,或者要删除了,就不再创建
	if s3bucket.IsAvailable() || s3bucket.GetDeletionTimestamp() != nil || controllerutil.ContainsFinalizer(s3bucket, stackv1alpha1.S3BucketFinalizer) {
		return nil
	}
	if s3bucket.Spec.Reference != "" {
		if s3bucket.Spec.Name == "" {
			return errors.New("bucket name is empty")
		}
		s3Connection := &stackv1alpha1.S3Connection{}
		err := r.Get(ctx, client.ObjectKey{Namespace: s3bucket.Namespace, Name: s3bucket.Spec.Reference}, s3Connection)
		if err != nil {
			return err
		}
		adminConfig, err := GetAdminConfig(ctx, r.Client, s3Connection)
		if err != nil {
			return err
		}
		initializer, err := NewMinioInitializer(adminConfig)
		if err != nil {
			return err
		}

		accessKey := fmt.Sprintf("%s%s", s3bucket.Name, util.GenerateRandomStr(5))
		secretKey := util.GenerateRandomStr(10)
		bucketName := s3bucket.Spec.Name
		policyName := fmt.Sprintf("zncdata_%s_%s", accessKey, bucketName)
		newUserConfig := &Config{
			Endpoint:  adminConfig.Endpoint,
			AccessKey: accessKey,
			SecretKey: secretKey,
			Region:    adminConfig.Region,
			SSL:       adminConfig.SSL,
		}

		err = initializer.createBucket(ctx, s3bucket.Spec.Name, adminConfig.Region)
		if err != nil {
			return err
		}
		err = initializer.createUser(ctx, newUserConfig)
		if err != nil {
			return err
		}
		err = initializer.createUserPolicy(ctx, policyName, accessKey, bucketName)
		if err != nil {
			return err
		}
		err = initializer.bindPolicy(ctx, accessKey, policyName)
		if err != nil {
			return err
		}
		sData := make(map[string][]byte)
		sData["accessKey"] = []byte(accessKey)
		sData["secretKey"] = []byte(secretKey)

		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-secret", s3bucket.Name),
				Namespace: s3Connection.Namespace,
			},
			Type: corev1.SecretTypeOpaque,
			Data: sData,
		}
		err = util.CreateOrUpdate(ctx, r.Client, secret)
		if err != nil {
			return err
		}
		s3bucket.Spec.Credential = &stackv1alpha1.S3BucketCredential{
			ExistingSecret: secretKey,
		}
		err = r.Update(ctx, s3bucket)
		if err != nil {
			return err
		}
	} else {
		return errors.New("no reference found")
	}

	return nil
}

func (r *S3BucketReconciler) finalizeDatabase(ctx context.Context, s3bucket *stackv1alpha1.S3Bucket) error {
	if s3bucket.Spec.Credential != nil && s3bucket.Spec.Credential.ExistingSecret != "" {
		s3Connection := &stackv1alpha1.S3Connection{}
		err := r.Get(ctx, client.ObjectKey{Namespace: s3bucket.Namespace, Name: s3bucket.Spec.Credential.ExistingSecret}, s3Connection)
		if err != nil {
			return err
		}
		adminConfig, err := GetAdminConfig(ctx, r.Client, s3Connection)
		if err != nil {
			return err
		}
		initializer, err := NewMinioInitializer(adminConfig)
		if err != nil {
			return err
		}
		config := &Config{
			Endpoint: adminConfig.Endpoint,
			Region:   adminConfig.Region,
			SSL:      adminConfig.SSL,
			Token:    adminConfig.Token,
		}
		if s3bucket.Spec.Credential.ExistingSecret != "" {
			secret := corev1.Secret{}
			err := r.Get(ctx, client.ObjectKey{Namespace: s3bucket.Namespace, Name: s3bucket.Spec.Credential.ExistingSecret}, &secret)
			if err != nil {
				return err
			}
			if accessKey, ok := secret.Data["accessKey"]; ok {
				decodeString, err := base64.StdEncoding.DecodeString(string(accessKey))
				if err != nil {
					return err
				}
				config.AccessKey = string(decodeString)
			}
			if secretKey, ok := secret.Data["secretKey"]; ok {
				decodeString, err := base64.StdEncoding.DecodeString(string(secretKey))
				if err != nil {
					return err
				}
				config.SecretKey = string(decodeString)
			}

		} else {
			return errors.New("no secret found")
		}

		if s3bucket.Spec.Name == "" {
			return errors.New("bucket name is empty")
		}
		err = initializer.removeUser(ctx, config.AccessKey)
		if err != nil {
			return err
		}
		err = initializer.removeBucket(ctx, s3bucket.Spec.Name)
		if err != nil {
			return err
		}

	} else {
		return errors.New("no reference found")
	}

	return nil

}
