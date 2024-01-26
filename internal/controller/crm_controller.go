/*
Copyright 2024.

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
	"github.com/LL-res/CRM/common/consts"
	"github.com/LL-res/CRM/common/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"

	elasticv1 "github.com/LL-res/CRM/api/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	defaultErrorRetryPeriod = 10 * time.Second
)

// CRMReconciler reconciles a CRM object
type CRMReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=elastic.github.com,resources=crms,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=elastic.github.com,resources=crms/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=elastic.github.com,resources=crms/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the CRM object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *CRMReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.Logger.WithName("reconcile")

	instance := &elasticv1.CRM{}
	err := r.Client.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			logger.Info("instance deleted")
			return reconcile.Result{}, nil
		}
		logger.Error(err, "failed to get instance")
		return reconcile.Result{RequeueAfter: defaultErrorRetryPeriod}, err
	}
	ctx = context.WithValue(ctx, consts.NAMESPACE, req.Namespace)
	ctx = context.WithValue(ctx, consts.NAME, req.Name)
	controller := NewController(instance, r)

	if err := controller.Handle(ctx); err != nil {
		return reconcile.Result{RequeueAfter: defaultErrorRetryPeriod}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CRMReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&elasticv1.CRM{}).
		Complete(r)
}
