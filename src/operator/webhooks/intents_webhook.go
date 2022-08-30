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

package webhooks

import (
	"context"
	"fmt"
	"github.com/otterize/intents-operator/operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

type IntentsValidator struct {
	client.Client
}

func (v *IntentsValidator) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&v1alpha1.Intents{}).
		WithValidator(v).
		Complete()
}

func NewIntentsValidator(c client.Client) *IntentsValidator {
	return &IntentsValidator{
		Client: c,
	}
}

//+kubebuilder:webhook:path=/validate-otterize-com-v1alpha1-intents,mutating=false,failurePolicy=fail,sideEffects=None,groups=otterize.com,resources=intents,verbs=create;update,versions=v1alpha1,name=vintents.kb.io,admissionReviewVersions=v1

var _ webhook.CustomValidator = &IntentsValidator{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (v *IntentsValidator) ValidateCreate(ctx context.Context, obj runtime.Object) error {
	intentsObj := obj.(*v1alpha1.Intents)
	return v.validateNoDuplicateClients(ctx, intentsObj)

}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (v *IntentsValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) error {
	intentsObj := oldObj.(*v1alpha1.Intents)
	return v.validateNoDuplicateClients(ctx, intentsObj)

}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (v *IntentsValidator) ValidateDelete(ctx context.Context, obj runtime.Object) error {
	return nil
}

func (v *IntentsValidator) validateNoDuplicateClients(ctx context.Context, intentsObj *v1alpha1.Intents) error {
	intentsList := &v1alpha1.IntentsList{}
	if err := v.List(ctx, intentsList, &client.ListOptions{Namespace: intentsObj.Namespace}); err != nil {
		return err
	}
	desiredClientName := intentsObj.GetServiceName()
	for _, intent := range intentsList.Items {
		// Intents for this client already exists, deny admission
		if intent.GetServiceName() == desiredClientName {
			// This is relevant for update - skip the comparison if it's the same object
			if intent.Name == intentsObj.Name {
				continue
			}

			gvk := intentsObj.GroupVersionKind()
			return errors.NewInvalid(
				schema.GroupKind{Group: gvk.Group, Kind: gvk.Kind},
				intentsObj.Name,
				field.ErrorList{
					&field.Error{
						Type:     field.ErrorTypeDuplicate,
						Field:    "name",
						BadValue: desiredClientName,
						Detail: fmt.Sprintf(
							"Intents for client %s already exist in resource %s", desiredClientName, intent.Name),
					},
				})
		}
	}

	return nil
}
