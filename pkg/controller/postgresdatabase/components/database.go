/*
Copyright 2019 Ridecell, Inc.

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
	"fmt"

	"github.com/lib/pq"
	"k8s.io/apimachinery/pkg/runtime"

	dbv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/db/v1beta1"
	"github.com/Ridecell/ridecell-operator/pkg/components"
	"github.com/Ridecell/ridecell-operator/pkg/components/postgres"
	"github.com/Ridecell/ridecell-operator/pkg/errors"
)

type databaseComponent struct{}

func NewDatabase() *databaseComponent {
	return &databaseComponent{}
}

func (_ *databaseComponent) WatchTypes() []runtime.Object {
	return []runtime.Object{}
}

func (_ *databaseComponent) IsReconcilable(ctx *components.ComponentContext) bool {
	instance := ctx.Top.(*dbv1beta1.PostgresDatabase)
	return instance.Status.DatabaseStatus == dbv1beta1.StatusReady && (instance.Spec.SkipUser || instance.Status.UserStatus == dbv1beta1.StatusReady)
}

func (comp *databaseComponent) Reconcile(ctx *components.ComponentContext) (components.Result, error) {
	instance := ctx.Top.(*dbv1beta1.PostgresDatabase)

	conn := instance.Status.Connection.DeepCopy()
	conn.Database = "postgres"
	db, err := postgres.Open(ctx, conn)
	if err != nil {
		return components.Result{}, err
	}

	row := db.QueryRow(`SELECT COUNT(*) FROM pg_catalog.pg_database WHERE datname = $1`, instance.Spec.DatabaseName)
	var count int
	err = row.Scan(&count)
	if err != nil {
		return components.Result{}, errors.Wrap(err, "database: error running db check query")
	}

	if count == 0 {
		// Got to make the database.
		_, err := db.Exec(fmt.Sprintf(`CREATE DATABASE %s WITH OWNER = $1`, pq.QuoteIdentifier(instance.Spec.DatabaseName)), instance.Spec.Owner)
		if err != nil {
			return components.Result{}, errors.Wrap(err, "database: error creating database")
		}
	}

	dbName := instance.Spec.DatabaseName
	return components.Result{StatusModifier: func(obj runtime.Object) error {
		instance := obj.(*dbv1beta1.PostgresDatabase)
		instance.Status.Status = dbv1beta1.StatusReady
		instance.Status.Message = fmt.Sprintf("Created database %s", dbName)
		return nil
	}}, nil
}
