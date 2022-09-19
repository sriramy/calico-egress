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
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/projectcalico/api/pkg/client/clientset_generated/clientset"

	egressv1 "github.com/sriramy/calico-egress/pkg/api/v1"
)

// EgressReconciler reconciles a Egress object
type EgressReconciler struct {
	client.Client
	Scheme       *runtime.Scheme
	ctx          context.Context
	mgr          ctrl.Manager
	calicoClient *clientset.Clientset
}

func NewEgressReconciler(ctx context.Context, calicoClient *clientset.Clientset,
	mgr ctrl.Manager) *EgressReconciler {
	return &EgressReconciler{
		Client:       mgr.GetClient(),
		Scheme:       mgr.GetScheme(),
		ctx:          ctx,
		calicoClient: calicoClient,
		mgr:          mgr,
	}
}

//+kubebuilder:rbac:groups="egress.github.com",resources=egresses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="egress.github.com",resources=egresses/status,verbs=get;update;patch
//+kubebuilder:rbac:groups="egress.github.com",resources=egresses/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;update;patch;create;delete
//+kubebuilder:rbac:groups="",resources=pods/status,verbs=get

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

	log.Info("EgressReconciler", "Name", req.NamespacedName)

	egress := &egressv1.Egress{}
	err := r.Get(ctx, req.NamespacedName, egress)
	if errors.IsNotFound(err) || !egress.GetDeletionTimestamp().IsZero() {
		log.Info("Reconciler", "skipping", "Cannot find egress object")
		return ctrl.Result{}, nil
	} else if err != nil {
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
	err = r.setupEgress(ctx, req.NamespacedName, egress, pods.Items)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *EgressReconciler) setupEgress(ctx context.Context, name types.NamespacedName, egress *egressv1.Egress, pods []corev1.Pod) error {
	var endpoints []egressv1.Endpoint
	for _, pod := range pods {
		endpoints = append(endpoints,
			egressv1.Endpoint{Name: pod.Name, IP: pod.Status.PodIP})

		list, err := r.calicoClient.ProjectcalicoV3().GlobalNetworkPolicies().List(context.Background(), metav1.ListOptions{})
		if err != nil {
			if !errors.IsNotFound(err) {
				panic(err)
			}
		}
		for _, gnp := range list.Items {
			fmt.Printf("%#v\n", gnp)
		}

		// if pod.Annotations == nil {
		// 	pod.Annotations = make(map[string]string)
		// }
		// pod.Annotations["egress.github.com/egressIP"] = egress.Spec.EgressIP
		// r.Update(ctx, &pod)
	}
	egress.Status.Endpoints = endpoints
	return r.Status().Update(ctx, egress)
}

// SetupWithManager sets up the controller with the Manager.
func (r *EgressReconciler) SetupWithManager() error {
	return ctrl.NewControllerManagedBy(r.mgr).
		For(&egressv1.Egress{}).
		Watches(&source.Kind{Type: &corev1.Pod{}}, handler.EnqueueRequestsFromMapFunc(r.getEgresses)).
		Complete(r)
}

func (r *EgressReconciler) getEgresses(pod client.Object) []reconcile.Request {
	egresses := &egressv1.EgressList{}
	listOps := &client.ListOptions{
		Namespace: pod.GetNamespace(),
	}
	err := r.List(context.TODO(), egresses, listOps)
	if err != nil {
		return []reconcile.Request{}
	}

	requests := make([]reconcile.Request, len(egresses.Items))
	for i, item := range egresses.Items {
		podSelector, _ := metav1.LabelSelectorAsSelector(item.Spec.PodSelector)
		if podSelector.Matches(labels.Set(pod.GetLabels())) {
			requests[i] = reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      item.GetName(),
					Namespace: item.GetNamespace(),
				},
			}
		}
	}
	return requests
}
