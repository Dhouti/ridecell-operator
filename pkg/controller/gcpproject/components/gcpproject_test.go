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

package components_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/Ridecell/ridecell-operator/pkg/components"
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/googleapi"

	gppcomponents "github.com/Ridecell/ridecell-operator/pkg/controller/gcpproject/components"
	. "github.com/Ridecell/ridecell-operator/pkg/test_helpers/matchers"
)

var _ = Describe("gcpproject gcpproject Component", func() {
	comp := gppcomponents.NewGCPProject()
	var crmmock *gppcomponents.GCPCloudResourceManagerMock
	BeforeEach(func() {
		comp = gppcomponents.NewGCPProject()
		crmmock = &gppcomponents.GCPCloudResourceManagerMock{
			CreateFunc: func(_ *components.ComponentContext, _ string) (*cloudresourcemanager.Operation, error) {
				return &cloudresourcemanager.Operation{}, nil
			},
			GetFunc: func(_ string) (*cloudresourcemanager.Project, error) {
				return &cloudresourcemanager.Project{}, nil
			},
			GetOperationFunc: func(_ string) (*cloudresourcemanager.Operation, error) {
				return &cloudresourcemanager.Operation{}, nil
			},
		}

		comp.InjectCRM(crmmock)
	})

	Describe("IsReconcilable", func() {
		It("returns true", func() {
			Expect(comp.IsReconcilable(ctx)).To(BeTrue())
		})
	})

	It("does nothing if the project already exists", func() {
		Expect(comp).To(ReconcileContext(ctx))
		Expect(crmmock.CreateCalls()).To(HaveLen(0))
	})

	It("creates the project if it doesn't exist", func() {
		crmmock.CreateFunc = func(_ *components.ComponentContext, _ string) (*cloudresourcemanager.Operation, error) {
			return &cloudresourcemanager.Operation{Done: false, Name: ""}, nil
		}
		crmmock.GetFunc = func(_ string) (*cloudresourcemanager.Project, error) {
			return nil, &googleapi.Error{Code: 404}
		}

		res, err := comp.Reconcile(ctx)
		Expect(err).ToNot(HaveOccurred())
		Expect(res.RequeueAfter).To(Equal(time.Minute))
		Expect(crmmock.GetCalls()).To(HaveLen(1))
		Expect(crmmock.GetOperationCalls()).To(HaveLen(0))
		Expect(crmmock.CreateCalls()).To(HaveLen(1))

		// Set status to an expected name
		instance.Status.ProjectOperationName = "creating-a-project"

		// Run again while operation not done, make sure nothing changes.
		res, err = comp.Reconcile(ctx)
		Expect(err).ToNot(HaveOccurred())
		Expect(res.RequeueAfter).To(Equal(time.Minute))
		Expect(crmmock.GetCalls()).To(HaveLen(2))
		Expect(crmmock.GetOperationCalls()).To(HaveLen(1))
		Expect(crmmock.CreateCalls()).To(HaveLen(1))

		// Mark operation as done
		crmmock.GetOperationFunc = func(operationName string) (*cloudresourcemanager.Operation, error) {
			if operationName == "creating-a-project" {
				return &cloudresourcemanager.Operation{Done: true}, nil
			}
			return nil, &googleapi.Error{Code: 404}
		}
		res, err = comp.Reconcile(ctx)
		Expect(err).ToNot(HaveOccurred())
		Expect(res.RequeueAfter).To(Equal(time.Second * 0))
		Expect(crmmock.GetCalls()).To(HaveLen(3))
		Expect(crmmock.GetOperationCalls()).To(HaveLen(2))
		Expect(crmmock.CreateCalls()).To(HaveLen(1))
	})
})
