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

package v1

import (
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var egresslog = logf.Log.WithName("egress-resource")

func (r *Egress) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-egress-github-com-v1-egress,mutating=true,failurePolicy=fail,sideEffects=None,groups=egress.github.com,resources=egresses,verbs=create;update,versions=v1,name=megress.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Egress{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Egress) Default() {
	egresslog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-egress-github-com-v1-egress,mutating=false,failurePolicy=fail,sideEffects=None,groups=egress.github.com,resources=egresses,verbs=create;update;delete,versions=v1,name=vegress.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Egress{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Egress) ValidateCreate() error {
	egresslog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Egress) ValidateUpdate(old runtime.Object) error {
	egresslog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Egress) ValidateDelete() error {
	egresslog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
