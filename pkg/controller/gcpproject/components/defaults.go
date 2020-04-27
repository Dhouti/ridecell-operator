/*
Copyright 2020 Ridecell, Inc.

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
	"os"

	"github.com/Ridecell/ridecell-operator/pkg/components"
	"github.com/Ridecell/ridecell-operator/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"

	gcpv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/gcp/v1beta1"
)

type defaultsComponent struct {
}

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
	instance := ctx.Top.(*gcpv1beta1.GCPProject)

	// Fill in defaults.
	if instance.Spec.EnableFirebase == nil {
		enableFirebaseDefault := false
		instance.Spec.EnableFirebase = &enableFirebaseDefault
	}

	if instance.Spec.EnableBilling == nil {
		enableBillingDefault := false
		instance.Spec.EnableBilling = &enableBillingDefault
	}

	if instance.Spec.EnableRealtimeDatabase == nil {
		enableRealtimeDatabaseDefault := false
		instance.Spec.EnableRealtimeDatabase = &enableRealtimeDatabaseDefault
	}

	if instance.Spec.RealtimeDatabaseRules == "" {
		defaultRules := os.Getenv("FIREBASE_DATABASE_DEFAULT_RULES")
		if defaultRules == "" {
			return components.Result{}, errors.New("gcpproject: FIREBASE_DATABASE_DEFAULT_RULES is not set")
		}
		instance.Spec.RealtimeDatabaseRules = defaultRules
	}

	return components.Result{}, nil
}
