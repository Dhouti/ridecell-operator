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

package components_test

import (
	"context"

	. "github.com/Ridecell/ridecell-operator/pkg/test_helpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/kms/kmsiface"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"

	encryptedsecretcomponents "github.com/Ridecell/ridecell-operator/pkg/controller/encryptedsecret/components"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type mockKMSClient struct {
	kmsiface.KMSAPI
}

var _ = Describe("encryptedsecret Component", func() {

	It("runs basic reconcile", func() {
		comp := encryptedsecretcomponents.NewEncryptedSecret()
		mockKMS := &mockKMSClient{}
		comp.InjectKMSAPI(mockKMS)

		instance.Data = map[string]string{
			"TEST_VALUE0": "test0",
			"TEST_VALUE1": "test1",
			"TEST_VALUE2": "test2",
			"test_value3": "TEST3",
		}

		Expect(comp).To(ReconcileContext(ctx))

		fetchSecret := &corev1.Secret{}
		err := ctx.Client.Get(ctx.Context, types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, fetchSecret)
		Expect(err).ToNot(HaveOccurred())

		Expect(string(fetchSecret.Data["TEST_VALUE0"])).To(Equal("test0"))
		Expect(string(fetchSecret.Data["TEST_VALUE1"])).To(Equal("test1"))
		Expect(string(fetchSecret.Data["TEST_VALUE2"])).To(Equal("test2"))
		Expect(string(fetchSecret.Data["test_value3"])).To(Equal("TEST3"))
	})

	It("updates an existing secret", func() {
		comp := encryptedsecretcomponents.NewEncryptedSecret()
		mockKMS := &mockKMSClient{}
		comp.InjectKMSAPI(mockKMS)

		newSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      instance.Name,
				Namespace: instance.Namespace,
			},
			Data: map[string][]byte{
				"test_value": []byte("test"),
			},
		}
		err := ctx.Create(context.TODO(), newSecret)
		Expect(err).ToNot(HaveOccurred())
		// Overwrite that secret with new one
		instance.Data = map[string]string{"new_value": "test1"}
		Expect(comp).To(ReconcileContext(ctx))

		fetchSecret := &corev1.Secret{}
		err = ctx.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, fetchSecret)
		Expect(err).ToNot(HaveOccurred())

		_, ok := fetchSecret.Data["test_value"]
		Expect(ok).To(Equal(false))

		val, ok := fetchSecret.Data["new_value"]
		Expect(ok).To(Equal(true))
		Expect(string(val)).To(Equal("test1"))
	})

})

func (m *mockKMSClient) Decrypt(input *kms.DecryptInput) (*kms.DecryptOutput, error) {
	if len(input.CiphertextBlob) < 0 {
		return &kms.DecryptOutput{}, awserr.New(kms.ErrCodeInvalidCiphertextException, "awsmock_decrypt: Invalid cipher text", errors.New(""))
	}
	return &kms.DecryptOutput{Plaintext: input.CiphertextBlob}, nil
}
