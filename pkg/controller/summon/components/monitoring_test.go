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
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/Ridecell/ridecell-operator/pkg/components"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	rmonitor "github.com/Ridecell/ridecell-operator/pkg/apis/monitoring/v1beta1"
	summoncomponents "github.com/Ridecell/ridecell-operator/pkg/controller/summon/components"
)

var _ = Describe("SummonPlatform monitoring Component", func() {
	var comp components.Component

	Context("Monitoring...", func() {
		BeforeEach(func() {
			comp = summoncomponents.NewMonitoring()

		})

		It("Is Reconciling? ", func() {
			val := true
			instance.Spec.Monitoring.Enabled = &val
			instance.Spec.Notifications.SlackChannel = "#test"
			instance.Spec.Notifications.Pagerdutyteam = "myteam"
			instance.Spec.MigrationOverrides.RabbitMQVhost = "oldone"
			Expect(comp).To(ReconcileContext(ctx))

			monitor := &rmonitor.Monitor{}
			err := ctx.Client.Get(context.TODO(), types.NamespacedName{Name: "foo-dev-monitoring", Namespace: "summon-dev"}, monitor)
			Expect(err).NotTo(HaveOccurred())
			Expect(monitor.Spec.Notify.Slack[0]).To(Equal("#test"))
			Expect(monitor.Spec.Notify.PagerdutyTeam).To(Equal("myteam"))
			Expect(len(monitor.Spec.MetricAlertRules)).Should(BeNumerically(">=", 1))
			Expect(monitor.Spec.MetricAlertRules[6].Expr).Should(ContainSubstring("oldone"))
		})

		It("Missing slack should Reconcile without err", func() {
			val := true
			instance.Spec.Monitoring.Enabled = &val
			Expect(comp).To(ReconcileContext(ctx))
			// This will not create kind: monitor
			monitor := &rmonitor.Monitor{}
			err := ctx.Client.Get(context.TODO(), types.NamespacedName{Name: "foo-dev-monitoring", Namespace: "summon-dev"}, monitor)
			Expect(err).To(HaveOccurred())
		})

		It("cleans up an existing Monitor object if monitoring is disabled", func() {
			val := false
			instance.Spec.Monitoring.Enabled = &val

			monitor := &rmonitor.Monitor{ObjectMeta: metav1.ObjectMeta{Name: "foo-dev-monitoring", Namespace: "summon-dev"}}
			ctx.Client = fake.NewFakeClient(instance, monitor)

			Expect(comp).To(ReconcileContext(ctx))

			err := ctx.Client.Get(context.TODO(), types.NamespacedName{Name: "foo-dev-monitoring", Namespace: "summon-dev"}, monitor)
			Expect(err).To(HaveOccurred())
			Expect(kerrors.IsNotFound(err)).To(BeTrue())
		})

	})
})
