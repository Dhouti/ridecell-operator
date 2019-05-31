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
	"fmt"
	"regexp"
	"strings"
	"time"

	. "github.com/Ridecell/ridecell-operator/pkg/test_helpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/rds/rdsiface"
	"github.com/pkg/errors"
	//"k8s.io/apimachinery/pkg/types"

	//dbv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/db/v1beta1"
	rdssnapshotcomponents "github.com/Ridecell/ridecell-operator/pkg/controller/rdssnapshot/components"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type mockRDSDBClient struct {
	rdsiface.RDSAPI

	snapshotExists  bool
	snapshotCreated bool
	snapshotDeleted bool

	snapshotStatus string
	snapshotTags   []*rds.Tag
}

var passwordSecret *corev1.Secret

var _ = Describe("rdssnapshot db Component", func() {
	comp := rdssnapshotcomponents.NewRDSSnapshot()
	var mockRDS *mockRDSDBClient

	BeforeEach(func() {
		comp = rdssnapshotcomponents.NewRDSSnapshot()
		mockRDS = &mockRDSDBClient{}
		comp.InjectRDSAPI(mockRDS)
		instance.Spec.RDSInstanceID = "fake-db"
		creationTimestamp := instance.ObjectMeta.CreationTimestamp.Add(time.Second * 0)
		curTimeString := time.Time.Format(creationTimestamp, rdssnapshotcomponents.CustomTimeLayout)
		instance.Spec.SnapshotID = fmt.Sprintf("%s-%s", instance.Name, curTimeString)
	})

	Describe("isReconcilable", func() {
		It("returns true", func() {
			Expect(comp.IsReconcilable(ctx)).To(BeTrue())
		})
	})

	It("creates a new rds snapshot with no TTL", func() {
		Expect(comp).To(ReconcileContext(ctx))
		Expect(instance.ObjectMeta.Finalizers[0]).To(Equal("rdssnapshot.finalizer"))
		Expect(mockRDS.snapshotCreated).To(BeTrue())
		Expect(mockRDS.snapshotTags).To(HaveLen(2))
	})

	It("adds deletiontimestamp to object", func() {
		instance.ObjectMeta.Finalizers = []string{"rdssnapshot.finalizer"}
		currentTime := metav1.Now()
		instance.ObjectMeta.SetDeletionTimestamp(&currentTime)
		Expect(comp).To(ReconcileContext(ctx))
		Expect(mockRDS.snapshotDeleted).To(BeTrue())
		Expect(instance.ObjectMeta.Finalizers).To(HaveLen(0))
	})

	It("reconciles with expired TTL", func() {
		instance.ObjectMeta.Finalizers = []string{"rdssnapshot.finalizer"}
		instance.Spec.TTL = time.Second * 5
		// Reconcile to add deletion timestamp
		Expect(comp).To(ReconcileContext(ctx))
		Expect(instance.ObjectMeta.DeletionTimestamp.IsZero()).To(BeFalse())
		// Handles actual deletion
		Expect(comp).To(ReconcileContext(ctx))
		Expect(mockRDS.snapshotDeleted).To(BeTrue())
		Expect(instance.ObjectMeta.Finalizers).To(HaveLen(0))
	})

	It("reconciles with non-expired TTL", func() {
		instance.ObjectMeta.Finalizers = []string{"rdssnapshot.finalizer"}
		// Reset the creation timestamp
		instance.ObjectMeta.CreationTimestamp = metav1.Now()
		instance.Spec.TTL = time.Second * 5
		Expect(comp).To(ReconcileContext(ctx))
		Expect(instance.ObjectMeta.DeletionTimestamp.IsZero()).To(BeTrue())
		Expect(mockRDS.snapshotDeleted).To(BeFalse())
	})
})

// Mock aws functions below

func (m *mockRDSDBClient) DescribeDBSnapshots(input *rds.DescribeDBSnapshotsInput) (*rds.DescribeDBSnapshotsOutput, error) {
	if m.snapshotExists {
		return &rds.DescribeDBSnapshotsOutput{
			DBSnapshots: []*rds.DBSnapshot{
				&rds.DBSnapshot{
					DBInstanceIdentifier: aws.String(instance.Spec.RDSInstanceID),
					DBSnapshotIdentifier: input.DBSnapshotIdentifier,
					Status:               aws.String(m.snapshotStatus),
				},
			},
		}, nil
	}
	return &rds.DescribeDBSnapshotsOutput{}, awserr.New(rds.ErrCodeDBSnapshotNotFoundFault, "", nil)
}

func (m *mockRDSDBClient) CreateDBSnapshot(input *rds.CreateDBSnapshotInput) (*rds.CreateDBSnapshotOutput, error) {
	match := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9-]*[a-zA-z]$`).MatchString(aws.StringValue(input.DBSnapshotIdentifier))
	if strings.Contains("--", aws.StringValue(input.DBSnapshotIdentifier)) || !match {
		return &rds.CreateDBSnapshotOutput{}, errors.New("mock_rds_snapshot: input snapshot id did not match regex")
	}

	m.snapshotCreated = true
	m.snapshotExists = true
	m.snapshotStatus = "pending"
	m.snapshotTags = input.Tags

	return &rds.CreateDBSnapshotOutput{
		DBSnapshot: &rds.DBSnapshot{
			DBInstanceIdentifier: aws.String(instance.Spec.RDSInstanceID),
			DBSnapshotIdentifier: input.DBSnapshotIdentifier,
			Status:               aws.String(m.snapshotStatus),
		},
	}, nil
}

func (m *mockRDSDBClient) DeleteDBSnapshot(input *rds.DeleteDBSnapshotInput) (*rds.DeleteDBSnapshotOutput, error) {
	m.snapshotDeleted = true
	if m.snapshotExists {
		return &rds.DeleteDBSnapshotOutput{}, nil
	}
	return &rds.DeleteDBSnapshotOutput{}, awserr.New(rds.ErrCodeDBSnapshotNotFoundFault, "", nil)
}
