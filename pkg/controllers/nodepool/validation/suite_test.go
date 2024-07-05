/*
Copyright The Kubernetes Authors.

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

package validation

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"sigs.k8s.io/karpenter/pkg/test/v1alpha1"

	"github.com/Pallinder/go-randomdata"
	v1 "k8s.io/api/core/v1"

	"github.com/awslabs/operatorpkg/status"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"sigs.k8s.io/karpenter/pkg/apis"
	"sigs.k8s.io/karpenter/pkg/apis/v1beta1"
	"sigs.k8s.io/karpenter/pkg/test"
	. "sigs.k8s.io/karpenter/pkg/test/expectations"
	. "sigs.k8s.io/karpenter/pkg/utils/testing"
)

var (
	nodePoolValidationController *Controller
	ctx                          context.Context
	env                          *test.Environment
	nodePool                     *v1beta1.NodePool
)

func TestAPIs(t *testing.T) {
	ctx = TestContextWithLogger(t)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Counter")
}

var _ = BeforeSuite(func() {
	env = test.NewEnvironment(test.WithCRDs(apis.CRDs...), test.WithCRDs(v1alpha1.CRDs...))
	nodePoolValidationController = NewController(env.Client)
})
var _ = AfterEach(func() {
	ExpectCleanedUp(ctx, env.Client)
})

var _ = AfterSuite(func() {
	Expect(env.Stop()).To(Succeed(), "Failed to stop environment")
})
var _ = Describe("Counter", func() {
	BeforeEach(func() {
		nodePool = test.NodePool()
		nodePool.StatusConditions().SetUnknown(v1beta1.ConditionTypeValidationSucceeded)
	})
	It("should set the NodePoolValidationSucceeded status condition to true if nodePool healthy checks succeed", func() {
		ExpectApplied(ctx, env.Client, nodePool)
		ExpectObjectReconciled(ctx, env.Client, nodePoolValidationController, nodePool)
		nodePool = ExpectExists(ctx, env.Client, nodePool)
		Expect(nodePool.StatusConditions().IsTrue(status.ConditionReady)).To(BeTrue())
		Expect(nodePool.StatusConditions().IsTrue(v1beta1.ConditionTypeValidationSucceeded)).To(BeTrue())
	})
	It("should set the NodePoolValidationSucceeded status condition to false if nodePool validation failed", func() {
		nodePool.Spec.Template.Spec.Taints = []v1.Taint{{Key: fmt.Sprintf("test.com.test.%s/test", strings.ToLower(randomdata.Alphanumeric(250))), Effect: v1.TaintEffectNoSchedule}}
		ExpectApplied(ctx, env.Client, nodePool)
		_ = ExpectObjectReconcileFailed(ctx, env.Client, nodePoolValidationController, nodePool)
		nodePool = ExpectExists(ctx, env.Client, nodePool)
		Expect(nodePool.StatusConditions().Get(status.ConditionReady).IsFalse()).To(BeTrue())
		Expect(nodePool.StatusConditions().Get(v1beta1.ConditionTypeValidationSucceeded).IsFalse()).To(BeTrue())
	})
})
