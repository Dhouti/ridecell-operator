/*
Copyright 2018 Ridecell, Inc.

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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	ducomponents "github.com/Ridecell/ridecell-operator/pkg/controller/djangouser/components"
	. "github.com/Ridecell/ridecell-operator/pkg/test_helpers/matchers"
)

var _ = Describe("DjangoUser Defaults Component", func() {
	It("does nothing on a filled out object", func() {
		instance.Spec.Email = "foo@bar.com"
		instance.Spec.PasswordSecret = "foo-credentials"

		comp := ducomponents.NewDefaults()
		Expect(comp).To(ReconcileContext(ctx))
		Expect(instance.Spec.Email).To(Equal("foo@bar.com"))
		Expect(instance.Spec.PasswordSecret).To(Equal("foo-credentials"))
	})

	It("sets a default password secret", func() {
		instance.Spec.Email = "foo@bar.com"

		comp := ducomponents.NewDefaults()
		Expect(comp).To(ReconcileContext(ctx))
		Expect(instance.Spec.Email).To(Equal("foo@bar.com"))
		Expect(instance.Spec.PasswordSecret).To(Equal("foo.bar.com-credentials"))
	})

	It("sets a default email", func() {
		instance.Spec.PasswordSecret = "foo-credentials"

		comp := ducomponents.NewDefaults()
		Expect(comp).To(ReconcileContext(ctx))
		Expect(instance.Spec.Email).To(Equal("foo@example.com"))
		Expect(instance.Spec.PasswordSecret).To(Equal("foo-credentials"))
	})

	It("sets a default email and password secret", func() {
		comp := ducomponents.NewDefaults()
		Expect(comp).To(ReconcileContext(ctx))
		Expect(instance.Spec.Email).To(Equal("foo@example.com"))
		Expect(instance.Spec.PasswordSecret).To(Equal("foo.example.com-credentials"))
	})
})
