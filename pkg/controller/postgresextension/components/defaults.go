/*
Copyright 2018-2019 Ridecell, Inc.

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

package components

import (
	"k8s.io/apimachinery/pkg/runtime"

	dbv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/db/v1beta1"
	"github.com/Ridecell/ridecell-operator/pkg/components"
)

type defaultsComponent struct{}

func NewDefaults() *defaultsComponent {
	return &defaultsComponent{}
}

func (_ *defaultsComponent) WatchTypes() []runtime.Object {
	return []runtime.Object{}
}

func (_ *defaultsComponent) IsReconcilable(_ *components.ComponentContext) bool {
	return true
}

func (comp *defaultsComponent) Reconcile(ctx *components.ComponentContext) (components.Result, error) {
	instance := ctx.Top.(*dbv1beta1.PostgresExtension)

	// Fill in defaults.
	if instance.Spec.ExtensionName == "" {
		// Default extension name is just the name of the resource.
		instance.Spec.ExtensionName = instance.Name
	}
	if instance.Spec.Database.Port == 0 {
		// Default Postgres port.
		instance.Spec.Database.Port = 5432
	}
	if instance.Spec.Database.PasswordSecretRef.Key == "" {
		// Use "password" as the default key.
		instance.Spec.Database.PasswordSecretRef.Key = "password"
	}

	return components.Result{}, nil
}
