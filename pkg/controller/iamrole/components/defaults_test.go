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
	"os"

	. "github.com/Ridecell/ridecell-operator/pkg/test_helpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	iamrolecomponents "github.com/Ridecell/ridecell-operator/pkg/controller/iamrole/components"
)

var _ = Describe("iamrole Defaults Component", func() {
	BeforeEach(func() {
		os.Setenv("DEFAULT_PERMISSIONS_BOUNDARY_ARN", "defaults-test")
	})

	It("does nothing on a filled out object", func() {
		comp := iamrolecomponents.NewDefaults()
		instance.Spec.RoleName = "test"
		instance.Spec.PermissionsBoundaryArn = "permboundary"

		Expect(comp).To(ReconcileContext(ctx))
		Expect(instance.Spec.RoleName).To(Equal("test"))
		Expect(instance.Spec.PermissionsBoundaryArn).To(Equal("permboundary"))

	})

	It("sets defaults", func() {
		comp := iamrolecomponents.NewDefaults()
		Expect(comp).To(ReconcileContext(ctx))

		Expect(instance.Spec.RoleName).To(Equal("test-role"))
		Expect(instance.Spec.PermissionsBoundaryArn).To(Equal("defaults-test"))
	})

})
