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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

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
//+kubebuilder:rbac:groups=egress.github.com,resources=pods,verbs=get;list;watch;update
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

	egress := &egressv1.Egress{}
	err := r.Get(ctx, req.NamespacedName, egress)
	if errors.IsNotFound(err) {
		log.Info("Reconciler", "skipping", "Cannot find egress object")
		return ctrl.Result{}, nil
	}
	if err != nil {
		return ctrl.Result{}, err
	}
	podSelector, err := metav1.LabelSelectorAsSelector(egress.Spec.PodSelector)
	if err != nil {
		return ctrl.Result{}, err
	}

	pods := &corev1.PodList{}
	err = r.List(ctx, pods, client.InNamespace(req.Namespace), client.MatchingLabelsSelector{Selector: podSelector})
	if err != nil {
		return ctrl.Result{}, err
	}

	if egress.Status.Endpoints != nil {
		egress.Status.Endpoints = nil
	}

	for _, pod := range pods.Items {
		egress.Status.Endpoints = append(egress.Status.Endpoints,
			egressv1.Endpoint{Name: pod.Name, IP: pod.Status.PodIP})
		pod.Annotations["egress.github.com/egressIP"] = egress.Spec.EgressIP
		r.Update(ctx, &pod)
	}
	r.Status().Update(ctx, egress)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *EgressReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&egressv1.Egress{}).
		Watches(
			&source.Kind{Type: &corev1.Pod{}},
			handler.EnqueueRequestsFromMapFunc(r.Map)).
		Complete(r)
}

// Map the passed in pod object to any matching Egress objects
func (r *EgressReconciler) Map(obj client.Object) []reconcile.Request {
	requests := make([]reconcile.Request, 0)

	for _, egress := range r.findMatchingEgresses(obj.GetNamespace(), obj.GetLabels()) {
		requests = append(requests, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      egress.Name,
				Namespace: egress.Namespace,
			},
		})
	}
	return requests
}

func (r *EgressReconciler) findMatchingEgresses(namespace string, labels map[string]string) []egressv1.Egress {
	// Find any Egress objects that match this pod.
	allEgresses := &egressv1.EgressList{}
	err := r.List(context.TODO(), allEgresses, client.InNamespace(namespace))
	if err != nil {
		logf.Log.Info("cannot find any egresses")
		return nil
	}

	egresses := make([]egressv1.Egress, 0)
	for _, egress := range allEgresses.Items {
		podSelector, err := metav1.LabelSelectorAsMap(egress.Spec.PodSelector)
		if err != nil {
			logf.Log.Error(err, "cannot get pod selector")
			return nil
		}
		for key, selector := range podSelector {
			if label, found := labels[key]; found && label == selector {
				egresses = append(egresses, egress)
				break
			}
		}
	}
	return egresses
}
