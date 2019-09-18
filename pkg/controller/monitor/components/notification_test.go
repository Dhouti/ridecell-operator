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
	"encoding/json"
	"os"

	. "github.com/Ridecell/ridecell-operator/pkg/test_helpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	alertmconfig "github.com/prometheus/alertmanager/config"
	"gopkg.in/yaml.v2"

	"k8s.io/apimachinery/pkg/types"

	monitoringv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/monitoring/v1beta1"
	mcomponents "github.com/Ridecell/ridecell-operator/pkg/controller/monitor/components"
	"github.com/Ridecell/ridecell-operator/pkg/test_helpers/fake_pagerduty"
)

var _ = Describe("Monitor Notification Component", func() {
	comp := mcomponents.NewNotification()
	BeforeEach(func() {
		os.Setenv("PG_MOCK_URL", "http://localhost:8082")
		fake_pagerduty.Run()
	})

	It("Is reconcilable?", func() {
		//instance.Spec.ServiceName = "dev-foo-service"
		instance.Spec.Notify = monitoringv1beta1.Notify{
			Slack: []string{
				"#test-alert",
				"#test",
			},
			PagerdutyTeam: "myteam",
		}

		Expect(comp).To(ReconcileContext(ctx))
		config := &monitoringv1beta1.AlertManagerConfig{}
		err := ctx.Get(context.Background(), types.NamespacedName{Name: "alertmanagerconfig-foo", Namespace: "default"}, config)
		// Check receiver correct slack channel name
		receiver := &alertmconfig.Receiver{}
		err = yaml.Unmarshal([]byte(config.Spec.Receivers[0]), receiver)
		Expect(err).ToNot(HaveOccurred())
		Expect(receiver.SlackConfigs[0].Channel).To(Equal("#test-alert"))
		// Check pd receiver
		err = json.Unmarshal([]byte(config.Spec.Receivers[1]), receiver)
		Expect(err).ToNot(HaveOccurred())
		Expect(receiver.PagerdutyConfigs[0].Severity).To(ContainSubstring("CommonLabels.severity"))
		// Check Route have correct Receiver name
		route := &alertmconfig.Route{}
		err = yaml.Unmarshal([]byte(config.Spec.Route), route)
		Expect(err).ToNot(HaveOccurred())
		// Check correct & default route condition present
		Expect(route.MatchRE["servicename"]).Should(ContainSubstring(instance.Spec.ServiceName))
	})

})
