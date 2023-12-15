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
	"github.com/go-logr/logr"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	stackv1alpha1 "github.com/zncdata-labs/zncdata-stack-operator/api/v1alpha1"
)

// S3ConnectionReconciler reconciles a S3Connection object
type S3ConnectionReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

//+kubebuilder:rbac:groups=stack.zncdata.net,resources=s3connections,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=stack.zncdata.net,resources=s3connections/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=stack.zncdata.net,resources=s3connections/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the S3Connection object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *S3ConnectionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	s3Connection := &stackv1alpha1.S3Connection{}
	if err := r.Get(ctx, req.NamespacedName, s3Connection); err != nil {
		if client.IgnoreNotFound(err) != nil {
			r.Log.Error(err, "unable to fetch instance")
			return ctrl.Result{}, err
		}
		r.Log.Info("s3Connection resource not found. Ignoring since object must be deleted")
		return ctrl.Result{}, nil
	}
	//// Get the status condition, if it exists and its generation is not the
	////same as the s3Connection's generation, reset the status conditions
	readCondition := apimeta.FindStatusCondition(s3Connection.Status.Conditions, stackv1alpha1.ConditionTypeProgressing)
	if readCondition == nil || readCondition.ObservedGeneration != s3Connection.GetGeneration() {
		s3Connection.InitStatusConditions()
		if err := r.UpdateStatus(ctx, s3Connection); err != nil {
			return ctrl.Result{}, err
		}
	}

	_, err := CreateS3AdminClient(ctx, r.Client, s3Connection)
	if err != nil {
		r.Log.Error(err, "create s3 admin client", "err: ", err)
		return ctrl.Result{}, err
	}

	r.Log.Info("s3Connection is found", "Name", s3Connection.Name)

	if s3Connection.IsAvailable() {
		return ctrl.Result{}, err
	}

	s3Connection.SetStatusCondition(metav1.Condition{
		Type:               stackv1alpha1.ConditionTypeAvailable,
		Status:             metav1.ConditionTrue,
		Reason:             stackv1alpha1.ConditionReasonRunning,
		Message:            "s3Connection is running",
		ObservedGeneration: s3Connection.GetGeneration(),
	})

	if err := r.UpdateStatus(ctx, s3Connection); err != nil {
		r.Log.Error(err, "Failed to update status")
		return ctrl.Result{}, err
	}

	r.Log.Info("Successfully updated status")

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *S3ConnectionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&stackv1alpha1.S3Connection{}).
		Complete(r)
}

func (r *S3ConnectionReconciler) UpdateStatus(ctx context.Context, instance *stackv1alpha1.S3Connection) error {
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
