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

	egressv1 "github.com/sriramy/calico-egress/pkg/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// PodReconciler reconciles a Egress object
type PodReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	reconciler *EgressReconciler
	mgr        ctrl.Manager
	stopCh     chan struct{}
}

func NewPodReconciler(mgr ctrl.Manager, egressReconciler *EgressReconciler) *PodReconciler {
	return &PodReconciler{
		Client:     mgr.GetClient(),
		Scheme:     mgr.GetScheme(),
		reconciler: egressReconciler,
		mgr:        mgr,
		stopCh:     make(chan struct{}),
	}
}

//+kubebuilder:rbac:groups=egress.github.com,resources=pods,verbs=get;list;watch;update
//+kubebuilder:rbac:groups=egress.github.com,resources=pods/status,verbs=get

// Start sets up the controller with the Manager.
func (p *PodReconciler) Start(egress *egressv1.Egress) error {
	nsSelectorPredicate := predicate.NewPredicateFuncs(func(o client.Object) bool {
		return egress.GetNamespace() == o.GetNamespace()
	})

	labelSelectorPredicate, err := predicate.LabelSelectorPredicate(*egress.Spec.PodSelector)
	if err != nil {
		return err
	}
	controller, err := controller.NewUnmanaged(egress.GetName()+"-reconciler", p.mgr,
		controller.Options{
			Reconciler: p.reconciler,
		})
	if err != nil {
		return err
	}

	controller.Watch(&source.Kind{Type: &corev1.Pod{}},
		&EnqueueRequestForEgress{Name: egress.GetName(), Namespace: egress.GetNamespace()},
		predicate.And(labelSelectorPredicate, nsSelectorPredicate))

	return controller.Start(p.getContext())
}

func (p *PodReconciler) Stop() {
	close(p.stopCh)
}

// getContext returns
func (p *PodReconciler) getContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		<-p.stopCh
		cancel()
	}()

	return ctx
}
