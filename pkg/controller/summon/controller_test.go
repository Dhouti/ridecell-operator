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

package summon_test

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	postgresv1 "github.com/zalando-incubator/postgres-operator/pkg/apis/acid.zalan.do/v1"
	"golang.org/x/net/context"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	dbv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/db/v1beta1"
	secretsv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/secrets/v1beta1"
	summonv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/summon/v1beta1"
	"github.com/Ridecell/ridecell-operator/pkg/test_helpers"
)

const timeout = time.Second * 30

var _ = Describe("Summon controller", func() {
	var helpers *test_helpers.PerTestHelpers

	BeforeEach(func() {
		helpers = testHelpers.SetupTest()
		pullSecret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "pull-secret", Namespace: helpers.OperatorNamespace}, Type: "kubernetes.io/dockerconfigjson", StringData: map[string]string{".dockerconfigjson": "{\"auths\": {}}"}}
		err := helpers.Client.Create(context.TODO(), pullSecret)
		Expect(err).NotTo(HaveOccurred())
		appSecrets := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "testsecret", Namespace: helpers.Namespace},
			Data: map[string][]byte{
				"filler": []byte{}}}
		err = helpers.Client.Create(context.TODO(), appSecrets)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		helpers.TeardownTest()
	})

	// Minimal test, service component has no deps so it should always immediately get created.
	It("creates a service", func() {
		c := helpers.Client
		instance := &summonv1beta1.SummonPlatform{ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: helpers.Namespace}}
		depKey := types.NamespacedName{Name: "foo-web", Namespace: helpers.Namespace}

		// Create the Summon object and expect the Reconcile and Service to be created
		err := c.Create(context.TODO(), instance)
		if apierrors.IsInvalid(err) {
			Fail(fmt.Sprintf("failed to create object, got an invalid object error: %v", err))
		}
		Expect(err).NotTo(HaveOccurred())

		service := &corev1.Service{}
		Eventually(func() error { return c.Get(context.TODO(), depKey, service) }, timeout).Should(Succeed())

		// Delete the Service and expect Reconcile to be called for Service deletion
		Expect(c.Delete(context.TODO(), service)).NotTo(HaveOccurred())
		Eventually(func() error { return c.Get(context.TODO(), depKey, service) }, timeout).Should(Succeed())
	})

	It("runs a basic reconcile", func() {
		c := helpers.Client
		instance := &summonv1beta1.SummonPlatform{ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: helpers.Namespace}, Spec: summonv1beta1.SummonPlatformSpec{
			Version: "1.2.3",
			Secret:  "testsecret",
		}}

		// Create the SummonPlatform object and expect the Reconcile to be created.
		err := c.Create(context.TODO(), instance)
		Expect(err).NotTo(HaveOccurred())
		err = c.Status().Update(context.TODO(), instance)
		Expect(err).NotTo(HaveOccurred())

		// Check the pull_secret object.
		pullsecret := &secretsv1beta1.PullSecret{}
		Eventually(func() error {
			return c.Get(context.TODO(), types.NamespacedName{Name: "foo-pullsecret", Namespace: helpers.Namespace}, pullsecret)
		}, timeout).
			Should(Succeed())
		pullsecret.Status.Status = secretsv1beta1.StatusReady
		err = c.Status().Update(context.TODO(), pullsecret)
		Expect(err).NotTo(HaveOccurred())

		// Check the Postgresql object.
		postgres := &postgresv1.Postgresql{}
		Eventually(func() error {
			return c.Get(context.TODO(), types.NamespacedName{Name: "foo-database", Namespace: helpers.Namespace}, postgres)
		}, timeout).
			Should(Succeed())
		Expect(postgres.Spec.Databases["summon"]).To(Equal("ridecell-admin"))

		// Create a fake credentials secret.
		dbSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "summon.foo-database.credentials", Namespace: helpers.Namespace},
			StringData: map[string]string{
				"password": "secretdbpass",
			},
		}
		err = c.Create(context.TODO(), dbSecret)
		Expect(err).NotTo(HaveOccurred())

		// Set the status of the DB to ready.
		postgres.Status = postgresv1.ClusterStatusRunning
		err = c.Status().Update(context.TODO(), postgres)
		Expect(err).NotTo(HaveOccurred())

		// Set the Postgres extensions to ready.
		ext := &dbv1beta1.PostgresExtension{}
		Eventually(func() error {
			return c.Get(context.TODO(), types.NamespacedName{Name: "foo-postgis", Namespace: helpers.Namespace}, ext)
		}, timeout).Should(Succeed())
		ext.Status.Status = dbv1beta1.StatusReady
		err = c.Status().Update(context.TODO(), ext)
		Expect(err).NotTo(HaveOccurred())
		Eventually(func() error {
			return c.Get(context.TODO(), types.NamespacedName{Name: "foo-postgis-topology", Namespace: helpers.Namespace}, ext)
		}, timeout).Should(Succeed())
		ext.Status.Status = dbv1beta1.StatusReady
		err = c.Status().Update(context.TODO(), ext)
		Expect(err).NotTo(HaveOccurred())

		// Check that a migration Job was created.
		job := &batchv1.Job{}
		Eventually(func() error {
			return c.Get(context.TODO(), types.NamespacedName{Name: "foo-migrations", Namespace: helpers.Namespace}, job)
		}, timeout).Should(Succeed())

		// Mark the migrations as successful.
		job.Status.Succeeded = 1
		err = c.Status().Update(context.TODO(), job)
		Expect(err).NotTo(HaveOccurred())

		// Check the web Deployment object.
		deploy := &appsv1.Deployment{}
		Eventually(func() error {
			return c.Get(context.TODO(), types.NamespacedName{Name: "foo-web", Namespace: helpers.Namespace}, deploy)
		}, timeout).Should(Succeed())
		Expect(deploy.Spec.Replicas).To(PointTo(BeEquivalentTo(1)))
		Expect(deploy.Spec.Template.Spec.Containers[0].Command).To(Equal([]string{"python", "-m", "gunicorn.app.wsgiapp", "-b", "0.0.0.0:8000", "summon_platform.wsgi", "--log-level=debug"}))
		Expect(deploy.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort).To(BeEquivalentTo(8000))

		// Check the web Service object.
		service := &corev1.Service{}
		Eventually(func() error {
			return c.Get(context.TODO(), types.NamespacedName{Name: "foo-web", Namespace: helpers.Namespace}, service)
		}, timeout).
			Should(Succeed())
		Expect(service.Spec.Ports[0].Port).To(BeEquivalentTo(8000))

		// Check the web Ingress object.
		ingress := &extv1beta1.Ingress{}
		Eventually(func() error {
			return c.Get(context.TODO(), types.NamespacedName{Name: "foo-web", Namespace: helpers.Namespace}, ingress)
		}, timeout).
			Should(Succeed())
		Expect(ingress.Spec.TLS[0].SecretName).To(Equal("testsecret-tls"))

		// Delete the Deployment and expect it to come back.
		Expect(c.Delete(context.TODO(), deploy)).NotTo(HaveOccurred())
		Eventually(func() error {
			return c.Get(context.TODO(), types.NamespacedName{Name: "foo-web", Namespace: helpers.Namespace}, deploy)
		}, timeout).
			Should(Succeed())
	})

	It("reconciles labels", func() {
		c := helpers.Client
		instance := &summonv1beta1.SummonPlatform{ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: helpers.Namespace}, Spec: summonv1beta1.SummonPlatformSpec{
			Version: "1.2.3",
			Secret:  "testsecret",
		}, Status: summonv1beta1.SummonPlatformStatus{
			MigrateVersion: "1.2.3",
		}}

		// Create the SummonPlatform object.
		err := c.Create(context.TODO(), instance)
		Expect(err).NotTo(HaveOccurred())
		err = c.Status().Update(context.TODO(), instance)
		Expect(err).NotTo(HaveOccurred())

		// Create a fake credentials secret.
		dbSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "summon.foo-database.credentials", Namespace: helpers.Namespace},
			StringData: map[string]string{
				"password": "secretdbpass",
			},
		}
		err = c.Create(context.TODO(), dbSecret)
		Expect(err).NotTo(HaveOccurred())

		// Set the status of PullSecret to ready.
		pullsecret := &secretsv1beta1.PullSecret{}
		Eventually(func() error {
			return c.Get(context.TODO(), types.NamespacedName{Name: "foo-pullsecret", Namespace: helpers.Namespace}, pullsecret)
		}, timeout).
			Should(Succeed())
		pullsecret.Status.Status = secretsv1beta1.StatusReady
		err = c.Status().Update(context.TODO(), pullsecret)
		Expect(err).NotTo(HaveOccurred())

		// Set the status of the DB to ready.
		postgres := &postgresv1.Postgresql{}
		Eventually(func() error {
			return c.Get(context.TODO(), types.NamespacedName{Name: "foo-database", Namespace: helpers.Namespace}, postgres)
		}, timeout).
			Should(Succeed())
		postgres.Status = postgresv1.ClusterStatusRunning
		err = c.Status().Update(context.TODO(), postgres)
		Expect(err).NotTo(HaveOccurred())

		// Set the Postgres extensions to ready.
		ext := &dbv1beta1.PostgresExtension{}
		Eventually(func() error {
			return c.Get(context.TODO(), types.NamespacedName{Name: "foo-postgis", Namespace: helpers.Namespace}, ext)
		}, timeout).Should(Succeed())
		ext.Status.Status = dbv1beta1.StatusReady
		err = c.Status().Update(context.TODO(), ext)
		Expect(err).NotTo(HaveOccurred())
		Eventually(func() error {
			return c.Get(context.TODO(), types.NamespacedName{Name: "foo-postgis-topology", Namespace: helpers.Namespace}, ext)
		}, timeout).Should(Succeed())
		ext.Status.Status = dbv1beta1.StatusReady
		err = c.Status().Update(context.TODO(), ext)
		Expect(err).NotTo(HaveOccurred())

		// Fetch the Deployment and check the initial label.
		deploy := &appsv1.Deployment{}
		Eventually(func() error {
			return c.Get(context.TODO(), types.NamespacedName{Name: "foo-web", Namespace: helpers.Namespace}, deploy)
		}, timeout).
			Should(Succeed())
		Expect(deploy.Labels["app.kubernetes.io/instance"]).To(Equal("foo-web"))
		Expect(deploy.Labels).ToNot(HaveKey("other"))

		// Modify the labels.
		deploy.Labels["app.kubernetes.io/instance"] = "boom"
		deploy.Labels["other"] = "foo"
		err = c.Update(context.TODO(), deploy)
		Expect(err).NotTo(HaveOccurred())

		// Check that the labels end up correct.
		Eventually(func() error {
			err := c.Get(context.TODO(), types.NamespacedName{Name: "foo-web", Namespace: helpers.Namespace}, deploy)
			if err != nil {
				return err
			}
			value, ok := deploy.Labels["other"]
			if !ok || value != "foo" {
				return fmt.Errorf("No update yet (other)")
			}
			return nil
		}, timeout).Should(Succeed())
		Eventually(func() error {
			err := c.Get(context.TODO(), types.NamespacedName{Name: "foo-web", Namespace: helpers.Namespace}, deploy)
			if err != nil {
				return err
			}
			value, ok := deploy.Labels["app.kubernetes.io/instance"]
			if !ok || value != "foo-web" {
				return fmt.Errorf("No update yet (app.kubernetes.io/instance)")
			}
			return nil
		}, timeout).Should(Succeed())
	})

	It("manages the status correctly", func() {
		c := helpers.Client

		// Create a SummonPlatform and related objects.
		instance := &summonv1beta1.SummonPlatform{
			ObjectMeta: metav1.ObjectMeta{Name: "statustester", Namespace: helpers.Namespace},
			Spec: summonv1beta1.SummonPlatformSpec{
				Version: "1-abcdef1-master",
				Secret:  "statustester",
			},
		}
		err := c.Create(context.TODO(), instance)
		Expect(err).NotTo(HaveOccurred())
		dbSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "summon.statustester-database.credentials", Namespace: helpers.Namespace},
			StringData: map[string]string{
				"password": "secretdbpass",
			},
		}
		err = c.Create(context.TODO(), dbSecret)
		Expect(err).NotTo(HaveOccurred())
		inSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "statustester", Namespace: helpers.Namespace},
			StringData: map[string]string{},
		}
		err = c.Create(context.TODO(), inSecret)
		Expect(err).NotTo(HaveOccurred())

		// Wait for the database to be created.
		postgres := &postgresv1.Postgresql{}
		Eventually(func() error {
			return c.Get(context.TODO(), types.NamespacedName{Name: "statustester-database", Namespace: helpers.Namespace}, postgres)
		}, timeout).Should(Succeed())

		// Check the status. Should not be set yet.
		assertStatus := func(status string) {
			Eventually(func() (string, error) {
				err := c.Get(context.TODO(), types.NamespacedName{Name: "statustester", Namespace: helpers.Namespace}, instance)
				return instance.Status.Status, err
			}, timeout).Should(Equal(status))
		}
		assertStatus("")

		// Set the database to Creating
		postgres.Status = postgresv1.ClusterStatusCreating
		err = c.Status().Update(context.TODO(), postgres)
		Expect(err).NotTo(HaveOccurred())

		// Check the status again. Should be Initializing.
		assertStatus(summonv1beta1.StatusInitializing)

		// Set the database to Running
		postgres.Status = postgresv1.ClusterStatusRunning
		err = c.Status().Update(context.TODO(), postgres)
		Expect(err).NotTo(HaveOccurred())

		// Check the status again. Should still be Initializing.
		assertStatus(summonv1beta1.StatusInitializing)

		// Set the postgres extensions to ready.
		ext := &dbv1beta1.PostgresExtension{}
		Eventually(func() error {
			return c.Get(context.TODO(), types.NamespacedName{Name: "statustester-postgis", Namespace: helpers.Namespace}, ext)
		}, timeout).Should(Succeed())
		ext.Status.Status = dbv1beta1.StatusReady
		err = c.Status().Update(context.TODO(), ext)
		Expect(err).NotTo(HaveOccurred())
		Eventually(func() error {
			return c.Get(context.TODO(), types.NamespacedName{Name: "statustester-postgis-topology", Namespace: helpers.Namespace}, ext)
		}, timeout).Should(Succeed())
		ext.Status.Status = dbv1beta1.StatusReady
		err = c.Status().Update(context.TODO(), ext)
		Expect(err).NotTo(HaveOccurred())

		// Set the pull secret to ready.
		pullSecret := &secretsv1beta1.PullSecret{}
		Eventually(func() error {
			return c.Get(context.TODO(), types.NamespacedName{Name: "statustester-pullsecret", Namespace: helpers.Namespace}, pullSecret)
		}, timeout).Should(Succeed())
		pullSecret.Status.Status = secretsv1beta1.StatusReady
		err = c.Status().Update(context.TODO(), pullSecret)
		Expect(err).NotTo(HaveOccurred())

		// Check the status again. Should be Migrating.
		assertStatus(summonv1beta1.StatusMigrating)

		// Mark the migration as a success.
		job := &batchv1.Job{}
		Eventually(func() error {
			return c.Get(context.TODO(), types.NamespacedName{Name: "statustester-migrations", Namespace: helpers.Namespace}, job)
		}, timeout).Should(Succeed())
		job.Status.Succeeded = 1
		err = c.Status().Update(context.TODO(), job)
		Expect(err).NotTo(HaveOccurred())

		// Check the status again. Should be Deploying.
		assertStatus(summonv1beta1.StatusDeploying)

		// Set deployments and statefulsets to ready.
		updateDeployment := func(s string) {
			deployment := &appsv1.Deployment{}
			Eventually(func() error {
				return c.Get(context.TODO(), types.NamespacedName{Name: "statustester-" + s, Namespace: helpers.Namespace}, deployment)
			}, timeout).Should(Succeed())
			deployment.Status.Replicas = 1
			deployment.Status.ReadyReplicas = 1
			deployment.Status.AvailableReplicas = 1
			err = c.Status().Update(context.TODO(), deployment)
			Expect(err).NotTo(HaveOccurred())
		}
		updateStatefulSet := func(s string) {
			statefulset := &appsv1.StatefulSet{}
			Eventually(func() error {
				return c.Get(context.TODO(), types.NamespacedName{Name: "statustester-" + s, Namespace: helpers.Namespace}, statefulset)
			}, timeout).Should(Succeed())
			Expect(err).NotTo(HaveOccurred())
			statefulset.Status.Replicas = 1
			statefulset.Status.ReadyReplicas = 1
			err = c.Status().Update(context.TODO(), statefulset)
			Expect(err).NotTo(HaveOccurred())
		}
		updateDeployment("web")
		updateDeployment("daphne")
		updateDeployment("celeryd")
		updateDeployment("channelworker")
		updateDeployment("static")
		updateStatefulSet("celerybeat")

		// Check the status again. Should be Deploying.
		assertStatus(summonv1beta1.StatusReady)
	})
})
