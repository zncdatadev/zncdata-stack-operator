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
	"encoding/base64"
	"errors"
	"github.com/go-logr/logr"
	stackv1alpha1 "github.com/zncdata-labs/zncdata-stack-operator/api/v1alpha1"
	"github.com/zncdata-labs/zncdata-stack-operator/internal/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"strings"
)

// DatabaseReconciler reconciles a Database object
type DatabaseReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

//+kubebuilder:rbac:groups=stack.zncdata.net,resources=databases,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=stack.zncdata.net,resources=databases/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=stack.zncdata.net,resources=databases/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Database object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *DatabaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	obj := &stackv1alpha1.Database{}
	// 1. Get the Database object
	if err := r.Get(ctx, req.NamespacedName, obj); err != nil {
		if err := client.IgnoreNotFound(err); err != nil {
			r.Log.Error(err, "unable to fetch Database")
			return ctrl.Result{}, err
		}
		r.Log.Info("Database not found, Ignoring it since it has been deleted")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// 2. Get the DatabaseConnection object, if it exists; use the default if it doesn't exist.
	connection, err := r.getReferenceDatabaseConnection(ctx, obj)
	if err != nil {
		r.Log.Error(err, "unable to fetch DatabaseConnection")
		return ctrl.Result{}, err
	}

	// 3. Get the DatabaseSecret object, if it exists; use the default if it doesn't exist.
	err = r.generateDatabaseSecret(ctx, obj, connection)
	if err != nil {
		r.Log.Error(err, "unable to generate DatabaseSecret")
		return ctrl.Result{}, err
	}
	r.Log.Info("DatabaseSecret generated")

	isDatabaseMarkedToBeDeleted := obj.GetDeletionTimestamp() != nil
	if isDatabaseMarkedToBeDeleted {
		if controllerutil.ContainsFinalizer(obj, stackv1alpha1.DatabaseFinalizer) {
			if err := r.finalizeDatabase(ctx, obj); err != nil {
				return ctrl.Result{}, err
			}
			controllerutil.RemoveFinalizer(obj, stackv1alpha1.DatabaseFinalizer)
			if err := r.Update(ctx, obj); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// Add finalizer for this CR
	if !controllerutil.ContainsFinalizer(obj, stackv1alpha1.DatabaseFinalizer) {
		controllerutil.AddFinalizer(obj, stackv1alpha1.DatabaseFinalizer)
		err = r.Update(ctx, obj)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	if obj.IsAvailable() {
		return ctrl.Result{}, nil
	}
	obj.SetStatusCondition(metav1.Condition{
		Type:               stackv1alpha1.ConditionTypeAvailable,
		Status:             metav1.ConditionTrue,
		Reason:             stackv1alpha1.ConditionReasonRunning,
		Message:            "DatabaseConnection is running",
		ObservedGeneration: obj.GetGeneration(),
	})

	if err := r.UpdateStatus(ctx, obj); err != nil {
		r.Log.Error(err, "Failed to update status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *DatabaseReconciler) UpdateStatus(ctx context.Context, instance *stackv1alpha1.Database) error {
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

// SetupWithManager sets up the controller with the Manager.
func (r *DatabaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&stackv1alpha1.Database{}).
		Complete(r)
}

func (r *DatabaseReconciler) getReferenceDatabaseConnection(ctx context.Context, obj *stackv1alpha1.Database) (*stackv1alpha1.DatabaseConnection, error) {
	if obj.Spec.Reference != "" {
		connection := &stackv1alpha1.DatabaseConnection{}
		err := r.Get(ctx, client.ObjectKey{Namespace: obj.Namespace, Name: obj.Spec.Reference}, connection)
		if err != nil {
			return nil, err
		}
		return connection, nil
	}
	return nil, errors.New("no reference found")
}

func (r *DatabaseReconciler) generateDatabaseSecret(ctx context.Context, obj *stackv1alpha1.Database, connection *stackv1alpha1.DatabaseConnection) error {
	// 如果可用,或者要删除了,就不再创建
	if obj.IsAvailable() || obj.GetDeletionTimestamp() != nil || controllerutil.ContainsFinalizer(obj, stackv1alpha1.DatabaseFinalizer) {
		return nil
	}
	dbName := obj.Spec.Name
	username := strings.ToLower(util.RemoveSpecialCharacter(obj.GetName() + util.GenerateRandomStr(5)))
	password := util.GenerateRandomStr(10)
	dsn, err := getDSNFromConnection(ctx, r.Client, connection)
	if err != nil {
		return err
	}
	initializer, err := NewDBInitializer(dsn)
	if err != nil {
		return err
	}
	// 初始化用户
	if err = initializer.initUser(username, password); err != nil {
		return err
	}
	// 初始化数据库
	if err = initializer.initDatabase(username, dbName); err != nil {
		return err
	}

	// 测试生成的数据库连通性
	newDsn := &DSN{
		Host:     dsn.Host,
		Port:     dsn.Port,
		Driver:   dsn.Driver,
		SSLMode:  dsn.SSLMode,
		Username: username,
		Password: password,
		Database: dbName,
	}

	dbInitializer, err := NewDBInitializer(newDsn)
	if err != nil {
		return err
	}
	// 检测连通性
	err = dbInitializer.ping()
	if err != nil {
		return err
	}
	// 创建secret
	data := make(map[string][]byte)
	data["username"] = []byte(username)
	data["password"] = []byte(password)

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      obj.GetNameWithSuffix("secret"),
			Namespace: obj.Namespace,
		},
		Type: corev1.SecretTypeOpaque,
		Data: data,
	}
	err = util.CreateOrUpdate(ctx, r.Client, secret)
	if err != nil {
		return err
	}
	return nil
}

func (r *DatabaseReconciler) finalizeDatabase(ctx context.Context, obj *stackv1alpha1.Database) error {

	// 1. 获取secret
	reference := obj.Spec.Reference
	connection := &stackv1alpha1.DatabaseConnection{}
	err := r.Get(ctx, client.ObjectKey{Namespace: obj.Namespace, Name: reference}, connection)
	if err != nil {
		return err
	}
	adminDSN, err := getDSNFromConnection(ctx, r.Client, connection)
	if err != nil {
		return err
	}

	admin, err := NewDBInitializer(adminDSN)
	if err != nil {
		return err
	}
	err = admin.dropDatabase(obj.Spec.Name)
	if err != nil {
		return err
	}

	username := obj.Spec.Credential.Username
	if obj.Spec.Credential.ExistSecret != "" {
		secret := &corev1.Secret{}
		err = r.Get(ctx, client.ObjectKey{Namespace: obj.Namespace, Name: obj.Spec.Credential.ExistSecret}, secret)
		if err != nil {
			return err
		}
		if ub, ok := secret.Data["username"]; ok {
			decodeString, err := base64.StdEncoding.DecodeString(string(ub))
			if err != nil {
				return err
			}
			username = string(decodeString)
		}

		err = r.Delete(ctx, secret)
		if err != nil {
			return err
		}
	}
	err = admin.dropUser(username)
	if err != nil {
		return err
	}
	return nil
}
