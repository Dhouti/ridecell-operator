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

package elasticsearch_test

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/Ridecell/ridecell-operator/pkg/test_helpers"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	es "github.com/aws/aws-sdk-go/service/elasticsearchservice"
	"github.com/aws/aws-sdk-go/service/sts"

	awsv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/aws/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var sess *session.Session
var essvc *es.ElasticsearchService
var esInstance *awsv1beta1.ElasticSearch

var _ = Describe("ElasticSearch controller", func() {
	var helpers *test_helpers.PerTestHelpers

	BeforeEach(func() {
		os.Setenv("ENABLE_FINALIZERS", "true")
		helpers = testHelpers.SetupTest()
		if os.Getenv("AWS_TESTING_ACCOUNT_ID") == "" {
			Skip("$AWS_TESTING_ACCOUNT_ID not set, skipping elasticsearch integration tests")
		}
		if os.Getenv("AWS_TEST_ACCOUNT_PERMISSIONS_BOUNDARY_ARN") == "" {
			Skip("$AWS_TEST_ACCOUNT_PERMISSIONS_BOUNDARY_ARN not set, skipping elasticsearch integration tests")
		}

		var err error
		sess, err = session.NewSession()
		Expect(err).NotTo(HaveOccurred())

		// Check if this being run on the testing account
		stssvc := sts.New(sess)
		getCallerIdentityOutput, err := stssvc.GetCallerIdentity(&sts.GetCallerIdentityInput{})
		Expect(err).NotTo(HaveOccurred())
		if aws.StringValue(getCallerIdentityOutput.Account) != os.Getenv("AWS_TESTING_ACCOUNT_ID") {
			Skip("These tests should only be run on the testing account.")
		}

		essvc = es.New(sess)

		esInstance = &awsv1beta1.ElasticSearch{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: helpers.Namespace,
			},
			Spec: awsv1beta1.ElasticSearchSpec{
				SecurityGroupId: "sg-036cc3bcdd664a325", // for testing only
			},
		}
	})

	AfterEach(func() {
		// Delete ES domain and see if it cleans up on its own
		c := helpers.TestClient
		c.Delete(esInstance)
		Eventually(func() error { return esDomainExists() }, time.Second*10).ShouldNot(Succeed())

		// Make sure the object is deleted
		fetchESInstance := &awsv1beta1.ElasticSearch{}
		Eventually(func() error {
			return helpers.Client.Get(context.TODO(), helpers.Name(esInstance.Name), fetchESInstance)
		}, time.Second*30).ShouldNot(Succeed())

		helpers.TeardownTest()
	})

	It("runs a basic reconcile", func() {
		c := helpers.TestClient
		c.Create(esInstance)

		fetchESInstance := &awsv1beta1.ElasticSearch{}
		c.EventuallyGet(helpers.Name("test"), fetchESInstance, c.EventuallyStatus("Ready"))

		Expect(fetchESInstance.ObjectMeta.Finalizers).To(HaveLen(1))
		Expect(fetchESInstance.ObjectMeta.DeletionTimestamp.IsZero()).To(BeTrue())
		esDomainConfig, err := describeESDomain()
		Expect(err).ToNot(HaveOccurred())
		Expect(aws.BoolValue(esDomainConfig.DomainStatus.ElasticsearchClusterConfig.DedicatedMasterEnabled)).To(BeFalse())
		Expect(esDomainHasValidTag()).To(BeTrue())

		// update deployment type
		esInstance.Spec.DeploymentType = "Production"

		c.Update(esInstance)
		fetchESInstance = &awsv1beta1.ElasticSearch{}
		c.EventuallyGet(helpers.Name("test"), fetchESInstance, c.EventuallyStatus("Ready"))

		// check for updated config
		esDomainConfig, err = describeESDomain()
		Expect(err).ToNot(HaveOccurred())
		Expect(aws.BoolValue(esDomainConfig.DomainStatus.ElasticsearchClusterConfig.DedicatedMasterEnabled)).To(BeTrue())

	})

})

func describeESDomain() (*es.DescribeElasticsearchDomainOutput, error) {
	return essvc.DescribeElasticsearchDomain(&es.DescribeElasticsearchDomainInput{DomainName: aws.String(strings.ToLower(esInstance.Name))})
}

func esDomainExists() error {
	_, err := describeESDomain()
	return err
}

func esDomainHasValidTag() bool {
	describeElasticsearchDomainOutput, err := essvc.DescribeElasticsearchDomain(&es.DescribeElasticsearchDomainInput{DomainName: aws.String(strings.ToLower(esInstance.Name))})
	if err != nil {
		return false
	}
	listTagsOutput, err := essvc.ListTags(&es.ListTagsInput{
		ARN: describeElasticsearchDomainOutput.DomainStatus.ARN,
	})
	if err != nil {
		return false
	}

	for _, tag := range listTagsOutput.TagList {
		if aws.StringValue(tag.Key) == "Ridecell-Operator" {
			return true
		}
	}
	return false
}
