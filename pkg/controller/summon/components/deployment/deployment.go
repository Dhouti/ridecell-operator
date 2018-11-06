/*
Copyright 2018 Ridecell, Inc..

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

package deployment

import (
	postgresv1 "github.com/zalando-incubator/postgres-operator/pkg/apis/acid.zalan.do/v1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	summonv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/summon/v1beta1"
	"github.com/Ridecell/ridecell-operator/pkg/components"
)

type deploymentComponent struct {
	templatePath    string
	waitForDatabase bool
}

func New(templatePath string, waitForDatabase bool) *deploymentComponent {
	return &deploymentComponent{templatePath: templatePath, waitForDatabase: waitForDatabase}
}

func (comp *deploymentComponent) WatchTypes() []runtime.Object {
	return []runtime.Object{
		&appsv1.Deployment{},
	}
}

func (comp *deploymentComponent) IsReconcilable(ctx *components.ComponentContext) bool {
	instance := ctx.Top.(*summonv1beta1.SummonPlatform)
	// Check on the pull secret. Not technically needed in some cases, but just wait.
	if instance.Status.PullSecretStatus != summonv1beta1.StatusReady {
		return false
	}
	// If we need the database, make sure that exists. Otherwise, always ready.
	if comp.waitForDatabase {
		return instance.Status.PostgresStatus != nil && *instance.Status.PostgresStatus == postgresv1.ClusterStatusRunning && instance.Spec.Version == instance.Status.MigrateVersion
	} else {
		return true
	}
}

func (comp *deploymentComponent) Reconcile(ctx *components.ComponentContext) (reconcile.Result, error) {
	res, _, err := ctx.CreateOrUpdate(comp.templatePath, func(goalObj, existingObj runtime.Object) error {
		goal := goalObj.(*appsv1.Deployment)
		existing := existingObj.(*appsv1.Deployment)
		// Copy the Spec over.
		existing.Spec = goal.Spec
		return nil
	})
	return res, err
}
