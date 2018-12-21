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
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	dbv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/db/v1beta1"
	summonv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/summon/v1beta1"
	summoncomponents "github.com/Ridecell/ridecell-operator/pkg/controller/summon/components"
)

var _ = Describe("SummonPlatform PostgresExtensions Component", func() {
	It("creates all the extensions", func() {
		comp := summoncomponents.NewPostgresExtensions()
		_, err := comp.Reconcile(ctx)
		Expect(err).NotTo(HaveOccurred())

		ext := &dbv1beta1.PostgresExtension{}
		err = ctx.Client.Get(context.TODO(), types.NamespacedName{Name: "foo-postgis", Namespace: "default"}, ext)
		Expect(err).NotTo(HaveOccurred())
		Expect(ext.Spec.ExtensionName).To(Equal("postgis"))

		err = ctx.Client.Get(context.TODO(), types.NamespacedName{Name: "foo-postgis_topology", Namespace: "default"}, ext)
		Expect(err).NotTo(HaveOccurred())
		Expect(ext.Spec.ExtensionName).To(Equal("postgis_topology"))

		ctx.Client.Get(context.TODO(), types.NamespacedName{Name: "foo", Namespace: "default"}, instance)
		Expect(err).NotTo(HaveOccurred())
		Expect(instance.Status.Status).To(Equal(""))
		Expect(instance.Status.Message).To(Equal(""))
		Expect(instance.Status.PostgresExtensionStatus).To(Equal(""))
	})

	It("handles an error in the postgis extension", func() {
		ext := &dbv1beta1.PostgresExtension{
			ObjectMeta: metav1.ObjectMeta{Name: "foo-postgis", Namespace: "default"},
			Status: dbv1beta1.PostgresExtensionStatus{
				Status:  dbv1beta1.StatusError,
				Message: "Unable to floop the foobar",
			},
		}
		ctx.Client = fake.NewFakeClient(instance, ext)

		comp := summoncomponents.NewPostgresExtensions()
		_, err := comp.Reconcile(ctx)
		Expect(err).NotTo(HaveOccurred())

		ctx.Client.Get(context.TODO(), types.NamespacedName{Name: "foo", Namespace: "default"}, instance)
		Expect(err).NotTo(HaveOccurred())
		Expect(instance.Status.Status).To(Equal(summonv1beta1.StatusError))
		Expect(instance.Status.Message).To(Equal("postgis: Unable to floop the foobar"))
		Expect(instance.Status.PostgresExtensionStatus).To(Equal(summonv1beta1.StatusError))
	})

	It("handles an error in the postgis_topology extension", func() {
		ext := &dbv1beta1.PostgresExtension{
			ObjectMeta: metav1.ObjectMeta{Name: "foo-postgis_topology", Namespace: "default"},
			Status: dbv1beta1.PostgresExtensionStatus{
				Status:  dbv1beta1.StatusError,
				Message: "Unable to floop the other foobar",
			},
		}
		ctx.Client = fake.NewFakeClient(instance, ext)

		comp := summoncomponents.NewPostgresExtensions()
		_, err := comp.Reconcile(ctx)
		Expect(err).NotTo(HaveOccurred())

		ctx.Client.Get(context.TODO(), types.NamespacedName{Name: "foo", Namespace: "default"}, instance)
		Expect(err).NotTo(HaveOccurred())
		Expect(instance.Status.Status).To(Equal(summonv1beta1.StatusError))
		Expect(instance.Status.Message).To(Equal("postgis_topology: Unable to floop the other foobar"))
		Expect(instance.Status.PostgresExtensionStatus).To(Equal(summonv1beta1.StatusError))
	})

	It("handles one extension being ready", func() {
		ext := &dbv1beta1.PostgresExtension{
			ObjectMeta: metav1.ObjectMeta{Name: "foo-postgis", Namespace: "default"},
			Status: dbv1beta1.PostgresExtensionStatus{
				Status:  dbv1beta1.StatusReady,
				Message: "",
			},
		}
		ctx.Client = fake.NewFakeClient(instance, ext)

		comp := summoncomponents.NewPostgresExtensions()
		_, err := comp.Reconcile(ctx)
		Expect(err).NotTo(HaveOccurred())

		ctx.Client.Get(context.TODO(), types.NamespacedName{Name: "foo", Namespace: "default"}, instance)
		Expect(err).NotTo(HaveOccurred())
		Expect(instance.Status.Status).To(Equal(""))
		Expect(instance.Status.Message).To(Equal(""))
		Expect(instance.Status.PostgresExtensionStatus).To(Equal(""))
	})

	It("handles both extensions being ready", func() {
		ext := &dbv1beta1.PostgresExtension{
			ObjectMeta: metav1.ObjectMeta{Name: "foo-postgis", Namespace: "default"},
			Status: dbv1beta1.PostgresExtensionStatus{
				Status:  dbv1beta1.StatusReady,
				Message: "",
			},
		}
		ext2 := &dbv1beta1.PostgresExtension{
			ObjectMeta: metav1.ObjectMeta{Name: "foo-postgis_topology", Namespace: "default"},
			Status: dbv1beta1.PostgresExtensionStatus{
				Status:  dbv1beta1.StatusReady,
				Message: "",
			},
		}
		ctx.Client = fake.NewFakeClient(instance, ext, ext2)

		comp := summoncomponents.NewPostgresExtensions()
		_, err := comp.Reconcile(ctx)
		Expect(err).NotTo(HaveOccurred())

		ctx.Client.Get(context.TODO(), types.NamespacedName{Name: "foo", Namespace: "default"}, instance)
		Expect(err).NotTo(HaveOccurred())
		Expect(instance.Status.Status).To(Equal(""))
		Expect(instance.Status.Message).To(Equal(""))
		Expect(instance.Status.PostgresExtensionStatus).To(Equal(summonv1beta1.StatusReady))
	})
})
