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
	"time"

	. "github.com/Ridecell/ridecell-operator/pkg/test_helpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	//summonv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/summon/v1beta1"
	summoncomponents "github.com/Ridecell/ridecell-operator/pkg/controller/summon/components"
	gcr "github.com/Ridecell/ridecell-operator/pkg/utils/gcr"
)

// Will take suggestions on a better way to handle mock logic around registry hub and caching...
var MockTags []string

// Dont actually need parameter, but mock func definition required to match real func definition for injection.
func MockGetSummonTags() {
	elapsed := time.Since(gcr.LastCacheUpdate)
	// Fetch tags if cache expired
	if elapsed >= gcr.CacheExpiry {
		// Instead of actually fetching tags, we use mock ones.
		gcr.CachedTags = MockTags
		gcr.LastCacheUpdate = time.Now()
	}
}

var _ = FDescribe("SummonPlatform AutoDeploy Component", func() {
	comp := summoncomponents.NewAutoDeploy()

	BeforeEach(func() {
		// Start each test case off with some test tags and reset cache timestamp to zero
		MockTags = []string{"1-abc1234-test-branch", "2-def5678-test-branch", "1-abc1234-other-branch"}
		gcr.LastCacheUpdate = time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)

		comp.InjectMockTagFetcher(MockGetSummonTags)
	})

	Describe("isReconcilable", func() {
		It("returns false if autoDeploy is not set", func() {
			Expect(comp.IsReconcilable(ctx)).To(BeFalse())
		})

		It("returns true if autoDeploy is set", func() {
			instance.Spec.AutoDeploy = "test-branch"
			Expect(comp.IsReconcilable(ctx)).To(BeTrue())
		})
	})

	It("sets the image version to the latest tag in the tag cache matching the branch name", func() {
		instance.Spec.AutoDeploy = "test-branch"
		Expect(comp).To(ReconcileContext(ctx))
		Expect(instance.Spec.Version).To(Equal("2-def5678-test-branch"))
		instance.Spec.AutoDeploy = "other-branch"
		Expect(comp).To(ReconcileContext(ctx))
		Expect(instance.Spec.Version).To(Equal("1-abc1234-other-branch"))
	})

	// tag cache update is handled by tagFetcher function which tracks time, and tagFetcher is triggered by Reconciler
	It("uses the latest image from the updated tag cache", func() {
		instance.Spec.AutoDeploy = "test-branch"
		// set LastCacheUpdate time to 5 mins in the past confirm cache update occurs
		gcr.LastCacheUpdate = time.Now().Add(time.Minute * -5)
		// Mock updated tag cache values
		MockTags = []string{"1-abc1234-test-branch", "2-def5678-test-branch", "3-ghi9101112-test-branch", "1-abc1234-other-branch", "2-def5678-other-branch"}
		Expect(comp).To(ReconcileContext(ctx))
		Expect(gcr.CachedTags).To(Equal(MockTags))
		Expect(instance.Spec.Version).To(Equal("3-ghi9101112-test-branch"))
	})

	It("leaves Spec.Version as is if no matching image found", func() {
		instance.Spec.AutoDeploy = "nonexistent-branch"
		_, err := comp.Reconcile(ctx)
		Expect(err).To(MatchError("autodeploy: no matching branch image for nonexistent-branch"))
		Expect(instance.Spec.Version).To(Equal("1.2.3"))
	})

	It("overrides Spec.Version with latest match found", func() {
		instance.Spec.AutoDeploy = "test-branch"
		_, err := comp.Reconcile(ctx)
		Expect(err).To(Not(HaveOccurred()))
		Expect(instance.Spec.Version).To(Equal("2-def5678-test-branch"))
	})
})