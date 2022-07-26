/*
Copyright 2022.

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

package controllers

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	egressv1 "github.com/sriramy/calico-egress/pkg/api/v1"
)

// EgressReconciler reconciles a Egress object
type EgressReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=egress.github.com,resources=egresses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=egress.github.com,resources=egresses/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=egress.github.com,resources=egresses/finalizers,verbs=update
//+kubebuilder:rbac:groups=egress.github.com,resources=pods,verbs=get;list;watch
//+kubebuilder:rbac:groups=egress.github.com,resources=pods/status,verbs=get
//+kubebuilder:rbac:groups=egress.github.com,resources=namespaces,verbs=get;list;watch
//+kubebuilder:rbac:groups=egress.github.com,resources=namespaces/status,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Egress object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.1/pkg/reconcile
func (r *EgressReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	log.Info("Reconciler", "try", req.NamespacedName)

	egressObject := &egressv1.Egress{}
	err := r.Get(ctx, req.NamespacedName, egressObject)
	if errors.IsNotFound(err) {
		log.Info("Reconciler", "skipping", "Cannot find egress object")
		return ctrl.Result{}, nil
	}
	if err != nil {
		return ctrl.Result{}, err
	}
	podSelector, err := metav1.LabelSelectorAsSelector(egressObject.Spec.PodSelector)
	if err != nil {
		return ctrl.Result{}, err
	}

	nsObject := &corev1.Namespace{}
	err = r.Get(ctx, types.NamespacedName{Name: req.Namespace}, nsObject)
	if err != nil {
		return ctrl.Result{}, err
	}

	nsSelector, err := metav1.LabelSelectorAsSelector(egressObject.Spec.NamespaceSelector)
	if err != nil {
		return ctrl.Result{}, err
	}
	if !nsSelector.Empty() && !nsSelector.Matches(labels.Set(nsObject.Labels)) {
		log.Info("Reconciler", "skipping", "Egress object doesn't match the namespace selector")
		return ctrl.Result{}, nil
	}

	pods := &corev1.PodList{}
	err = r.List(ctx, pods, client.InNamespace(req.Namespace), client.MatchingLabelsSelector{Selector: podSelector})
	if err != nil {
		return ctrl.Result{}, err
	}

	log.Info("Reconciler", "pods found", len(pods.Items))

	if len(pods.Items) > 0 {
		if egressObject.Status.PodList == nil {
			egressObject.Status.PodList = make([]string, 1)
		}
		egressObject.Status.PodList[0] = pods.Items[0].Name + ":" + pods.Items[0].Name
		r.Update(ctx, egressObject)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *EgressReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&egressv1.Egress{}).
		Complete(r)
}
