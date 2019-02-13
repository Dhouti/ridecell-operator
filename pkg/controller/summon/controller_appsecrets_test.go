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

package summon_test

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	postgresv1 "github.com/zalando-incubator/postgres-operator/pkg/apis/acid.zalan.do/v1"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	summonv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/summon/v1beta1"
	"github.com/Ridecell/ridecell-operator/pkg/test_helpers"
)

var _ = Describe("Summon controller appsecrets", func() {
	var helpers *test_helpers.PerTestHelpers
	var instance *summonv1beta1.SummonPlatform

	BeforeEach(func() {
		helpers = testHelpers.SetupTest()

		// Set up the instance object for other tests.
		instance = &summonv1beta1.SummonPlatform{
			ObjectMeta: metav1.ObjectMeta{Name: "appsecretstest", Namespace: helpers.Namespace},
			Spec: summonv1beta1.SummonPlatformSpec{
				Version: "80813-eb6b515-master",
				Secrets: []string{},
				Database: summonv1beta1.DatabaseSpec{
					ExclusiveDatabase: true,
				},
			},
		}
	})

	AfterEach(func() {
		// Display some debugging info if the test failed.
		if CurrentGinkgoTestDescription().Failed {
			summons := &summonv1beta1.SummonPlatformList{}
			err := helpers.Client.List(context.Background(), nil, summons)
			if err != nil {
				fmt.Printf("!!!!!! %s\n", err)
			} else {
				fmt.Print("Failed instances:\n")
				for _, item := range summons.Items {
					if item.Namespace == helpers.Namespace {
						fmt.Printf("\t%s %#v\n", item.Name, item.Status)
					}
				}
			}
		}

		helpers.TeardownTest()
	})

	It("creates the app secret if all inputs exist already", func() {
		c := helpers.TestClient

		// Create all the input secrets.
		inputSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "testsecret", Namespace: helpers.Namespace},
			Data: map[string][]byte{
				"filler": []byte{}}}
		c.Create(inputSecret)
		awsSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "appsecretstest.aws-credentials", Namespace: helpers.Namespace},
			StringData: map[string]string{
				"AWS_ACCESS_KEY_ID":     "AKIAtest",
				"AWS_SECRET_ACCESS_KEY": "test",
			},
		}
		c.Create(awsSecret)
		dbSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "summon.appsecretstest-database.credentials", Namespace: helpers.Namespace},
			StringData: map[string]string{
				"password": "secretdbpass",
			},
		}
		c.Create(dbSecret)

		// Create the instance.
		instance.Spec.Secrets = []string{"testsecret"}
		c.Create(instance)

		// Advance postgres to running.
		postgres := &postgresv1.Postgresql{}
		c.EventuallyGet(helpers.Name("appsecretstest-database"), postgres)
		postgres.Status = postgresv1.ClusterStatusRunning
		c.Status().Update(postgres)

		// Get the output app secrets.
		appSecret := &corev1.Secret{}
		c.EventuallyGet(helpers.Name("appsecretstest.app-secrets"), appSecret)

		// Parse the YAML to check it.
		data := map[string]interface{}{}
		err := yaml.Unmarshal(appSecret.Data["summon-platform.yml"], &data)
		Expect(err).ToNot(HaveOccurred())
		Expect(data["DATABASE_URL"]).To(Equal("postgis://summon:secretdbpass@appsecretstest-database/summon"))
	})

	It("creates the app secret the database secret is created afterwards", func() {
		c := helpers.TestClient

		// Create some of the input secrets.
		inputSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "testsecret", Namespace: helpers.Namespace},
			Data: map[string][]byte{
				"filler": []byte{}}}
		c.Create(inputSecret)
		awsSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "appsecretstest.aws-credentials", Namespace: helpers.Namespace},
			StringData: map[string]string{
				"AWS_ACCESS_KEY_ID":     "AKIAtest",
				"AWS_SECRET_ACCESS_KEY": "test",
			},
		}
		c.Create(awsSecret)

		// Create the instance.
		instance.Spec.Secrets = []string{"testsecret"}
		c.Create(instance)

		// Advance postgres to running.
		postgres := &postgresv1.Postgresql{}
		c.EventuallyGet(helpers.Name("appsecretstest-database"), postgres)
		postgres.Status = postgresv1.ClusterStatusRunning
		c.Status().Update(postgres)

		// Create the DB secret later than where it would normally be created.
		time.Sleep(2 * time.Second)
		dbSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "summon.appsecretstest-database.credentials", Namespace: helpers.Namespace},
			StringData: map[string]string{
				"password": "secretdbpass",
			},
		}
		c.Create(dbSecret)

		// Get the output app secrets.
		appSecret := &corev1.Secret{}
		c.EventuallyGet(helpers.Name("appsecretstest.app-secrets"), appSecret)

		// Parse the YAML to check it.
		data := map[string]interface{}{}
		err := yaml.Unmarshal(appSecret.Data["summon-platform.yml"], &data)
		Expect(err).ToNot(HaveOccurred())
		Expect(data["DATABASE_URL"]).To(Equal("postgis://summon:secretdbpass@appsecretstest-database/summon"))
	})

	It("creates the app secret the database secret is changed afterwards", func() {
		c := helpers.TestClient

		// Create some of the input secrets.
		inputSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "testsecret", Namespace: helpers.Namespace},
			Data: map[string][]byte{
				"filler": []byte{}}}
		c.Create(inputSecret)
		dbSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "summon.appsecretstest-database.credentials", Namespace: helpers.Namespace},
			StringData: map[string]string{
				"password": "secretdbpass",
			},
		}
		c.Create(dbSecret)
		awsSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "appsecretstest.aws-credentials", Namespace: helpers.Namespace},
			StringData: map[string]string{
				"AWS_ACCESS_KEY_ID":     "AKIAtest",
				"AWS_SECRET_ACCESS_KEY": "test",
			},
		}
		c.Create(awsSecret)

		// Create the instance.
		instance.Spec.Secrets = []string{"testsecret"}
		c.Create(instance)

		// Advance postgres to running.
		postgres := &postgresv1.Postgresql{}
		c.EventuallyGet(helpers.Name("appsecretstest-database"), postgres)
		postgres.Status = postgresv1.ClusterStatusRunning
		c.Status().Update(postgres)

		// Change the DB secret
		time.Sleep(10 * time.Second)
		dbSecret.StringData["password"] = "other"
		c.Update(dbSecret)

		// Get the output app secrets.
		appSecret := &corev1.Secret{}
		data := map[string]interface{}{}
		c.EventuallyGet(helpers.Name("appsecretstest.app-secrets"), appSecret, c.EventuallyValue("postgis://summon:other@appsecretstest-database/summon", func(obj runtime.Object) (interface{}, error) {
			err := yaml.Unmarshal(appSecret.Data["summon-platform.yml"], &data)
			if err != nil {
				return nil, err
			}
			return data["DATABASE_URL"], nil
		}))
	})
})
