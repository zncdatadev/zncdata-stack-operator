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

package controller

import (
	"context"
	"errors"
	"github.com/go-logr/logr"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	stackv1alpha1 "github.com/zncdata-labs/zncdata-stack-operator/api/v1alpha1"
)

// DatabaseConnectionReconciler reconciles a DatabaseConnection object
type DatabaseConnectionReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

//+kubebuilder:rbac:groups=stack.zncdata.net,resources=databaseconnections,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=stack.zncdata.net,resources=databaseconnections/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=stack.zncdata.net,resources=databaseconnections/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the DatabaseConnection object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *DatabaseConnectionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	databaseConnection := &stackv1alpha1.DatabaseConnection{}

	if err := r.Get(ctx, req.NamespacedName, databaseConnection); err != nil {
		if client.IgnoreNotFound(err) != nil {
			r.Log.Error(err, "unable to fetch instance")
			return ctrl.Result{}, err
		}
		r.Log.Info("DatabaseConnection resource not found. Ignoring since object must be deleted")
		return ctrl.Result{}, nil
	}

	//// Get the status condition, if it exists and its generation is not the
	////same as the DatabaseConnection's generation, reset the status conditions
	readCondition := apimeta.FindStatusCondition(databaseConnection.Status.Conditions, stackv1alpha1.ConditionTypeProgressing)
	if readCondition == nil || readCondition.ObservedGeneration != databaseConnection.GetGeneration() {
		databaseConnection.InitStatusConditions()
		if err := r.UpdateStatus(ctx, databaseConnection); err != nil {
			return ctrl.Result{}, err
		}
	}

	r.Log.Info("DatabaseConnection is found", "Name", databaseConnection.Name)

	if err := r.checkDefault(ctx, databaseConnection); err != nil {
		r.Log.Error(err, "Failed to check default")
		databaseConnection.SetStatusCondition(metav1.Condition{
			Type:               stackv1alpha1.ConditionTypeReconcile,
			Status:             metav1.ConditionFalse,
			Reason:             stackv1alpha1.ConditionReasonPreparing,
			Message:            "DatabaseConnection has exist default.",
			ObservedGeneration: databaseConnection.GetGeneration(),
		})
		if err := r.UpdateStatus(ctx, databaseConnection); err != nil {
			r.Log.Error(err, "Failed to update status")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, err
	}

	if err := r.checkConnection(ctx, databaseConnection); err != nil {
		r.Log.Error(err, "Failed to check connection")
		return ctrl.Result{}, err
	}

	if databaseConnection.IsAvailable() {
		return ctrl.Result{}, nil
	}

	databaseConnection.SetStatusCondition(metav1.Condition{
		Type:               stackv1alpha1.ConditionTypeAvailable,
		Status:             metav1.ConditionTrue,
		Reason:             stackv1alpha1.ConditionReasonRunning,
		Message:            "DatabaseConnection is running",
		ObservedGeneration: databaseConnection.GetGeneration(),
	})

	if err := r.UpdateStatus(ctx, databaseConnection); err != nil {
		r.Log.Error(err, "Failed to update status")
		return ctrl.Result{}, err
	}

	r.Log.Info("Successfully updated status")

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DatabaseConnectionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&stackv1alpha1.DatabaseConnection{}).
		Complete(r)
}

func (r *DatabaseConnectionReconciler) UpdateStatus(ctx context.Context, instance *stackv1alpha1.DatabaseConnection) error {
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

func (r *DatabaseConnectionReconciler) checkDefault(ctx context.Context, connection *stackv1alpha1.DatabaseConnection) error {
	if !connection.Spec.Default {
		return nil
	}
	list := stackv1alpha1.DatabaseConnectionList{}
	ops := &client.ListOptions{
		Namespace: connection.Namespace,
	}
	err := r.Client.List(ctx, &list, ops)
	if err != nil {
		return err
	}
	for _, item := range list.Items { // 同一个driver类型的default只能有一个
		if item.Spec.Default &&
			item.Name != connection.Name &&
			item.Spec.Provider.Driver == connection.Spec.Provider.Driver {
			return errors.New("default connection already exists")
		}
	}
	return nil
}
